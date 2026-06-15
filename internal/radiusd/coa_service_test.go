package radiusd

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
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
	hasMessageAuth bool // request carried a Message-Authenticator attribute
	messageAuthOK  bool // the carried Message-Authenticator verified against the secret
}

// fakeNAS is an in-process UDP RADIUS responder that emulates a NAS receiving
// CoA/Disconnect requests, used to drive the CoAService send path end to end.
type fakeNAS struct {
	addr   string
	server *radius.PacketServer
	secret string

	mu        sync.Mutex
	replyCode radius.Code        // 0 => drop the request (no reply)
	errCause  rfc3576.ErrorCause // added to NAK replies when non-zero
	dropFirst int                // drop the first N requests, then use replyCode
	replyAuth replyAuthMode      // how the reply carries a Message-Authenticator
	received  []capturedReq
}

// replyAuthMode controls how the fake NAS signs its CoA/Disconnect reply, so
// tests can exercise the RFC 5176 §3.4 reply-side Message-Authenticator paths.
type replyAuthMode int

const (
	// replyAuthNone leaves the reply unsigned (no Message-Authenticator). It is
	// the zero value and the default behavior.
	replyAuthNone replyAuthMode = iota
	// replyAuthSigned attaches a valid Message-Authenticator (RFC 5176 §3.4).
	replyAuthSigned
	// replyAuthCorrupt attaches a present but invalid Message-Authenticator by
	// signing with the wrong secret, so the attribute is structurally valid
	// (sixteen octets) yet fails the client's check.
	replyAuthCorrupt
)

// signCoAReply signs a CoA/Disconnect reply with a Message-Authenticator the way
// a NAS does per RFC 5176 §3.4: the request authenticator (carried into
// resp.Authenticator by r.Response) sits in the Authenticator field, the
// attribute value is treated as sixteen zero octets for the HMAC-MD5, and the
// result is stored before the transport computes the Response Authenticator.
func signCoAReply(resp *radius.Packet, secret string) {
	_ = rfc2869.MessageAuthenticator_Set(resp, make([]byte, 16))
	if b, err := resp.MarshalBinary(); err == nil {
		mac := hmac.New(md5.New, []byte(secret))
		mac.Write(b)
		_ = rfc2869.MessageAuthenticator_Set(resp, mac.Sum(nil))
	}
}

func newFakeNAS(t *testing.T, secret string, replyCode radius.Code) *fakeNAS {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake nas: %v", err)
	}
	fn := &fakeNAS{
		addr:      pc.LocalAddr().String(),
		secret:    secret,
		replyCode: replyCode,
	}
	fn.server = &radius.PacketServer{
		Handler:      radius.HandlerFunc(fn.handle),
		SecretSource: radius.StaticSecretSource([]byte(secret)),
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
	// Act as a strict NAS (RFC 5176 §3.4): verify the request Message-Authenticator.
	cap.hasMessageAuth, cap.messageAuthOK = verifyRequestMessageAuth(r.Packet, fn.secret)

	fn.mu.Lock()
	fn.received = append(fn.received, cap)
	drop := fn.dropFirst > 0
	if drop {
		fn.dropFirst--
	}
	code := fn.replyCode
	cause := fn.errCause
	authMode := fn.replyAuth
	fn.mu.Unlock()

	if drop || code == 0 {
		return // no response => the client times out
	}
	resp := r.Response(code)
	if cause != 0 && (code == radius.CodeDisconnectNAK || code == radius.CodeCoANAK) {
		_ = rfc3576.ErrorCause_Add(resp, cause)
	}
	switch authMode {
	case replyAuthSigned:
		signCoAReply(resp, fn.secret)
	case replyAuthCorrupt:
		signCoAReply(resp, fn.secret+"-wrong")
	}
	_ = w.Write(resp)
}

// verifyRequestMessageAuth reports whether an inbound CoA/Disconnect request
// carries a Message-Authenticator and, if so, whether it verifies per RFC 5176
// §3.4: HMAC-MD5 over the packet with the Request Authenticator field and the
// Message-Authenticator value treated as sixteen zero octets.
func verifyRequestMessageAuth(p *radius.Packet, secret string) (present, valid bool) {
	got, err := rfc2869.MessageAuthenticator_Lookup(p)
	if err != nil {
		return false, false
	}
	saved := append([]byte(nil), got...)
	_ = rfc2869.MessageAuthenticator_Set(p, make([]byte, 16))
	b, mErr := p.MarshalBinary()
	_ = rfc2869.MessageAuthenticator_Set(p, saved)
	if mErr != nil {
		return true, false
	}
	for i := 4; i < 20; i++ {
		b[i] = 0 // zero the Request Authenticator field
	}
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write(b)
	return true, hmac.Equal(mac.Sum(nil), saved)
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

// setReplyAuth selects how the fake NAS signs its CoA/Disconnect reply so tests
// can drive the RFC 5176 §3.4 reply-side validation paths.
func (fn *fakeNAS) setReplyAuth(mode replyAuthMode) {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	fn.replyAuth = mode
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
	if !first.hasMessageAuth || !first.messageAuthOK {
		t.Errorf("request Message-Authenticator: present=%v valid=%v, want both true", first.hasMessageAuth, first.messageAuthOK)
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

func TestCoAServiceSessionHelpersRequireRadiusService(t *testing.T) {
	// The *Session helpers resolve a live session through the repositories, so a
	// nil embedded RadiusService must be reported rather than panicking. The
	// end-to-end DB-backed path (session -> NAS -> send) is covered by the M2.2
	// Admin API tests and the M2.6 integration acceptance case.
	svc := NewCoAService(nil)
	if _, err := svc.DisconnectSession(context.Background(), "sess"); err == nil {
		t.Error("DisconnectSession with nil RadiusService: expected error, got nil")
	}
	if _, err := svc.CoASession(context.Background(), "sess"); err == nil {
		t.Error("CoASession with nil RadiusService: expected error, got nil")
	}
}

// TestCoAServiceRequestMessageAuthenticator asserts that every outbound
// CoA/Disconnect request carries a correctly computed Message-Authenticator
// (RFC 5176 §3.4). The fake NAS verifies the attribute and records the result.
func TestCoAServiceRequestMessageAuthenticator(t *testing.T) {
	const secret = "testing123"

	t.Run("disconnect", func(t *testing.T) {
		nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)
		svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(1))
		target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

		result, err := svc.Disconnect(context.Background(), target, SessionIdentity{Username: "alice", AcctSessionID: "s1"})
		if err != nil {
			t.Fatalf("Disconnect returned error: %v", err)
		}
		if !result.Success {
			t.Fatalf("expected success, got %+v", result)
		}
		got := nas.snapshot()
		if len(got) != 1 {
			t.Fatalf("received %d packets, want 1", len(got))
		}
		if !got[0].hasMessageAuth {
			t.Fatal("Disconnect-Request did not carry a Message-Authenticator")
		}
		if !got[0].messageAuthOK {
			t.Error("Disconnect-Request Message-Authenticator did not verify")
		}
	})

	t.Run("coa", func(t *testing.T) {
		nas := newFakeNAS(t, secret, radius.CodeCoAACK)
		svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(1))
		target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

		result, err := svc.CoA(context.Background(), target, SessionIdentity{Username: "carol", AcctSessionID: "s2"}, WithSessionTimeout(60))
		if err != nil {
			t.Fatalf("CoA returned error: %v", err)
		}
		if !result.Success {
			t.Fatalf("expected success, got %+v", result)
		}
		got := nas.snapshot()
		if len(got) != 1 {
			t.Fatalf("received %d packets, want 1", len(got))
		}
		if !got[0].hasMessageAuth || !got[0].messageAuthOK {
			t.Errorf("CoA-Request Message-Authenticator: present=%v valid=%v, want both true", got[0].hasMessageAuth, got[0].messageAuthOK)
		}
	})
}

// TestCoAServiceResponseMessageAuthenticator drives the RFC 5176 §3.4 reply-side
// validation through the full send path: a correctly signed reply is accepted, an
// unsigned reply is accepted (the attribute is OPTIONAL on responses), and a reply
// whose Message-Authenticator does not verify is silently discarded.
func TestCoAServiceResponseMessageAuthenticator(t *testing.T) {
const secret = "testing123"

t.Run("signed reply accepted", func(t *testing.T) {
nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)
nas.setReplyAuth(replyAuthSigned)
svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(2))
target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

result, err := svc.Disconnect(context.Background(), target, SessionIdentity{Username: "alice", AcctSessionID: "s1"})
if err != nil {
t.Fatalf("Disconnect returned error: %v", err)
}
if !result.Success {
t.Fatalf("expected success with a valid reply Message-Authenticator, got %+v", result)
}
if result.ResponseCode != "Disconnect-ACK" {
t.Errorf("response code = %q, want Disconnect-ACK", result.ResponseCode)
}
if result.Attempts != 1 {
t.Errorf("attempts = %d, want 1", result.Attempts)
}
})

t.Run("unsigned reply accepted", func(t *testing.T) {
nas := newFakeNAS(t, secret, radius.CodeCoAACK)
nas.setReplyAuth(replyAuthNone) // reply carries no Message-Authenticator
svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(2))
target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

result, err := svc.CoA(context.Background(), target, SessionIdentity{Username: "bob", AcctSessionID: "s2"}, WithSessionTimeout(60))
if err != nil {
t.Fatalf("CoA returned error: %v", err)
}
if !result.Success {
t.Fatalf("expected success with an unsigned reply (reply MA is OPTIONAL), got %+v", result)
}
if result.Attempts != 1 {
t.Errorf("attempts = %d, want 1", result.Attempts)
}
})

t.Run("invalid reply discarded", func(t *testing.T) {
nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)
nas.setReplyAuth(replyAuthCorrupt)
svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(2))
target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

result, err := svc.Disconnect(context.Background(), target, SessionIdentity{Username: "carol", AcctSessionID: "s3"})
if err != nil {
t.Fatalf("Disconnect returned error: %v", err)
}
if result.Success {
t.Fatalf("expected the forged reply to be discarded, got success: %+v", result)
}
if result.TimedOut {
t.Error("a discarded reply must not be reported as a timeout")
}
if result.ResponseCode != "" {
t.Errorf("response code = %q, want empty (no reply accepted)", result.ResponseCode)
}
if result.Err == "" {
t.Error("expected Err to describe the discard")
}
// Every discarded reply is treated as unanswered, so the full budget is
// spent: initial transmission + 2 retransmissions.
if result.Attempts != 3 {
t.Errorf("attempts = %d, want 3 (initial + 2 retransmissions)", result.Attempts)
}
if got := len(nas.snapshot()); got != 3 {
t.Errorf("nas received %d packets, want 3", got)
}
})
}

// TestVerifyResponseMessageAuthenticator exercises the pure RFC 5176 §3.4 reply
// validator against signed, unsigned, malformed, and forged vectors.
func TestVerifyResponseMessageAuthenticator(t *testing.T) {
const secret = "respsecret"
var reqAuth [16]byte
for i := range reqAuth {
reqAuth[i] = byte(i + 1)
}

// build returns a Disconnect-ACK signed the way a NAS does (RFC 5176 §3.4):
// keyed on reqAuth with the Message-Authenticator value zeroed during HMAC.
build := func() *radius.Packet {
p := radius.New(radius.CodeDisconnectACK, []byte(secret))
p.Identifier = 7
p.Authenticator = reqAuth
if err := rfc2869.MessageAuthenticator_Set(p, make([]byte, 16)); err != nil {
t.Fatalf("set placeholder MA: %v", err)
}
b, err := p.MarshalBinary()
if err != nil {
t.Fatalf("marshal: %v", err)
}
mac := hmac.New(md5.New, []byte(secret))
mac.Write(b)
if err := rfc2869.MessageAuthenticator_Set(p, mac.Sum(nil)); err != nil {
t.Fatalf("set MA: %v", err)
}
return p
}

t.Run("valid", func(t *testing.T) {
if got := verifyResponseMessageAuthenticator(build(), reqAuth, []byte(secret)); got != msgAuthValid {
t.Errorf("got %v, want msgAuthValid", got)
}
})

t.Run("keyed on request authenticator not the carried one", func(t *testing.T) {
p := build()
// Simulate the parsed reply carrying a Response Authenticator that differs
// from the request authenticator; verification must still key on reqAuth.
for i := range p.Authenticator {
p.Authenticator[i] = 0xAA
}
if got := verifyResponseMessageAuthenticator(p, reqAuth, []byte(secret)); got != msgAuthValid {
t.Errorf("got %v, want msgAuthValid", got)
}
})

t.Run("absent accepted", func(t *testing.T) {
p := radius.New(radius.CodeDisconnectACK, []byte(secret))
if got := verifyResponseMessageAuthenticator(p, reqAuth, []byte(secret)); got != msgAuthAbsent {
t.Errorf("got %v, want msgAuthAbsent", got)
}
})

t.Run("nil accepted", func(t *testing.T) {
if got := verifyResponseMessageAuthenticator(nil, reqAuth, []byte(secret)); got != msgAuthAbsent {
t.Errorf("got %v, want msgAuthAbsent", got)
}
})

t.Run("tampered value discarded", func(t *testing.T) {
p := build()
v, err := rfc2869.MessageAuthenticator_Lookup(p)
if err != nil {
t.Fatalf("lookup MA: %v", err)
}
tampered := append([]byte(nil), v...)
tampered[0] ^= 0xFF
if err := rfc2869.MessageAuthenticator_Set(p, tampered); err != nil {
t.Fatalf("set tampered MA: %v", err)
}
if got := verifyResponseMessageAuthenticator(p, reqAuth, []byte(secret)); got != msgAuthInvalid {
t.Errorf("got %v, want msgAuthInvalid", got)
}
})

t.Run("wrong request authenticator discarded", func(t *testing.T) {
var other [16]byte
for i := range other {
other[i] = 0x99
}
if got := verifyResponseMessageAuthenticator(build(), other, []byte(secret)); got != msgAuthInvalid {
t.Errorf("got %v, want msgAuthInvalid", got)
}
})

t.Run("wrong length discarded", func(t *testing.T) {
p := build()
if err := rfc2869.MessageAuthenticator_Set(p, make([]byte, 8)); err != nil {
t.Fatalf("set short MA: %v", err)
}
if got := verifyResponseMessageAuthenticator(p, reqAuth, []byte(secret)); got != msgAuthInvalid {
t.Errorf("got %v, want msgAuthInvalid", got)
}
})

t.Run("duplicate attribute discarded", func(t *testing.T) {
p := build()
if err := rfc2869.MessageAuthenticator_Add(p, make([]byte, 16)); err != nil {
t.Fatalf("add duplicate MA: %v", err)
}
if got := verifyResponseMessageAuthenticator(p, reqAuth, []byte(secret)); got != msgAuthInvalid {
t.Errorf("got %v, want msgAuthInvalid", got)
}
})

t.Run("present but no secret discarded", func(t *testing.T) {
if got := verifyResponseMessageAuthenticator(build(), reqAuth, nil); got != msgAuthInvalid {
t.Errorf("got %v, want msgAuthInvalid", got)
}
})
}
