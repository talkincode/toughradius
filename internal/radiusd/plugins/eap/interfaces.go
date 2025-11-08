package eap

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"layeh.com/radius"
)

// EAP Code 常量
const (
	CodeRequest  = 1 // EAP Request message
	CodeResponse = 2 // EAP Response message
	CodeSuccess  = 3 // Indicates successful authentication
	CodeFailure  = 4 // Indicates failed authentication
)

// EAP Type 常量
const (
	TypeIdentity     = 1  // Identity
	TypeNotification = 2  // Notification
	TypeNak          = 3  // Response only (Legacy Nak)
	TypeMD5Challenge = 4  // MD5-Challenge
	TypeOTP          = 5  // One-Time Password
	TypeGTC          = 6  // Generic Token Card
	TypeTLS          = 13 // EAP-TLS
	TypeMSCHAPv2     = 26 // EAP-MSCHAPv2
)

// EAPState EAP 状态数据
type EAPState struct {
	Username  string                 // 用户名
	Challenge []byte                 // Challenge 数据
	StateID   string                 // 状态ID (RADIUS State 属性值)
	Method    string                 // EAP 方法名称 (eap-md5, eap-mschapv2, etc.)
	Success   bool                   // 是否认证成功
	Data      map[string]interface{} // 额外数据存储
}

// EAPContext EAP 认证上下文
type EAPContext struct {
	Context        context.Context
	Request        *radius.Request
	ResponseWriter radius.ResponseWriter // RADIUS 响应写入器
	Response       *radius.Packet
	User           *domain.RadiusUser
	NAS            *domain.NetNas
	EAPMessage     *EAPMessage
	EAPState       *EAPState
	IsMacAuth      bool
	Secret         string // RADIUS Secret
	StateManager   EAPStateManager
	PwdProvider    PasswordProvider
}

// EAPMessage EAP 消息结构
type EAPMessage struct {
	Code       uint8  // EAP Code
	Identifier uint8  // EAP Identifier
	Length     uint16 // EAP Length
	Type       uint8  // EAP Type
	Data       []byte // EAP Data
}

// EAPHandler EAP 认证处理器接口
type EAPHandler interface {
	// Name 返回处理器名称 (如 "eap-md5", "eap-mschapv2")
	Name() string

	// EAPType 返回处理的 EAP 类型码
	EAPType() uint8

	// CanHandle 判断是否可以处理该 EAP 消息
	CanHandle(ctx *EAPContext) bool

	// HandleIdentity 处理 EAP-Response/Identity，发送 Challenge
	// 返回 true 表示已处理并发送响应，false 表示不处理
	HandleIdentity(ctx *EAPContext) (bool, error)

	// HandleResponse 处理 EAP-Response (Challenge Response)
	// 返回 true 表示认证成功，false 表示认证失败
	HandleResponse(ctx *EAPContext) (bool, error)
}

// EAPStateManager EAP 状态管理器接口
type EAPStateManager interface {
	// GetState 获取 EAP 状态
	GetState(stateID string) (*EAPState, error)

	// SetState 设置 EAP 状态
	SetState(stateID string, state *EAPState) error

	// DeleteState 删除 EAP 状态
	DeleteState(stateID string) error
}

// PasswordProvider 密码提供者接口
type PasswordProvider interface {
	// GetPassword 获取用户密码（明文或加密）
	GetPassword(user *domain.RadiusUser, isMacAuth bool) (string, error)
}
