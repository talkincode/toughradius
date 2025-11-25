package radiusd

import (
	"strings"

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
func NewEAPAuthHelper(radiusService *RadiusService, allowedHandlers []string) *EAPAuthHelper {
	// Create state manager
	stateManager := statemanager.NewMemoryStateManager()

	// Create password provider
	pwdProvider := eap.NewDefaultPasswordProvider()

	// get handler registry
	var handlerRegistry eap.HandlerRegistry = registry.GetGlobalRegistry()
	if len(allowedHandlers) > 0 {
		handlerRegistry = newFilteredHandlerRegistry(handlerRegistry, allowedHandlers)
	}

	// Get debug flag from config
	debug := radiusService.Config().Radiusd.Debug

	// Create coordinator with debug flag
	coordinator := eap.NewCoordinator(stateManager, pwdProvider, handlerRegistry, debug)

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

type filteredHandlerRegistry struct {
	base    eap.HandlerRegistry
	allowed map[string]struct{}
}

func newFilteredHandlerRegistry(base eap.HandlerRegistry, allowed []string) eap.HandlerRegistry {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		norm := strings.ToLower(strings.TrimSpace(name))
		if norm == "" {
			continue
		}
		allowedSet[norm] = struct{}{}
	}
	if len(allowedSet) == 0 {
		return base
	}
	return &filteredHandlerRegistry{base: base, allowed: allowedSet}
}

func (f *filteredHandlerRegistry) GetHandler(eapType uint8) (eap.EAPHandler, bool) {
	if f == nil || f.base == nil {
		return nil, false
	}
	handler, ok := f.base.GetHandler(eapType)
	if !ok {
		return nil, false
	}
	if len(f.allowed) == 0 {
		return handler, true
	}
	if _, allowed := f.allowed[strings.ToLower(handler.Name())]; allowed {
		return handler, true
	}
	return nil, false
}
