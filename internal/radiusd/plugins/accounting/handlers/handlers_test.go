package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
	"layeh.com/radius/rfc2866"
)

// mockSessionRepository is a test mock for SessionRepository
type mockSessionRepository struct {
	sessions         map[string]*domain.RadiusOnline
	createErr        error
	updateErr        error
	deleteErr        error
	batchDeleteErr   error
	userSessionCount map[string]int
}

func newMockSessionRepo() *mockSessionRepository {
	return &mockSessionRepository{
		sessions:         make(map[string]*domain.RadiusOnline),
		userSessionCount: make(map[string]int),
	}
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.RadiusOnline) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.sessions[session.AcctSessionId] = session
	return nil
}

func (m *mockSessionRepository) Update(ctx context.Context, session *domain.RadiusOnline) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.sessions[session.AcctSessionId]; ok {
		existing.AcctSessionTime = session.AcctSessionTime
		existing.AcctInputTotal = session.AcctInputTotal
		existing.AcctOutputTotal = session.AcctOutputTotal
		existing.AcctInputPackets = session.AcctInputPackets
		existing.AcctOutputPackets = session.AcctOutputPackets
		existing.LastUpdate = session.LastUpdate
	}
	return nil
}

func (m *mockSessionRepository) Delete(ctx context.Context, sessionId string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.sessions, sessionId)
	return nil
}

func (m *mockSessionRepository) GetBySessionId(ctx context.Context, sessionId string) (*domain.RadiusOnline, error) {
	if session, ok := m.sessions[sessionId]; ok {
		return session, nil
	}
	return nil, errors.New("session not found")
}

func (m *mockSessionRepository) CountByUsername(ctx context.Context, username string) (int, error) {
	if count, ok := m.userSessionCount[username]; ok {
		return count, nil
	}
	return 0, nil
}

func (m *mockSessionRepository) Exists(ctx context.Context, sessionId string) (bool, error) {
	_, ok := m.sessions[sessionId]
	return ok, nil
}

func (m *mockSessionRepository) BatchDelete(ctx context.Context, ids []string) error {
	if m.batchDeleteErr != nil {
		return m.batchDeleteErr
	}
	for _, id := range ids {
		delete(m.sessions, id)
	}
	return nil
}

func (m *mockSessionRepository) BatchDeleteByNas(ctx context.Context, nasAddr, nasId string) error {
	if m.batchDeleteErr != nil {
		return m.batchDeleteErr
	}
	// Delete all sessions matching the NAS
	for id, session := range m.sessions {
		if session.NasAddr == nasAddr || session.NasId == nasId {
			delete(m.sessions, id)
		}
	}
	return nil
}

// mockAccountingRepository is a test mock for AccountingRepository
type mockAccountingRepository struct {
	records       map[string]*domain.RadiusAccounting
	createErr     error
	updateStopErr error
}

func newMockAccountingRepo() *mockAccountingRepository {
	return &mockAccountingRepository{
		records: make(map[string]*domain.RadiusAccounting),
	}
}

func (m *mockAccountingRepository) Create(ctx context.Context, acct *domain.RadiusAccounting) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.records[acct.AcctSessionId] = acct
	return nil
}

func (m *mockAccountingRepository) UpdateStop(ctx context.Context, sessionId string, acct *domain.RadiusAccounting) error {
	if m.updateStopErr != nil {
		return m.updateStopErr
	}
	if existing, ok := m.records[sessionId]; ok {
		existing.AcctStopTime = time.Now()
		existing.AcctInputTotal = acct.AcctInputTotal
		existing.AcctOutputTotal = acct.AcctOutputTotal
		existing.AcctSessionTime = acct.AcctSessionTime
	}
	return nil
}

// Helper to create a mock accounting context
func createMockAccountingContext(statusType int) *accounting.AccountingContext {
	// Create a minimal RADIUS request with required attributes
	packet := radius.New(radius.CodeAccountingRequest, []byte("secret123456"))
	_ = rfc2866.AcctStatusType_Set(packet, rfc2866.AcctStatusType(statusType)) //nolint:errcheck
	_ = rfc2866.AcctSessionID_SetString(packet, "test-session-123")            //nolint:errcheck

	return &accounting.AccountingContext{
		Context:    context.Background(),
		Request:    &radius.Request{Packet: packet},
		VendorReq:  &vendorparserspkg.VendorRequest{},
		Username:   "testuser",
		NAS:        &domain.NetNas{ID: 1, Ipaddr: "192.168.1.1", Identifier: "nas-01"},
		NASIP:      "192.168.1.1",
		StatusType: statusType,
	}
}

// ============ StartHandler Tests ============

func TestNewStartHandler(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()

	handler := NewStartHandler(sessionRepo, acctRepo)
	assert.NotNil(t, handler)
}

func TestStartHandler_Name(t *testing.T) {
	handler := NewStartHandler(nil, nil)
	assert.Equal(t, "StartHandler", handler.Name())
}

func TestStartHandler_CanHandle(t *testing.T) {
	handler := NewStartHandler(nil, nil)

	tests := []struct {
		name       string
		statusType int
		expected   bool
	}{
		{"Start", int(rfc2866.AcctStatusType_Value_Start), true},
		{"Stop", int(rfc2866.AcctStatusType_Value_Stop), false},
		{"InterimUpdate", int(rfc2866.AcctStatusType_Value_InterimUpdate), false},
		{"AccountingOn", int(rfc2866.AcctStatusType_Value_AccountingOn), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createMockAccountingContext(tt.statusType)
			assert.Equal(t, tt.expected, handler.CanHandle(ctx))
		})
	}
}

func TestStartHandler_Handle_Success(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()
	handler := NewStartHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Start))
	err := handler.Handle(ctx)

	assert.NoError(t, err)
	assert.Len(t, sessionRepo.sessions, 1)
	assert.Len(t, acctRepo.records, 1)
}

func TestStartHandler_Handle_SessionCreateError(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	sessionRepo.createErr = errors.New("database error")
	acctRepo := newMockAccountingRepo()
	handler := NewStartHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Start))
	err := handler.Handle(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestStartHandler_Handle_AccountingCreateError(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()
	acctRepo.createErr = errors.New("accounting error")
	handler := NewStartHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Start))
	err := handler.Handle(ctx)

	assert.Error(t, err)
	// Session was created, but accounting failed
	assert.Len(t, sessionRepo.sessions, 1)
}

func TestStartHandler_Handle_NilVendorRequest(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()
	handler := NewStartHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Start))
	ctx.VendorReq = nil // Explicitly nil

	err := handler.Handle(ctx)
	assert.NoError(t, err)
}

// ============ StopHandler Tests ============

func TestNewStopHandler(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()

	handler := NewStopHandler(sessionRepo, acctRepo)
	assert.NotNil(t, handler)
}

func TestStopHandler_Name(t *testing.T) {
	handler := NewStopHandler(nil, nil)
	assert.Equal(t, "StopHandler", handler.Name())
}

func TestStopHandler_CanHandle(t *testing.T) {
	handler := NewStopHandler(nil, nil)

	tests := []struct {
		name       string
		statusType int
		expected   bool
	}{
		{"Stop", int(rfc2866.AcctStatusType_Value_Stop), true},
		{"Start", int(rfc2866.AcctStatusType_Value_Start), false},
		{"InterimUpdate", int(rfc2866.AcctStatusType_Value_InterimUpdate), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createMockAccountingContext(tt.statusType)
			assert.Equal(t, tt.expected, handler.CanHandle(ctx))
		})
	}
}

func TestStopHandler_Handle_Success(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	sessionRepo.sessions["test-session-123"] = &domain.RadiusOnline{
		AcctSessionId: "test-session-123",
		Username:      "testuser",
	}
	acctRepo := newMockAccountingRepo()
	acctRepo.records["test-session-123"] = &domain.RadiusAccounting{
		AcctSessionId: "test-session-123",
	}
	handler := NewStopHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Stop))
	err := handler.Handle(ctx)

	assert.NoError(t, err)
	assert.Empty(t, sessionRepo.sessions) // Session should be deleted
}

func TestStopHandler_Handle_DeleteError(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	sessionRepo.deleteErr = errors.New("delete failed")
	acctRepo := newMockAccountingRepo()
	handler := NewStopHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Stop))
	err := handler.Handle(ctx)

	assert.Error(t, err)
}

func TestStopHandler_Handle_NilVendorRequest(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()
	handler := NewStopHandler(sessionRepo, acctRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_Stop))
	ctx.VendorReq = nil

	err := handler.Handle(ctx)
	assert.NoError(t, err)
}

// ============ UpdateHandler Tests ============

func TestNewUpdateHandler(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	handler := NewUpdateHandler(sessionRepo)
	assert.NotNil(t, handler)
}

func TestUpdateHandler_Name(t *testing.T) {
	handler := NewUpdateHandler(nil)
	assert.Equal(t, "UpdateHandler", handler.Name())
}

func TestUpdateHandler_CanHandle(t *testing.T) {
	handler := NewUpdateHandler(nil)

	tests := []struct {
		name       string
		statusType int
		expected   bool
	}{
		{"InterimUpdate", int(rfc2866.AcctStatusType_Value_InterimUpdate), true},
		{"Start", int(rfc2866.AcctStatusType_Value_Start), false},
		{"Stop", int(rfc2866.AcctStatusType_Value_Stop), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createMockAccountingContext(tt.statusType)
			assert.Equal(t, tt.expected, handler.CanHandle(ctx))
		})
	}
}

func TestUpdateHandler_Handle_Success(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	sessionRepo.sessions["test-session-123"] = &domain.RadiusOnline{
		AcctSessionId: "test-session-123",
		Username:      "testuser",
	}
	handler := NewUpdateHandler(sessionRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_InterimUpdate))
	err := handler.Handle(ctx)

	assert.NoError(t, err)
}

func TestUpdateHandler_Handle_UpdateError(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	sessionRepo.updateErr = errors.New("update failed")
	handler := NewUpdateHandler(sessionRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_InterimUpdate))
	err := handler.Handle(ctx)

	assert.Error(t, err)
}

func TestUpdateHandler_Handle_NilVendorRequest(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	handler := NewUpdateHandler(sessionRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_InterimUpdate))
	ctx.VendorReq = nil

	err := handler.Handle(ctx)
	assert.NoError(t, err)
}

// ============ NasStateHandler Tests ============

func TestNewNasStateHandler(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	handler := NewNasStateHandler(sessionRepo)
	assert.NotNil(t, handler)
}

func TestNasStateHandler_Name(t *testing.T) {
	handler := NewNasStateHandler(nil)
	assert.Equal(t, "NasStateHandler", handler.Name())
}

func TestNasStateHandler_CanHandle(t *testing.T) {
	handler := NewNasStateHandler(nil)

	tests := []struct {
		name       string
		statusType int
		expected   bool
	}{
		{"AccountingOn", int(rfc2866.AcctStatusType_Value_AccountingOn), true},
		{"AccountingOff", int(rfc2866.AcctStatusType_Value_AccountingOff), true},
		{"Start", int(rfc2866.AcctStatusType_Value_Start), false},
		{"Stop", int(rfc2866.AcctStatusType_Value_Stop), false},
		{"InterimUpdate", int(rfc2866.AcctStatusType_Value_InterimUpdate), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createMockAccountingContext(tt.statusType)
			assert.Equal(t, tt.expected, handler.CanHandle(ctx))
		})
	}
}

func TestNasStateHandler_Handle_Success(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	// Add some sessions for the NAS
	sessionRepo.sessions["session-1"] = &domain.RadiusOnline{
		AcctSessionId: "session-1",
		NasAddr:       "192.168.1.1",
		NasId:         "nas-01",
	}
	sessionRepo.sessions["session-2"] = &domain.RadiusOnline{
		AcctSessionId: "session-2",
		NasAddr:       "192.168.1.1",
		NasId:         "nas-01",
	}
	handler := NewNasStateHandler(sessionRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_AccountingOn))
	err := handler.Handle(ctx)

	assert.NoError(t, err)
	assert.Empty(t, sessionRepo.sessions) // All sessions should be cleared
}

func TestNasStateHandler_Handle_NilSessionRepo(t *testing.T) {
	handler := NewNasStateHandler(nil)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_AccountingOn))
	err := handler.Handle(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session repository is not available")
}

func TestNasStateHandler_Handle_NilNAS(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	handler := NewNasStateHandler(sessionRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_AccountingOn))
	ctx.NAS = nil

	err := handler.Handle(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nas information is missing")
}

func TestNasStateHandler_Handle_BatchDeleteError(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	sessionRepo.batchDeleteErr = errors.New("batch delete failed")
	handler := NewNasStateHandler(sessionRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_AccountingOff))
	err := handler.Handle(ctx)

	assert.Error(t, err)
}

// ============ Integration-like Tests ============

func TestHandlers_ImplementAccountingHandler(t *testing.T) {
	sessionRepo := newMockSessionRepo()
	acctRepo := newMockAccountingRepo()

	// Verify all handlers implement AccountingHandler interface
	handlers := []accounting.AccountingHandler{
		NewStartHandler(sessionRepo, acctRepo),
		NewStopHandler(sessionRepo, acctRepo),
		NewUpdateHandler(sessionRepo),
		NewNasStateHandler(sessionRepo),
	}

	require.Len(t, handlers, 4)
	for _, h := range handlers {
		assert.NotEmpty(t, h.Name())
	}
}

func TestBuildOnlineFromRequest(t *testing.T) {
	// Test the helper function indirectly through UpdateHandler
	sessionRepo := newMockSessionRepo()
	sessionRepo.sessions["test-session-123"] = &domain.RadiusOnline{
		AcctSessionId: "test-session-123",
	}
	handler := NewUpdateHandler(sessionRepo)

	ctx := createMockAccountingContext(int(rfc2866.AcctStatusType_Value_InterimUpdate))
	err := handler.Handle(ctx)

	assert.NoError(t, err)
	// Verify session was updated
	session := sessionRepo.sessions["test-session-123"]
	assert.NotNil(t, session)
}
