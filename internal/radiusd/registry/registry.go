package registry

import (
	"sort"
	"sync"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/accounting"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	vendorparserspkg "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
)

// Registry 插件注册中心
type Registry struct {
	passwordValidators map[string]auth.PasswordValidator
	policyCheckers     []auth.PolicyChecker
	responseEnhancers  []auth.ResponseEnhancer
	authGuards         []auth.Guard
	vendorParsers      map[string]vendorparserspkg.VendorParser
	vendorBuilders     map[string]vendorparserspkg.VendorResponseBuilder
	acctHandlers       []accounting.AccountingHandler
	eapHandlers        map[uint8]eap.EAPHandler // EAP处理器，按EAPType索引
	mu                 sync.RWMutex
}

var globalRegistry = newRegistry()

func newRegistry() *Registry {
	return &Registry{
		passwordValidators: make(map[string]auth.PasswordValidator),
		policyCheckers:     make([]auth.PolicyChecker, 0),
		responseEnhancers:  make([]auth.ResponseEnhancer, 0),
		authGuards:         make([]auth.Guard, 0),
		vendorParsers:      make(map[string]vendorparserspkg.VendorParser),
		vendorBuilders:     make(map[string]vendorparserspkg.VendorResponseBuilder),
		acctHandlers:       make([]accounting.AccountingHandler, 0),
		eapHandlers:        make(map[uint8]eap.EAPHandler),
	}
}

// RegisterPasswordValidator 注册密码验证器
func RegisterPasswordValidator(validator auth.PasswordValidator) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.passwordValidators[validator.Name()] = validator
}

// GetPasswordValidators 获取所有密码验证器
func GetPasswordValidators() []auth.PasswordValidator {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	validators := make([]auth.PasswordValidator, 0, len(globalRegistry.passwordValidators))
	for _, v := range globalRegistry.passwordValidators {
		validators = append(validators, v)
	}
	return validators
}

// GetPasswordValidator 根据名称获取密码验证器
func GetPasswordValidator(name string) (auth.PasswordValidator, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	v, ok := globalRegistry.passwordValidators[name]
	return v, ok
}

// RegisterPolicyChecker 注册策略检查器
func RegisterPolicyChecker(checker auth.PolicyChecker) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.policyCheckers = append(globalRegistry.policyCheckers, checker)
	// 按Order排序
	sort.Slice(globalRegistry.policyCheckers, func(i, j int) bool {
		return globalRegistry.policyCheckers[i].Order() < globalRegistry.policyCheckers[j].Order()
	})
}

// GetPolicyCheckers 获取所有策略检查器（已按Order排序）
func GetPolicyCheckers() []auth.PolicyChecker {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	// 返回副本
	checkers := make([]auth.PolicyChecker, len(globalRegistry.policyCheckers))
	copy(checkers, globalRegistry.policyCheckers)
	return checkers
}

// RegisterResponseEnhancer 注册响应增强器
func RegisterResponseEnhancer(enhancer auth.ResponseEnhancer) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.responseEnhancers = append(globalRegistry.responseEnhancers, enhancer)
}

// GetResponseEnhancers 获取所有响应增强器
func GetResponseEnhancers() []auth.ResponseEnhancer {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	enhancers := make([]auth.ResponseEnhancer, len(globalRegistry.responseEnhancers))
	copy(enhancers, globalRegistry.responseEnhancers)
	return enhancers
}

// RegisterAuthGuard 注册认证守卫
func RegisterAuthGuard(guard auth.Guard) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.authGuards = append(globalRegistry.authGuards, guard)
}

// GetAuthGuards 获取所有认证守卫
func GetAuthGuards() []auth.Guard {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	guards := make([]auth.Guard, len(globalRegistry.authGuards))
	copy(guards, globalRegistry.authGuards)
	return guards
}

// RegisterVendorParser 注册厂商解析器
func RegisterVendorParser(parser vendorparserspkg.VendorParser) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.vendorParsers[parser.VendorCode()] = parser
}

// GetVendorParser 获取厂商解析器
func GetVendorParser(vendorCode string) (vendorparserspkg.VendorParser, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	parser, ok := globalRegistry.vendorParsers[vendorCode]
	if !ok {
		// 返回默认解析器
		parser, ok = globalRegistry.vendorParsers["default"]
	}
	return parser, ok
}

// RegisterVendorResponseBuilder 注册厂商响应构建器
func RegisterVendorResponseBuilder(builder vendorparserspkg.VendorResponseBuilder) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.vendorBuilders[builder.VendorCode()] = builder
}

// GetVendorResponseBuilder 获取厂商响应构建器
func GetVendorResponseBuilder(vendorCode string) (vendorparserspkg.VendorResponseBuilder, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	builder, ok := globalRegistry.vendorBuilders[vendorCode]
	return builder, ok
}

// RegisterAccountingHandler 注册计费处理器
func RegisterAccountingHandler(handler accounting.AccountingHandler) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.acctHandlers = append(globalRegistry.acctHandlers, handler)
}

// GetAccountingHandlers 获取所有计费处理器
func GetAccountingHandlers() []accounting.AccountingHandler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	handlers := make([]accounting.AccountingHandler, len(globalRegistry.acctHandlers))
	copy(handlers, globalRegistry.acctHandlers)
	return handlers
}

// RegisterEAPHandler 注册 EAP 处理器
func RegisterEAPHandler(handler eap.EAPHandler) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.eapHandlers[handler.EAPType()] = handler
}

// GetEAPHandler 根据 EAP 类型获取处理器
func GetEAPHandler(eapType uint8) (eap.EAPHandler, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	handler, ok := globalRegistry.eapHandlers[eapType]
	return handler, ok
}

// GetAllEAPHandlers 获取所有 EAP 处理器
func GetAllEAPHandlers() map[uint8]eap.EAPHandler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	handlers := make(map[uint8]eap.EAPHandler, len(globalRegistry.eapHandlers))
	for k, v := range globalRegistry.eapHandlers {
		handlers[k] = v
	}
	return handlers
}

// GetHandler 实现 eap.HandlerRegistry 接口
func (r *Registry) GetHandler(eapType uint8) (eap.EAPHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.eapHandlers[eapType]
	return handler, ok
}

// GetGlobalRegistry 获取全局注册表实例
// 用于实现 eap.HandlerRegistry 接口
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// ResetForTest clears the registry. Test helper only.
func ResetForTest() {
	globalRegistry = newRegistry()
}
