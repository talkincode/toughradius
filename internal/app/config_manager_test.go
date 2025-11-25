package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
)

// Test basic functionality
func TestConfigManager_Basic(t *testing.T) {
	app := &Application{
		appConfig: &config.AppConfig{},
	}

	cm := &ConfigManager{
		app:     app,
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// Testget
	cm.configs["test.key1"] = "value1"
	value := cm.Get("test", "key1")
	assert.Equal(t, "value1", value)

	// Test integers
	cm.configs["test.intkey"] = "123"
	intVal := cm.GetInt("test", "intkey")
	assert.Equal(t, int64(123), intVal)

	// Test booleans
	cm.configs["test.boolkey"] = "true"
	boolVal := cm.GetBool("test", "boolkey")
	assert.True(t, boolVal)
}

// Test configuration registration and defaults
func TestConfigManager_SchemaDefaults(t *testing.T) {
	cm := &ConfigManager{
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// Register configuration
	schema := &ConfigSchema{
		Key:         "test.sample",
		Type:        TypeString,
		Default:     "default_value",
		Description: "Sample configuration",
	}
	cm.register(schema)

	// Test default values
	value := cm.Get("test", "sample")
	assert.Equal(t, "default_value", value)

	// Test retrieving all schemas
	schemas := cm.GetAllSchemas()
	assert.Contains(t, schemas, "test.sample")
	assert.Equal(t, "default_value", schemas["test.sample"].Default)
}

// Test configuration validation
func TestConfigManager_Validation(t *testing.T) {
	cm := &ConfigManager{
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// Register configuration with enums
	enumSchema := &ConfigSchema{
		Key:  "test.enum",
		Type: TypeString,
		Enum: []string{"option1", "option2", "option3"},
	}
	cm.schemas["test.enum"] = enumSchema

	// Test valid enum value
	err := cm.validate(enumSchema, "option1")
	assert.NoError(t, err)

	// Test invalid enum value
	err = cm.validate(enumSchema, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")

	// Register integer range configuration
	intSchema := &ConfigSchema{
		Key:  "test.intrange",
		Type: TypeInt,
		Min:  func() *int64 { v := int64(1); return &v }(),
		Max:  func() *int64 { v := int64(100); return &v }(),
	}
	cm.schemas["test.intrange"] = intSchema

	// Test valid integer
	err = cm.validate(intSchema, "50")
	assert.NoError(t, err)

	// Test out-of-range integer
	err = cm.validate(intSchema, "200")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be <=")

	err = cm.validate(intSchema, "0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be >=")

	// Test invalid integer format
	err = cm.validate(intSchema, "not_number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid integer")
}

// Test boolean validation
func TestConfigManager_BoolValidation(t *testing.T) {
	cm := &ConfigManager{
		schemas: make(map[string]*ConfigSchema),
	}

	boolSchema := &ConfigSchema{
		Key:  "test.bool",
		Type: TypeBool,
	}

	// Test valid boolean values
	validBools := []string{"true", "false", "enabled", "disabled", "1", "0"}
	for _, val := range validBools {
		err := cm.validate(boolSchema, val)
		assert.NoError(t, err, "Should accept boolean value: %s", val)
	}

	// Test invalid boolean values
	err := cm.validate(boolSchema, "maybe")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid boolean")
}

// Test JSON configuration loading
func TestConfigManagerJSON(t *testing.T) {
	cm := &ConfigManager{
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// Test loading configuration definitions from JSON
	err := cm.loadSchemasFromJSON()
	assert.NoError(t, err, "Should load configuration definitions from JSON")

	// Validate that configuration definitions loaded successfully
	assert.Greater(t, len(cm.schemas), 0, "Configuration definitions should be loaded")

	// Test specific configuration entries
	radiusEapSchema, exists := cm.schemas["radius.EapMethod"]
	assert.True(t, exists, "radius.EapMethod configuration should exist")
	assert.Equal(t, TypeString, radiusEapSchema.Type, "radius.EapMethod should be string type")
	assert.Equal(t, "eap-md5", radiusEapSchema.Default, "radius.EapMethod default should be eap-md5")
	assert.Contains(t, radiusEapSchema.Enum, "eap-md5", "Enum values should include eap-md5")
	assert.Contains(t, radiusEapSchema.Enum, "eap-mschapv2", "Enum values should include eap-mschapv2")
	assert.Equal(t, "EAP Method", radiusEapSchema.Title)
	assert.Equal(t, "config.radius.eap_method.title", radiusEapSchema.TitleI18n)
	assert.Equal(t, "config.radius.eap_method.description", radiusEapSchema.DescI18n)

	// Ensure EAP handler enable list exists
	enabledHandlerSchema, exists := cm.schemas["radius.EapEnabledHandlers"]
	assert.True(t, exists, "radius.EapEnabledHandlers configuration should exist")
	assert.Equal(t, TypeString, enabledHandlerSchema.Type, "radius.EapEnabledHandlers should be string type")
	assert.Equal(t, "*", enabledHandlerSchema.Default, "radius.EapEnabledHandlers default should allow all handlers")

	// Test integer configuration
	interimSchema, exists := cm.schemas["radius.AcctInterimInterval"]
	assert.True(t, exists, "radius.AcctInterimInterval configuration should exist")
	assert.Equal(t, TypeInt, interimSchema.Type, "radius.AcctInterimInterval should be integer type")
	assert.NotNil(t, interimSchema.Min, "Minimum value should be set")
	assert.NotNil(t, interimSchema.Max, "Maximum value should be set")
	assert.Equal(t, int64(60), *interimSchema.Min, "Minimum value should be 60")
	assert.Equal(t, int64(3600), *interimSchema.Max, "Maximum value should be 3600")

	// Test boolean configuration
	ignorePassSchema, exists := cm.schemas["radius.IgnorePassword"]
	assert.True(t, exists, "radius.IgnorePassword configuration should exist")
	assert.Equal(t, TypeBool, ignorePassSchema.Type, "radius.IgnorePassword should be boolean type")
	assert.Equal(t, "false", ignorePassSchema.Default, "Default should be false")
	assert.Equal(t, "config.radius.ignore_password.title", ignorePassSchema.TitleI18n)

	maxRejectSchema, exists := cm.schemas["radius.RejectDelayMaxRejects"]
	assert.True(t, exists, "radius.RejectDelayMaxRejects configuration should exist")
	assert.Equal(t, TypeInt, maxRejectSchema.Type)
	assert.Equal(t, "7", maxRejectSchema.Default)
	if assert.NotNil(t, maxRejectSchema.Min) {
		assert.Equal(t, int64(1), *maxRejectSchema.Min)
	}
	if assert.NotNil(t, maxRejectSchema.Max) {
		assert.Equal(t, int64(1000), *maxRejectSchema.Max)
	}

	windowSchema, exists := cm.schemas["radius.RejectDelayWindowSeconds"]
	assert.True(t, exists, "radius.RejectDelayWindowSeconds configuration should exist")
	assert.Equal(t, TypeInt, windowSchema.Type)
	assert.Equal(t, "10", windowSchema.Default)
	if assert.NotNil(t, windowSchema.Min) {
		assert.Equal(t, int64(1), *windowSchema.Min)
	}
	if assert.NotNil(t, windowSchema.Max) {
		assert.Equal(t, int64(3600), *windowSchema.Max)
	}
}

// Test configuration type parsing
func TestParseConfigType(t *testing.T) {
	cm := &ConfigManager{}

	tests := []struct {
		input    string
		expected ConfigType
	}{
		{"string", TypeString},
		{"int", TypeInt},
		{"bool", TypeBool},
		{"duration", TypeDuration},
		{"json", TypeJSON},
		{"unknown", TypeString}, // Unknown types should default to string
	}

	for _, test := range tests {
		result := cm.parseConfigType(test.input)
		assert.Equal(t, test.expected, result, "Parsing %s should yield %v", test.input, test.expected)
	}
}
