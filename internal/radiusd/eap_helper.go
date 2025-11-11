package radiusd

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"layeh.com/radius"
)

// EAPAuthHelper EAP authentication helper
type EAPAuthHelper struct {
	coordinator *eap.Coordinator
}

// NewEAPAuthHelper Create EAP authentication helper
func NewEAPAuthHelper() *EAPAuthHelper {
	// Createstate manager
	stateManager := statemanager.NewMemoryStateManager()

	// Createpassword provider
	pwdProvider := eap.NewDefaultPasswordProvider()

	// gethandler registry
	var handlerRegistry eap.HandlerRegistry = registry.GetGlobalRegistry()

	// Createcoordinator
	coordinator := eap.NewCoordinator(stateManager, pwdProvider, handlerRegistry)

	return &EAPAuthHelper{
		coordinator: coordinator,
	}
}

// HandleEAPAuthentication Handle EAP authentication
// Returns (handled bool, success bool, err error)
func (h *EAPAuthHelper) HandleEAPAuthentication(
	w radius.ResponseWriter,
	r *radius.Request,
	user *domain.RadiusUser,
	nas *domain.NetNas,
	vendorReq *vendorparsers.VendorRequest,
	response *radius.Packet,
	eapMethod string,
) (handled bool, success bool, err error) {

	// Check if is MAC authentication
	isMacAuth := vendorReq.MacAddr != "" && vendorReq.MacAddr == user.Username

	// Call coordinator to handle EAP request
	handled, success, err = h.coordinator.HandleEAPRequest(
		w, r, user, nas, response, nas.Secret, isMacAuth, eapMethod,
	)

	return handled, success, err
}

// SendEAPSuccess Send EAP Success response
func (h *EAPAuthHelper) SendEAPSuccess(
	w radius.ResponseWriter,
	r *radius.Request,
	response *radius.Packet,
	secret string,
) error {
	return h.coordinator.SendEAPSuccess(w, r, response, secret)
}

// SendEAPFailure Send EAP Failure response
func (h *EAPAuthHelper) SendEAPFailure(
	w radius.ResponseWriter,
	r *radius.Request,
	secret string,
	reason error,
) error {
	return h.coordinator.SendEAPFailure(w, r, secret, reason)
}

// CleanupState Cleanup EAP Status
func (h *EAPAuthHelper) CleanupState(r *radius.Request) {
	h.coordinator.CleanupState(r)
}

// GetCoordinator Get underlying coordinator(for advanced usage)
func (h *EAPAuthHelper) GetCoordinator() *eap.Coordinator {
	return h.coordinator
}
