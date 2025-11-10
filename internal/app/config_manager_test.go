package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/config"
)

// 测试基础功能
func TestConfigManager_Basic(t *testing.T) {
	app := &Application{
		appConfig: &config.AppConfig{},
	}

	cm := &ConfigManager{
		app:     app,
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// 测试获取
	cm.configs["test.key1"] = "value1"
	value := cm.Get("test", "key1")
	assert.Equal(t, "value1", value)

	// 测试整数
	cm.configs["test.intkey"] = "123"
	intVal := cm.GetInt("test", "intkey")
	assert.Equal(t, int64(123), intVal)

	// 测试布尔
	cm.configs["test.boolkey"] = "true"
	boolVal := cm.GetBool("test", "boolkey")
	assert.True(t, boolVal)
}

// 测试配置注册和默认值
func TestConfigManager_SchemaDefaults(t *testing.T) {
	cm := &ConfigManager{
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// 注册配置
	schema := &ConfigSchema{
		Key:         "test.sample",
		Type:        TypeString,
		Default:     "default_value",
		Description: "测试配置",
	}
	cm.register(schema)

	// 测试默认值
	value := cm.Get("test", "sample")
	assert.Equal(t, "default_value", value)

	// 测试获取所有 schemas
	schemas := cm.GetAllSchemas()
	assert.Contains(t, schemas, "test.sample")
	assert.Equal(t, "default_value", schemas["test.sample"].Default)
}

// 测试配置校验
func TestConfigManager_Validation(t *testing.T) {
	cm := &ConfigManager{
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// 注册带枚举的配置
	enumSchema := &ConfigSchema{
		Key:  "test.enum",
		Type: TypeString,
		Enum: []string{"option1", "option2", "option3"},
	}
	cm.schemas["test.enum"] = enumSchema

	// 测试有效枚举值
	err := cm.validate(enumSchema, "option1")
	assert.NoError(t, err)

	// 测试无效枚举值
	err = cm.validate(enumSchema, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")

	// 注册整数范围配置
	intSchema := &ConfigSchema{
		Key:  "test.intrange",
		Type: TypeInt,
		Min:  func() *int64 { v := int64(1); return &v }(),
		Max:  func() *int64 { v := int64(100); return &v }(),
	}
	cm.schemas["test.intrange"] = intSchema

	// 测试有效整数
	err = cm.validate(intSchema, "50")
	assert.NoError(t, err)

	// 测试超出范围的整数
	err = cm.validate(intSchema, "200")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be <=")

	err = cm.validate(intSchema, "0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be >=")

	// 测试无效整数格式
	err = cm.validate(intSchema, "not_number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid integer")
}

// 测试布尔值校验
func TestConfigManager_BoolValidation(t *testing.T) {
	cm := &ConfigManager{
		schemas: make(map[string]*ConfigSchema),
	}

	boolSchema := &ConfigSchema{
		Key:  "test.bool",
		Type: TypeBool,
	}

	// 测试有效布尔值
	validBools := []string{"true", "false", "enabled", "disabled", "1", "0"}
	for _, val := range validBools {
		err := cm.validate(boolSchema, val)
		assert.NoError(t, err, "应该接受布尔值: %s", val)
	}

	// 测试无效布尔值
	err := cm.validate(boolSchema, "maybe")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid boolean")
}

// 测试 JSON 配置加载
func TestConfigManagerJSON(t *testing.T) {
	cm := &ConfigManager{
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// 测试从 JSON 加载配置定义
	err := cm.loadSchemasFromJSON()
	assert.NoError(t, err, "应该能够从 JSON 加载配置定义")

	// 验证配置定义是否正确加载
	assert.Greater(t, len(cm.schemas), 0, "应该加载了配置定义")

	// 测试特定的配置项
	radiusEapSchema, exists := cm.schemas["radius.EapMethod"]
	assert.True(t, exists, "radius.EapMethod 配置应该存在")
	assert.Equal(t, TypeString, radiusEapSchema.Type, "radius.EapMethod 应该是字符串类型")
	assert.Equal(t, "eap-md5", radiusEapSchema.Default, "radius.EapMethod 默认值应该是 eap-md5")
	assert.Contains(t, radiusEapSchema.Enum, "eap-md5", "枚举值应该包含 eap-md5")
	assert.Contains(t, radiusEapSchema.Enum, "eap-mschapv2", "枚举值应该包含 eap-mschapv2")

	// 测试整数配置
	interimSchema, exists := cm.schemas["radius.AcctInterimInterval"]
	assert.True(t, exists, "radius.AcctInterimInterval 配置应该存在")
	assert.Equal(t, TypeInt, interimSchema.Type, "radius.AcctInterimInterval 应该是整数类型")
	assert.NotNil(t, interimSchema.Min, "应该设置了最小值")
	assert.NotNil(t, interimSchema.Max, "应该设置了最大值")
	assert.Equal(t, int64(60), *interimSchema.Min, "最小值应该是60")
	assert.Equal(t, int64(3600), *interimSchema.Max, "最大值应该是3600")

	// 测试布尔配置
	ignorePassSchema, exists := cm.schemas["radius.IgnorePassword"]
	assert.True(t, exists, "radius.IgnorePassword 配置应该存在")
	assert.Equal(t, TypeBool, ignorePassSchema.Type, "radius.IgnorePassword 应该是布尔类型")
	assert.Equal(t, "false", ignorePassSchema.Default, "默认值应该是 false")
}

// 测试配置类型解析
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
		{"unknown", TypeString}, // 未知类型应该默认为字符串
	}

	for _, test := range tests {
		result := cm.parseConfigType(test.input)
		assert.Equal(t, test.expected, result, "解析 %s 应该得到 %v", test.input, test.expected)
	}
}
