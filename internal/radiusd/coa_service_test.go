package radiusd

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc3576"
)

// capturedReq holds the identity-relevant fields extracted from a request the
// fake NAS received, copied out under the handler lock so the test goroutine
// never reads a radius.Packet owned by the server goroutine.
type capturedReq struct {
	code           radius.Code
	identifier     byte
	username       string
	acctSessionID  string
	nasIP          string
	framedIP       string
	filterID       string
	sessionTimeout uint32
	hasTimeout     bool
}

// fakeNAS is an in-process UDP RADIUS responder that emulates a NAS receiving
// CoA/Disconnect requests, used to drive the CoAService send path end to end.
type fakeNAS struct {
	addr   string
	server *radius.PacketServer

	mu        sync.Mutex
	replyCode radius.Code        // 0 => drop the request (no reply)
	errCause  rfc3576.ErrorCause // added to NAK replies when non-zero
	dropFirst int                // drop the first N requests, then use replyCode
	received  []capturedReq
}

func newFakeNAS(t *testing.T, secret string, replyCode radius.Code) *fakeNAS {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake nas: %v", err)
	}
	fn := &fakeNAS{
		addr:      pc.LocalAddr().String(),
		replyCode: replyCode,
	}
	fn.server = &radius.PacketServer{
		Handler:            radius.HandlerFunc(fn.handle),
		SecretSource:       radius.StaticSecretSource([]byte(secret)),
		InsecureSkipVerify: true,
	}
	go func() { _ = fn.server.Serve(pc) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = fn.server.Shutdown(ctx)
	})
	return fn
}

func (fn *fakeNAS) handle(w radius.ResponseWriter, r *radius.Request) {
	cap := capturedReq{
		code:          r.Code,
		identifier:    r.Identifier,
		username:      rfc2865.UserName_GetString(r.Packet),
		acctSessionID: rfc2866.AcctSessionID_GetString(r.Packet),
		filterID:      rfc2865.FilterID_GetString(r.Packet),
	}
	if ip := rfc2865.NASIPAddress_Get(r.Packet); ip != nil {
		cap.nasIP = ip.String()
	}
	if ip := rfc2865.FramedIPAddress_Get(r.Packet); ip != nil {
		cap.framedIP = ip.String()
	}
	if st, err := rfc2865.SessionTimeout_Lookup(r.Packet); err == nil {
		cap.hasTimeout = true
		cap.sessionTimeout = uint32(st)
	}

	fn.mu.Lock()
	fn.received = append(fn.received, cap)
	drop := fn.dropFirst > 0
	if drop {
		fn.dropFirst--
	}
	code := fn.replyCode
	cause := fn.errCause
	fn.mu.Unlock()

	if drop || code == 0 {
		return // no response => the client times out
	}
	resp := r.Response(code)
	if cause != 0 && (code == radius.CodeDisconnectNAK || code == radius.CodeCoANAK) {
		_ = rfc3576.ErrorCause_Add(resp, cause)
	}
	_ = w.Write(resp)
}

func (fn *fakeNAS) port(t *testing.T) int {
	t.Helper()
	_, portStr, err := net.SplitHostPort(fn.addr)
	if err != nil {
		t.Fatalf("split fake nas addr: %v", err)
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse fake nas port: %v", err)
	}
	return p
}

func (fn *fakeNAS) snapshot() []capturedReq {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	out := make([]capturedReq, len(fn.received))
	copy(out, fn.received)
	return out
}

func (fn *fakeNAS) setBehavior(replyCode radius.Code, cause rfc3576.ErrorCause, dropFirst int) {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	fn.replyCode = replyCode
	fn.errCause = cause
	fn.dropFirst = dropFirst
}

func TestCoAServiceDisconnectACK(t *testing.T) {
	const secret = "testing123"
	nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)
	svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(2))

	nasPort := uint32(7)
	id := SessionIdentity{
		Username:       "alice",
		NasIP:          "10.0.0.1",
		AcctSessionID:  "sess-ack-1",
		FramedIP:       "100.64.0.9",
		CallingStation: "AA:BB:CC:00:11:22",
		NasPort:        &nasPort,
		NasPortID:      "gi0/1",
	}
	target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

	result, err := svc.Disconnect(context.Background(), target, id)
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got result %+v", result)
	}
	if result.Action != CoAActionDisconnect {
		t.Errorf("action = %q, want %q", result.Action, CoAActionDisconnect)
	}
	if result.ResponseCode != "Disconnect-ACK" {
		t.Errorf("response code = %q, want Disconnect-ACK", result.ResponseCode)
	}
	if result.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", result.Attempts)
	}
	if result.ErrorCause != 0 {
		t.Errorf("error cause = %d, want 0", result.ErrorCause)
	}
	if result.TimedOut {
		t.Error("unexpected TimedOut=true")
	}

	got := nas.snapshot()
	if len(got) != 1 {
		t.Fatalf("fake nas received %d packets, want 1", len(got))
	}
	first := got[0]
	if first.code != radius.CodeDisconnectRequest {
		t.Errorf("nas saw code %v, want Disconnect-Request", first.code)
	}
	if first.username != "alice" || first.acctSessionID != "sess-ack-1" {
		t.Errorf("identity mismatch: username=%q acctSessionID=%q", first.username, first.acctSessionID)
	}
	if first.nasIP != "10.0.0.1" || first.framedIP != "100.64.0.9" {
		t.Errorf("address attrs mismatch: nasIP=%q framedIP=%q", first.nasIP, first.framedIP)
	}
}

func TestCoAServiceDisconnectNAK(t *testing.T) {
	const secret = "testing123"
	nas := newFakeNAS(t, secret, radius.CodeDisconnectNAK)
	nas.setBehavior(radius.CodeDisconnectNAK, rfc3576.ErrorCause_Value_SessionContextNotFound, 0)
	svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(2))

	id := SessionIdentity{Username: "bob", NasIP: "10.0.0.1", AcctSessionID: "missing-sess"}
	target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

	result, err := svc.Disconnect(context.Background(), target, id)
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if result.Success {
		t.Fatalf("expected failure, got success: %+v", result)
	}
	if result.ResponseCode != "Disconnect-NAK" {
		t.Errorf("response code = %q, want Disconnect-NAK", result.ResponseCode)
	}
	if result.ErrorCause != int(rfc3576.ErrorCause_Value_SessionContextNotFound) {
		t.Errorf("error cause = %d, want %d", result.ErrorCause, rfc3576.ErrorCause_Value_SessionContextNotFound)
	}
	if result.ErrorCauseText != "Session-Context-Not-Found" {
		t.Errorf("error cause text = %q, want Session-Context-Not-Found", result.ErrorCauseText)
	}
}

func TestCoAServiceCoAACKWithChanges(t *testing.T) {
	const secret = "testing123"
	nas := newFakeNAS(t, secret, radius.CodeCoAACK)
	svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(1))

	id := SessionIdentity{Username: "carol", NasIP: "10.0.0.2", AcctSessionID: "sess-coa-1"}
	target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

	result, err := svc.CoA(context.Background(), target, id, WithSessionTimeout(3600), WithFilterID("throttled"))
	if err != nil {
		t.Fatalf("CoA returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got %+v", result)
	}
	if result.Action != CoAActionCoA {
		t.Errorf("action = %q, want %q", result.Action, CoAActionCoA)
	}
	if result.ResponseCode != "CoA-ACK" {
		t.Errorf("response code = %q, want CoA-ACK", result.ResponseCode)
	}

	got := nas.snapshot()
	if len(got) != 1 {
		t.Fatalf("fake nas received %d packets, want 1", len(got))
	}
	change := got[0]
	if change.code != radius.CodeCoARequest {
		t.Errorf("nas saw code %v, want CoA-Request", change.code)
	}
	if !change.hasTimeout || change.sessionTimeout != 3600 {
		t.Errorf("Session-Timeout not applied: has=%v value=%d", change.hasTimeout, change.sessionTimeout)
	}
	if change.filterID != "throttled" {
		t.Errorf("Filter-Id = %q, want throttled", change.filterID)
	}
}

func TestCoAServiceTimeoutAndRetry(t *testing.T) {
	const secret = "testing123"
	nas := newFakeNAS(t, secret, 0) // drop everything => always times out
	svc := NewCoAService(nil, WithCoATimeout(150*time.Millisecond), WithCoARetries(2))

	id := SessionIdentity{Username: "dave", NasIP: "10.0.0.1", AcctSessionID: "sess-timeout"}
	target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

	start := time.Now()
	result, err := svc.Disconnect(context.Background(), target, id)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if result.Success {
		t.Fatalf("expected failure on timeout, got %+v", result)
	}
	if !result.TimedOut {
		t.Error("expected TimedOut=true")
	}
	if result.Attempts != 3 {
		t.Errorf("attempts = %d, want 3 (initial + 2 retries)", result.Attempts)
	}
	if elapsed < 300*time.Millisecond {
		t.Errorf("elapsed %v too short for 3 attempts of 150ms", elapsed)
	}

	// All retransmissions must reuse the same Identifier (RFC 5176 §2.3).
	got := nas.snapshot()
	if len(got) == 0 {
		t.Fatal("fake nas received no packets")
	}
	for i, r := range got {
		if int(r.identifier) != result.Identifier {
			t.Errorf("packet %d identifier = %d, want %d (retransmissions must match)", i, r.identifier, result.Identifier)
		}
	}
}

func TestCoAServiceRetryThenSuccess(t *testing.T) {
	const secret = "testing123"
	nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)
	nas.setBehavior(radius.CodeDisconnectACK, 0, 1) // drop the first, ACK the rest
	svc := NewCoAService(nil, WithCoATimeout(150*time.Millisecond), WithCoARetries(2))

	id := SessionIdentity{Username: "erin", NasIP: "10.0.0.1", AcctSessionID: "sess-retry"}
	target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

	result, err := svc.Disconnect(context.Background(), target, id)
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success after retry, got %+v", result)
	}
	if result.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", result.Attempts)
	}
}

func TestCoAServiceNonTimeoutErrorNoRetry(t *testing.T) {
	svc := NewCoAService(nil, WithCoATimeout(time.Second), WithCoARetries(3))
	sentinel := errors.New("dial failure")
	calls := 0
	svc.exchange = func(_ context.Context, _ *radius.Packet, _ string) (*radius.Packet, error) {
		calls++
		return nil, sentinel
	}

	id := SessionIdentity{Username: "frank", AcctSessionID: "sess-x"}
	target := CoATarget{Addr: "127.0.0.1", Secret: "s", Port: 3799}

	result, err := svc.Disconnect(context.Background(), target, id)
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if result.Success || result.TimedOut {
		t.Errorf("unexpected success/timeout flags: %+v", result)
	}
	if result.Attempts != 1 {
		t.Errorf("attempts = %d, want 1 (non-timeout error must not retry)", result.Attempts)
	}
	if calls != 1 {
		t.Errorf("exchange called %d times, want 1", calls)
	}
	if result.Err != sentinel.Error() {
		t.Errorf("err = %q, want %q", result.Err, sentinel.Error())
	}
}

func TestCoAServiceMissingTargetAddr(t *testing.T) {
	svc := NewCoAService(nil)
	if _, err := svc.Disconnect(context.Background(), CoATarget{}, SessionIdentity{}); !errors.Is(err, ErrNoTarget) {
		t.Errorf("Disconnect err = %v, want ErrNoTarget", err)
	}
	if _, err := svc.CoA(context.Background(), CoATarget{}, SessionIdentity{}); !errors.Is(err, ErrNoTarget) {
		t.Errorf("CoA err = %v, want ErrNoTarget", err)
	}
}

func TestCoATargetEndpointPortFallback(t *testing.T) {
	if got := (CoATarget{Addr: "192.0.2.10"}).endpoint(); got != "192.0.2.10:3799" {
		t.Errorf("endpoint with zero port = %q, want 192.0.2.10:3799", got)
	}
	if got := (CoATarget{Addr: "192.0.2.10", Port: 1700}).endpoint(); got != "192.0.2.10:1700" {
		t.Errorf("endpoint with explicit port = %q, want 192.0.2.10:1700", got)
	}
}

func TestSessionIdentityFromOnline(t *testing.T) {
	online := &domain.RadiusOnline{
		Username:      "user1",
		NasAddr:       "10.0.0.1",
		NasId:         "nas-a",
		AcctSessionId: "s-1",
		FramedIpaddr:  "100.64.0.5",
		MacAddr:       "AA:BB:CC:DD:EE:FF",
		NasPort:       42,
		NasPortId:     "port-1",
	}
	id := SessionIdentityFromOnline(online)
	if id.Username != "user1" || id.NasIP != "10.0.0.1" || id.NasIdentifier != "nas-a" {
		t.Errorf("nas/user mapping wrong: %+v", id)
	}
	if id.AcctSessionID != "s-1" || id.FramedIP != "100.64.0.5" || id.CallingStation != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("session mapping wrong: %+v", id)
	}
	if id.NasPortID != "port-1" {
		t.Errorf("NasPortID = %q, want port-1", id.NasPortID)
	}
	if id.NasPort == nil || *id.NasPort != 42 {
		t.Errorf("NasPort = %v, want 42", id.NasPort)
	}

	zero := SessionIdentityFromOnline(&domain.RadiusOnline{NasPort: 0})
	if zero.NasPort != nil {
		t.Errorf("NasPort for stored 0 = %v, want nil (omitted)", zero.NasPort)
	}
}

func TestCoATargetFromNas(t *testing.T) {
	target := CoATargetFromNas(&domain.NetNas{Ipaddr: "10.0.0.1", Secret: "sec", CoaPort: 3799})
	if target.Addr != "10.0.0.1" || target.Secret != "sec" || target.Port != 3799 {
		t.Errorf("CoATargetFromNas = %+v", target)
	}
}

func TestCoAServiceDisconnectSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB-backed CoA test in short mode")
	}
	appCtx, _ := setupTestEnv(t)
	defer appCtx.Release()

	radiusService := NewRadiusService(appCtx)
	defer radiusService.Release()

	const secret = "db-secret"
	nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)

	db := appCtx.DB()
	netNas := &domain.NetNas{
		ID:         101,
		Identifier: "db-nas",
		Ipaddr:     "127.0.0.1",
		Secret:     secret,
		CoaPort:    nas.port(t),
		VendorCode: "0",
		Status:     common.ENABLED,
	}
	if err := db.Create(netNas).Error; err != nil {
		t.Fatalf("seed nas: %v", err)
	}
	online := &domain.RadiusOnline{
		ID:            201,
		Username:      "dbuser",
		NasId:         "db-nas",
		NasAddr:       "127.0.0.1",
		AcctSessionId: "db-sess-1",
		FramedIpaddr:  "100.64.0.20",
		AcctStartTime: time.Now(),
		LastUpdate:    time.Now(),
	}
	if err := db.Create(online).Error; err != nil {
		t.Fatalf("seed online session: %v", err)
	}

	svc := NewCoAService(radiusService, WithCoATimeout(2*time.Second), WithCoARetries(1))

	result, err := svc.DisconnectSession(context.Background(), "db-sess-1")
	if err != nil {
		t.Fatalf("DisconnectSession error: %v", err)
	}
	if !result.Success || result.ResponseCode != "Disconnect-ACK" {
		t.Fatalf("expected Disconnect-ACK success, got %+v", result)
	}
	if result.Username != "dbuser" || result.AcctSessionID != "db-sess-1" {
		t.Errorf("result identity mismatch: %+v", result)
	}

	got := nas.snapshot()
	if len(got) != 1 || got[0].username != "dbuser" || got[0].acctSessionID != "db-sess-1" {
		t.Errorf("fake nas did not receive expected disconnect: %+v", got)
	}

	if _, err := svc.DisconnectSession(context.Background(), "no-such-session"); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("missing session err = %v, want ErrSessionNotFound", err)
	}
}
