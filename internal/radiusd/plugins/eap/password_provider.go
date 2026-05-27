package eap

import (
	"github.com/talkincode/toughradius/v9/internal/domain"
)

// DefaultPasswordProvider is the default password provider implementation
type DefaultPasswordProvider struct{}

// NewDefaultPasswordProvider creates the default password provider
func NewDefaultPasswordProvider() *DefaultPasswordProvider {
	return &DefaultPasswordProvider{}
}

// GetPassword returns the user's password
// For MAC authentication, return the MAC address as the password
// For regular users, return the plaintext password
func (p *DefaultPasswordProvider) GetPassword(user *domain.RadiusUser, isMacAuth bool) (string, error) {
	if isMacAuth {
		// Use the MAC address as the password (remove separators)
		if user.MacAddr != "" {
			return user.MacAddr, nil
		}
		return user.Username, nil
	}

	// Regular authentication returns the plaintext password
	return user.Password, nil
}
