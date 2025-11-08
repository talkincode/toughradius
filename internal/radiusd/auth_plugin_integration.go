package radiusd

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"layeh.com/radius"
)

// authPluginOptions 定义插件认证可选项
type authPluginOptions struct {
	skipPasswordValidation bool
}

// AuthPluginOption 认证选项函数
type AuthPluginOption func(*authPluginOptions)

// SkipPasswordValidation 跳过密码验证（用于已通过其他方式认证的场景，例如 EAP）
func SkipPasswordValidation() AuthPluginOption {
	return func(opts *authPluginOptions) {
		opts.skipPasswordValidation = true
	}
}

// AuthenticateUserWithPlugins 使用插件系统进行用户认证
func (s *AuthService) AuthenticateUserWithPlugins(
	ctx context.Context,
	r *radius.Request,
	response *radius.Packet,
	user *domain.RadiusUser,
	nas *domain.NetNas,
	vendorReq *vendorparsers.VendorRequest,
	isMacAuth bool,
	opts ...AuthPluginOption,
) error {
	// 解析可选参数
	options := &authPluginOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	// 创建认证上下文
	authCtx := &auth.AuthContext{
		Request:       r,
		Response:      response,
		User:          user,
		Nas:           nas,
		VendorRequest: vendorReq,
		IsMacAuth:     isMacAuth,
		Metadata:      make(map[string]interface{}),
	}

	var password string
	var err error

	// 1. 执行密码验证（使用插件）
	if !isMacAuth && !options.skipPasswordValidation {
		password, err = s.GetLocalPassword(user, isMacAuth)
		if err != nil {
			return errors.WrapError("radus_reject_passwd_error", err)
		}

		if err := s.validatePasswordWithPlugins(ctx, authCtx, password); err != nil {
			return err
		}
	}

	// 2. 执行策略检查（使用插件）
	if !isMacAuth {
		if err := s.checkPoliciesWithPlugins(ctx, authCtx); err != nil {
			return err
		}
	}

	return nil
}

// validatePasswordWithPlugins 使用密码验证器插件进行密码验证
func (s *AuthService) validatePasswordWithPlugins(
	ctx context.Context,
	authCtx *auth.AuthContext,
	password string,
) error {
	// 获取所有已注册的密码验证器
	validators := registry.GetPasswordValidators()

	// 遍历验证器，找到能处理当前请求的验证器
	for _, validator := range validators {
		if validator.CanHandle(authCtx) {
			return validator.Validate(ctx, authCtx, password)
		}
	}

	// 没有找到合适的验证器，返回错误
	return errors.NewAuthError("radus_reject_other", "no suitable password validator found")
}

// checkPoliciesWithPlugins 使用策略检查器插件进行策略验证
func (s *AuthService) checkPoliciesWithPlugins(
	ctx context.Context,
	authCtx *auth.AuthContext,
) error {
	// 获取所有已注册的策略检查器（已按Order排序）
	checkers := registry.GetPolicyCheckers()

	// 按顺序执行所有策略检查器
	for _, checker := range checkers {
		if err := checker.Check(ctx, authCtx); err != nil {
			return err
		}
	}

	return nil
}
