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

// ConfigType represents the configuration value type
type ConfigType int

const (
	TypeString ConfigType = iota
	TypeInt
	TypeBool
	TypeDuration
	TypeJSON
)

// ConfigSchemaJSON defines the JSON structure for configuration definitions
type ConfigSchemaJSON struct {
	Key         string   `json:"key"`         // Configuration key "category.name"
	Type        string   `json:"type"`        // Value type
	Default     string   `json:"default"`     // Default value
	Enum        []string `json:"enum"`        // Enum constraints
	Min         *int64   `json:"min"`         // Minimum value
	Max         *int64   `json:"max"`         // Maximum value
	Description string   `json:"description"` // Description
}

// ConfigSchemasJSON groups configuration definitions
type ConfigSchemasJSON struct {
	Schemas []ConfigSchemaJSON `json:"schemas"`
}

// ConfigSchema defines a configuration entry
type ConfigSchema struct {
	Key         string             // Configuration key "category.name"
	Type        ConfigType         // Value type
	Default     string             // Default value
	Enum        []string           // Enum constraints
	Min         *int64             // Minimum value
	Max         *int64             // Maximum value
	Description string             // Description
	Validator   func(string) error // Custom validator
}

//go:embed config_schemas.json
var configSchemasData []byte

// ConfigManager is a lightweight configuration manager (memory-first with database backup)
type ConfigManager struct {
	app     *Application
	mu      sync.RWMutex
	configs map[string]string        // configuration storage: "category.name" -> "value"
	schemas map[string]*ConfigSchema // configuration definitions
}

// NewConfigManager creates a configuration manager
func NewConfigManager(app *Application) *ConfigManager {
	cm := &ConfigManager{
		app:     app,
		configs: make(map[string]string),
		schemas: make(map[string]*ConfigSchema),
	}

	// 1. Register configuration definitions (including defaults)
	cm.registerSchemas()

	// 2. Load configuration from the database
	cm.loadFromDatabase()

	return cm
}

// registerSchemas registers configuration definitions
func (cm *ConfigManager) registerSchemas() {
	// Load configuration definitions from the embedded JSON file
	if err := cm.loadSchemasFromJSON(); err != nil {
		zap.L().Error("failed to load schemas from JSON, falling back to hardcoded", zap.Error(err))
		cm.registerHardcodedSchemas()
	}
}

// loadSchemasFromJSON loads configuration definitions from the embedded JSON file
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

// parseConfigType parses a configuration type string
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

// registerHardcodedSchemas registers hardcoded configuration definitions (fallback)
// Only used when JSON configuration loading fails, includes the most essential configurations
func (cm *ConfigManager) registerHardcodedSchemas() {
	zap.L().Warn("using hardcoded fallback schemas")

		// Only register the most critical configuration entries as a fallback
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

// register is an internal registration helper
func (cm *ConfigManager) register(schema *ConfigSchema) {
	cm.schemas[schema.Key] = schema
		// Set default values into memory
	cm.configs[schema.Key] = schema.Default
}

// loadFromDatabase loads configuration from the database
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

		// Only load registered configurations
		if _, exists := cm.schemas[key]; exists {
			cm.configs[key] = value
			loaded++
		}
	}

	zap.L().Info("config loaded from database", zap.Int("count", loaded))
}

// Get retrieves a configuration value
func (cm *ConfigManager) Get(category, name string) string {
	key := category + "." + name

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if value, exists := cm.configs[key]; exists {
		return value
	}

	// e.g., if the configuration is missing, use the schema default value
	if schema, exists := cm.schemas[key]; exists {
		return schema.Default
	}

	return ""
}

// Set updates a configuration value
func (cm *ConfigManager) Set(category, name, value string) error {
	key := category + "." + name

	// Validateconfiguration
	schema, exists := cm.schemas[key]
	if !exists {
		return fmt.Errorf("config %s not registered", key)
	}

	if err := cm.validate(schema, value); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

		// Get the previous value
	oldValue := cm.Get(category, name)

		// Update the in-memory cache
	cm.mu.Lock()
	cm.configs[key] = value
	cm.mu.Unlock()

		// Update the database
	err := cm.app.gormDB.Exec(`
		INSERT INTO sys_config (type, name, value, updated_at) 
		VALUES (?, ?, ?, ?) 
		ON CONFLICT (type, name) 
		DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		category, name, value, time.Now()).Error

	if err != nil {
		// Rollback the in-memory cache
		cm.mu.Lock()
		cm.configs[key] = oldValue
		cm.mu.Unlock()
		return fmt.Errorf("failed to save to database: %w", err)
	}

	zap.L().Info("config updated", zap.String("key", key), zap.String("new", value))
	return nil
}

// GetString retrieves a string configuration
func (cm *ConfigManager) GetString(category, name string) string {
	return cm.Get(category, name)
}

// GetInt retrieves an integer configuration (legacy interface)
func (cm *ConfigManager) GetInt(category, name string) int64 {
	value := cm.Get(category, name)
	if value == "" {
		return 0
	}

	// Simple integer conversion
	var result int64
	fmt.Sscanf(value, "%d", &result)
	return result
}

// GetInt64 retrieves an int64 configuration (same as GetInt)
func (cm *ConfigManager) GetInt64(category, name string) int64 {
	return cm.GetInt(category, name)
}

// GetBool retrieves a boolean configuration
func (cm *ConfigManager) GetBool(category, name string) bool {
	value := cm.Get(category, name)
	return value == "true" || value == "enabled" || value == "1"
}

// ReloadAll reloads all configurations
func (cm *ConfigManager) ReloadAll() {
	cm.loadFromDatabase()
	zap.L().Info("all configs reloaded")
}

// GetAllSchemas returns all configuration definitions
func (cm *ConfigManager) GetAllSchemas() map[string]*ConfigSchema {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*ConfigSchema, len(cm.schemas))
	for k, v := range cm.schemas {
		result[k] = v
	}
	return result
}

// validate validates configuration values
func (cm *ConfigManager) validate(schema *ConfigSchema, value string) error {
	// Check enum values
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

	// Check types
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

	// Custom validation
	if schema.Validator != nil {
		return schema.Validator(value)
	}

	return nil
}
