//go:build integration

package integration

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"layeh.com/radius"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3576"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// coaActionBody mirrors the adminapi coaActionResponse JSON returned by the
// /sessions/:id/disconnect and /sessions/:id/coa endpoints. It is duplicated
// here (rather than imported) because the integration package exercises the API
// strictly over HTTP, as a real client would.
type coaActionBody struct {
	Action         string `json:"action"`
	Target         string `json:"target"`
	Username       string `json:"username"`
	AcctSessionID  string `json:"acct_session_id"`
	Success        bool   `json:"success"`
	ResponseCode   string `json:"response_code"`
	ErrorCause     int    `json:"error_cause"`
	ErrorCauseText string `json:"error_cause_text"`
	Attempts       int    `json:"attempts"`
	TimedOut       bool   `json:"timed_out"`
}

// itFakeNAS is an in-process UDP RADIUS responder standing in for a real NAS, so
// the CoA/Disconnect endpoints are exercised end to end over a real socket: the
// HTTP handler builds and signs an RFC 5176 request, transmits it, and parses
// the reply this responder returns.
type itFakeNAS struct {
	addr   string
	server *radius.PacketServer
	secret string

	mu        sync.Mutex
	replyCode radius.Code // 0 => drop the request (force a client timeout)
	errCause  rfc3576.ErrorCause
	received  []radius.Code
	badAuth   int // count of requests missing/failing Message-Authenticator
}

func newITFakeNAS(t *testing.T, secret string, replyCode radius.Code, cause rfc3576.ErrorCause) *itFakeNAS {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	fn := &itFakeNAS{addr: pc.LocalAddr().String(), secret: secret, replyCode: replyCode, errCause: cause}
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

func (fn *itFakeNAS) handle(w radius.ResponseWriter, r *radius.Request) {
	// Strict NAS (RFC 5176 §3.4): require a valid Message-Authenticator on every
	// CoA/Disconnect request. A request that lacks or fails it is silently
	// dropped, which surfaces as a client timeout — exactly how a real
	// FreeRADIUS-style NAS behaves and what makes the omission observable.
	present, valid := itVerifyRequestMessageAuth(r.Packet, fn.secret)

	fn.mu.Lock()
	fn.received = append(fn.received, r.Code)
	if !present || !valid {
		fn.badAuth++
	}
	code := fn.replyCode
	cause := fn.errCause
	fn.mu.Unlock()

	if !present || !valid {
		return
	}
	if code == 0 {
		return // no response => the client times out
	}
	resp := r.Response(code)
	if cause != 0 && (code == radius.CodeDisconnectNAK || code == radius.CodeCoANAK) {
		_ = rfc3576.ErrorCause_Add(resp, cause)
	}
	// Sign the reply (RFC 5176 §3.4) so the Dynamic Authorization Client can
	// authenticate it. resp.Authenticator is the request authenticator, which is
	// the value the client recomputes the reply digest with.
	_ = rfc2869.MessageAuthenticator_Set(resp, make([]byte, 16))
	if b, err := resp.MarshalBinary(); err == nil {
		mac := hmac.New(md5.New, resp.Secret)
		mac.Write(b)
		_ = rfc2869.MessageAuthenticator_Set(resp, mac.Sum(nil))
	}
	_ = w.Write(resp)
}

// itVerifyRequestMessageAuth verifies the Message-Authenticator on an inbound
// CoA/Disconnect request per RFC 5176 §3.4 (Request Authenticator field and the
// attribute value treated as sixteen zero octets during the HMAC-MD5).
func itVerifyRequestMessageAuth(p *radius.Packet, secret string) (present, valid bool) {
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
		b[i] = 0
	}
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write(b)
	return true, hmac.Equal(mac.Sum(nil), saved)
}

func (fn *itFakeNAS) port(t *testing.T) int {
	t.Helper()
	_, portStr, err := net.SplitHostPort(fn.addr)
	require.NoError(t, err)
	p, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	return p
}

func (fn *itFakeNAS) requestCount() int {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	return len(fn.received)
}

// badAuthCount reports how many received requests lacked a valid
// Message-Authenticator (RFC 5176 §3.4).
func (fn *itFakeNAS) badAuthCount() int {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	return fn.badAuth
}

// seedCoAOnlineSession inserts a NAS (reachable at 127.0.0.1:coaPort) and an
// online session bound to it, returning the session primary key and identity so
// the test can target it via the real Admin API.
//
// The Disconnect/CoA handlers resolve the NAS by `ipaddr = session.NasAddr`, so
// every CoA test shares the 127.0.0.1 loopback address. To keep First() lookups
// deterministic on the suite's shared database, these tests must run serially
// (no t.Parallel): the helper removes any stale 127.0.0.1 NAS before seeding and
// deletes its own rows on cleanup, guaranteeing exactly one match at call time.
func seedCoAOnlineSession(t *testing.T, secret string, coaPort int) (id int64, username, acctSessionID string) {
	t.Helper()
	require.NoError(t, h.appCtx.DB().Where("ipaddr = ?", "127.0.0.1").Delete(&domain.NetNas{}).Error)

	suffix := uniqueSuffix()
	username = "it-coa-" + suffix
	acctSessionID = "it-coa-sess-" + suffix

	nas := &domain.NetNas{
		ID:         common.UUIDint64(),
		Name:       "it-coa-nas-" + suffix,
		Identifier: "it-coa-nas-id-" + suffix,
		Ipaddr:     "127.0.0.1",
		Secret:     secret,
		CoaPort:    coaPort,
		Status:     common.ENABLED,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(nas).Error)

	session := &domain.RadiusOnline{
		Username:      username,
		NasId:         nas.Identifier,
		NasAddr:       "127.0.0.1",
		FramedIpaddr:  "100.64.0.10",
		MacAddr:       "AA:BB:CC:DD:EE:01",
		NasPort:       7,
		NasPortId:     "gi0/1",
		AcctSessionId: acctSessionID,
		AcctStartTime: time.Now().Add(-time.Minute),
		LastUpdate:    time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(session).Error)

	t.Cleanup(func() {
		_ = h.appCtx.DB().Delete(&domain.NetNas{}, nas.ID).Error
		_ = h.appCtx.DB().Delete(&domain.RadiusOnline{}, session.ID).Error
	})
	return session.ID, username, acctSessionID
}

func latestAuditForSession(t *testing.T, sessionID int64) domain.RadiusSessionActionAudit {
	t.Helper()
	var audit domain.RadiusSessionActionAudit
	require.NoError(t, h.appCtx.DB().
		Where("session_id = ?", sessionID).
		Order("id desc").
		First(&audit).Error)
	return audit
}

// TestSessionDisconnectEndToEnd_ACK drives the full Disconnect path as a client:
// authenticated HTTP POST -> Admin API handler -> CoAService -> real RFC 5176
// Disconnect-Request over UDP -> NAS ACK -> structured response, and proves the
// audit record lands in PostgreSQL.
func TestSessionDisconnectEndToEnd_ACK(t *testing.T) {
	const secret = "it-coa-secret-ack"
	nas := newITFakeNAS(t, secret, radius.CodeDisconnectACK, 0)
	id, username, acctSessionID := seedCoAOnlineSession(t, secret, nas.port(t))

	c := newAPIClient(t)
	status, body := c.post(t, "/api/v1/sessions/"+strconv.FormatInt(id, 10)+"/disconnect", nil)
	require.Equalf(t, 200, status, "disconnect body: %s", string(body))

	var action coaActionBody
	unwrapData(t, body, &action)
	assert.True(t, action.Success, "expected Disconnect-ACK success")
	assert.Equal(t, "disconnect", action.Action)
	assert.Equal(t, "Disconnect-ACK", action.ResponseCode)
	assert.Equal(t, username, action.Username)
	assert.Equal(t, acctSessionID, action.AcctSessionID)
	assert.False(t, action.TimedOut)
	assert.GreaterOrEqual(t, action.Attempts, 1)
	assert.Equal(t, 1, nas.requestCount(), "NAS should have received exactly one request")
	assert.Equal(t, 0, nas.badAuthCount(), "Disconnect-Request must carry a valid Message-Authenticator (RFC 5176 §3.4)")

	audit := latestAuditForSession(t, id)
	assert.Equal(t, "disconnect", audit.Action)
	assert.True(t, audit.Success)
	assert.Equal(t, "Disconnect-ACK", audit.ResponseCode)
	assert.Equal(t, username, audit.Username)
	assert.Equal(t, h.adminUser, audit.OperatorName)
	assert.False(t, audit.TimedOut)
}

// TestSessionCoAEndToEnd_NAK drives the CoA path to a NAS rejection: the NAS
// replies CoA-NAK with an Error-Cause. The exchange completed, so the API
// returns HTTP 200 with success=false and surfaces the Error-Cause, and the
// failure is recorded in the PostgreSQL audit trail.
func TestSessionCoAEndToEnd_NAK(t *testing.T) {
	const secret = "it-coa-secret-nak"
	nas := newITFakeNAS(t, secret, radius.CodeCoANAK, rfc3576.ErrorCause_Value_UnsupportedAttribute)
	id, _, _ := seedCoAOnlineSession(t, secret, nas.port(t))

	c := newAPIClient(t)
	reqBody, _ := json.Marshal(map[string]interface{}{"session_timeout": 7200, "filter_id": "it-throttle"})
	status, body := c.post(t, "/api/v1/sessions/"+strconv.FormatInt(id, 10)+"/coa", reqBody)
	require.Equalf(t, 200, status, "coa body: %s", string(body))

	var action coaActionBody
	unwrapData(t, body, &action)
	assert.False(t, action.Success, "CoA-NAK must report success=false")
	assert.Equal(t, "coa", action.Action)
	assert.Equal(t, "CoA-NAK", action.ResponseCode)
	assert.Equal(t, int(rfc3576.ErrorCause_Value_UnsupportedAttribute), action.ErrorCause)
	assert.Equal(t, "Unsupported-Attribute", action.ErrorCauseText)
	assert.False(t, action.TimedOut)

	audit := latestAuditForSession(t, id)
	assert.Equal(t, "coa", audit.Action)
	assert.False(t, audit.Success)
	assert.Equal(t, "CoA-NAK", audit.ResponseCode)
	assert.Equal(t, int(rfc3576.ErrorCause_Value_UnsupportedAttribute), audit.ErrorCause)
	assert.Equal(t, h.adminUser, audit.OperatorName)
}
