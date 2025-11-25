package eap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

// Mock implementations for testing

// mockResponseWriter captures RADIUS responses for testing
type mockResponseWriter struct {
	response *radius.Packet
	err      error
}

func (m *mockResponseWriter) Write(p *radius.Packet) error {
	m.response = p
	return m.err
}

// mockStateManager provides in-memory state management for testing
type mockStateManager struct {
	states map[string]*EAPState
	setErr error
	getErr error
	delErr error
}

func newMockStateManager() *mockStateManager {
	return &mockStateManager{
		states: make(map[string]*EAPState),
	}
}

func (m *mockStateManager) GetState(stateID string) (*EAPState, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	state, ok := m.states[stateID]
	if !ok {
		return nil, ErrStateNotFound
	}
	return state, nil
}

func (m *mockStateManager) SetState(stateID string, state *EAPState) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.states[stateID] = state
	return nil
}

func (m *mockStateManager) DeleteState(stateID string) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.states, stateID)
	return nil
}

// mockPasswordProvider provides test passwords
type mockPasswordProvider struct {
	password string
	err      error
}

func (m *mockPasswordProvider) GetPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.password != "" {
		return m.password, nil
	}
	return user.Password, nil
}

// mockEAPHandler simulates an EAP handler for testing
type mockEAPHandler struct {
	name              string
	eapType           uint8
	canHandle         bool
	handleIdentityOk  bool
	handleIdentityErr error
	handleResponseOk  bool
	handleResponseErr error
}

func (m *mockEAPHandler) Name() string {
	return m.name
}

func (m *mockEAPHandler) EAPType() uint8 {
	return m.eapType
}

func (m *mockEAPHandler) CanHandle(ctx *EAPContext) bool {
	return m.canHandle
}

func (m *mockEAPHandler) HandleIdentity(ctx *EAPContext) (bool, error) {
	if m.handleIdentityErr != nil {
		return false, m.handleIdentityErr
	}
	// Simulate sending a challenge response
	if ctx.ResponseWriter != nil {
		response := ctx.Request.Response(radius.CodeAccessChallenge)
		eapChallenge := []byte{CodeRequest, ctx.EAPMessage.Identifier, 0, 10, m.eapType, 1, 2, 3, 4, 5}
		SetEAPMessageAndAuth(response, eapChallenge, ctx.Secret)
		_ = rfc2865.State_SetString(response, "test-state-id") //nolint:errcheck
		_ = ctx.ResponseWriter.Write(response)                 //nolint:errcheck
	}
	return m.handleIdentityOk, nil
}

func (m *mockEAPHandler) HandleResponse(ctx *EAPContext) (bool, error) {
	return m.handleResponseOk, m.handleResponseErr
}

// mockHandlerRegistry provides mock handler lookup
type mockHandlerRegistry struct {
	handlers map[uint8]EAPHandler
}

func newMockHandlerRegistry() *mockHandlerRegistry {
	return &mockHandlerRegistry{
		handlers: make(map[uint8]EAPHandler),
	}
}

func (m *mockHandlerRegistry) GetHandler(eapType uint8) (EAPHandler, bool) {
	h, ok := m.handlers[eapType]
	return h, ok
}

func (m *mockHandlerRegistry) Register(handler EAPHandler) {
	m.handlers[handler.EAPType()] = handler
}

// Helper functions for creating test packets

func createEAPIdentityResponse(identifier uint8, identity string) *radius.Packet {
	p := radius.New(radius.CodeAccessRequest, []byte("secret"))
	_ = rfc2865.UserName_SetString(p, identity) //nolint:errcheck

	// EAP-Response/Identity: Code=2, Type=1, Data=identity
	identityBytes := []byte(identity)
	eapLen := 5 + len(identityBytes)
	eapMsg := make([]byte, eapLen)
	eapMsg[0] = CodeResponse
	eapMsg[1] = identifier
	eapMsg[2] = byte(eapLen >> 8)
	eapMsg[3] = byte(eapLen)
	eapMsg[4] = TypeIdentity
	copy(eapMsg[5:], identityBytes)
	_ = rfc2869.EAPMessage_Set(p, eapMsg) //nolint:errcheck

	return p
}

func createEAPNakResponse(identifier uint8, suggestedTypes ...uint8) *radius.Packet {
	p := radius.New(radius.CodeAccessRequest, []byte("secret"))

	// EAP-Response/Nak: Code=2, Type=3, Data=suggested types
	eapLen := 5 + len(suggestedTypes)
	eapMsg := make([]byte, eapLen)
	eapMsg[0] = CodeResponse
	eapMsg[1] = identifier
	eapMsg[2] = byte(eapLen >> 8)
	eapMsg[3] = byte(eapLen)
	eapMsg[4] = TypeNak
	copy(eapMsg[5:], suggestedTypes)
	_ = rfc2869.EAPMessage_Set(p, eapMsg) //nolint:errcheck

	return p
}

func createEAPChallengeResponse(identifier uint8, eapType uint8, data []byte) *radius.Packet {
	p := radius.New(radius.CodeAccessRequest, []byte("secret"))

	// EAP-Response with specified type
	eapLen := 5 + len(data)
	eapMsg := make([]byte, eapLen)
	eapMsg[0] = CodeResponse
	eapMsg[1] = identifier
	eapMsg[2] = byte(eapLen >> 8)
	eapMsg[3] = byte(eapLen)
	eapMsg[4] = eapType
	copy(eapMsg[5:], data)
	_ = rfc2869.EAPMessage_Set(p, eapMsg) //nolint:errcheck

	return p
}

func createEAPRequestPacket(identifier uint8) *radius.Packet {
	p := radius.New(radius.CodeAccessRequest, []byte("secret"))

	// EAP-Request (Code=1)
	eapMsg := []byte{CodeRequest, identifier, 0, 5, TypeIdentity}
	_ = rfc2869.EAPMessage_Set(p, eapMsg) //nolint:errcheck

	return p
}

// Tests for NewCoordinator

func TestNewCoordinator(t *testing.T) {
	stateManager := newMockStateManager()
	pwdProvider := &mockPasswordProvider{password: "test"}
	registry := newMockHandlerRegistry()

	coordinator := NewCoordinator(stateManager, pwdProvider, registry, true)

	assert.NotNil(t, coordinator)
	assert.Equal(t, stateManager, coordinator.stateManager)
	assert.Equal(t, pwdProvider, coordinator.pwdProvider)
	assert.Equal(t, registry, coordinator.handlerRegistry)
	assert.True(t, coordinator.debug)
}

func TestNewCoordinator_DebugFalse(t *testing.T) {
	coordinator := NewCoordinator(nil, nil, nil, false)

	assert.NotNil(t, coordinator)
	assert.False(t, coordinator.debug)
}

// Tests for HandleEAPRequest

func TestHandleEAPRequest_NoEAPMessage(t *testing.T) {
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), false)
	writer := &mockResponseWriter{}

	// Packet without EAP-Message attribute
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.NoError(t, err)
}

func TestHandleEAPRequest_IdentityResponse_MD5(t *testing.T) {
	stateManager := newMockStateManager()
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:             "eap-md5",
		eapType:          TypeMD5Challenge,
		canHandle:        true,
		handleIdentityOk: true,
	})

	coordinator := NewCoordinator(stateManager, &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(1, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{Username: "testuser"}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.False(t, success) // Identity phase doesn't return success
	assert.NoError(t, err)
	assert.NotNil(t, writer.response)
	assert.Equal(t, radius.CodeAccessChallenge, writer.response.Code)
}

func TestHandleEAPRequest_IdentityResponse_MSCHAPv2(t *testing.T) {
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:             "eap-mschapv2",
		eapType:          TypeMSCHAPv2,
		canHandle:        true,
		handleIdentityOk: true,
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(1, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-mschapv2",
	)

	assert.True(t, handled)
	assert.False(t, success)
	assert.NoError(t, err)
}

func TestHandleEAPRequest_IdentityResponse_OTP(t *testing.T) {
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:             "eap-otp",
		eapType:          TypeOTP,
		canHandle:        true,
		handleIdentityOk: true,
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(1, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-otp",
	)

	assert.True(t, handled)
	assert.False(t, success)
	assert.NoError(t, err)
}

func TestHandleEAPRequest_IdentityResponse_DefaultToMD5(t *testing.T) {
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:             "eap-md5",
		eapType:          TypeMD5Challenge,
		canHandle:        true,
		handleIdentityOk: true,
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(1, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	// Use unknown method, should default to MD5
	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "unknown-method",
	)

	assert.True(t, handled)
	assert.False(t, success)
	assert.NoError(t, err)
}

func TestHandleEAPRequest_IdentityResponse_HandlerNotFound(t *testing.T) {
	// Empty registry - no handlers registered
	registry := newMockHandlerRegistry()

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(1, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler not found")
}

func TestHandleEAPRequest_IdentityResponse_HandlerError(t *testing.T) {
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:              "eap-md5",
		eapType:           TypeMD5Challenge,
		canHandle:         true,
		handleIdentityErr: errors.New("handler error"),
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(1, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
}

func TestHandleEAPRequest_NakResponse(t *testing.T) {
	registry := newMockHandlerRegistry()
	// Client suggests MSCHAPv2 instead
	registry.Register(&mockEAPHandler{
		name:             "eap-mschapv2",
		eapType:          TypeMSCHAPv2,
		canHandle:        true,
		handleIdentityOk: true,
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPNakResponse(1, TypeMSCHAPv2)
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.False(t, success) // Nak response triggers new challenge
	assert.NoError(t, err)
}

func TestHandleEAPRequest_NakResponse_NoAlternative(t *testing.T) {
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), false)
	writer := &mockResponseWriter{}

	// Nak with empty data (no suggested methods)
	packet := createEAPNakResponse(1)
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no alternative methods")
}

func TestHandleEAPRequest_NakResponse_UnsupportedType(t *testing.T) {
	// Registry has no handler for the suggested type
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), false)
	writer := &mockResponseWriter{}

	packet := createEAPNakResponse(1, TypeTLS) // Suggest TLS, but no handler
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported EAP type")
}

func TestHandleEAPRequest_ChallengeResponse_Success(t *testing.T) {
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:             "eap-md5",
		eapType:          TypeMD5Challenge,
		canHandle:        true,
		handleResponseOk: true,
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{password: "testpass"}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPChallengeResponse(1, TypeMD5Challenge, []byte{16, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{Password: "testpass"}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.True(t, success)
	assert.NoError(t, err)
}

func TestHandleEAPRequest_ChallengeResponse_Failure(t *testing.T) {
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:              "eap-md5",
		eapType:           TypeMD5Challenge,
		canHandle:         true,
		handleResponseOk:  false,
		handleResponseErr: ErrPasswordMismatch,
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPChallengeResponse(1, TypeMD5Challenge, []byte{16, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Equal(t, ErrPasswordMismatch, err)
}

func TestHandleEAPRequest_ChallengeResponse_UnsupportedType(t *testing.T) {
	// No handler for the EAP type
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), false)
	writer := &mockResponseWriter{}

	packet := createEAPChallengeResponse(1, TypeMD5Challenge, []byte{16, 1, 2, 3})
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported EAP type")
}

func TestHandleEAPRequest_ChallengeResponse_HandlerCannotHandle(t *testing.T) {
	registry := newMockHandlerRegistry()
	registry.Register(&mockEAPHandler{
		name:      "eap-md5",
		eapType:   TypeMD5Challenge,
		canHandle: false, // Handler says it cannot handle
	})

	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, registry, false)
	writer := &mockResponseWriter{}

	packet := createEAPChallengeResponse(1, TypeMD5Challenge, []byte{16, 1, 2, 3})
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot handle")
}

func TestHandleEAPRequest_UnsupportedEAPCode(t *testing.T) {
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), false)
	writer := &mockResponseWriter{}

	// Create packet with EAP-Request code (not Response)
	packet := createEAPRequestPacket(1)
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer, req, &domain.RadiusUser{}, &domain.NetNas{}, response, "secret", false, "eap-md5",
	)

	assert.False(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported EAP code")
}

// Tests for SendEAPSuccess

func TestSendEAPSuccess(t *testing.T) {
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), true)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(5, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	err := coordinator.SendEAPSuccess(writer, req, response, "secret")

	require.NoError(t, err)
	assert.NotNil(t, writer.response)

	// Verify EAP-Success message
	eapMsg, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	assert.Equal(t, uint8(CodeSuccess), eapMsg[0])
	assert.Equal(t, uint8(5), eapMsg[1]) // Identifier should match
	assert.Equal(t, uint16(4), uint16(eapMsg[2])<<8|uint16(eapMsg[3]))
}

func TestSendEAPSuccess_DebugMode(t *testing.T) {
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), true)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(10, "testuser")
	req := &radius.Request{Packet: packet}
	response := req.Response(radius.CodeAccessAccept)

	err := coordinator.SendEAPSuccess(writer, req, response, "secret")

	require.NoError(t, err)
	assert.NotNil(t, writer.response)
}

// Tests for SendEAPFailure

func TestSendEAPFailure(t *testing.T) {
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(7, "testuser")
	req := &radius.Request{Packet: packet}

	err := coordinator.SendEAPFailure(writer, req, "secret", errors.New("auth failed"))

	require.NoError(t, err)
	assert.NotNil(t, writer.response)
	assert.Equal(t, radius.CodeAccessReject, writer.response.Code)

	// Verify EAP-Failure message
	eapMsg, err := rfc2869.EAPMessage_Lookup(writer.response)
	require.NoError(t, err)
	assert.Equal(t, uint8(CodeFailure), eapMsg[0])
	assert.Equal(t, uint8(7), eapMsg[1]) // Identifier should match
	assert.Equal(t, uint16(4), uint16(eapMsg[2])<<8|uint16(eapMsg[3]))
}

func TestSendEAPFailure_WithNilReason(t *testing.T) {
	coordinator := NewCoordinator(newMockStateManager(), &mockPasswordProvider{}, newMockHandlerRegistry(), false)
	writer := &mockResponseWriter{}

	packet := createEAPIdentityResponse(1, "testuser")
	req := &radius.Request{Packet: packet}

	err := coordinator.SendEAPFailure(writer, req, "secret", nil)

	require.NoError(t, err)
	assert.NotNil(t, writer.response)
	assert.Equal(t, radius.CodeAccessReject, writer.response.Code)
}

// Tests for CleanupState

func TestCleanupState_WithState(t *testing.T) {
	stateManager := newMockStateManager()
	stateManager.states["test-state-123"] = &EAPState{
		Username: "testuser",
		StateID:  "test-state-123",
	}

	coordinator := NewCoordinator(stateManager, &mockPasswordProvider{}, newMockHandlerRegistry(), false)

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	_ = rfc2865.State_SetString(packet, "test-state-123") //nolint:errcheck
	req := &radius.Request{Packet: packet}

	coordinator.CleanupState(req)

	// State should be deleted
	_, err := stateManager.GetState("test-state-123")
	assert.Error(t, err)
	assert.Equal(t, ErrStateNotFound, err)
}

func TestCleanupState_WithoutState(t *testing.T) {
	stateManager := newMockStateManager()
	coordinator := NewCoordinator(stateManager, &mockPasswordProvider{}, newMockHandlerRegistry(), false)

	// Packet without State attribute
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	req := &radius.Request{Packet: packet}

	// Should not panic
	coordinator.CleanupState(req)
}

func TestCleanupState_StateNotInManager(t *testing.T) {
	stateManager := newMockStateManager()
	coordinator := NewCoordinator(stateManager, &mockPasswordProvider{}, newMockHandlerRegistry(), false)

	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	_ = rfc2865.State_SetString(packet, "non-existent-state") //nolint:errcheck
	req := &radius.Request{Packet: packet}

	// Should not panic even if state doesn't exist
	coordinator.CleanupState(req)
}

// Integration test: Full EAP-MD5 authentication flow simulation

func TestCoordinator_FullEAPMD5Flow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	stateManager := newMockStateManager()
	pwdProvider := &mockPasswordProvider{password: "testpassword"}
	registry := newMockHandlerRegistry()

	// Register a mock MD5 handler that simulates the full flow
	md5Handler := &mockEAPHandler{
		name:             "eap-md5",
		eapType:          TypeMD5Challenge,
		canHandle:        true,
		handleIdentityOk: true,
		handleResponseOk: true,
	}
	registry.Register(md5Handler)

	coordinator := NewCoordinator(stateManager, pwdProvider, registry, true)

	user := &domain.RadiusUser{
		Username: "testuser",
		Password: "testpassword",
	}
	nas := &domain.NetNas{
		Identifier: "test-nas",
	}

	// Phase 1: EAP-Response/Identity
	writer1 := &mockResponseWriter{}
	identityPacket := createEAPIdentityResponse(1, "testuser")
	req1 := &radius.Request{Packet: identityPacket}
	response1 := req1.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer1, req1, user, nas, response1, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.False(t, success) // Identity phase returns false for success
	assert.NoError(t, err)
	assert.NotNil(t, writer1.response)
	assert.Equal(t, radius.CodeAccessChallenge, writer1.response.Code)

	// Phase 2: EAP-Response (Challenge Response)
	writer2 := &mockResponseWriter{}
	challengeResponsePacket := createEAPChallengeResponse(2, TypeMD5Challenge, []byte{16, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	req2 := &radius.Request{Packet: challengeResponsePacket}
	response2 := req2.Response(radius.CodeAccessAccept)

	handled, success, err = coordinator.HandleEAPRequest(
		writer2, req2, user, nas, response2, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.True(t, success)
	assert.NoError(t, err)

	// Phase 3: Send EAP-Success
	writer3 := &mockResponseWriter{}
	err = coordinator.SendEAPSuccess(writer3, req2, response2, "secret")

	assert.NoError(t, err)
	assert.NotNil(t, writer3.response)

	// Verify EAP-Success was sent
	eapMsg, _ := rfc2869.EAPMessage_Lookup(writer3.response)
	assert.Equal(t, uint8(CodeSuccess), eapMsg[0])
}

// Integration test: EAP authentication failure flow

func TestCoordinator_EAPAuthFailureFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	stateManager := newMockStateManager()
	pwdProvider := &mockPasswordProvider{password: "wrongpassword"}
	registry := newMockHandlerRegistry()

	md5Handler := &mockEAPHandler{
		name:              "eap-md5",
		eapType:           TypeMD5Challenge,
		canHandle:         true,
		handleIdentityOk:  true,
		handleResponseOk:  false,
		handleResponseErr: ErrPasswordMismatch,
	}
	registry.Register(md5Handler)

	coordinator := NewCoordinator(stateManager, pwdProvider, registry, false)

	user := &domain.RadiusUser{
		Username: "testuser",
		Password: "correctpassword",
	}

	// Phase 1: Identity
	writer1 := &mockResponseWriter{}
	identityPacket := createEAPIdentityResponse(1, "testuser")
	req1 := &radius.Request{Packet: identityPacket}
	response1 := req1.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer1, req1, user, &domain.NetNas{}, response1, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.False(t, success)
	assert.NoError(t, err)

	// Phase 2: Challenge response with wrong password
	writer2 := &mockResponseWriter{}
	challengePacket := createEAPChallengeResponse(2, TypeMD5Challenge, []byte{16, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	req2 := &radius.Request{Packet: challengePacket}
	response2 := req2.Response(radius.CodeAccessAccept)

	handled, success, err = coordinator.HandleEAPRequest(
		writer2, req2, user, &domain.NetNas{}, response2, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.False(t, success)
	assert.Error(t, err)
	assert.Equal(t, ErrPasswordMismatch, err)

	// Phase 3: Send EAP-Failure
	writer3 := &mockResponseWriter{}
	err = coordinator.SendEAPFailure(writer3, req2, "secret", ErrPasswordMismatch)

	assert.NoError(t, err)
	assert.Equal(t, radius.CodeAccessReject, writer3.response.Code)

	// Verify EAP-Failure was sent
	eapMsg, _ := rfc2869.EAPMessage_Lookup(writer3.response)
	assert.Equal(t, uint8(CodeFailure), eapMsg[0])
}

// Test Nak negotiation flow

func TestCoordinator_EAPNakNegotiationFlow(t *testing.T) {
	stateManager := newMockStateManager()
	registry := newMockHandlerRegistry()

	// Server offers MD5, but client prefers MSCHAPv2
	registry.Register(&mockEAPHandler{
		name:             "eap-md5",
		eapType:          TypeMD5Challenge,
		canHandle:        true,
		handleIdentityOk: true,
	})
	registry.Register(&mockEAPHandler{
		name:             "eap-mschapv2",
		eapType:          TypeMSCHAPv2,
		canHandle:        true,
		handleIdentityOk: true,
		handleResponseOk: true,
	})

	coordinator := NewCoordinator(stateManager, &mockPasswordProvider{password: "test"}, registry, false)

	// Phase 1: Identity
	writer1 := &mockResponseWriter{}
	identityPacket := createEAPIdentityResponse(1, "testuser")
	req1 := &radius.Request{Packet: identityPacket}
	response1 := req1.Response(radius.CodeAccessAccept)

	handled, _, err := coordinator.HandleEAPRequest(
		writer1, req1, &domain.RadiusUser{}, &domain.NetNas{}, response1, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.NoError(t, err)

	// Phase 2: Client sends Nak, suggesting MSCHAPv2
	writer2 := &mockResponseWriter{}
	nakPacket := createEAPNakResponse(2, TypeMSCHAPv2)
	req2 := &radius.Request{Packet: nakPacket}
	response2 := req2.Response(radius.CodeAccessAccept)

	handled, success, err := coordinator.HandleEAPRequest(
		writer2, req2, &domain.RadiusUser{}, &domain.NetNas{}, response2, "secret", false, "eap-md5",
	)

	assert.True(t, handled)
	assert.False(t, success) // New challenge sent
	assert.NoError(t, err)
	assert.NotNil(t, writer2.response)
	assert.Equal(t, radius.CodeAccessChallenge, writer2.response.Code)

	// Phase 3: Client responds to MSCHAPv2 challenge
	writer3 := &mockResponseWriter{}
	mschapPacket := createEAPChallengeResponse(3, TypeMSCHAPv2, []byte{1, 2, 3, 4})
	req3 := &radius.Request{Packet: mschapPacket}
	response3 := req3.Response(radius.CodeAccessAccept)

	handled, success, err = coordinator.HandleEAPRequest(
		writer3, req3, &domain.RadiusUser{}, &domain.NetNas{}, response3, "secret", false, "eap-mschapv2",
	)

	assert.True(t, handled)
	assert.True(t, success)
	assert.NoError(t, err)
}

// Test interface compliance

func TestCoordinator_Interfaces(t *testing.T) {
	// Verify mock implementations satisfy interfaces
	var _ EAPStateManager = (*mockStateManager)(nil)
	var _ PasswordProvider = (*mockPasswordProvider)(nil)
	var _ EAPHandler = (*mockEAPHandler)(nil)
	var _ HandlerRegistry = (*mockHandlerRegistry)(nil)
	var _ radius.ResponseWriter = (*mockResponseWriter)(nil)
}
