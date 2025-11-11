package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator is a custom Echo validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a validator instance
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom validation rules
	registerCustomValidations(v)

	return &CustomValidator{validator: v}
}

// Validate implements the echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(400, formatValidationError(err))
	}
	return nil
}

// formatValidationError formats validation errors into user-friendly messages
func formatValidationError(err error) map[string]interface{} {
	errors := make(map[string]string)

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]interface{}{
			"error":   "VALIDATION_ERROR",
			"message": err.Error(),
		}
	}

	for _, err := range validationErrors {
		field := strings.ToLower(err.Field())
		switch err.Tag() {
		case "required":
			errors[field] = fmt.Sprintf("%s 不能为空", err.Field())
		case "email":
			errors[field] = "请输入有效的邮箱地址"
		case "min":
			errors[field] = fmt.Sprintf("%s 最小长度为 %s", err.Field(), err.Param())
		case "max":
			errors[field] = fmt.Sprintf("%s 最大长度为 %s", err.Field(), err.Param())
		case "gte":
			errors[field] = fmt.Sprintf("%s 必须大于等于 %s", err.Field(), err.Param())
		case "lte":
			errors[field] = fmt.Sprintf("%s 必须小于等于 %s", err.Field(), err.Param())
		case "gt":
			errors[field] = fmt.Sprintf("%s 必须大于 %s", err.Field(), err.Param())
		case "lt":
			errors[field] = fmt.Sprintf("%s 必须小于 %s", err.Field(), err.Param())
		case "oneof":
			errors[field] = fmt.Sprintf("%s 必须是以下值之一: %s", err.Field(), err.Param())
		case "ip":
			errors[field] = fmt.Sprintf("%s 必须是有效的 IP 地址", err.Field())
		case "ipv4":
			errors[field] = fmt.Sprintf("%s 必须是有效的 IPv4 地址", err.Field())
		case "ipv6":
			errors[field] = fmt.Sprintf("%s 必须是有效的 IPv6 地址", err.Field())
		case "cidr":
			errors[field] = fmt.Sprintf("%s 必须是有效的 CIDR 格式", err.Field())
		case "cidrv4":
			errors[field] = fmt.Sprintf("%s 必须是有效的 IPv4 CIDR 格式", err.Field())
		case "cidrv6":
			errors[field] = fmt.Sprintf("%s 必须是有效的 IPv6 CIDR 格式", err.Field())
		case "mac":
			errors[field] = fmt.Sprintf("%s 必须是有效的 MAC 地址", err.Field())
		case "url":
			errors[field] = fmt.Sprintf("%s 必须是有效的 URL", err.Field())
		case "alphanum":
			errors[field] = fmt.Sprintf("%s 只能包含字母和数字", err.Field())
		case "addrpool":
			errors[field] = fmt.Sprintf("%s 必须是有效的地址池格式（CIDR）", err.Field())
		case "radiusstatus":
			errors[field] = fmt.Sprintf("%s 必须是 enabled 或 disabled", err.Field())
		default:
			errors[field] = fmt.Sprintf("%s 验证失败: %s", err.Field(), err.Tag())
		}
	}

	return map[string]interface{}{
		"error":   "VALIDATION_ERROR",
		"message": "请求参数验证失败",
		"details": errors,
	}
}

// registerCustomValidations registers custom validation rules
func registerCustomValidations(v *validator.Validate) {
	// ValidateAddress pool format (simple CIDR check)
	v.RegisterValidation("addrpool", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true // Allow empty values; use the required tag for mandatory fields
		}
		// Ensure the value contains a CIDR separator
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			return false
		}
		// Could add stricter IP and mask validation here
		return true
	})

	// Validate RADIUS status value
	v.RegisterValidation("radiusstatus", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return value == "enabled" || value == "disabled" || value == ""
	})

	// Validate username format (letters, digits, underscore, hyphen)
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		// Allow letters, digits, underscore, hyphen, @, and dot
		for _, c := range value {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '@' || c == '.') {
				return false
			}
		}
		return true
	})

	// Validate port numbers
	v.RegisterValidation("port", func(fl validator.FieldLevel) bool {
		port := fl.Field().Int()
		return port >= 1 && port <= 65535
	})
}
