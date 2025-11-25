package registry

import (
	"context"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
)

// Mock implementations for auth interfaces

type mockPasswordValidator struct {
	name string
}

func (m *mockPasswordValidator) Name() string { return m.name }
func (m *mockPasswordValidator) CanHandle(ctx *auth.AuthContext) bool {
	return true
}
func (m *mockPasswordValidator) Validate(ctx context.Context, authCtx *auth.AuthContext, password string) error {
	return nil
}

type mockPolicyChecker struct {
	name  string
	order int
}

func (m *mockPolicyChecker) Name() string                                               { return m.name }
func (m *mockPolicyChecker) Check(ctx context.Context, authCtx *auth.AuthContext) error { return nil }
func (m *mockPolicyChecker) Order() int                                                 { return m.order }

type mockResponseEnhancer struct {
	name string
}

func (m *mockResponseEnhancer) Name() string { return m.name }
func (m *mockResponseEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	return nil
}

type mockGuard struct {
	name string
}

func (m *mockGuard) Name() string { return m.name }
func (m *mockGuard) OnError(ctx context.Context, authCtx *auth.AuthContext, stage string, err error) error {
	return err
}

// Mock implementations for accounting interfaces

type mockAccountingHandler struct {
	name string
}

func (m *mockAccountingHandler) Name() string { return m.name }
func (m *mockAccountingHandler) CanHandle(ctx *accounting.AccountingContext) bool {
	return true
}
func (m *mockAccountingHandler) Handle(ctx *accounting.AccountingContext) error {
	return nil
}

// Mock implementations for EAP interfaces

type mockEAPHandler struct {
	name    string
	eapType uint8
}

func (m *mockEAPHandler) Name() string                                     { return m.name }
func (m *mockEAPHandler) EAPType() uint8                                   { return m.eapType }
func (m *mockEAPHandler) CanHandle(ctx *eap.EAPContext) bool               { return true }
func (m *mockEAPHandler) HandleIdentity(ctx *eap.EAPContext) (bool, error) { return true, nil }
func (m *mockEAPHandler) HandleResponse(ctx *eap.EAPContext) (bool, error) { return true, nil }

// Tests for password validators

func TestRegisterPasswordValidator(t *testing.T) {
	ResetForTest()

	validator := &mockPasswordValidator{name: "test-validator"}
	RegisterPasswordValidator(validator)

	result, ok := GetPasswordValidator("test-validator")
	if !ok {
		t.Fatal("expected validator to be registered")
	}
	if result.Name() != "test-validator" {
		t.Errorf("expected name 'test-validator', got '%s'", result.Name())
	}
}

func TestGetPasswordValidator_NotFound(t *testing.T) {
	ResetForTest()

	_, ok := GetPasswordValidator("non-existent")
	if ok {
		t.Error("expected false for non-existent validator")
	}
}

func TestGetPasswordValidators(t *testing.T) {
	ResetForTest()

	validator1 := &mockPasswordValidator{name: "validator-1"}
	validator2 := &mockPasswordValidator{name: "validator-2"}

	RegisterPasswordValidator(validator1)
	RegisterPasswordValidator(validator2)

	validators := GetPasswordValidators()
	if len(validators) != 2 {
		t.Errorf("expected 2 validators, got %d", len(validators))
	}
}

// Tests for policy checkers

func TestRegisterPolicyChecker(t *testing.T) {
	ResetForTest()

	checker := &mockPolicyChecker{name: "test-checker", order: 10}
	RegisterPolicyChecker(checker)

	checkers := GetPolicyCheckers()
	if len(checkers) != 1 {
		t.Fatal("expected checker to be registered")
	}
	if checkers[0].Name() != "test-checker" {
		t.Errorf("expected name 'test-checker', got '%s'", checkers[0].Name())
	}
}

func TestGetPolicyCheckers_SortedByOrder(t *testing.T) {
	ResetForTest()

	checker1 := &mockPolicyChecker{name: "checker-1", order: 10}
	checker2 := &mockPolicyChecker{name: "checker-2", order: 5}
	checker3 := &mockPolicyChecker{name: "checker-3", order: 20}

	RegisterPolicyChecker(checker1)
	RegisterPolicyChecker(checker2)
	RegisterPolicyChecker(checker3)

	checkers := GetPolicyCheckers()
	if len(checkers) != 3 {
		t.Errorf("expected 3 checkers, got %d", len(checkers))
	}

	// Verify sorted by order
	if checkers[0].Order() != 5 {
		t.Errorf("expected first checker order 5, got %d", checkers[0].Order())
	}
	if checkers[1].Order() != 10 {
		t.Errorf("expected second checker order 10, got %d", checkers[1].Order())
	}
	if checkers[2].Order() != 20 {
		t.Errorf("expected third checker order 20, got %d", checkers[2].Order())
	}
}

// Tests for response enhancers

func TestRegisterResponseEnhancer(t *testing.T) {
	ResetForTest()

	enhancer := &mockResponseEnhancer{name: "test-enhancer"}
	RegisterResponseEnhancer(enhancer)

	enhancers := GetResponseEnhancers()
	if len(enhancers) != 1 {
		t.Fatal("expected enhancer to be registered")
	}
	if enhancers[0].Name() != "test-enhancer" {
		t.Errorf("expected name 'test-enhancer', got '%s'", enhancers[0].Name())
	}
}

func TestGetResponseEnhancers(t *testing.T) {
	ResetForTest()

	enhancer1 := &mockResponseEnhancer{name: "enhancer-1"}
	enhancer2 := &mockResponseEnhancer{name: "enhancer-2"}

	RegisterResponseEnhancer(enhancer1)
	RegisterResponseEnhancer(enhancer2)

	enhancers := GetResponseEnhancers()
	if len(enhancers) != 2 {
		t.Errorf("expected 2 enhancers, got %d", len(enhancers))
	}
}

// Tests for guards

func TestRegisterAuthGuard(t *testing.T) {
	ResetForTest()

	guard := &mockGuard{name: "test-guard"}
	RegisterAuthGuard(guard)

	guards := GetAuthGuards()
	if len(guards) != 1 {
		t.Fatal("expected guard to be registered")
	}
	if guards[0].Name() != "test-guard" {
		t.Errorf("expected name 'test-guard', got '%s'", guards[0].Name())
	}
}

func TestGetAuthGuards(t *testing.T) {
	ResetForTest()

	guard1 := &mockGuard{name: "guard-1"}
	guard2 := &mockGuard{name: "guard-2"}

	RegisterAuthGuard(guard1)
	RegisterAuthGuard(guard2)

	allGuards := GetAuthGuards()
	if len(allGuards) != 2 {
		t.Errorf("expected 2 guards, got %d", len(allGuards))
	}
}

// Tests for accounting handlers

func TestRegisterAccountingHandler(t *testing.T) {
	ResetForTest()

	handler := &mockAccountingHandler{name: "test-handler"}
	RegisterAccountingHandler(handler)

	handlers := GetAccountingHandlers()
	if len(handlers) != 1 {
		t.Fatal("expected handler to be registered")
	}
	if handlers[0].Name() != "test-handler" {
		t.Errorf("expected name 'test-handler', got '%s'", handlers[0].Name())
	}
}

func TestGetAccountingHandlers(t *testing.T) {
	ResetForTest()

	handler1 := &mockAccountingHandler{name: "handler-1"}
	handler2 := &mockAccountingHandler{name: "handler-2"}

	RegisterAccountingHandler(handler1)
	RegisterAccountingHandler(handler2)

	handlers := GetAccountingHandlers()
	if len(handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(handlers))
	}
}

// Tests for EAP handlers

func TestRegisterEAPHandler(t *testing.T) {
	ResetForTest()

	handler := &mockEAPHandler{name: "eap-md5", eapType: 4}
	RegisterEAPHandler(handler)

	result, ok := GetEAPHandler(4)
	if !ok {
		t.Fatal("expected handler to be registered")
	}
	if result.Name() != "eap-md5" {
		t.Errorf("expected name 'eap-md5', got '%s'", result.Name())
	}
	if result.EAPType() != 4 {
		t.Errorf("expected EAPType 4, got %d", result.EAPType())
	}
}

func TestGetEAPHandler_NotFound(t *testing.T) {
	ResetForTest()

	_, ok := GetEAPHandler(99)
	if ok {
		t.Error("expected false for non-existent EAP type")
	}
}

func TestGetAllEAPHandlers(t *testing.T) {
	ResetForTest()

	handler1 := &mockEAPHandler{name: "eap-md5", eapType: 4}
	handler2 := &mockEAPHandler{name: "eap-mschapv2", eapType: 26}

	RegisterEAPHandler(handler1)
	RegisterEAPHandler(handler2)

	handlers := GetAllEAPHandlers()
	if len(handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(handlers))
	}
}

// Tests for Registry methods

func TestRegistry_GetHandler(t *testing.T) {
	ResetForTest()

	handler := &mockEAPHandler{name: "eap-md5", eapType: 4}
	RegisterEAPHandler(handler)

	registry := GetGlobalRegistry()
	result, ok := registry.GetHandler(4)
	if !ok {
		t.Fatal("expected handler to be found via registry")
	}
	if result.Name() != "eap-md5" {
		t.Errorf("expected name 'eap-md5', got '%s'", result.Name())
	}
}

func TestRegistry_GetHandler_NotFound(t *testing.T) {
	ResetForTest()

	registry := GetGlobalRegistry()
	_, ok := registry.GetHandler(99)
	if ok {
		t.Error("expected false for non-existent handler")
	}
}

// Tests for overwriting existing registrations

func TestRegisterPasswordValidator_Overwrite(t *testing.T) {
	ResetForTest()

	validator1 := &mockPasswordValidator{name: "validator"}
	validator2 := &mockPasswordValidator{name: "validator"}

	RegisterPasswordValidator(validator1)
	RegisterPasswordValidator(validator2)

	// Should have replaced the first one
	validators := GetPasswordValidators()
	if len(validators) != 1 {
		t.Errorf("expected 1 validator (overwritten), got %d", len(validators))
	}
}

func TestRegisterEAPHandler_Overwrite(t *testing.T) {
	ResetForTest()

	handler1 := &mockEAPHandler{name: "handler-1", eapType: 4}
	handler2 := &mockEAPHandler{name: "handler-2", eapType: 4}

	RegisterEAPHandler(handler1)
	RegisterEAPHandler(handler2)

	// Should have replaced the first one
	handlers := GetAllEAPHandlers()
	if len(handlers) != 1 {
		t.Errorf("expected 1 handler (overwritten), got %d", len(handlers))
	}

	// Verify it's the second one
	result, _ := GetEAPHandler(4)
	if result.Name() != "handler-2" {
		t.Errorf("expected name 'handler-2', got '%s'", result.Name())
	}
}

// Test GetGlobalRegistry

func TestGetGlobalRegistry(t *testing.T) {
	registry := GetGlobalRegistry()
	if registry == nil {
		t.Fatal("expected global registry to be non-nil")
	}
}

// Test ResetForTest

func TestResetForTest(t *testing.T) {
	// Register some stuff
	validator := &mockPasswordValidator{name: "test"}
	RegisterPasswordValidator(validator)

	// Reset
	ResetForTest()

	// Should be empty
	validators := GetPasswordValidators()
	if len(validators) != 0 {
		t.Errorf("expected 0 validators after reset, got %d", len(validators))
	}
}

// Test that returned slices are copies

func TestGetPolicyCheckers_ReturnsCopy(t *testing.T) {
	ResetForTest()

	checker := &mockPolicyChecker{name: "checker", order: 1}
	RegisterPolicyChecker(checker)

	checkers1 := GetPolicyCheckers()
	checkers2 := GetPolicyCheckers()

	// Modify first slice
	if len(checkers1) > 0 {
		checkers1[0] = nil
	}

	// Second slice should be unaffected
	if checkers2[0] == nil {
		t.Error("expected returned slices to be independent copies")
	}
}

func TestGetResponseEnhancers_ReturnsCopy(t *testing.T) {
	ResetForTest()

	enhancer := &mockResponseEnhancer{name: "enhancer"}
	RegisterResponseEnhancer(enhancer)

	enhancers1 := GetResponseEnhancers()
	enhancers2 := GetResponseEnhancers()

	// Modify first slice
	if len(enhancers1) > 0 {
		enhancers1[0] = nil
	}

	// Second slice should be unaffected
	if enhancers2[0] == nil {
		t.Error("expected returned slices to be independent copies")
	}
}

func TestGetAuthGuards_ReturnsCopy(t *testing.T) {
	ResetForTest()

	guard := &mockGuard{name: "guard"}
	RegisterAuthGuard(guard)

	guards1 := GetAuthGuards()
	guards2 := GetAuthGuards()

	// Modify first slice
	if len(guards1) > 0 {
		guards1[0] = nil
	}

	// Second slice should be unaffected
	if guards2[0] == nil {
		t.Error("expected returned slices to be independent copies")
	}
}

func TestGetAccountingHandlers_ReturnsCopy(t *testing.T) {
	ResetForTest()

	handler := &mockAccountingHandler{name: "handler"}
	RegisterAccountingHandler(handler)

	handlers1 := GetAccountingHandlers()
	handlers2 := GetAccountingHandlers()

	// Modify first slice
	if len(handlers1) > 0 {
		handlers1[0] = nil
	}

	// Second slice should be unaffected
	if handlers2[0] == nil {
		t.Error("expected returned slices to be independent copies")
	}
}

func TestGetAllEAPHandlers_ReturnsCopy(t *testing.T) {
	ResetForTest()

	handler := &mockEAPHandler{name: "eap-md5", eapType: 4}
	RegisterEAPHandler(handler)

	handlers1 := GetAllEAPHandlers()
	handlers2 := GetAllEAPHandlers()

	// Delete from first map
	delete(handlers1, 4)

	// Second map should be unaffected
	if _, ok := handlers2[4]; !ok {
		t.Error("expected returned maps to be independent copies")
	}
}
