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

// replyAuthMode controls whether the fake NAS signs its reply with a
// Message-Authenticator (RFC 5176 §3.4) and, when it does, whether the signature
// is valid or deliberately corrupted.
type replyAuthMode int

const (
	replyAuthNone    replyAuthMode = iota // omit Message-Authenticator on the reply
	replyAuthValid                        // include a correctly computed Message-Authenticator
	replyAuthCorrupt                      // include a Message-Authenticator with a wrong value
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
	replyAuth replyAuthMode      // how the reply is signed (default: valid)
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
		secret:    secret,
		replyCode: replyCode,
		replyAuth: replyAuthValid,
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
	replyAuth := fn.replyAuth
	fn.mu.Unlock()

	if drop || code == 0 {
		return // no response => the client times out
	}
	resp := r.Response(code)
	if cause != 0 && (code == radius.CodeDisconnectNAK || code == radius.CodeCoANAK) {
		_ = rfc3576.ErrorCause_Add(resp, cause)
	}
	signReply(resp, replyAuth)
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

// signReply attaches a Message-Authenticator to an ACK/NAK reply according to
// mode. resp.Authenticator is the request authenticator (copied by Response),
// which is exactly the value the client recomputes the reply digest with.
func signReply(resp *radius.Packet, mode replyAuthMode) {
	switch mode {
	case replyAuthValid:
		_ = rfc2869.MessageAuthenticator_Set(resp, make([]byte, 16))
		b, err := resp.MarshalBinary()
		if err != nil {
			return
		}
		mac := hmac.New(md5.New, resp.Secret)
		mac.Write(b)
		_ = rfc2869.MessageAuthenticator_Set(resp, mac.Sum(nil))
	case replyAuthCorrupt:
		bad := make([]byte, 16)
		bad[0] = 0xFF
		_ = rfc2869.MessageAuthenticator_Set(resp, bad)
	case replyAuthNone:
		// leave the reply unsigned
	}
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

// TestCoAServiceRejectsInvalidReplyMessageAuthenticator asserts that a reply
// carrying a Message-Authenticator that does not verify is discarded
// (RFC 5176 §3.4): the result is not reported as a successful ACK.
func TestCoAServiceRejectsInvalidReplyMessageAuthenticator(t *testing.T) {
	const secret = "testing123"
	nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)
	nas.setReplyAuth(replyAuthCorrupt)
	svc := NewCoAService(nil, WithCoATimeout(150*time.Millisecond), WithCoARetries(1))
	target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

	result, err := svc.Disconnect(context.Background(), target, SessionIdentity{Username: "alice", AcctSessionID: "s1"})
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if result.Success {
		t.Fatalf("reply with bad Message-Authenticator must not be treated as success: %+v", result)
	}
	if result.Err != errInvalidMessageAuthenticator.Error() {
		t.Errorf("result.Err = %q, want %q", result.Err, errInvalidMessageAuthenticator.Error())
	}
}

// TestCoAServiceAcceptsUnsignedReply asserts that a reply without a
// Message-Authenticator is accepted, since the attribute is OPTIONAL (0-1) in
// replies (RFC 5176 §3.4).
func TestCoAServiceAcceptsUnsignedReply(t *testing.T) {
	const secret = "testing123"
	nas := newFakeNAS(t, secret, radius.CodeDisconnectACK)
	nas.setReplyAuth(replyAuthNone)
	svc := NewCoAService(nil, WithCoATimeout(2*time.Second), WithCoARetries(1))
	target := CoATarget{Addr: "127.0.0.1", Secret: secret, Port: nas.port(t)}

	result, err := svc.Disconnect(context.Background(), target, SessionIdentity{Username: "alice", AcctSessionID: "s1"})
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("unsigned reply must still be accepted: %+v", result)
	}
}
