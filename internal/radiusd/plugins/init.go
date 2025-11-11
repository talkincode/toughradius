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

// InitPlugins initializes all plugins
// sessionRepo and accountingRepo must be supplied externally to support dependency injection for plugins
func InitPlugins(sessionRepo repository.SessionRepository, accountingRepo repository.AccountingRepository) {
	// Register password validators (stateless plugins)
	registry.RegisterPasswordValidator(&validators.PAPValidator{})
	registry.RegisterPasswordValidator(&validators.CHAPValidator{})
	registry.RegisterPasswordValidator(&validators.MSCHAPValidator{})

	// Register profile checkers (mostly stateless)
	registry.RegisterPolicyChecker(&checkers.StatusChecker{})
	registry.RegisterPolicyChecker(&checkers.ExpireChecker{})
	registry.RegisterPolicyChecker(&checkers.MacBindChecker{})
	registry.RegisterPolicyChecker(&checkers.VlanBindChecker{})

	// Checkers that require dependency injection
	if sessionRepo != nil {
		registry.RegisterPolicyChecker(checkers.NewOnlineCountChecker(sessionRepo))
	}

	// Register response enhancers
	registry.RegisterResponseEnhancer(enhancers.NewDefaultAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewHuaweiAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewH3CAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewZTEAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewMikrotikAcceptEnhancer())
	registry.RegisterResponseEnhancer(enhancers.NewIkuaiAcceptEnhancer())

	// Register authentication guards
	registry.RegisterAuthGuard(guards.NewRejectDelayGuard())

	// Register accounting handlers (dependency injection required)
	if sessionRepo != nil && accountingRepo != nil {
		registry.RegisterAccountingHandler(handlers.NewStartHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewUpdateHandler(sessionRepo))
		registry.RegisterAccountingHandler(handlers.NewStopHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewNasStateHandler(sessionRepo))
	}

	// Register EAP handlers
	registry.RegisterEAPHandler(eaphandlers.NewMD5Handler())
	registry.RegisterEAPHandler(eaphandlers.NewOTPHandler())
	registry.RegisterEAPHandler(eaphandlers.NewMSCHAPv2Handler())

	// Vendor parsers under vendor/parsers register themselves via init()
}
