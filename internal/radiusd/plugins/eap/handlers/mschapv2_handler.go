package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/microsoft"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius"
	"layeh.com/radius/rfc2759"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc3079"
)

const (
	// MSCHAPv2 OpCodes
	MSCHAPv2Challenge = 1
	MSCHAPv2Response  = 2
	MSCHAPv2Success   = 3
	MSCHAPv2Failure   = 4

	// MSCHAPv2 constants
	MSCHAPChallengeSize = 16
	MSCHAPResponseSize  = 49 // PeerChallenge(16) + Reserved(8) + NTResponse(24) + Flags(1)
	EAPMethodMSCHAPv2   = "eap-mschapv2"
	ServerName          = "toughradius"
)

// MSCHAPv2Handler EAP-MSCHAPv2 authenticationhandler
type MSCHAPv2Handler struct{}

// NewMSCHAPv2Handler Create EAP-MSCHAPv2 handler
func NewMSCHAPv2Handler() *MSCHAPv2Handler {
	return &MSCHAPv2Handler{}
}

// Name Returnshandlernames
func (h *MSCHAPv2Handler) Name() string {
	return EAPMethodMSCHAPv2
}

// EAPType returns the EAP type code
func (h *MSCHAPv2Handler) EAPType() uint8 {
	return eap.TypeMSCHAPv2
}

// CanHandle checks whether this handler can process the EAP message
func (h *MSCHAPv2Handler) CanHandle(ctx *eap.EAPContext) bool {
	if ctx.EAPMessage == nil {
		return false
	}
	return ctx.EAPMessage.Type == eap.TypeMSCHAPv2
}

// HandleIdentity Handle EAP-Response/Identity，Send MSCHAPv2 Challenge
func (h *MSCHAPv2Handler) HandleIdentity(ctx *eap.EAPContext) (bool, error) {
	// Generate a random Authenticator Challenge (16 bytes)
	challenge, err := eap.GenerateRandomBytes(MSCHAPChallengeSize)
	if err != nil {
		return false, fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Build the EAP-MSCHAPv2 Challenge Request
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
		Method:    EAPMethodMSCHAPv2,
		Success:   false,
		Data:      make(map[string]interface{}),
	}

	// Save the MSIdentifier for later use
	state.Data["ms_identifier"] = ctx.EAPMessage.Identifier

	if err := ctx.StateManager.SetState(stateID, state); err != nil {
		return false, fmt.Errorf("failed to save state: %w", err)
	}

	// Set the State attribute
	_ = rfc2865.State_SetString(response, stateID) //nolint:errcheck

	// Set the EAP-Message and Message-Authenticator
	eap.SetEAPMessageAndAuth(response, eapData, ctx.Secret)

	// Sendresponse
	return true, ctx.ResponseWriter.Write(response)
}

// HandleResponse Handle EAP-Response (MSCHAPv2 Response)
func (h *MSCHAPv2Handler) HandleResponse(ctx *eap.EAPContext) (bool, error) {
	// getStatus
	stateID := rfc2865.State_GetString(ctx.Request.Packet)
	if stateID == "" {
		return false, eap.ErrStateNotFound
	}

	state, err := ctx.StateManager.GetState(stateID)
	if err != nil {
		return false, err
	}

	// Parse MSCHAPv2 Response
	msResp, err := h.parseResponse(ctx.EAPMessage.Data)
	if err != nil {
		return false, fmt.Errorf("failed to parse MSCHAPv2 response: %w", err)
	}

	// getPassword
	password, err := ctx.PwdProvider.GetPassword(ctx.User, ctx.IsMacAuth)
	if err != nil {
		return false, err
	}

	// Validate MSCHAPv2 Response
	success, err := h.verifyResponse(
		ctx.User.Username,
		password,
		state.Challenge,
		msResp.PeerChallenge,
		msResp.NTResponse,
		ctx.Response,
		msResp.MsIdentifier,
	)

	if err != nil {
		return false, err
	}

	if !success {
		return false, eap.ErrPasswordMismatch
	}

	// Mark authentication as successful
	state.Success = true
	_ = ctx.StateManager.SetState(stateID, state) //nolint:errcheck

	return true, nil
}

// buildChallengeRequest constructs the MSCHAPv2 Challenge Request
// EAP-MSCHAPv2 Challenge format:
// Code(1) | Identifier(1) | Length(2) | Type(1) | OpCode(1) | MS-CHAPv2-ID(1) |
// MS-Length(2) | Value-Size(1) | Challenge(16) | Name(variable)
func (h *MSCHAPv2Handler) buildChallengeRequest(identifier uint8, challenge []byte) []byte {
	serverName := []byte(ServerName)

	// Compute the MS-CHAPv2 data length
	// OpCode(1) + MS-CHAPv2-ID(1) + MS-Length(2) + Value-Size(1) + Challenge(16) + Name(len)
	msDataLen := 1 + 1 + 2 + 1 + MSCHAPChallengeSize + len(serverName)

	// Compute the total EAP length
	// EAP Header(4) + Type(1) + MS-CHAPv2 Data
	totalLen := 4 + 1 + msDataLen

	buffer := make([]byte, totalLen)

	// EAP Header
	buffer[0] = eap.CodeRequest
	buffer[1] = identifier
	binary.BigEndian.PutUint16(buffer[2:4], uint16(totalLen))

	// EAP Type
	buffer[4] = eap.TypeMSCHAPv2

	// MS-CHAPv2 Data
	offset := 5
	buffer[offset] = MSCHAPv2Challenge                                       // OpCode
	buffer[offset+1] = identifier                                            // MS-CHAPv2-ID (matches the EAP identifier)
	binary.BigEndian.PutUint16(buffer[offset+2:offset+4], uint16(msDataLen)) // MS-Length
	buffer[offset+4] = MSCHAPChallengeSize                                   // Value-Size
	copy(buffer[offset+5:offset+5+MSCHAPChallengeSize], challenge)           // Challenge
	copy(buffer[offset+5+MSCHAPChallengeSize:], serverName)                  // Name

	return buffer
}

// MSCHAPv2ResponseData defines the MSCHAPv2 response structure
type MSCHAPv2ResponseData struct {
	OpCode        uint8
	MsIdentifier  uint8
	MsLength      uint16
	ValueSize     uint8
	PeerChallenge []byte // 16 bytes
	Reserved      []byte // 8 bytes
	NTResponse    []byte // 24 bytes
	Flags         uint8
	Name          []byte
}

// parseResponse Parse MSCHAPv2 Response
// EAP-MSCHAPv2 Response format:
// OpCode(1) | MS-CHAPv2-ID(1) | MS-Length(2) | Value-Size(1) |
// Peer-Challenge(16) | Reserved(8) | NT-Response(24) | Flags(1) | Name(variable)
func (h *MSCHAPv2Handler) parseResponse(data []byte) (*MSCHAPv2ResponseData, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("MSCHAPv2 response too short: %d bytes", len(data))
	}

	resp := &MSCHAPv2ResponseData{}
	offset := 0

	resp.OpCode = data[offset]
	resp.MsIdentifier = data[offset+1]
	resp.MsLength = binary.BigEndian.Uint16(data[offset+2 : offset+4])
	resp.ValueSize = data[offset+4]
	offset += 5

	// Check the opcode and value size
	if resp.OpCode != MSCHAPv2Response {
		return nil, fmt.Errorf("invalid OpCode: expected %d, got %d", MSCHAPv2Response, resp.OpCode)
	}

	if resp.ValueSize != MSCHAPResponseSize {
		return nil, fmt.Errorf("invalid ValueSize: expected %d, got %d", MSCHAPResponseSize, resp.ValueSize)
	}

	// Check the remaining data length
	if len(data) < offset+int(resp.ValueSize) {
		return nil, fmt.Errorf("insufficient data for response value: need %d, have %d",
			offset+int(resp.ValueSize), len(data))
	}

	// Parse Peer-Challenge, Reserved, NT-Response, Flags
	resp.PeerChallenge = data[offset : offset+16]
	resp.Reserved = data[offset+16 : offset+24]
	resp.NTResponse = data[offset+24 : offset+48]
	resp.Flags = data[offset+48]
	offset += int(resp.ValueSize)

	// Name (remaining portion)
	if len(data) > offset {
		resp.Name = data[offset:]
	}

	return resp, nil
}

// verifyResponse validates the MSCHAPv2 response and generates encryption keys
func (h *MSCHAPv2Handler) verifyResponse(
	username string,
	password string,
	authChallenge []byte,
	peerChallenge []byte,
	ntResponse []byte,
	response *radius.Packet,
	msIdentifier uint8,
) (bool, error) {
	byteUser := []byte(username)
	bytePwd := []byte(password)

	// Using RFC 2759 Generate NT-Response
	expectedNTResponse, err := rfc2759.GenerateNTResponse(
		authChallenge,
		peerChallenge,
		byteUser,
		bytePwd,
	)
	if err != nil {
		return false, fmt.Errorf("failed to generate NT-Response: %w", err)
	}

	// Validate NT-Response
	if !bytes.Equal(expectedNTResponse, ntResponse) {
		return false, nil
	}

	// Generate MPPE keys
	recvKey, err := rfc3079.MakeKey(expectedNTResponse, bytePwd, false)
	if err != nil {
		return false, fmt.Errorf("failed to generate recv key: %w", err)
	}

	sendKey, err := rfc3079.MakeKey(expectedNTResponse, bytePwd, true)
	if err != nil {
		return false, fmt.Errorf("failed to generate send key: %w", err)
	}

	// Generate Authenticator Response (RFC 2759)
	authenticatorResponse, err := rfc2759.GenerateAuthenticatorResponse(
		authChallenge,
		peerChallenge,
		expectedNTResponse,
		byteUser,
		bytePwd,
	)
	if err != nil {
		return false, fmt.Errorf("failed to generate authenticator response: %w", err)
	}

	// Construct the MSCHAPv2-Success attribute value
	// format: Ident(1) + Authenticator-Response(42)
	success := make([]byte, 43)
	success[0] = msIdentifier
	copy(success[1:], authenticatorResponse)

	// Add Microsoft-specific attributes to the response
	_ = microsoft.MSCHAP2Success_Add(response, success) //nolint:errcheck
	_ = microsoft.MSMPPERecvKey_Add(response, recvKey)  //nolint:errcheck
	_ = microsoft.MSMPPESendKey_Add(response, sendKey)  //nolint:errcheck
	_ = microsoft.MSMPPEEncryptionPolicy_Add(response,  //nolint:errcheck
		microsoft.MSMPPEEncryptionPolicy_Value_EncryptionAllowed)
	_ = microsoft.MSMPPEEncryptionTypes_Add(response, //nolint:errcheck
		microsoft.MSMPPEEncryptionTypes_Value_RC440or128BitAllowed)

	return true, nil
}
