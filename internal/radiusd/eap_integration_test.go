package radiusd

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins"
	eap "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// eapTestClient is a minimal EAP supplicant used to drive a full EAP
// challenge/response handshake against a live RADIUS auth server in tests.
type eapTestClient struct {
	serverAddr    string
	secret        string
	username      string
	nasIdentifier string
	nasIP         net.IP
}

// eapExchangeTimeout bounds each RADIUS round-trip so a missing response (e.g.
// the server failing to bind or a handler stopping without writing a reply)
// fails the test promptly instead of hanging the package.
const eapExchangeTimeout = 5 * time.Second

// exchange performs a single bounded RADIUS request/response round-trip.
func (c *eapTestClient) exchange(t *testing.T, packet *radius.Packet) (*radius.Packet, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), eapExchangeTimeout)
	defer cancel()
	return radius.Exchange(ctx, packet, c.serverAddr)
}

// newAccessRequest builds a fresh Access-Request carrying the mandatory
// identity attributes shared by every round of the handshake.
func (c *eapTestClient) newAccessRequest() *radius.Packet {
	packet := radius.New(radius.CodeAccessRequest, []byte(c.secret))
	_ = rfc2865.UserName_SetString(packet, c.username)           //nolint:errcheck
	_ = rfc2865.NASIdentifier_SetString(packet, c.nasIdentifier) //nolint:errcheck
	_ = rfc2865.NASIPAddress_Set(packet, c.nasIP)                //nolint:errcheck
	return packet
}

// sendIdentity performs the first round: EAP-Response/Identity. The server is
// expected to answer with an Access-Challenge carrying the EAP request and a
// State attribute that must be echoed back in subsequent rounds.
func (c *eapTestClient) sendIdentity(t *testing.T) (*eap.EAPMessage, []byte) {
	t.Helper()

	identity := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: 1,
		Type:       eap.TypeIdentity,
		Data:       []byte(c.username),
	}

	packet := c.newAccessRequest()
	eap.SetEAPMessageAndAuth(packet, identity.Encode(), c.secret)

	resp, err := c.exchange(t, packet)
	if err != nil {
		t.Fatalf("identity exchange failed: %v", err)
	}
	if resp.Code != radius.CodeAccessChallenge {
		t.Fatalf("expected Access-Challenge after identity, got %v", resp.Code)
	}

	challenge, err := eap.ParseEAPMessage(resp)
	if err != nil {
		t.Fatalf("failed to parse EAP challenge: %v", err)
	}
	if challenge.Code != eap.CodeRequest {
		t.Fatalf("unexpected EAP challenge code=%d type=%d", challenge.Code, challenge.Type)
	}

	state := rfc2865.State_Get(resp)
	if len(state) == 0 {
		t.Fatal("challenge response missing State attribute")
	}

	return challenge, state
}

// buildMD5Response computes the EAP-MD5 response payload for the given
// challenge using MD5(identifier || password || challenge).
func buildMD5Response(identifier uint8, password string, challenge []byte) []byte {
	hash := md5.New()
	hash.Write([]byte{identifier})
	hash.Write([]byte(password))
	hash.Write(challenge)
	digest := hash.Sum(nil)

	// EAP-MD5 value format: Value-Size (1 byte) || Value (16 bytes)
	data := make([]byte, 0, 1+len(digest))
	data = append(data, byte(md5.Size))
	data = append(data, digest...)
	return data
}

// extractMD5Challenge pulls the raw challenge bytes out of the EAP-MD5 request.
// Format: Value-Size (1 byte) || Value (Value-Size bytes).
func extractMD5Challenge(t *testing.T, msg *eap.EAPMessage) []byte {
	t.Helper()
	if len(msg.Data) < 1 {
		t.Fatal("MD5 challenge has no value-size byte")
	}
	valueSize := int(msg.Data[0])
	if len(msg.Data) < 1+valueSize {
		t.Fatalf("MD5 challenge truncated: want %d value bytes, have %d", valueSize, len(msg.Data)-1)
	}
	return msg.Data[1 : 1+valueSize]
}

// sendMD5Response performs the second round: EAP-Response/MD5-Challenge.
func (c *eapTestClient) sendMD5Response(t *testing.T, challenge *eap.EAPMessage, state []byte, password string) *radius.Packet {
	t.Helper()

	challengeValue := extractMD5Challenge(t, challenge)
	response := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: challenge.Identifier,
		Type:       eap.TypeMD5Challenge,
		Data:       buildMD5Response(challenge.Identifier, password, challengeValue),
	}

	packet := c.newAccessRequest()
	_ = rfc2865.State_Set(packet, state) //nolint:errcheck
	eap.SetEAPMessageAndAuth(packet, response.Encode(), c.secret)

	resp, err := c.exchange(t, packet)
	if err != nil {
		t.Fatalf("md5 response exchange failed: %v", err)
	}
	return resp
}

// sendNak performs an EAP-Response/Nak suggesting an alternative method.
func (c *eapTestClient) sendNak(t *testing.T, identifier uint8, state []byte, suggestedType uint8) *radius.Packet {
	t.Helper()

	nak := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: identifier,
		Type:       eap.TypeNak,
		Data:       []byte{suggestedType},
	}

	packet := c.newAccessRequest()
	_ = rfc2865.State_Set(packet, state) //nolint:errcheck
	eap.SetEAPMessageAndAuth(packet, nak.Encode(), c.secret)

	resp, err := c.exchange(t, packet)
	if err != nil {
		t.Fatalf("nak exchange failed: %v", err)
	}
	return resp
}

// MSCHAPv2 EAP wire-format constants used by the test supplicant.
const (
	mschapv2OpChallenge = 1
	mschapv2OpResponse  = 2
	mschapv2ChallengeSz = 16
	mschapv2ResponseSz  = 49 // PeerChallenge(16) + Reserved(8) + NTResponse(24) + Flags(1)
)

// parseMSCHAPv2Challenge extracts the authenticator challenge and MS-CHAPv2-ID
// from an EAP-MSCHAPv2 challenge request.
// Data layout: OpCode(1) | MS-CHAPv2-ID(1) | MS-Length(2) | Value-Size(1) | Challenge(16) | Name(variable)
func parseMSCHAPv2Challenge(t *testing.T, msg *eap.EAPMessage) (msID uint8, authChallenge []byte) {
	t.Helper()
	if msg.Type != eap.TypeMSCHAPv2 {
		t.Fatalf("expected MSCHAPv2 EAP type, got %d", msg.Type)
	}
	if len(msg.Data) < 5+mschapv2ChallengeSz {
		t.Fatalf("MSCHAPv2 challenge too short: %d bytes", len(msg.Data))
	}
	if msg.Data[0] != mschapv2OpChallenge {
		t.Fatalf("expected MSCHAPv2 Challenge opcode, got %d", msg.Data[0])
	}
	msID = msg.Data[1]
	valueSize := int(msg.Data[4])
	if valueSize != mschapv2ChallengeSz {
		t.Fatalf("unexpected MSCHAPv2 challenge value-size: %d", valueSize)
	}
	authChallenge = msg.Data[5 : 5+mschapv2ChallengeSz]
	return msID, authChallenge
}

// buildMSCHAPv2Response constructs the EAP-MSCHAPv2 Response payload (the bytes
// following the EAP Type field) for the supplied challenge and credentials.
func buildMSCHAPv2Response(t *testing.T, username, password string, msID uint8, authChallenge []byte) []byte {
	t.Helper()

	peerChallenge := make([]byte, mschapv2ChallengeSz)
	if _, err := rand.Read(peerChallenge); err != nil {
		t.Fatalf("failed to generate peer challenge: %v", err)
	}

	ntResponse, err := rfc2759.GenerateNTResponse(authChallenge, peerChallenge, []byte(username), []byte(password))
	if err != nil {
		t.Fatalf("failed to generate NT-Response: %v", err)
	}

	// Value: Peer-Challenge(16) | Reserved(8) | NT-Response(24) | Flags(1)
	value := make([]byte, 0, mschapv2ResponseSz)
	value = append(value, peerChallenge...)
	value = append(value, make([]byte, 8)...)
	value = append(value, ntResponse...)
	value = append(value, 0) // Flags

	name := []byte(username)
	msLen := 1 + 1 + 2 + 1 + len(value) + len(name) // OpCode+ID+Length+ValueSize+Value+Name

	data := make([]byte, 0, msLen)
	data = append(data, mschapv2OpResponse) // OpCode
	data = append(data, msID)               // MS-CHAPv2-ID
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(msLen)) //nolint:gosec // bounded by EAP packet size
	data = append(data, lenBuf...)
	data = append(data, byte(mschapv2ResponseSz)) // Value-Size
	data = append(data, value...)
	data = append(data, name...)
	return data
}

// sendMSCHAPv2Response performs the second round: EAP-Response/MSCHAPv2.
func (c *eapTestClient) sendMSCHAPv2Response(t *testing.T, challenge *eap.EAPMessage, state []byte, password string) *radius.Packet {
	t.Helper()

	msID, authChallenge := parseMSCHAPv2Challenge(t, challenge)
	response := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: challenge.Identifier,
		Type:       eap.TypeMSCHAPv2,
		Data:       buildMSCHAPv2Response(t, c.username, password, msID, authChallenge),
	}

	packet := c.newAccessRequest()
	_ = rfc2865.State_Set(packet, state) //nolint:errcheck
	eap.SetEAPMessageAndAuth(packet, response.Encode(), c.secret)

	resp, err := c.exchange(t, packet)
	if err != nil {
		t.Fatalf("mschapv2 response exchange failed: %v", err)
	}
	return resp
}

// startEAPTestServer boots a live auth server backed by an in-memory SQLite DB
// and returns a configured EAP test client plus the database handle so callers
// can seed users and NAS records.
func startEAPTestServer(t *testing.T) (*eapTestClient, *RadiusService) {
	t.Helper()

	appCtx, cfg := setupTestEnv(t)
	t.Cleanup(appCtx.Release)

	registry.ResetForTest()
	t.Cleanup(registry.ResetForTest)
	reRegisterVendorParsers()

	radiusService := NewRadiusService(appCtx)
	t.Cleanup(radiusService.Release)
	plugins.InitPlugins(appCtx, radiusService.SessionRepo, radiusService.AccountingRepo)
	authService := NewAuthService(radiusService)

	go func() {
		if err := ListenRadiusAuthServer(appCtx, authService); err != nil {
			t.Logf("Auth server stopped: %v", err)
		}
	}()

	// Give the UDP listener time to bind before exchanging packets.
	time.Sleep(time.Second)

	client := &eapTestClient{
		serverAddr:    fmt.Sprintf("127.0.0.1:%d", cfg.Radiusd.AuthPort),
		secret:        "secret",
		nasIdentifier: "eap-nas",
		nasIP:         net.ParseIP("10.0.0.1"),
	}

	return client, radiusService
}

// seedEAPUser inserts a NAS and user record used by the EAP handshake tests.
func seedEAPUser(t *testing.T, rs *RadiusService, username, password string) {
	t.Helper()

	db := rs.AppContext().DB()

	nas := &domain.NetNas{
		Identifier: "eap-nas",
		Ipaddr:     "10.0.0.1",
		Secret:     "secret",
		VendorCode: "0",
		Status:     common.ENABLED,
		Remark:     "EAP Test NAS",
	}
	if err := db.Create(nas).Error; err != nil {
		t.Fatalf("failed to create NAS: %v", err)
	}

	user := &domain.RadiusUser{
		Username:   username,
		Password:   password,
		Status:     common.ENABLED,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
}

func TestEAPMD5IntegrationSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, rs := startEAPTestServer(t)
	seedEAPUser(t, rs, "eapuser", "eappass")
	client.username = "eapuser"

	challenge, state := client.sendIdentity(t)
	resp := client.sendMD5Response(t, challenge, state, "eappass")

	if resp.Code != radius.CodeAccessAccept {
		t.Fatalf("expected Access-Accept, got %v", resp.Code)
	}

	eapAttr := rfc2869.EAPMessage_Get(resp)
	if len(eapAttr) < 1 {
		t.Fatal("Access-Accept missing EAP-Message attribute")
	}
	if eapAttr[0] != eap.CodeSuccess {
		t.Fatalf("expected EAP-Success (code %d), got code %d", eap.CodeSuccess, eapAttr[0])
	}
}

func TestEAPMD5IntegrationWrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, rs := startEAPTestServer(t)
	seedEAPUser(t, rs, "eapuser", "eappass")
	client.username = "eapuser"

	challenge, state := client.sendIdentity(t)
	resp := client.sendMD5Response(t, challenge, state, "wrong-password")

	if resp.Code != radius.CodeAccessReject {
		t.Fatalf("expected Access-Reject for wrong password, got %v", resp.Code)
	}

	eapAttr := rfc2869.EAPMessage_Get(resp)
	if len(eapAttr) < 1 {
		t.Fatal("Access-Reject missing EAP-Message attribute")
	}
	if eapAttr[0] != eap.CodeFailure {
		t.Fatalf("expected EAP-Failure (code %d), got code %d", eap.CodeFailure, eapAttr[0])
	}
	if reply := rfc2865.ReplyMessage_GetString(resp); reply == "" {
		t.Fatal("expected Reply-Message to contain EAP failure reason")
	}
}

func TestEAPMD5IntegrationNakUnsupportedMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, rs := startEAPTestServer(t)
	seedEAPUser(t, rs, "eapuser", "eappass")
	client.username = "eapuser"

	challenge, state := client.sendIdentity(t)

	// Suggest EAP-GTC, which has no registered handler; the server must reject.
	resp := client.sendNak(t, challenge.Identifier, state, eap.TypeGTC)

	if resp.Code != radius.CodeAccessReject {
		t.Fatalf("expected Access-Reject after Nak to unsupported method, got %v", resp.Code)
	}
}

// setEapMethod switches the server-side configured EAP method for a test.
func setEapMethod(t *testing.T, rs *RadiusService, method string) {
	t.Helper()
	if err := rs.AppContext().ConfigMgr().Set("radius", "EapMethod", method); err != nil {
		t.Fatalf("failed to set EapMethod=%s: %v", method, err)
	}
}

// TestEAPTLSIntegrationStartThenSafeReject verifies the M1.1 EAP-TLS skeleton is
// wired end-to-end: selecting eap-tls makes the server answer identity with an
// EAP-TLS Start (RFC 5216 §2.1.1), and a subsequent handshake response is
// rejected because the handshake is not implemented yet (no auth bypass).
func TestEAPTLSIntegrationStartThenSafeReject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, rs := startEAPTestServer(t)
	seedEAPUser(t, rs, "eapuser", "eappass")
	setEapMethod(t, rs, "eap-tls")
	client.username = "eapuser"

	challenge, state := client.sendIdentity(t)
	if challenge.Type != eap.TypeTLS {
		t.Fatalf("expected EAP-TLS challenge, got EAP type %d", challenge.Type)
	}
	if len(challenge.Data) < 1 || challenge.Data[0]&0x20 == 0 {
		t.Fatalf("expected EAP-TLS Start (S) flag set, got data %x", challenge.Data)
	}

	// Client replies with a (dummy) handshake fragment; the skeleton must reject.
	tlsResp := &eap.EAPMessage{
		Code:       eap.CodeResponse,
		Identifier: challenge.Identifier,
		Type:       eap.TypeTLS,
		Data:       []byte{0x00}, // Flags only, no TLS data
	}
	packet := client.newAccessRequest()
	_ = rfc2865.State_Set(packet, state) //nolint:errcheck
	eap.SetEAPMessageAndAuth(packet, tlsResp.Encode(), client.secret)

	resp, err := client.exchange(t, packet)
	if err != nil {
		t.Fatalf("tls response exchange failed: %v", err)
	}
	if resp.Code != radius.CodeAccessReject {
		t.Fatalf("expected Access-Reject from EAP-TLS skeleton, got %v", resp.Code)
	}
	reply := strings.ToLower(rfc2865.ReplyMessage_GetString(resp))
	if !strings.Contains(reply, "eap-tls trust configuration missing") {
		t.Fatalf("expected explicit EAP-TLS failure reason, got %q", reply)
	}
}

func TestEAPMSCHAPv2IntegrationSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, rs := startEAPTestServer(t)
	seedEAPUser(t, rs, "eapuser", "eappass")
	setEapMethod(t, rs, "eap-mschapv2")
	client.username = "eapuser"

	challenge, state := client.sendIdentity(t)
	if challenge.Type != eap.TypeMSCHAPv2 {
		t.Fatalf("expected MSCHAPv2 challenge, got EAP type %d", challenge.Type)
	}

	resp := client.sendMSCHAPv2Response(t, challenge, state, "eappass")

	if resp.Code != radius.CodeAccessAccept {
		t.Fatalf("expected Access-Accept, got %v", resp.Code)
	}

	eapAttr := rfc2869.EAPMessage_Get(resp)
	if len(eapAttr) < 1 || eapAttr[0] != eap.CodeSuccess {
		t.Fatalf("expected EAP-Success in response, got %x", eapAttr)
	}

	if msSuccess := microsoft.MSCHAP2Success_Get(resp); len(msSuccess) == 0 {
		t.Fatal("Access-Accept missing MS-CHAP2-Success attribute")
	}
}

func TestEAPMSCHAPv2IntegrationWrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, rs := startEAPTestServer(t)
	seedEAPUser(t, rs, "eapuser", "eappass")
	setEapMethod(t, rs, "eap-mschapv2")
	client.username = "eapuser"

	challenge, state := client.sendIdentity(t)
	resp := client.sendMSCHAPv2Response(t, challenge, state, "wrong-password")

	if resp.Code != radius.CodeAccessReject {
		t.Fatalf("expected Access-Reject for wrong password, got %v", resp.Code)
	}

	eapAttr := rfc2869.EAPMessage_Get(resp)
	if len(eapAttr) < 1 || eapAttr[0] != eap.CodeFailure {
		t.Fatalf("expected EAP-Failure in response, got %x", eapAttr)
	}
}

// TestEAPNakNegotiationToMSCHAPv2 exercises the Nak path: the server is
// configured for EAP-MD5 but the supplicant rejects it and negotiates
// EAP-MSCHAPv2, completing the handshake successfully.
func TestEAPNakNegotiationToMSCHAPv2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, rs := startEAPTestServer(t)
	seedEAPUser(t, rs, "eapuser", "eappass")
	client.username = "eapuser"

	// Round 1: Identity -> server offers EAP-MD5 (default method).
	md5Challenge, state := client.sendIdentity(t)
	if md5Challenge.Type != eap.TypeMD5Challenge {
		t.Fatalf("expected MD5 challenge from default method, got EAP type %d", md5Challenge.Type)
	}

	// Round 2: Nak suggesting MSCHAPv2 -> server replies with MSCHAPv2 challenge.
	nakResp := client.sendNak(t, md5Challenge.Identifier, state, eap.TypeMSCHAPv2)
	if nakResp.Code != radius.CodeAccessChallenge {
		t.Fatalf("expected Access-Challenge after Nak, got %v", nakResp.Code)
	}

	mschapChallenge, err := eap.ParseEAPMessage(nakResp)
	if err != nil {
		t.Fatalf("failed to parse MSCHAPv2 challenge after Nak: %v", err)
	}
	if mschapChallenge.Type != eap.TypeMSCHAPv2 {
		t.Fatalf("expected MSCHAPv2 challenge after Nak, got EAP type %d", mschapChallenge.Type)
	}
	nakState := rfc2865.State_Get(nakResp)
	if len(nakState) == 0 {
		t.Fatal("MSCHAPv2 challenge missing State attribute")
	}

	// Round 3: MSCHAPv2 response -> Access-Accept.
	resp := client.sendMSCHAPv2Response(t, mschapChallenge, nakState, "eappass")
	if resp.Code != radius.CodeAccessAccept {
		t.Fatalf("expected Access-Accept after MSCHAPv2 response, got %v", resp.Code)
	}

	eapAttr := rfc2869.EAPMessage_Get(resp)
	if len(eapAttr) < 1 || eapAttr[0] != eap.CodeSuccess {
		t.Fatalf("expected EAP-Success after Nak negotiation, got %x", eapAttr)
	}
}
