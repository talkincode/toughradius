package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
)

func TestInitPlugins_WithNilDependencies(t *testing.T) {
	registry.ResetForTest()
	defer registry.ResetForTest()

	assert.NotPanics(t, func() {
		InitPlugins(nil, nil, nil)
	})

	validators := registry.GetPasswordValidators()
	assert.NotEmpty(t, validators, "password validators should be registered")

	checkers := registry.GetPolicyCheckers()
	assert.NotEmpty(t, checkers, "policy checkers should be registered")

	enhancers := registry.GetResponseEnhancers()
	assert.NotEmpty(t, enhancers, "response enhancers should be registered")

	guards := registry.GetAuthGuards()
	assert.NotEmpty(t, guards, "auth guards should be registered")

	eapHandlers := registry.GetAllEAPHandlers()
	assert.NotEmpty(t, eapHandlers, "EAP handlers should be registered")
}

func TestInitPlugins_PasswordValidators(t *testing.T) {
	registry.ResetForTest()
	defer registry.ResetForTest()

	InitPlugins(nil, nil, nil)

	// Actual names returned by Name() method: "pap", "chap", "mschap"
	expectedValidators := []string{"pap", "chap", "mschap"}
	validators := registry.GetPasswordValidators()

	validatorNames := make(map[string]bool)
	for _, v := range validators {
		validatorNames[v.Name()] = true
	}

	for _, name := range expectedValidators {
		assert.True(t, validatorNames[name], "validator %s should be registered", name)
	}
}

func TestInitPlugins_PolicyCheckers(t *testing.T) {
	registry.ResetForTest()
	defer registry.ResetForTest()

	InitPlugins(nil, nil, nil)

	checkers := registry.GetPolicyCheckers()
	assert.GreaterOrEqual(t, len(checkers), 4)

	checkerNames := make(map[string]bool)
	for _, c := range checkers {
		checkerNames[c.Name()] = true
	}

	// Actual names returned by Name() method: "status", "expire", "mac_bind", "vlan_bind"
	expectedCheckers := []string{"status", "expire", "mac_bind", "vlan_bind"}
	for _, name := range expectedCheckers {
		assert.True(t, checkerNames[name], "checker %s should be registered", name)
	}
}

func TestInitPlugins_ResponseEnhancers(t *testing.T) {
	registry.ResetForTest()
	defer registry.ResetForTest()

	InitPlugins(nil, nil, nil)

	enhancers := registry.GetResponseEnhancers()
	assert.GreaterOrEqual(t, len(enhancers), 5)
}

func TestInitPlugins_EAPHandlers(t *testing.T) {
	registry.ResetForTest()
	defer registry.ResetForTest()

	InitPlugins(nil, nil, nil)

	eapHandlers := registry.GetAllEAPHandlers()

	assert.NotNil(t, eapHandlers[4], "EAP-MD5 should be registered")
	assert.NotNil(t, eapHandlers[5], "EAP-OTP should be registered")
	assert.NotNil(t, eapHandlers[26], "EAP-MSCHAPv2 should be registered")
}

func TestInitPlugins_NoAccountingHandlersWithNilRepos(t *testing.T) {
	registry.ResetForTest()
	defer registry.ResetForTest()

	InitPlugins(nil, nil, nil)

	handlers := registry.GetAccountingHandlers()
	assert.Empty(t, handlers)
}
