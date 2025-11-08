package auth

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
)

// AuthContext 认证上下文
type AuthContext struct {
	Request       *radius.Request
	Response      *radius.Packet
	User          *domain.RadiusUser
	Nas           *domain.NetNas
	VendorRequest interface{}
	IsMacAuth     bool                   // 是否为MAC认证
	Metadata      map[string]interface{} // 额外的元数据
}

// PasswordValidator 密码验证器接口
type PasswordValidator interface {
	// Name 返回验证器名称 (pap, chap, mschap, eap-md5, etc.)
	Name() string

	// CanHandle 判断是否可以处理该请求
	CanHandle(ctx *AuthContext) bool

	// Validate 执行密码验证
	Validate(ctx context.Context, authCtx *AuthContext, password string) error
}

// PolicyChecker 策略检查器接口
type PolicyChecker interface {
	// Name 返回检查器名称
	Name() string

	// Check 执行策略检查
	Check(ctx context.Context, authCtx *AuthContext) error

	// Order 返回执行顺序（数字越小越先执行）
	Order() int
}

// ResponseEnhancer 响应增强器接口
type ResponseEnhancer interface {
	// Name 返回增强器名称
	Name() string

	// Enhance 增强响应内容（添加厂商属性等）
	Enhance(ctx context.Context, authCtx *AuthContext) error
}

// Guard 认证守卫，用于统一处理认证过程中的错误（如拒绝延迟、黑名单等）
type Guard interface {
	// Name 返回守卫名称
	Name() string

	// OnError 在认证流程发生错误时被调用，可以返回新的错误中断流程
	OnError(ctx context.Context, authCtx *AuthContext, stage string, err error) error
}
