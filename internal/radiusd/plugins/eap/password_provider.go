package eap

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
)

// DefaultPasswordProvider 默认密码提供者实现
type DefaultPasswordProvider struct{}

// NewDefaultPasswordProvider 创建默认密码提供者
func NewDefaultPasswordProvider() *DefaultPasswordProvider {
	return &DefaultPasswordProvider{}
}

// GetPassword 获取用户密码
// 对于 MAC 认证,返回 MAC 地址作为密码
// 对于普通用户,返回明文密码
func (p *DefaultPasswordProvider) GetPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
	if isMacAuth {
		// MAC 认证使用 MAC 地址作为密码(去除分隔符)
		if user.MacAddr != "" {
			return user.MacAddr, nil
		}
		return user.Username, nil
	}

	// 普通认证返回明文密码
	return user.Password, nil
}
