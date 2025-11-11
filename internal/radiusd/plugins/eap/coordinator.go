package eap

import (
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// HandlerRegistry defines the EAP handler registry interface
type HandlerRegistry interface {
	GetHandler(eapType uint8) (EAPHandler, bool)
}

// Coordinator orchestrates EAP message dispatching and handling
type Coordinator struct {
	stateManager    EAPStateManager
	pwdProvider     PasswordProvider
	handlerRegistry HandlerRegistry
}

// NewCoordinator creates a new EAP coordinator
func NewCoordinator(stateManager EAPStateManager, pwdProvider PasswordProvider, handlerRegistry HandlerRegistry) *Coordinator {
	return &Coordinator{
		stateManager:    stateManager,
		pwdProvider:     pwdProvider,
		handlerRegistry: handlerRegistry,
	}
}

// HandleEAPRequest Handle EAP request
// Returns: handled bool, success bool, err error
// handled: whether the request was handled
// success: whether authentication succeeded (only meaningful when handled=true)
// err: error occurred during handling
func (c *Coordinator) HandleEAPRequest(
	w radius.ResponseWriter,
	r *radius.Request,
	user *domain.RadiusUser,
	nas *domain.NetNas,
	response *radius.Packet,
	secret string,
	isMacAuth bool,
	configuredMethod string, // Configured EAP method (eap-md5, eap-mschapv2, etc.)
) (handled bool, success bool, err error) {

	// Parse the EAP message
	eapMsg, err := ParseEAPMessage(r.Packet)
	if err != nil {
		// Not an EAP request
		return false, false, nil
	}

	// Create the EAP context
	ctx := &EAPContext{
		Request:        r,
		ResponseWriter: w,
		Response:       response,
		User:           user,
		NAS:            nas,
		EAPMessage:     eapMsg,
		Secret:         secret,
		IsMacAuth:      isMacAuth,
		StateManager:   c.stateManager,
		PwdProvider:    c.pwdProvider,
	}

	// Handle EAP-Response/Identity
	if eapMsg.Code == CodeResponse && eapMsg.Type == TypeIdentity {
		return c.handleIdentityResponse(ctx, configuredMethod)
	}

	// Handle EAP-Response/Nak
	if eapMsg.Code == CodeResponse && eapMsg.Type == TypeNak {
		return c.handleNak(ctx)
	}

	// Handle EAP-Response (Challenge Response)
	if eapMsg.Code == CodeResponse {
		return c.handleChallengeResponse(ctx)
	}

	return false, false, fmt.Errorf("unsupported EAP code: %d", eapMsg.Code)
}

// handleIdentityResponse processes EAP-Response/Identity and sends the configured challenge
func (c *Coordinator) handleIdentityResponse(ctx *EAPContext, configuredMethod string) (bool, bool, error) {
	// Select the handler for the configured method
	var handler EAPHandler

	switch configuredMethod {
	case "eap-md5":
		handler, _ = c.handlerRegistry.GetHandler(TypeMD5Challenge)
	case "eap-mschapv2":
		handler, _ = c.handlerRegistry.GetHandler(TypeMSCHAPv2)
	case "eap-otp":
		handler, _ = c.handlerRegistry.GetHandler(TypeOTP)
	default:
		// Default to MD5
		handler, _ = c.handlerRegistry.GetHandler(TypeMD5Challenge)
	}

	if handler == nil {
		return false, false, fmt.Errorf("EAP handler not found for method: %s", configuredMethod)
	}

	// Invoke the handler to send the challenge
	handled, err := handler.HandleIdentity(ctx)
	if err != nil {
		zap.L().Error("EAP HandleIdentity failed",
			zap.String("method", configuredMethod),
			zap.Error(err))
	}

	// Identity phase does not return success; wait for the challenge response
	return handled, false, err
}

// handleNak handles EAP-Response/Nak when the client rejects the current method
func (c *Coordinator) handleNak(ctx *EAPContext) (bool, bool, error) {
	if len(ctx.EAPMessage.Data) == 0 {
		return false, false, fmt.Errorf("Nak message has no alternative methods")
	}

	// Take the first method suggested by the client
	suggestedType := ctx.EAPMessage.Data[0]

	handler, ok := c.handlerRegistry.GetHandler(suggestedType)
	if !ok {
		return false, false, fmt.Errorf("unsupported EAP type: %d", suggestedType)
	}

	// Send the challenge using the suggested method
	handled, err := handler.HandleIdentity(ctx)
	return handled, false, err
}

// handleChallengeResponse processes EAP-Response (Challenge Response)
func (c *Coordinator) handleChallengeResponse(ctx *EAPContext) (bool, bool, error) {
	// Get the handler based on the EAP type
	handler, ok := c.handlerRegistry.GetHandler(ctx.EAPMessage.Type)
	if !ok {
		return false, false, fmt.Errorf("unsupported EAP type: %d", ctx.EAPMessage.Type)
	}

	// Check whether the handler can process this message
	if !handler.CanHandle(ctx) {
		return false, false, fmt.Errorf("handler cannot handle this EAP message")
	}

	// Invoke the handler to validate the response
	success, err := handler.HandleResponse(ctx)
	if err != nil {
		zap.L().Error("EAP HandleResponse failed",
			zap.String("method", handler.Name()),
			zap.Error(err))
		return true, false, err
	}

	return true, success, nil
}

// SendEAPSuccess Send EAP-Success response
func (c *Coordinator) SendEAPSuccess(w radius.ResponseWriter, r *radius.Request, response *radius.Packet, secret string) error {
	eapMsg, _ := ParseEAPMessage(r.Packet)
	identifier := eapMsg.Identifier

	// Create the EAP-Success message
	eapSuccess := EncodeEAPHeader(CodeSuccess, identifier)

	// Set the EAP-Message and Message-Authenticator
	SetEAPMessageAndAuth(response, eapSuccess, secret)

	// Sendresponse
	if app.GConfig().Radiusd.Debug {
		zap.L().Info("Sending EAP-Success",
			zap.Uint8("identifier", identifier))
	}

	return w.Write(response)
}

// SendEAPFailure Send EAP-Failure response
func (c *Coordinator) SendEAPFailure(w radius.ResponseWriter, r *radius.Request, secret string, reason error) error {
	eapMsg, _ := ParseEAPMessage(r.Packet)
	identifier := eapMsg.Identifier

	// Create the EAP-Failure message
	eapFailure := EncodeEAPHeader(CodeFailure, identifier)

	// Create RADIUS Reject response
	response := r.Response(radius.CodeAccessReject)

	// Set the EAP-Message and Message-Authenticator
	SetEAPMessageAndAuth(response, eapFailure, secret)

	// Log the failure event
	zap.L().Warn("Sending EAP-Failure",
		zap.Uint8("identifier", identifier),
		zap.Error(reason))

	return w.Write(response)
}

// CleanupState Cleanup EAP Status
func (c *Coordinator) CleanupState(r *radius.Request) {
	stateID := rfc2865.State_GetString(r.Packet)
	if stateID != "" {
		c.stateManager.DeleteState(stateID)
	}
}
