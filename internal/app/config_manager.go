package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConfigType 配置值类型
type ConfigType int

const (
	TypeString ConfigType = iota
	TypeInt
	TypeBool
	TypeDuration
	TypeJSON
)

// ConfigSchemaJSON JSON配置定义结构
type ConfigSchemaJSON struct {
	Key         string   `json:"key"`         // 配置键 "category.name"
	Type        string   `json:"type"`        // 值类型
	Default     string   `json:"default"`     // 默认值
	Enum        []string `json:"enum"`        // 枚举值限制
	Min         *int64   `json:"min"`         // 最小值
	Max         *int64   `json:"max"`         // 最大值
	Description string   `json:"description"` // 描述
}

// ConfigSchemasJSON 配置定义集合
type ConfigSchemasJSON struct {
	Schemas []ConfigSchemaJSON `json:"schemas"`
}

// ConfigSchema 配置项定义
type ConfigSchema struct {
	Key         string             // 配置键 "category.name"
	Type        ConfigType         // 值类型
	Default     string             // 默认值
	Enum        []string           // 枚举值限制
	Min         *int64             // 最小值
	Max         *int64             // 最大值
	Description string             // 描述
	Validator   func(string) error // 自定义验证器
}

//go:embed config_schemas.json
var configSchemasData []byte

// ConfigManager 精简版配置管理器 - 内存为主 + 数据库备份
type ConfigManager struct {
	app     *Application
	mu      sync.RWMutex
	configs map[string]string        // 配置存储: "category.name" -> "value"
	schemas map[string]*ConfigSchema // 配置定义
}

// NewConfigManager 创建配置管理器
func NewConfigManager(app *Application) *ConfigManager {
	cm := &ConfigManager{
		app:     app,
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// 1. 注册配置定义(包含默认值)
	cm.registerSchemas()

	// 2. 从数据库加载配置
	cm.loadFromDatabase()

	return cm
}

// registerSchemas 注册配置定义
func (cm *ConfigManager) registerSchemas() {
	// 从嵌入的 JSON 文件加载配置定义
	if err := cm.loadSchemasFromJSON(); err != nil {
		zap.L().Error("failed to load schemas from JSON, falling back to hardcoded", zap.Error(err))
		cm.registerHardcodedSchemas()
	}
}

// loadSchemasFromJSON 从嵌入的 JSON 文件加载配置定义
func (cm *ConfigManager) loadSchemasFromJSON() error {
	var schemasData ConfigSchemasJSON
	if err := json.Unmarshal(configSchemasData, &schemasData); err != nil {
		return fmt.Errorf("failed to unmarshal schemas JSON: %w", err)
	}

	for _, schemaJSON := range schemasData.Schemas {
		schema := &ConfigSchema{
			Key:         schemaJSON.Key,
			Type:        cm.parseConfigType(schemaJSON.Type),
			Default:     schemaJSON.Default,
			Enum:        schemaJSON.Enum,
			Min:         schemaJSON.Min,
			Max:         schemaJSON.Max,
			Description: schemaJSON.Description,
		}
		cm.register(schema)
	}

	zap.L().Info("config schemas loaded from JSON", zap.Int("count", len(schemasData.Schemas)))
	return nil
}

// parseConfigType 解析配置类型字符串
func (cm *ConfigManager) parseConfigType(typeStr string) ConfigType {
	switch typeStr {
	case "string":
		return TypeString
	case "int":
		return TypeInt
	case "bool":
		return TypeBool
	case "duration":
		return TypeDuration
	case "json":
		return TypeJSON
	default:
		return TypeString
	}
}

// registerHardcodedSchemas 注册硬编码的配置定义（兜底方案）
// 只在 JSON 配置加载失败时使用，包含最基本的配置
func (cm *ConfigManager) registerHardcodedSchemas() {
	zap.L().Warn("using hardcoded fallback schemas")

	// 只注册最关键的配置项作为兜底
	cm.register(&ConfigSchema{
		Key:         "radius.EapMethod",
		Type:        TypeString,
		Default:     "eap-md5",
		Enum:        []string{"eap-md5", "eap-mschapv2"},
		Description: "EAP authentication method",
	})

	cm.register(&ConfigSchema{
		Key:         "radius.IgnorePassword",
		Type:        TypeBool,
		Default:     "false",
		Description: "Ignore password check",
	})

	cm.register(&ConfigSchema{
		Key:         "radius.AccountingHistoryDays",
		Type:        TypeInt,
		Default:     "90",
		Min:         func() *int64 { v := int64(0); return &v }(),
		Description: "Accounting log retention days (0=disabled)",
	})

	zap.L().Info("hardcoded fallback schemas registered", zap.Int("count", 3))
}

// register 内部注册方法
func (cm *ConfigManager) register(schema *ConfigSchema) {
	cm.schemas[schema.Key] = schema
	// 设置默认值到内存
	cm.configs[schema.Key] = schema.Default
}

// loadFromDatabase 从数据库加载配置
func (cm *ConfigManager) loadFromDatabase() {
	rows, err := cm.app.gormDB.Raw("SELECT CONCAT(type, '.', name) as key, value FROM sys_config").Rows()
	if err != nil {
		zap.L().Warn("failed to load config from database", zap.Error(err))
		return
	}
	defer rows.Close()

	loaded := 0
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}

		// 只加载已注册的配置
		if _, exists := cm.schemas[key]; exists {
			cm.configs[key] = value
			loaded++
		}
	}

	zap.L().Info("config loaded from database", zap.Int("count", loaded))
}

// Get 获取配置值
func (cm *ConfigManager) Get(category, name string) string {
	key := category + "." + name

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if value, exists := cm.configs[key]; exists {
		return value
	}

	// 如果配置不存在，检查 schema 默认值
	if schema, exists := cm.schemas[key]; exists {
		return schema.Default
	}

	return ""
}

// Set 设置配置值
func (cm *ConfigManager) Set(category, name, value string) error {
	key := category + "." + name

	// 验证配置
	schema, exists := cm.schemas[key]
	if !exists {
		return fmt.Errorf("config %s not registered", key)
	}

	if err := cm.validate(schema, value); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 获取旧值
	oldValue := cm.Get(category, name)

	// 更新内存
	cm.mu.Lock()
	cm.configs[key] = value
	cm.mu.Unlock()

	// 更新数据库
	err := cm.app.gormDB.Exec(`
		INSERT INTO sys_config (type, name, value, updated_at) 
		VALUES (?, ?, ?, ?) 
		ON CONFLICT (type, name) 
		DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		category, name, value, time.Now()).Error

	if err != nil {
		// 回滚内存
		cm.mu.Lock()
		cm.configs[key] = oldValue
		cm.mu.Unlock()
		return fmt.Errorf("failed to save to database: %w", err)
	}

	zap.L().Info("config updated", zap.String("key", key), zap.String("new", value))
	return nil
}

// GetString 获取字符串配置
func (cm *ConfigManager) GetString(category, name string) string {
	return cm.Get(category, name)
}

// GetInt 获取整数配置（兼容原有接口）
func (cm *ConfigManager) GetInt(category, name string) int64 {
	value := cm.Get(category, name)
	if value == "" {
		return 0
	}

	// 简单的整数转换
	var result int64
	fmt.Sscanf(value, "%d", &result)
	return result
}

// GetInt64 获取int64配置（兼容性方法，与GetInt相同）
func (cm *ConfigManager) GetInt64(category, name string) int64 {
	return cm.GetInt(category, name)
} // GetBool 获取布尔配置
func (cm *ConfigManager) GetBool(category, name string) bool {
	value := cm.Get(category, name)
	return value == "true" || value == "enabled" || value == "1"
}

// ReloadAll 重载所有配置
func (cm *ConfigManager) ReloadAll() {
	cm.loadFromDatabase()
	zap.L().Info("all configs reloaded")
}

// GetAllSchemas 获取所有配置定义
func (cm *ConfigManager) GetAllSchemas() map[string]*ConfigSchema {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*ConfigSchema, len(cm.schemas))
	for k, v := range cm.schemas {
		result[k] = v
	}
	return result
}

// validate 验证配置值
func (cm *ConfigManager) validate(schema *ConfigSchema, value string) error {
	// 枚举检查
	if len(schema.Enum) > 0 {
		valid := false
		for _, enum := range schema.Enum {
			if value == enum {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("value must be one of %v", schema.Enum)
		}
	}

	// 类型检查
	switch schema.Type {
	case TypeInt:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %s", value)
		}
		if schema.Min != nil && intVal < *schema.Min {
			return fmt.Errorf("value must be >= %d", *schema.Min)
		}
		if schema.Max != nil && intVal > *schema.Max {
			return fmt.Errorf("value must be <= %d", *schema.Max)
		}

	case TypeBool:
		if value != "true" && value != "false" &&
			value != "enabled" && value != "disabled" &&
			value != "1" && value != "0" {
			return fmt.Errorf("invalid boolean: %s", value)
		}

	case TypeDuration:
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid duration: %s", value)
		}

	case TypeJSON:
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("invalid JSON format")
		}
	}

	// 自定义验证
	if schema.Validator != nil {
		return schema.Validator(value)
	}

	return nil
}
