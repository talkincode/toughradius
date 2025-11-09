package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator Echo 自定义验证器
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator 创建验证器实例
func NewValidator() *CustomValidator {
	v := validator.New()

	// 注册自定义验证规则
	registerCustomValidations(v)

	return &CustomValidator{validator: v}
}

// Validate 实现 echo.Validator 接口
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(400, formatValidationError(err))
	}
	return nil
}

// formatValidationError 格式化验证错误为友好消息
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

// registerCustomValidations 注册自定义验证规则
func registerCustomValidations(v *validator.Validate) {
	// 验证地址池格式 (简单的 CIDR 检查)
	v.RegisterValidation("addrpool", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true // 允许空值，使用 required 标签控制必填
		}
		// 简单验证是否包含 CIDR 标记
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			return false
		}
		// 可以添加更严格的 IP 和掩码验证
		return true
	})

	// 验证 RADIUS 状态值
	v.RegisterValidation("radiusstatus", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return value == "enabled" || value == "disabled" || value == ""
	})

	// 验证用户名格式（字母数字、下划线、中划线）
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		// 允许字母、数字、下划线、中划线、@符号
		for _, c := range value {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '@' || c == '.') {
				return false
			}
		}
		return true
	})

	// 验证端口号
	v.RegisterValidation("port", func(fl validator.FieldLevel) bool {
		port := fl.Field().Int()
		return port >= 1 && port <= 65535
	})
}
