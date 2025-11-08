package plugins

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/checkers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/enhancers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/guards"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/validators"
	eaphandlers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// InitPlugins 初始化所有插件
// sessionRepo 和 accountingRepo 需要外部传入，用于需要依赖注入的插件
func InitPlugins(sessionRepo repository.SessionRepository, accountingRepo repository.AccountingRepository) {
	// 注册密码验证器（无状态插件）
	registry.RegisterPasswordValidator(&validators.PAPValidator{})
	registry.RegisterPasswordValidator(&validators.CHAPValidator{})
	registry.RegisterPasswordValidator(&validators.MSCHAPValidator{})

	// 注册策略检查器（大部分无状态）
	registry.RegisterPolicyChecker(&checkers.StatusChecker{})
	registry.RegisterPolicyChecker(&checkers.ExpireChecker{})
	registry.RegisterPolicyChecker(&checkers.MacBindChecker{})
	registry.RegisterPolicyChecker(&checkers.VlanBindChecker{})

	// 需要依赖注入的检查器
	if sessionRepo != nil {
		registry.RegisterPolicyChecker(checkers.NewOnlineCountChecker(sessionRepo))
	}

	// 注册响应增强器
	registry.RegisterResponseEnhancer(enhancers.NewDefaultAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewHuaweiAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewH3CAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewZTEAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewMikrotikAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewIkuaiAcceptEnhancer())

	// 注册认证守卫
	registry.RegisterAuthGuard(guards.NewRejectDelayGuard())

	// 注册计费处理器（需要依赖注入）
	if sessionRepo != nil && accountingRepo != nil {
		registry.RegisterAccountingHandler(handlers.NewStartHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewUpdateHandler(sessionRepo))
		registry.RegisterAccountingHandler(handlers.NewStopHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewNasStateHandler(sessionRepo))
	}

	// 注册 EAP 处理器
	registry.RegisterEAPHandler(eaphandlers.NewMD5Handler())
	registry.RegisterEAPHandler(eaphandlers.NewOTPHandler())
	registry.RegisterEAPHandler(eaphandlers.NewMSCHAPv2Handler())

	// 厂商解析器在 vendor/parsers 包中通过 init() 自动注册
}
