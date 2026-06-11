package adminapi

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"gorm.io/gorm"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc3576"
)

const coaTestSecret = "coa-secret-123"

// fakeCoAReq holds the identity/change attributes the fake NAS extracted from a
// received CoA/Disconnect request, copied under lock so the test goroutine never
// reads a packet owned by the server goroutine.
type fakeCoAReq struct {
	code           radius.Code
	username       string
	acctSessionID  string
	filterID       string
	sessionTimeout uint32
	hasTimeout     bool
}

// fakeCoANAS is an in-process UDP RADIUS responder emulating a NAS receiving the
// Admin API's CoA/Disconnect requests, so the endpoints are exercised over a real
// socket (the same approach as the radiusd CoAService unit tests).
type fakeCoANAS struct {
	addr   string
	server *radius.PacketServer

	mu        sync.Mutex
	replyCode radius.Code // 0 => drop the request (force a client timeout)
	errCause  rfc3576.ErrorCause
	received  []fakeCoAReq
}

func newFakeCoANAS(t *testing.T, secret string, replyCode radius.Code, cause rfc3576.ErrorCause) *fakeCoANAS {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	fn := &fakeCoANAS{addr: pc.LocalAddr().String(), replyCode: replyCode, errCause: cause}
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

func (fn *fakeCoANAS) handle(w radius.ResponseWriter, r *radius.Request) {
	rec := fakeCoAReq{
		code:          r.Code,
		username:      rfc2865.UserName_GetString(r.Packet),
		acctSessionID: rfc2866.AcctSessionID_GetString(r.Packet),
		filterID:      rfc2865.FilterID_GetString(r.Packet),
	}
	if st, err := rfc2865.SessionTimeout_Lookup(r.Packet); err == nil {
		rec.hasTimeout = true
		rec.sessionTimeout = uint32(st)
	}

	fn.mu.Lock()
	fn.received = append(fn.received, rec)
	code := fn.replyCode
	cause := fn.errCause
	fn.mu.Unlock()

	if code == 0 {
		return // no response => the client times out
	}
	resp := r.Response(code)
	if cause != 0 && (code == radius.CodeDisconnectNAK || code == radius.CodeCoANAK) {
		_ = rfc3576.ErrorCause_Add(resp, cause)
	}
	_ = w.Write(resp)
}

func (fn *fakeCoANAS) port(t *testing.T) int {
	t.Helper()
	_, portStr, err := net.SplitHostPort(fn.addr)
	require.NoError(t, err)
	p, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	return p
}

func (fn *fakeCoANAS) snapshot() []fakeCoAReq {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	out := make([]fakeCoAReq, len(fn.received))
	copy(out, fn.received)
	return out
}

// seedCoASession inserts a NAS (reachable at 127.0.0.1:coaPort) and an online
// session bound to it, returning the session's primary key.
func seedCoASession(t *testing.T, db *gorm.DB, coaPort int) int64 {
	t.Helper()
	nas := &domain.NetNas{
		Name:       "fake-nas",
		Identifier: "fake-nas-id",
		Ipaddr:     "127.0.0.1",
		Secret:     coaTestSecret,
		CoaPort:    coaPort,
		Status:     "enabled",
	}
	require.NoError(t, db.Create(nas).Error)

	session := &domain.RadiusOnline{
		Username:      "alice",
		NasId:         "fake-nas-id",
		NasAddr:       "127.0.0.1",
		FramedIpaddr:  "100.64.0.10",
		MacAddr:       "AA:BB:CC:DD:EE:01",
		NasPort:       7,
		NasPortId:     "gi0/1",
		AcctSessionId: "coa-sess-" + strconv.FormatInt(testSessionSeq.Add(1), 10),
		AcctStartTime: time.Now().Add(-time.Minute),
		LastUpdate:    time.Now(),
	}
	require.NoError(t, db.Create(session).Error)
	return session.ID
}

// newSessionActionCtx builds an echo context for a session action request with
// the :id path parameter set and (by default) a super-admin operator injected.
func newSessionActionCtx(db *gorm.DB, appCtx app.AppContext, body, idParam string) (echo.Context, *httptest.ResponseRecorder) {
	e := setupTestEcho()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+idParam+"/action", reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)
	c.SetParamNames("id")
	c.SetParamValues(idParam)
	return c, rec
}

func decodeCoAAction(t *testing.T, rec *httptest.ResponseRecorder) coaActionResponse {
	t.Helper()
	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	raw, err := json.Marshal(resp.Data)
	require.NoError(t, err)
	var action coaActionResponse
	require.NoError(t, json.Unmarshal(raw, &action))
	return action
}

func decodeErr(t *testing.T, rec *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()
	var errResp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
	return errResp
}

func listSessionActionAudits(t *testing.T, db *gorm.DB) []domain.RadiusSessionActionAudit {
	t.Helper()
	var audits []domain.RadiusSessionActionAudit
	require.NoError(t, db.Order("id asc").Find(&audits).Error)
	return audits
}

func TestDisconnectOnlineSession_ACK(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, radius.CodeDisconnectACK, 0)
	id := seedCoASession(t, db, nas.port(t))

	c, rec := newSessionActionCtx(db, appCtx, "", strconv.FormatInt(id, 10))
	require.NoError(t, DisconnectOnlineSession(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	action := decodeCoAAction(t, rec)
	assert.True(t, action.Success)
	assert.Equal(t, "disconnect", action.Action)
	assert.Equal(t, "Disconnect-ACK", action.ResponseCode)
	assert.Equal(t, 1, action.Attempts)
	assert.False(t, action.TimedOut)
	assert.Equal(t, "alice", action.Username)

	got := nas.snapshot()
	require.Len(t, got, 1)
	assert.Equal(t, radius.CodeDisconnectRequest, got[0].code)
	assert.Equal(t, "alice", got[0].username)

	audits := listSessionActionAudits(t, db)
	require.Len(t, audits, 1)
	assert.Equal(t, id, audits[0].SessionID)
	assert.Equal(t, "disconnect", audits[0].Action)
	assert.Equal(t, "alice", audits[0].Username)
	assert.Equal(t, "superadmin", audits[0].OperatorName)
	assert.True(t, audits[0].Success)
	assert.Equal(t, "Disconnect-ACK", audits[0].ResponseCode)
}

func TestDisconnectOnlineSession_NAK(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, radius.CodeDisconnectNAK, rfc3576.ErrorCause_Value_SessionContextNotFound)
	id := seedCoASession(t, db, nas.port(t))

	c, rec := newSessionActionCtx(db, appCtx, "", strconv.FormatInt(id, 10))
	require.NoError(t, DisconnectOnlineSession(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	action := decodeCoAAction(t, rec)
	assert.False(t, action.Success)
	assert.Equal(t, "Disconnect-NAK", action.ResponseCode)
	assert.Equal(t, int(rfc3576.ErrorCause_Value_SessionContextNotFound), action.ErrorCause)
	assert.Equal(t, "Session-Context-Not-Found", action.ErrorCauseText)

	audits := listSessionActionAudits(t, db)
	require.Len(t, audits, 1)
	assert.Equal(t, "disconnect", audits[0].Action)
	assert.False(t, audits[0].Success)
	assert.Equal(t, "Disconnect-NAK", audits[0].ResponseCode)
	assert.Equal(t, int(rfc3576.ErrorCause_Value_SessionContextNotFound), audits[0].ErrorCause)
}

func TestChangeOnlineSessionAuthorization_ACK(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, radius.CodeCoAACK, 0)
	id := seedCoASession(t, db, nas.port(t))

	c, rec := newSessionActionCtx(db, appCtx, `{"session_timeout":3600,"filter_id":"throttled"}`, strconv.FormatInt(id, 10))
	require.NoError(t, ChangeOnlineSessionAuthorization(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	action := decodeCoAAction(t, rec)
	assert.True(t, action.Success)
	assert.Equal(t, "coa", action.Action)
	assert.Equal(t, "CoA-ACK", action.ResponseCode)

	got := nas.snapshot()
	require.Len(t, got, 1)
	assert.Equal(t, radius.CodeCoARequest, got[0].code)
	assert.True(t, got[0].hasTimeout)
	assert.Equal(t, uint32(3600), got[0].sessionTimeout)
	assert.Equal(t, "throttled", got[0].filterID)

	audits := listSessionActionAudits(t, db)
	require.Len(t, audits, 1)
	assert.Equal(t, "coa", audits[0].Action)
	assert.True(t, audits[0].Success)
	assert.Equal(t, "CoA-ACK", audits[0].ResponseCode)
}

func TestChangeOnlineSessionAuthorization_NoChanges(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	id := seedCoASession(t, db, 3799)

	c, rec := newSessionActionCtx(db, appCtx, `{}`, strconv.FormatInt(id, 10))
	require.NoError(t, ChangeOnlineSessionAuthorization(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "NO_CHANGES", decodeErr(t, rec).Error)
}

func TestChangeOnlineSessionAuthorization_NAK(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, radius.CodeCoANAK, rfc3576.ErrorCause_Value_UnsupportedAttribute)
	id := seedCoASession(t, db, nas.port(t))

	c, rec := newSessionActionCtx(db, appCtx, `{"filter_id":"throttled"}`, strconv.FormatInt(id, 10))
	require.NoError(t, ChangeOnlineSessionAuthorization(c))
	// A NAK is a completed exchange: HTTP 200 with success=false, not an error status.
	assert.Equal(t, http.StatusOK, rec.Code)

	action := decodeCoAAction(t, rec)
	assert.False(t, action.Success)
	assert.Equal(t, "coa", action.Action)
	assert.Equal(t, "CoA-NAK", action.ResponseCode)
	assert.Equal(t, int(rfc3576.ErrorCause_Value_UnsupportedAttribute), action.ErrorCause)
	assert.Equal(t, "Unsupported-Attribute", action.ErrorCauseText)
	assert.False(t, action.TimedOut)

	got := nas.snapshot()
	require.Len(t, got, 1)
	assert.Equal(t, radius.CodeCoARequest, got[0].code)

	audits := listSessionActionAudits(t, db)
	require.Len(t, audits, 1)
	assert.Equal(t, "coa", audits[0].Action)
	assert.False(t, audits[0].Success)
	assert.Equal(t, "CoA-NAK", audits[0].ResponseCode)
	assert.Equal(t, int(rfc3576.ErrorCause_Value_UnsupportedAttribute), audits[0].ErrorCause)
}

func TestChangeOnlineSessionAuthorization_Timeout(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, 0, 0) // drop all => timeout
	id := seedCoASession(t, db, nas.port(t))

	restore := sessionCoAService
	sessionCoAService = func() *radiusd.CoAService {
		return radiusd.NewCoAService(nil,
			radiusd.WithCoATimeout(120*time.Millisecond),
			radiusd.WithCoARetries(1))
	}
	t.Cleanup(func() { sessionCoAService = restore })

	c, rec := newSessionActionCtx(db, appCtx, `{"session_timeout":600}`, strconv.FormatInt(id, 10))
	require.NoError(t, ChangeOnlineSessionAuthorization(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	action := decodeCoAAction(t, rec)
	assert.False(t, action.Success)
	assert.Equal(t, "coa", action.Action)
	assert.True(t, action.TimedOut)
	assert.Equal(t, 2, action.Attempts) // initial + 1 retry

	audits := listSessionActionAudits(t, db)
	require.Len(t, audits, 1)
	assert.Equal(t, "coa", audits[0].Action)
	assert.False(t, audits[0].Success)
	assert.True(t, audits[0].TimedOut)
	assert.Equal(t, 2, audits[0].Attempts)
}

func TestDisconnectOnlineSession_Timeout(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, 0, 0) // drop all => timeout
	id := seedCoASession(t, db, nas.port(t))

	restore := sessionCoAService
	sessionCoAService = func() *radiusd.CoAService {
		return radiusd.NewCoAService(nil,
			radiusd.WithCoATimeout(120*time.Millisecond),
			radiusd.WithCoARetries(1))
	}
	t.Cleanup(func() { sessionCoAService = restore })

	c, rec := newSessionActionCtx(db, appCtx, "", strconv.FormatInt(id, 10))
	require.NoError(t, DisconnectOnlineSession(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	action := decodeCoAAction(t, rec)
	assert.False(t, action.Success)
	assert.True(t, action.TimedOut)
	assert.Equal(t, 2, action.Attempts) // initial + 1 retry

	audits := listSessionActionAudits(t, db)
	require.Len(t, audits, 1)
	assert.Equal(t, "disconnect", audits[0].Action)
	assert.False(t, audits[0].Success)
	assert.True(t, audits[0].TimedOut)
	assert.Equal(t, 2, audits[0].Attempts)
}

func TestSessionAction_SessionNotFound(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	c, rec := newSessionActionCtx(db, appCtx, "", "999")
	require.NoError(t, DisconnectOnlineSession(c))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "NOT_FOUND", decodeErr(t, rec).Error)
}

func TestSessionAction_NASNotFound(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Online session whose NasAddr matches no configured NAS.
	session := &domain.RadiusOnline{
		Username:      "bob",
		NasAddr:       "10.99.99.99",
		AcctSessionId: "orphan-" + strconv.FormatInt(testSessionSeq.Add(1), 10),
		LastUpdate:    time.Now(),
	}
	require.NoError(t, db.Create(session).Error)

	c, rec := newSessionActionCtx(db, appCtx, "", strconv.FormatInt(session.ID, 10))
	require.NoError(t, DisconnectOnlineSession(c))
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, "NAS_NOT_FOUND", decodeErr(t, rec).Error)
}

func TestSessionAction_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	c, rec := newSessionActionCtx(db, appCtx, "", "not-a-number")
	require.NoError(t, DisconnectOnlineSession(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "INVALID_ID", decodeErr(t, rec).Error)

	c2, rec2 := newSessionActionCtx(db, appCtx, `{"session_timeout":60}`, "not-a-number")
	require.NoError(t, ChangeOnlineSessionAuthorization(c2))
	assert.Equal(t, http.StatusBadRequest, rec2.Code)
	assert.Equal(t, "INVALID_ID", decodeErr(t, rec2).Error)
}

// TestSessionActionsRequireAdmin verifies the authorization guard on the session
// action routes: operator-level accounts are rejected (403) before the handler
// runs, while admin/super accounts are allowed through to a real exchange.
func TestSessionActionsRequireAdmin(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, radius.CodeDisconnectACK, 0)
	id := seedCoASession(t, db, nas.port(t))
	guard := requireAdmin()

	cases := []struct {
		name    string
		level   string
		allowed bool
		status  int
	}{
		{"operator denied", LevelOperator, false, http.StatusForbidden},
		{"unknown level denied", "guest", false, http.StatusForbidden},
		{"admin allowed", LevelAdmin, true, http.StatusOK},
		{"super allowed", LevelSuper, true, http.StatusOK},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newSessionActionCtx(db, appCtx, "", strconv.FormatInt(id, 10))
			c.Set("current_operator", &domain.SysOpr{ID: 1, Level: tt.level, Status: "enabled"})

			ran := false
			wrapped := guard(func(c echo.Context) error {
				ran = true
				return DisconnectOnlineSession(c)
			})
			require.NoError(t, wrapped(c))

			assert.Equal(t, tt.allowed, ran)
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

// TestDeleteOnlineSessionRequireAdmin verifies that DELETE /sessions/:id
// (force-offline) is gated by requireAdmin, matching the disconnect/coa actions:
// operator-level (and unknown) accounts are rejected with 403 before the handler
// runs, while admin/super accounts are allowed through.
func TestDeleteOnlineSessionRequireAdmin(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	guard := requireAdmin()

	cases := []struct {
		name    string
		level   string
		allowed bool
		status  int
	}{
		{"operator denied", LevelOperator, false, http.StatusForbidden},
		{"unknown level denied", "guest", false, http.StatusForbidden},
		{"admin allowed", LevelAdmin, true, http.StatusOK},
		{"super allowed", LevelSuper, true, http.StatusOK},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// A fresh session per case: the handler deletes it on success.
			id := seedCoASession(t, db, 3799)

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/"+strconv.FormatInt(id, 10), nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(id, 10))
			c.Set("current_operator", &domain.SysOpr{ID: 1, Level: tt.level, Status: "enabled"})

			ran := false
			wrapped := guard(func(c echo.Context) error {
				ran = true
				return DeleteOnlineSession(c)
			})
			require.NoError(t, wrapped(c))

			assert.Equal(t, tt.allowed, ran)
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

// TestDeleteOnlineSessionAudited verifies that the force-offline DELETE route
// drives the RFC 5176 Disconnect through the shared CoAService (sending the full
// identity triplet to the NAS) and writes a durable M2.3 audit record, in
// addition to removing the local online session row.
func TestDeleteOnlineSessionAudited(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	nas := newFakeCoANAS(t, coaTestSecret, radius.CodeDisconnectACK, 0)
	id := seedCoASession(t, db, nas.port(t))

	e := setupTestEcho()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/"+strconv.FormatInt(id, 10), nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatInt(id, 10))

	require.NoError(t, DeleteOnlineSession(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	// The local online record is removed.
	var count int64
	db.Model(&domain.RadiusOnline{}).Where("id = ?", id).Count(&count)
	assert.Equal(t, int64(0), count, "online session should be deleted")

	// The NAS received a Disconnect-Request carrying the full identity triplet.
	got := nas.snapshot()
	require.Len(t, got, 1)
	assert.Equal(t, radius.CodeDisconnectRequest, got[0].code)
	assert.Equal(t, "alice", got[0].username)
	assert.NotEmpty(t, got[0].acctSessionID)

	// A durable audit record is persisted for the force-offline action.
	audits := listSessionActionAudits(t, db)
	require.Len(t, audits, 1)
	assert.Equal(t, id, audits[0].SessionID)
	assert.Equal(t, "disconnect", audits[0].Action)
	assert.Equal(t, "alice", audits[0].Username)
	assert.Equal(t, "superadmin", audits[0].OperatorName)
	assert.True(t, audits[0].Success)
	assert.Equal(t, "Disconnect-ACK", audits[0].ResponseCode)
}
