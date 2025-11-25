package validator

import (
	"fmt"
	"reflect"
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
			errors[field] = fmt.Sprintf("%s cannot be empty", err.Field())
		case "email":
			errors[field] = "Please provide a valid email address"
		case "min":
			errors[field] = fmt.Sprintf("%s minimum length is %s", err.Field(), err.Param())
		case "max":
			errors[field] = fmt.Sprintf("%s maximum length is %s", err.Field(), err.Param())
		case "gte":
			errors[field] = fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
		case "lte":
			errors[field] = fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
		case "gt":
			errors[field] = fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
		case "lt":
			errors[field] = fmt.Sprintf("%s must be less than %s", err.Field(), err.Param())
		case "oneof":
			errors[field] = fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
		case "ip":
			errors[field] = fmt.Sprintf("%s must be a valid IP address", err.Field())
		case "ipv4":
			errors[field] = fmt.Sprintf("%s must be a valid IPv4 address", err.Field())
		case "ipv6":
			errors[field] = fmt.Sprintf("%s must be a valid IPv6 address", err.Field())
		case "cidr":
			errors[field] = fmt.Sprintf("%s must be a valid CIDR", err.Field())
		case "cidrv4":
			errors[field] = fmt.Sprintf("%s must be a valid IPv4 CIDR", err.Field())
		case "cidrv6":
			errors[field] = fmt.Sprintf("%s must be a valid IPv6 CIDR", err.Field())
		case "mac":
			errors[field] = fmt.Sprintf("%s must be a valid MAC address", err.Field())
		case "url":
			errors[field] = fmt.Sprintf("%s must be a valid URL", err.Field())
		case "alphanum":
			errors[field] = fmt.Sprintf("%s may only contain letters and digits", err.Field())
		case "addrpool":
			errors[field] = fmt.Sprintf("%s must be a valid address pool format (CIDR)", err.Field())
		case "radiusstatus":
			errors[field] = fmt.Sprintf("%s must be enabled or disabled", err.Field())
		default:
			errors[field] = fmt.Sprintf("%s validation failed: %s", err.Field(), err.Tag())
		}
	}

	return map[string]interface{}{
		"error":   "VALIDATION_ERROR",
		"message": "Request parameter validation failed",
		"details": errors,
	}
}

// registerCustomValidations registers custom validation rules
func registerCustomValidations(v *validator.Validate) {
	// ValidateAddress pool format (simple CIDR check)
	_ = v.RegisterValidation("addrpool", func(fl validator.FieldLevel) bool { //nolint:errcheck
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
	_ = v.RegisterValidation("radiusstatus", func(fl validator.FieldLevel) bool { //nolint:errcheck
		value := fl.Field().String()
		return value == "enabled" || value == "disabled" || value == ""
	})

	// Validate username format (letters, digits, underscore, hyphen)
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool { //nolint:errcheck
		value := fl.Field().String()
		if value == "" {
			return true
		}
		// Allow letters, digits, underscore, hyphen, @, and dot
		for _, c := range value {
			//nolint:staticcheck // intentionally using verbose comparison for clarity
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '@' || c == '.') {
				return false
			}
		}
		return true
	})

	// Validate port numbers
	_ = v.RegisterValidation("port", func(fl validator.FieldLevel) bool { //nolint:errcheck
		field := fl.Field()
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return true
			}
			field = field.Elem()
		}
		if !field.IsValid() {
			return false
		}
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			port := field.Int()
			return port >= 1 && port <= 65535
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			port := field.Uint()
			return port >= 1 && port <= 65535
		default:
			return false
		}
	})
}
