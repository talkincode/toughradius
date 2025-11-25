package handlers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	OTPChallengeMessage = "Please enter a one-time password"
	EAPMethodOTP        = "eap-otp"
)

// OTPHandler EAP-OTP authenticationhandler
type OTPHandler struct{}

// NewOTPHandler Create EAP-OTP handler
func NewOTPHandler() *OTPHandler {
	return &OTPHandler{}
}

// Name Returnshandlernames
func (h *OTPHandler) Name() string {
	return EAPMethodOTP
}

// EAPType returns the EAP type code
func (h *OTPHandler) EAPType() uint8 {
	return eap.TypeOTP
}

// CanHandle checks whether this handler can process the EAP message
func (h *OTPHandler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeOTP
}

// HandleIdentity Handle EAP-Response/Identity，Send OTP Challenge
func (h *OTPHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	// The OTP challenge is a textual message
	challenge := []byte(OTPChallengeMessage)

	// Create Challenge Request
	eapData := h.buildChallengeRequest(ctx.EAPMessage.Identifier, challenge)

	// Create RADIUS Access-Challenge response
	response := ctx.Request.Response(radius.CodeAccessChallenge)

	// Generate and save the state
	stateID := common.UUID()
	username := rfc2865.UserName_GetString(ctx.Request.Packet)

	state := &eap.EAPState{
		Username:  username,
		Challenge: challenge,
		StateID:   stateID,
		Method:    EAPMethodOTP,
		Success:   false,
	}

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}

	// Set the State attribute
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck

	// Set the EAP-Message and Message-Authenticator
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)

	// Sendresponse
	return true, ctx.ResponseWriter.Write(response)
}

// HandleResponse Handle EAP-Response (OTP Response)
func (h *OTPHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	// getStatus
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	// Get the OTP password (simplified here; a real implementation should call an OTP validation service)
	// TODO: Integrate with a real OTP validation logic
	expectedOTP := "123456" // Sample fixed value; in reality retrieve from the validation service

	// Extract the user's entered OTP from the EAP data
	userOTP := string(ctx.EAPMessage.Data)

	// Validate OTP
	if userOTP != expectedOTP {
		return false, eap.ErrPasswordMismatch
	}

	// Mark authentication as successful
	state.Success = true
	_ = ctx.StateManager.SetState(stateID, state) //nolint:errcheck

	return true, nil
}

// buildChallengeRequest constructs the OTP Challenge Request
func (h *OTPHandler) buildChallengeRequest(identifier uint8, challenge []byte) []byte {
	// EAP-OTP format:
	// Code (1) | Identifier (1) | Length (2) | Type (1) | Challenge (variable)

	dataLen := len(challenge)
	totalLen := 5 + dataLen // EAP header (4) + Type (1) + challenge

	buffer := make([]byte, totalLen)
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	buffer[2] = byte(totalLen >> 8)
	buffer[3] = byte(totalLen)
	buffer[4] = eap.TypeOTP
	copy(buffer[5:], challenge)

	return buffer
}
