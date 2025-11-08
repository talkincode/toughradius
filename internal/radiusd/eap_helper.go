package radiusd

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"github.com/talkincode/toughradius/v9/internal/radiusd/registry"
	"layeh.com/radius"
)

// EAPAuthHelper EAP 认证辅助工具
type EAPAuthHelper struct {
	coordinator *eap.Coordinator
}

// NewEAPAuthHelper 创建 EAP 认证辅助工具
func NewEAPAuthHelper() *EAPAuthHelper {
	// 创建状态管理器
	stateManager := statemanager.NewMemoryStateManager()

	// 创建密码提供者
	pwdProvider := eap.NewDefaultPasswordProvider()

	// 获取处理器注册表
	var handlerRegistry eap.HandlerRegistry = registry.GetGlobalRegistry()

	// 创建协调器
	coordinator := eap.NewCoordinator(stateManager, pwdProvider, handlerRegistry)

	return &EAPAuthHelper{
		coordinator: coordinator,
	}
}

// HandleEAPAuthentication 处理 EAP 认证
// 返回 (handled bool, success bool, err error)
func (h *EAPAuthHelper) HandleEAPAuthentication(
	w radius.ResponseWriter,
	r *radius.Request,
	user *domain.RadiusUser,
	nas *domain.NetNas,
	vendorReq *vendorparsers.VendorRequest,
	response *radius.Packet,
	eapMethod string,
) (handled bool, success bool, err error) {

	// 判断是否为 MAC 认证
	isMacAuth := vendorReq.MacAddr != "" && vendorReq.MacAddr == user.Username

	// 调用协调器处理 EAP 请求
	handled, success, err = h.coordinator.HandleEAPRequest(
		w, r, user, nas, response, nas.Secret, isMacAuth, eapMethod,
	)

	return handled, success, err
}

// SendEAPSuccess 发送 EAP Success 响应
func (h *EAPAuthHelper) SendEAPSuccess(
	w radius.ResponseWriter,
	r *radius.Request,
	response *radius.Packet,
	secret string,
) error {
	return h.coordinator.SendEAPSuccess(w, r, response, secret)
}

// SendEAPFailure 发送 EAP Failure 响应
func (h *EAPAuthHelper) SendEAPFailure(
	w radius.ResponseWriter,
	r *radius.Request,
	secret string,
	reason error,
) error {
	return h.coordinator.SendEAPFailure(w, r, secret, reason)
}

// CleanupState 清理 EAP 状态
func (h *EAPAuthHelper) CleanupState(r *radius.Request) {
	h.coordinator.CleanupState(r)
}

// GetCoordinator 获取底层的协调器(用于高级用法)
func (h *EAPAuthHelper) GetCoordinator() *eap.Coordinator {
	return h.coordinator
}
