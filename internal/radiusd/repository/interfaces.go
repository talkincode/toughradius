package repository

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	// GetByUsername 根据用户名查询用户
	GetByUsername(ctx context.Context, username string) (*domain.RadiusUser, error)

	// GetByMacAddr 根据MAC地址查询用户
	GetByMacAddr(ctx context.Context, macAddr string) (*domain.RadiusUser, error)

	// UpdateMacAddr 更新用户MAC地址
	UpdateMacAddr(ctx context.Context, username, macAddr string) error

	// UpdateVlanId 更新用户VLAN ID
	UpdateVlanId(ctx context.Context, username string, vlanId1, vlanId2 int) error

	// UpdateLastOnline 更新最后在线时间
	UpdateLastOnline(ctx context.Context, username string) error

	// UpdateField 更新用户指定字段
	UpdateField(ctx context.Context, username string, field string, value interface{}) error
}

// SessionRepository 在线会话管理接口
type SessionRepository interface {
	// Create 创建在线会话
	Create(ctx context.Context, session *domain.RadiusOnline) error

	// Update 更新会话数据
	Update(ctx context.Context, session *domain.RadiusOnline) error

	// Delete 删除会话
	Delete(ctx context.Context, sessionId string) error

	// GetBySessionId 根据会话ID查询
	GetBySessionId(ctx context.Context, sessionId string) (*domain.RadiusOnline, error)

	// CountByUsername 统计用户在线数
	CountByUsername(ctx context.Context, username string) (int, error)

	// Exists 检查会话是否存在
	Exists(ctx context.Context, sessionId string) (bool, error)

	// BatchDelete 批量删除
	BatchDelete(ctx context.Context, ids []string) error

	// BatchDeleteByNas 根据NAS批量删除
	BatchDeleteByNas(ctx context.Context, nasAddr, nasId string) error
}

// AccountingRepository 计费记录接口
type AccountingRepository interface {
	// Create 创建计费记录
	Create(ctx context.Context, accounting *domain.RadiusAccounting) error

	// UpdateStop 更新停止时间和流量
	UpdateStop(ctx context.Context, sessionId string, accounting *domain.RadiusAccounting) error
}

// NasRepository NAS设备管理接口
type NasRepository interface {
	// GetByIP 根据IP查询NAS
	GetByIP(ctx context.Context, ip string) (*domain.NetNas, error)

	// GetByIdentifier 根据标识符查询NAS
	GetByIdentifier(ctx context.Context, identifier string) (*domain.NetNas, error)

	// GetByIPOrIdentifier 根据IP或标识符查询NAS
	GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error)
}

// ConfigRepository 配置访问接口
type ConfigRepository interface {
	// GetString 获取字符串配置
	GetString(ctx context.Context, category, key string) string

	// GetInt 获取整数配置
	GetInt(ctx context.Context, category, key string, defaultVal int64) int64

	// GetBool 获取布尔配置
	GetBool(ctx context.Context, category, key string) bool
}
