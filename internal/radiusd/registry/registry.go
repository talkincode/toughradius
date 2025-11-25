package registry

import (
	"sort"
	"sync"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
)

// Registry holds plugin registrations
type Registry struct {
	passwordValidators map[string]auth.PasswordValidator
	policyCheckers     []auth.PolicyChecker
	responseEnhancers  []auth.ResponseEnhancer
	authGuards         []auth.Guard
	acctHandlers       []accounting.AccountingHandler
	eapHandlers        map[uint8]eap.EAPHandler // EAP handlers indexed by EAP type
	mu                 sync.RWMutex
}

var globalRegistry = newRegistry()

func newRegistry() *Registry {
	return &Registry{
		passwordValidators: make(map[string]auth.PasswordValidator),
		policyCheckers:     make([]auth.PolicyChecker, 0),
		responseEnhancers:  make([]auth.ResponseEnhancer, 0),
		authGuards:         make([]auth.Guard, 0),
		acctHandlers:       make([]accounting.AccountingHandler, 0),
		eapHandlers:        make(map[uint8]eap.EAPHandler),
	}
}

// RegisterPasswordValidator registers a password validator
func RegisterPasswordValidator(validator auth.PasswordValidator) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.passwordValidators[validator.Name()] = validator
}

// GetPasswordValidators returns all password validators
func GetPasswordValidators() []auth.PasswordValidator {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	validators := make([]auth.PasswordValidator, 0, len(globalRegistry.passwordValidators))
	for _, v := range globalRegistry.passwordValidators {
		validators = append(validators, v)
	}
	return validators
}

// GetPasswordValidator returns a password validator by name
func GetPasswordValidator(name string) (auth.PasswordValidator, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	v, ok := globalRegistry.passwordValidators[name]
	return v, ok
}

// RegisterPolicyChecker registers a profile checker
func RegisterPolicyChecker(checker auth.PolicyChecker) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.policyCheckers = append(globalRegistry.policyCheckers, checker)
	// Sort by order
	sort.Slice(globalRegistry.policyCheckers, func(i, j int) bool {
		return globalRegistry.policyCheckers[i].Order() < globalRegistry.policyCheckers[j].Order()
	})
}

// GetPolicyCheckers returns all profile checkers (sorted by order)
func GetPolicyCheckers() []auth.PolicyChecker {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	// Returns a copy
	checkers := make([]auth.PolicyChecker, len(globalRegistry.policyCheckers))
	copy(checkers, globalRegistry.policyCheckers)
	return checkers
}

// RegisterResponseEnhancer registers a response enhancer
func RegisterResponseEnhancer(enhancer auth.ResponseEnhancer) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.responseEnhancers = append(globalRegistry.responseEnhancers, enhancer)
}

// GetResponseEnhancers returns all response enhancers
func GetResponseEnhancers() []auth.ResponseEnhancer {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	enhancers := make([]auth.ResponseEnhancer, len(globalRegistry.responseEnhancers))
	copy(enhancers, globalRegistry.responseEnhancers)
	return enhancers
}

// RegisterAuthGuard registers an authentication guard
func RegisterAuthGuard(guard auth.Guard) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.authGuards = append(globalRegistry.authGuards, guard)
}

// GetAuthGuards returns all authentication guards
func GetAuthGuards() []auth.Guard {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	guards := make([]auth.Guard, len(globalRegistry.authGuards))
	copy(guards, globalRegistry.authGuards)
	return guards
}

// RegisterVendorParser registers a vendor parser
func RegisterVendorParser(parser vendorparserspkg.VendorParser) {
	vendors.Register(&vendors.VendorInfo{
		Code:   parser.VendorCode(),
		Name:   parser.VendorName(),
		Parser: parser,
	})
}

// GetVendorParser returns a vendor parser
func GetVendorParser(vendorCode string) (vendorparserspkg.VendorParser, bool) {
	info, ok := vendors.Get(vendorCode)
	if ok && info.Parser != nil {
		return info.Parser, true
	}
	// Fallback to default
	info, ok = vendors.Get("default")
	if ok && info.Parser != nil {
		return info.Parser, true
	}
	return nil, false
}

// RegisterVendorResponseBuilder registers a vendor response builder
func RegisterVendorResponseBuilder(builder vendorparserspkg.VendorResponseBuilder) {
	vendors.Register(&vendors.VendorInfo{
		Code:    builder.VendorCode(),
		Builder: builder,
	})
}

// GetVendorResponseBuilder returns a vendor response builder
func GetVendorResponseBuilder(vendorCode string) (vendorparserspkg.VendorResponseBuilder, bool) {
	info, ok := vendors.Get(vendorCode)
	if ok && info.Builder != nil {
		return info.Builder, true
	}
	return nil, false
}

// RegisterAccountingHandler registers an accounting handler
func RegisterAccountingHandler(handler accounting.AccountingHandler) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.acctHandlers = append(globalRegistry.acctHandlers, handler)
}

// GetAccountingHandlers returns all accounting handlers
func GetAccountingHandlers() []accounting.AccountingHandler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	handlers := make([]accounting.AccountingHandler, len(globalRegistry.acctHandlers))
	copy(handlers, globalRegistry.acctHandlers)
	return handlers
}

// RegisterEAPHandler registers an EAP handler
func RegisterEAPHandler(handler eap.EAPHandler) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.eapHandlers[handler.EAPType()] = handler
}

// GetEAPHandler returns the handler for a given EAP type
func GetEAPHandler(eapType uint8) (eap.EAPHandler, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	handler, ok := globalRegistry.eapHandlers[eapType]
	return handler, ok
}

// GetAllEAPHandlers returns all EAP handlers
func GetAllEAPHandlers() map[uint8]eap.EAPHandler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	handlers := make(map[uint8]eap.EAPHandler, len(globalRegistry.eapHandlers))
	for k, v := range globalRegistry.eapHandlers {
		handlers[k] = v
	}
	return handlers
}

// GetHandler implements the eap.HandlerRegistry interface
func (r *Registry) GetHandler(eapType uint8) (eap.EAPHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.eapHandlers[eapType]
	return handler, ok
}

// GetGlobalRegistry returns the global registry instance
// Used to implement the eap.HandlerRegistry interface
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// ResetForTest clears the registry. Test helper only.
func ResetForTest() {
	globalRegistry = newRegistry()
}
