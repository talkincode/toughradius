package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/talkincode/toughradius/v9/internal/app"
)

// validateConfigSchemas validates the formatting and content of the configuration JSON file
func validateConfigSchemas(filePath string) error {
	data, err := os.ReadFile(filePath)
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

// printConfigSummary prints the configuration summary
func printConfigSummary(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var schemasData app.ConfigSchemasJSON
	if err := json.Unmarshal(data, &schemasData); err != nil {
		return err
	}

	fmt.Printf("\nConfiguration summary (%d entries):\n", len(schemasData.Schemas))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	categoryMap := make(map[string][]app.ConfigSchemaJSON)
	for _, schema := range schemasData.Schemas {
		// Group entries by category
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

func findDotIndex(s string) int {
	for i, r := range s {
		if r == '.' {
			return i
		}
	}
	return -1
}

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
