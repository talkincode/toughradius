package handlers

import (
	"crypto/md5"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	MD5ChallengeLength = 16
	EAPMethodMD5       = "eap-md5"
)

// MD5Handler EAP-MD5 authenticationhandler
type MD5Handler struct{}

// NewMD5Handler Create EAP-MD5 handler
func NewMD5Handler() *MD5Handler {
	return &MD5Handler{}
}

// Name Returnshandlernames
func (h *MD5Handler) Name() string {
	return EAPMethodMD5
}

// EAPType returns the EAP type code
func (h *MD5Handler) EAPType() uint8 {
	return eap.TypeMD5Challenge
}

// CanHandle checks whether this handler can process the EAP message
func (h *MD5Handler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeMD5Challenge
}

// HandleIdentity Handle EAP-Response/Identity，Send MD5 Challenge
func (h *MD5Handler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	// Generate a random challenge
	challenge, err := eap.GenerateRandomBytes(MD5ChallengeLength)
	if err != nil {
		return false, err
	}

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
		Method:    EAPMethodMD5,
		Success:   false,
	}

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, err
	}

	// Set the State attribute
	rfc2865.State_SetString(response, stateID)

	// Set the EAP-Message and Message-Authenticator
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)

	// Sendresponse
	return true, ctx.ResponseWriter.Write(response)
}

// HandleResponse Handle EAP-Response (Challenge Response)
func (h *MD5Handler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	// getStatus
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	// getPassword
	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return false, err
	}

	// Validate MD5 response
	if !h.verifyMD5Response(ctx.EAPMessage.Identifier, password, state.Challenge, ctx.EAPMessage.Data) {
		return false, eap.ErrPasswordMismatch
	}

	// Mark authentication as successful
	state.Success = true
	ctx.StateManager.SetState(stateID, state)

	return true, nil
}

// buildChallengeRequest constructs the MD5 Challenge Request
func (h *MD5Handler) buildChallengeRequest(identifier uint8, challenge []byte) []byte {
	// EAP-MD5 format:
	// Code (1) | Identifier (1) | Length (2) | Type (1) | Value-Size (1) | Value (16) | Name (optional)

	valueSize := byte(len(challenge))
	dataLen := 1 + len(challenge) // Value-Size + Value
	totalLen := 5 + dataLen       // EAP header (4) + Type (1) + data

	buffer := make([]byte, totalLen)
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	buffer[2] = byte(totalLen >> 8)
	buffer[3] = byte(totalLen)
	buffer[4] = eap.TypeMD5Challenge
	buffer[5] = valueSize
	copy(buffer[6:], challenge)

	return buffer
}

// verifyMD5Response Validate MD5 response
// MD5(identifier + password + challenge) == response
func (h *MD5Handler) verifyMD5Response(identifier uint8, password string, challenge, response []byte) bool {
	if len(response) < 1 {
		return false
	}

	// responseformat: Value-Size (1) + Value (16)
		// Skip the Value-Size
	actualResponse := response[1:]

		// Compute the expected MD5
	hash := md5.New()
	hash.Write([]byte{identifier})
	hash.Write([]byte(password))
	hash.Write(challenge)
	expectedResponse := hash.Sum(nil)

	return eap.VerifyMD5Hash(expectedResponse, actualResponse)
}
