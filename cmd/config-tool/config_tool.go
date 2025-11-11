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
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var schemasData app.ConfigSchemasJSON
	if err := json.Unmarshal(data, &schemasData); err != nil {
		return fmt.Errorf("JSON 格式错误: %w", err)
	}

		// Validate each configuration entry
	keyMap := make(map[string]bool)
	for i, schema := range schemasData.Schemas {
		// Check required fields
		if schema.Key == "" {
			return fmt.Errorf("配置项 %d: key 不能为空", i)
		}
		if schema.Type == "" {
			return fmt.Errorf("配置项 %d (%s): type 不能为空", i, schema.Key)
		}
		if schema.Default == "" {
			return fmt.Errorf("配置项 %d (%s): default 不能为空", i, schema.Key)
		}

		// Check for duplicate keys
		if keyMap[schema.Key] {
			return fmt.Errorf("配置项 %d (%s): key 重复", i, schema.Key)
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
			return fmt.Errorf("配置项 %d (%s): 无效的类型 %s，支持的类型: %v", i, schema.Key, schema.Type, validTypes)
		}

		// Validate integer ranges
		if schema.Type == "int" {
			if schema.Min != nil && schema.Max != nil && *schema.Min > *schema.Max {
				return fmt.Errorf("配置项 %d (%s): min 值不能大于 max 值", i, schema.Key)
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
				return fmt.Errorf("配置项 %d (%s): 默认值 %s 不在枚举列表中 %v", i, schema.Key, schema.Default, schema.Enum)
			}
		}
	}

	fmt.Printf("✓ 配置文件验证成功！共有 %d 个配置项\n", len(schemasData.Schemas))
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

	fmt.Printf("\n配置摘要 (共 %d 项):\n", len(schemasData.Schemas))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	categoryMap := make(map[string][]app.ConfigSchemaJSON)
	for _, schema := range schemasData.Schemas {
		// Group entries by category
		var category string
		if idx := findDotIndex(schema.Key); idx != -1 {
			category = schema.Key[:idx]
		} else {
			category = "其他"
		}
		categoryMap[category] = append(categoryMap[category], schema)
	}

	for category, schemas := range categoryMap {
		fmt.Printf("\n🔧 %s (%d 项):\n", category, len(schemas))
		for _, schema := range schemas {
			fmt.Printf("  • %-30s [%s] %s\n", schema.Key, schema.Type, schema.Description)
			if len(schema.Enum) > 0 {
				fmt.Printf("    └─ 枚举: %v\n", schema.Enum)
			}
			if schema.Min != nil || schema.Max != nil {
				rangeInfo := "    └─ 范围: "
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
		fmt.Println("使用方法:")
		fmt.Println("  go run config_tool.go validate <config_schemas.json>  # 验证配置文件")
		fmt.Println("  go run config_tool.go summary <config_schemas.json>  # 显示配置摘要")
		os.Exit(1)
	}

	command := os.Args[1]
	if len(os.Args) < 3 {
		fmt.Println("错误: 请提供配置文件路径")
		os.Exit(1)
	}

	filePath := os.Args[2]

	switch command {
	case "validate":
		if err := validateConfigSchemas(filePath); err != nil {
			fmt.Printf("❌ 验证失败: %v\n", err)
			os.Exit(1)
		}
	case "summary":
		if err := validateConfigSchemas(filePath); err != nil {
			fmt.Printf("❌ 验证失败: %v\n", err)
			os.Exit(1)
		}
		if err := printConfigSummary(filePath); err != nil {
			fmt.Printf("❌ 显示摘要失败: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("未知命令: %s\n", command)
		fmt.Println("支持的命令: validate, summary")
		os.Exit(1)
	}
}
