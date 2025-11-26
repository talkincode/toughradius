// Package main provides a command-line tool for validating and inspecting
// ToughRADIUS configuration schema files.
//
// This tool ensures configuration schemas (typically config_schemas.json) follow
// the required format and constraints before being loaded into the application.
// It performs structural validation, type checking, and constraint verification.
//
// Usage:
//
//	go run config_tool.go validate <config_schemas.json>  # Validate configuration file
//	go run config_tool.go summary <config_schemas.json>   # Display configuration summary
//
// The tool validates:
//   - JSON structure and syntax
//   - Required fields (key, type, default)
//   - Type validity (string, int, bool, duration, json)
//   - Duplicate key detection
//   - Integer range constraints (min/max)
//   - Enum value consistency
//
// Exit codes:
//   - 0: Validation successful
//   - 1: Validation failed or invalid usage
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/talkincode/toughradius/v9/internal/app"
)

// validateConfigSchemas validates the formatting and content of the configuration JSON file.
// It performs comprehensive checks on the configuration schema to ensure it meets
// all structural and semantic requirements before being used in production.
//
// Validation checks include:
//   - JSON syntax and structure
//   - Required fields presence (key, type, default)
//   - Key uniqueness across all configuration entries
//   - Type validity against supported types
//   - Integer range constraints (min <= max)
//   - Enum default value membership
//
// Parameters:
//   - filePath: Absolute or relative path to the configuration schema JSON file
//
// Returns:
//   - error: nil on successful validation, otherwise detailed error with line number and issue
//
// Supported configuration types:
//   - "string": Text values
//   - "int": Integer values with optional min/max constraints
//   - "bool": Boolean true/false values
//   - "duration": Time duration strings (e.g., "30s", "5m")
//   - "json": JSON-encoded complex values
//
// Example:
//
//	if err := validateConfigSchemas("config/config_schemas.json"); err != nil {
//	   log.Fatalf("Validation failed: %v", err)
//	}
func validateConfigSchemas(filePath string) error {
	data, err := os.ReadFile(filePath) //nolint:gosec // G304: path is user-specified config file
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var schemasData app.ConfigSchemasJSON
	if err := json.Unmarshal(data, &schemasData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Validate each configuration entry
	keyMap := make(map[string]bool)
	for i, schema := range schemasData.Schemas {
		// Check required fields
		if schema.Key == "" {
			return fmt.Errorf("config item %d: key cannot be empty", i)
		}
		if schema.Type == "" {
			return fmt.Errorf("config item %d (%s): type cannot be empty", i, schema.Key)
		}
		if schema.Default == "" {
			return fmt.Errorf("config item %d (%s): default cannot be empty", i, schema.Key)
		}

		// Check for duplicate keys
		if keyMap[schema.Key] {
			return fmt.Errorf("config item %d (%s): duplicate key", i, schema.Key)
		}
		keyMap[schema.Key] = true

		// Validate the type
		validTypes := []string{"string", "int", "bool", "duration", "json"}
		typeValid := false
		for _, validType := range validTypes {
			if schema.Type == validType {
				typeValid = true
				break
			}
		}
		if !typeValid {
			return fmt.Errorf("config item %d (%s): invalid type %s, supported: %v", i, schema.Key, schema.Type, validTypes)
		}

		// Validate integer ranges
		if schema.Type == "int" {
			if schema.Min != nil && schema.Max != nil && *schema.Min > *schema.Max {
				return fmt.Errorf("config item %d (%s): min value cannot be greater than max", i, schema.Key)
			}
		}

		// Validate enumeration values
		if len(schema.Enum) > 0 {
			defaultInEnum := false
			for _, enumVal := range schema.Enum {
				if enumVal == schema.Default {
					defaultInEnum = true
					break
				}
			}
			if !defaultInEnum {
				return fmt.Errorf("config item %d (%s): default %s is not in enum %v", i, schema.Key, schema.Default, schema.Enum)
			}
		}
	}

	fmt.Printf("✓ Configuration validation succeeded! %d entries found\n", len(schemasData.Schemas))
	return nil
}

// printConfigSummary prints a formatted, categorized summary of configuration schema entries.
// It reads the configuration file, validates its structure, and displays entries grouped
// by category (derived from the key prefix before the first dot).
//
// The summary includes:
//   - Total entry count
//   - Configuration entries grouped by category (e.g., "radius", "database", "system")
//   - For each entry: key, type, description
//   - Additional constraints: enum values, integer ranges (min/max)
//
// Parameters:
//   - filePath: Path to the configuration schema JSON file
//
// Returns:
//   - error: File read error or JSON parsing error (nil on success)
//
// Output format:
//
//	Configuration summary (42 entries):
//	━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//	🔧 radius (15 entries):
//	  • radius.AuthPort              [int]      Authentication port number
//	    └─ Range: min=1024 max=65535
//	  • radius.EapMethod             [string]   EAP authentication method
//	    └─ Enum: [PEAP TTLS]
//
// Example:
//
//	if err := printConfigSummary("config/config_schemas.json"); err != nil {
//	   return fmt.Errorf("failed to display summary: %w", err)
//	}
func printConfigSummary(filePath string) error {
	data, err := os.ReadFile(filePath) //nolint:gosec // G304: path is user-specified config file
	if err != nil {
		return err
	}

	var schemasData app.ConfigSchemasJSON
	if err := json.Unmarshal(data, &schemasData); err != nil {
		return err
	}

	fmt.Printf("\nConfiguration summary (%d entries):\n", len(schemasData.Schemas))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Group configuration entries by category (extracted from key prefix)
	categoryMap := make(map[string][]app.ConfigSchemaJSON)
	for _, schema := range schemasData.Schemas {
		var category string
		if idx := findDotIndex(schema.Key); idx != -1 {
			category = schema.Key[:idx]
		} else {
			category = "Other"
		}
		categoryMap[category] = append(categoryMap[category], schema)
	}

	for category, schemas := range categoryMap {
		fmt.Printf("\n🔧 %s (%d entries):\n", category, len(schemas))
		for _, schema := range schemas {
			fmt.Printf("  • %-30s [%s] %s\n", schema.Key, schema.Type, schema.Description)
			if len(schema.Enum) > 0 {
				fmt.Printf("    └─ Enum: %v\n", schema.Enum)
			}
			if schema.Min != nil || schema.Max != nil {
				rangeInfo := "    └─ Range: "
				if schema.Min != nil {
					rangeInfo += fmt.Sprintf("min=%d ", *schema.Min)
				}
				if schema.Max != nil {
					rangeInfo += fmt.Sprintf("max=%d", *schema.Max)
				}
				fmt.Println(rangeInfo)
			}
		}
	}

	return nil
}

// findDotIndex returns the index of the first dot character in the string.
// This helper function is used to extract category prefixes from configuration keys
// (e.g., "radius.AuthPort" → category is "radius").
//
// Parameters:
//   - s: String to search for dot character
//
// Returns:
//   - int: Zero-based index of the first dot, or -1 if no dot is found
//
// Example:
//
//	idx := findDotIndex("radius.AuthPort")  // Returns 6
//	idx := findDotIndex("standalone")        // Returns -1
func findDotIndex(s string) int {
	for i, r := range s {
		if r == '.' {
			return i
		}
	}
	return -1
}

// main is the entry point for the configuration tool CLI.
// It parses command-line arguments and dispatches to the appropriate validation
// or summary display function.
//
// Command-line usage:
//
//	config_tool validate <file>  - Validate configuration schema file
//	config_tool summary <file>   - Display categorized configuration summary
//
// Exit codes:
//   - 0: Operation completed successfully
//   - 1: Invalid arguments, validation failure, or file read error
//
// The tool requires exactly 2 arguments: command and file path.
// Both 'validate' and 'summary' commands perform validation first to ensure
// the configuration file is well-formed before displaying output.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run config_tool.go validate <config_schemas.json>  # validate the configuration file")
		fmt.Println("  go run config_tool.go summary <config_schemas.json>  # display the configuration summary")
		os.Exit(1)
	}

	command := os.Args[1]
	if len(os.Args) < 3 {
		fmt.Println("Error: Provide the configuration file path")
		os.Exit(1)
	}

	filePath := os.Args[2]

	switch command {
	case "validate":
		if err := validateConfigSchemas(filePath); err != nil {
			fmt.Printf("❌ Validation failed: %v\n", err)
			os.Exit(1)
		}
	case "summary":
		if err := validateConfigSchemas(filePath); err != nil {
			fmt.Printf("❌ Validation failed: %v\n", err)
			os.Exit(1)
		}
		if err := printConfigSummary(filePath); err != nil {
			fmt.Printf("❌ Failed to display summary: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Supported commands: validate, summary")
		os.Exit(1)
	}
}
