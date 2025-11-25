package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/app"
)

// TestValidateConfigSchemas tests the configuration validation logic
func TestValidateConfigSchemas(t *testing.T) {
	tests := []struct {
		name      string
		schemas   app.ConfigSchemasJSON
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid configuration with all types",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "radius.AuthPort",
						Type:        "int",
						Default:     "1812",
						Description: "Auth port",
						Min:         int64Ptr(1024),
						Max:         int64Ptr(65535),
					},
					{
						Key:         "radius.EapMethod",
						Type:        "string",
						Default:     "PEAP",
						Description: "EAP method",
						Enum:        []string{"PEAP", "TTLS"},
					},
					{
						Key:         "system.Debug",
						Type:        "bool",
						Default:     "false",
						Description: "Debug mode",
					},
					{
						Key:         "system.SessionTimeout",
						Type:        "duration",
						Default:     "30m",
						Description: "Session timeout",
					},
					{
						Key:         "advanced.CustomConfig",
						Type:        "json",
						Default:     "{}",
						Description: "Custom JSON config",
					},
				},
			},
			wantError: false,
		},
		{
			name: "empty key",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "",
						Type:        "string",
						Default:     "value",
						Description: "Empty key",
					},
				},
			},
			wantError: true,
			errorMsg:  "key cannot be empty",
		},
		{
			name: "empty type",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Key",
						Type:        "",
						Default:     "value",
						Description: "Empty type",
					},
				},
			},
			wantError: true,
			errorMsg:  "type cannot be empty",
		},
		{
			name: "empty default",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Key",
						Type:        "string",
						Default:     "",
						Description: "Empty default",
					},
				},
			},
			wantError: true,
			errorMsg:  "default cannot be empty",
		},
		{
			name: "duplicate keys",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Key",
						Type:        "string",
						Default:     "value1",
						Description: "First",
					},
					{
						Key:         "test.Key",
						Type:        "string",
						Default:     "value2",
						Description: "Duplicate",
					},
				},
			},
			wantError: true,
			errorMsg:  "duplicate key",
		},
		{
			name: "invalid type",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Key",
						Type:        "invalid_type",
						Default:     "value",
						Description: "Invalid type",
					},
				},
			},
			wantError: true,
			errorMsg:  "invalid type",
		},
		{
			name: "min greater than max",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Port",
						Type:        "int",
						Default:     "5000",
						Description: "Invalid range",
						Min:         int64Ptr(10000),
						Max:         int64Ptr(1000),
					},
				},
			},
			wantError: true,
			errorMsg:  "min value cannot be greater than max",
		},
		{
			name: "default not in enum",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Method",
						Type:        "string",
						Default:     "INVALID",
						Description: "Invalid enum default",
						Enum:        []string{"PEAP", "TTLS"},
					},
				},
			},
			wantError: true,
			errorMsg:  "default INVALID is not in enum",
		},
		{
			name: "valid enum with matching default",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Method",
						Type:        "string",
						Default:     "PEAP",
						Description: "Valid enum",
						Enum:        []string{"PEAP", "TTLS"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid int range",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Port",
						Type:        "int",
						Default:     "8080",
						Description: "Valid range",
						Min:         int64Ptr(1024),
						Max:         int64Ptr(65535),
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with test data
			tmpFile := createTempConfigFile(t, tt.schemas)
			defer func() { _ = os.Remove(tmpFile) }() //nolint:errcheck

			// Run validation
			err := validateConfigSchemas(tmpFile)

			if tt.wantError {
				if err == nil {
					t.Errorf("validateConfigSchemas() expected error but got nil")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateConfigSchemas() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateConfigSchemas() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateConfigSchemasFileErrors tests file handling errors
func TestValidateConfigSchemasFileErrors(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "non-existent file",
			filePath:  "/non/existent/path/config.json",
			wantError: true,
			errorMsg:  "failed to read file",
		},
		{
			name:      "invalid JSON",
			filePath:  createTempInvalidJSON(t),
			wantError: true,
			errorMsg:  "invalid JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filePath != "" && tt.filePath != "/non/existent/path/config.json" {
				defer func() { _ = os.Remove(tt.filePath) }() //nolint:errcheck
			}

			err := validateConfigSchemas(tt.filePath)

			if tt.wantError {
				if err == nil {
					t.Errorf("validateConfigSchemas() expected error but got nil")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateConfigSchemas() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateConfigSchemas() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestPrintConfigSummary tests the configuration summary display
func TestPrintConfigSummary(t *testing.T) {
	tests := []struct {
		name      string
		schemas   app.ConfigSchemasJSON
		wantError bool
	}{
		{
			name: "valid configuration with multiple categories",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "radius.AuthPort",
						Type:        "int",
						Default:     "1812",
						Description: "Authentication port",
						Min:         int64Ptr(1024),
						Max:         int64Ptr(65535),
					},
					{
						Key:         "radius.AcctPort",
						Type:        "int",
						Default:     "1813",
						Description: "Accounting port",
					},
					{
						Key:         "database.Host",
						Type:        "string",
						Default:     "localhost",
						Description: "Database host",
					},
					{
						Key:         "system.Debug",
						Type:        "bool",
						Default:     "false",
						Description: "Debug mode",
					},
				},
			},
			wantError: false,
		},
		{
			name: "configuration without category prefix",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "standalone",
						Type:        "string",
						Default:     "value",
						Description: "No category",
					},
				},
			},
			wantError: false,
		},
		{
			name: "empty configuration",
			schemas: app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempConfigFile(t, tt.schemas)
			defer func() { _ = os.Remove(tmpFile) }() //nolint:errcheck

			err := printConfigSummary(tmpFile)

			if tt.wantError {
				if err == nil {
					t.Errorf("printConfigSummary() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("printConfigSummary() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestPrintConfigSummaryFileErrors tests file handling in summary display
func TestPrintConfigSummaryFileErrors(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		wantError bool
	}{
		{
			name:      "non-existent file",
			filePath:  "/non/existent/path/config.json",
			wantError: true,
		},
		{
			name:      "invalid JSON",
			filePath:  createTempInvalidJSON(t),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filePath != "/non/existent/path/config.json" {
				defer func() { _ = os.Remove(tt.filePath) }() //nolint:errcheck
			}

			err := printConfigSummary(tt.filePath)

			if tt.wantError && err == nil {
				t.Errorf("printConfigSummary() expected error but got nil")
			}
		})
	}
}

// TestFindDotIndex tests the dot index finding helper
func TestFindDotIndex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "string with dot",
			input: "radius.AuthPort",
			want:  6,
		},
		{
			name:  "string with multiple dots",
			input: "system.database.host",
			want:  6, // Returns index of first dot
		},
		{
			name:  "string without dot",
			input: "standalone",
			want:  -1,
		},
		{
			name:  "empty string",
			input: "",
			want:  -1,
		},
		{
			name:  "dot at start",
			input: ".config",
			want:  0,
		},
		{
			name:  "dot at end",
			input: "config.",
			want:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDotIndex(tt.input)
			if got != tt.want {
				t.Errorf("findDotIndex(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestValidateAllTypes ensures all supported types pass validation
func TestValidateAllTypes(t *testing.T) {
	schemas := app.ConfigSchemasJSON{
		Schemas: []app.ConfigSchemaJSON{
			{
				Key:         "test.StringValue",
				Type:        "string",
				Default:     "default",
				Description: "String type",
			},
			{
				Key:         "test.IntValue",
				Type:        "int",
				Default:     "100",
				Description: "Int type",
			},
			{
				Key:         "test.BoolValue",
				Type:        "bool",
				Default:     "true",
				Description: "Bool type",
			},
			{
				Key:         "test.DurationValue",
				Type:        "duration",
				Default:     "5m",
				Description: "Duration type",
			},
			{
				Key:         "test.JSONValue",
				Type:        "json",
				Default:     "{}",
				Description: "JSON type",
			},
		},
	}

	tmpFile := createTempConfigFile(t, schemas)
	defer func() { _ = os.Remove(tmpFile) }() //nolint:errcheck

	err := validateConfigSchemas(tmpFile)
	if err != nil {
		t.Errorf("validateConfigSchemas() failed for all valid types: %v", err)
	}
}

// TestValidateIntegerConstraints tests integer min/max validation edge cases
func TestValidateIntegerConstraints(t *testing.T) {
	tests := []struct {
		name      string
		min       *int64
		max       *int64
		wantError bool
	}{
		{
			name:      "min and max equal",
			min:       int64Ptr(1000),
			max:       int64Ptr(1000),
			wantError: false,
		},
		{
			name:      "only min specified",
			min:       int64Ptr(1000),
			max:       nil,
			wantError: false,
		},
		{
			name:      "only max specified",
			min:       nil,
			max:       int64Ptr(1000),
			wantError: false,
		},
		{
			name:      "both nil",
			min:       nil,
			max:       nil,
			wantError: false,
		},
		{
			name:      "negative min and positive max",
			min:       int64Ptr(-100),
			max:       int64Ptr(100),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas := app.ConfigSchemasJSON{
				Schemas: []app.ConfigSchemaJSON{
					{
						Key:         "test.Value",
						Type:        "int",
						Default:     "0",
						Description: "Test",
						Min:         tt.min,
						Max:         tt.max,
					},
				},
			}

			tmpFile := createTempConfigFile(t, schemas)
			defer func() { _ = os.Remove(tmpFile) }() //nolint:errcheck

			err := validateConfigSchemas(tmpFile)

			if tt.wantError && err == nil {
				t.Errorf("validateConfigSchemas() expected error but got nil")
			} else if !tt.wantError && err != nil {
				t.Errorf("validateConfigSchemas() unexpected error = %v", err)
			}
		})
	}
}

// Helper functions

// int64Ptr returns a pointer to an int64 value
func int64Ptr(v int64) *int64 {
	return &v
}

// createTempConfigFile creates a temporary config file with the given schemas
func createTempConfigFile(t *testing.T, schemas app.ConfigSchemasJSON) string {
	t.Helper()

	data, err := json.Marshal(schemas)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "config_test.json")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tmpFile
}

// createTempInvalidJSON creates a temporary file with invalid JSON
func createTempInvalidJSON(t *testing.T) string {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(tmpFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tmpFile
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
