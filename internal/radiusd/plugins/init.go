package plugins

import (
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/checkers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/enhancers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/guards"

	// "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/guards"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth/validators"
	eaphandlers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/handlers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"github.com/talkincode/toughradius/v9/internal/radiusd/repository"
)

// InitPlugins initializes all plugins
// sessionRepo and accountingRepo must be supplied externally to support dependency injection for plugins
func InitPlugins(appCtx app.ConfigManagerProvider, sessionRepo repository.SessionRepository, accountingRepo repository.AccountingRepository) {
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
	var cfgGetter interface{ GetInt64(string, string) int64 }
	if appCtx != nil {
		cfgGetter = appCtx.ConfigMgr()
	}
	registry.RegisterAuthGuard(guards.NewRejectDelayGuard(cfgGetter))

	// Register accounting handlers (dependency injection required)
	if sessionRepo != nil && accountingRepo != nil {
		registry.RegisterAccountingHandler(handlers.NewStartHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewUpdateHandler(sessionRepo))
		registry.RegisterAccountingHandler(handlers.NewStopHandler(sessionRepo, accountingRepo))
		registry.RegisterAccountingHandler(handlers.NewNasStateHandler(sessionRepo))
	}

	// Register EAP handlers
	registry.RegisterEAPHandler(eaphandlers.NewMD5Handler())
	// EAP-OTP is intentionally not registered: its handler has no real OTP
	// validation backend yet. Registering it would expose an unauthenticated
	// EAP method. Re-enable only once a real validation service is wired in.
	registry.RegisterEAPHandler(eaphandlers.NewMSCHAPv2Handler())
	// EAP-TLS drives a full server-side TLS handshake with CA-chain
	// certificate validation and certificate-to-User-Name identity mapping
	// (milestone M1.3). Its certificate material is supplied from dynamic
	// settings (milestone M1.5): the provider returns a nil config — so the
	// handler rejects safely with eap.ErrTLSNotConfigured — until the
	// certificate/key/CA paths are configured, and it can never authenticate a
	// client without configured trust anchors. When no config manager is
	// available (e.g. unit tests with a nil appCtx), fall back to the
	// unconfigured handler which rejects identically.
	if appCtx != nil && appCtx.ConfigMgr() != nil {
		provider := eaphandlers.NewSettingsTLSConfigProvider(appCtx.ConfigMgr())
		registry.RegisterEAPHandler(eaphandlers.NewTLSHandlerWithConfig(provider))
	} else {
		registry.RegisterEAPHandler(eaphandlers.NewTLSHandler())
	}

	// PEAPv0 (EAP type 25) is a compatibility-first tunneled method for
	// Windows / Active Directory estates. M8.2 establishes the server-only outer
	// TLS tunnel with EAP-TLS framing, then rejects safely until the inner
	// EAP-MSCHAPv2 exchange (M8.3) is delivered.
	if appCtx != nil && appCtx.ConfigMgr() != nil {
		provider := eaphandlers.NewSettingsPEAPConfigProvider(appCtx.ConfigMgr())
		registry.RegisterEAPHandler(eaphandlers.NewPEAPHandlerWithConfig(provider))
	} else {
		registry.RegisterEAPHandler(eaphandlers.NewPEAPHandler())
	}

	// EAP-TTLS (EAP type 21) is a back-end-adaptation tunneled method (RFC 5281)
	// that protects legacy inner authentication (PAP / MS-CHAP-V2) with a
	// server-only TLS tunnel. M9.2 establishes the outer TLS tunnel with EAP-TLS
	// framing, then rejects safely with eap.ErrTTLSInnerNotImplemented until the
	// inner AVP authentication (M9.3+) is delivered, so it can never grant access
	// yet. When no config manager is available (e.g. unit tests with a nil
	// appCtx), fall back to the unconfigured handler which rejects identically
	// with eap.ErrTLSNotConfigured.
	if appCtx != nil && appCtx.ConfigMgr() != nil {
		provider := eaphandlers.NewSettingsTTLSConfigProvider(appCtx.ConfigMgr())
		registry.RegisterEAPHandler(eaphandlers.NewTTLSHandlerWithConfig(provider))
	} else {
		registry.RegisterEAPHandler(eaphandlers.NewTTLSHandler())
	}

	// Vendor parsers under vendor/parsers register themselves via init()
}
