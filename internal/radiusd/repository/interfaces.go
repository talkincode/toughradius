package repository

import (
	"context"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

// UserRepository defines user data access operations
type UserRepository interface {
	// GetByUsername finds a user by username
	GetByUsername(ctx context.Context, username string) (*domain.RadiusUser, error)

	// GetByMacAddr finds a user by MAC address
	GetByMacAddr(ctx context.Context, macAddr string) (*domain.RadiusUser, error)

	// UpdateMacAddr updates the user's MAC address
	UpdateMacAddr(ctx context.Context, username, macAddr string) error

	// UpdateVlanId updates the user's VLAN ID
	UpdateVlanId(ctx context.Context, username string, vlanId1, vlanId2 int) error

	// UpdateLastOnline updates the last online time
	UpdateLastOnline(ctx context.Context, username string) error

	// UpdateField updates a specified user field
	UpdateField(ctx context.Context, username string, field string, value interface{}) error
}

// SessionRepository manages online sessions
type SessionRepository interface {
	// Create Create online session
	Create(ctx context.Context, session *domain.RadiusOnline) error

	// Update updates session data
	Update(ctx context.Context, session *domain.RadiusOnline) error

	// Delete deletes a session
	Delete(ctx context.Context, sessionId string) error

	// GetBySessionId finds a session by its ID
	GetBySessionId(ctx context.Context, sessionId string) (*domain.RadiusOnline, error)

	// CountByUsername counts online sessions per user
	CountByUsername(ctx context.Context, username string) (int, error)

	// Exists checks whether the session exists
	Exists(ctx context.Context, sessionId string) (bool, error)

	// BatchDelete deletes sessions in bulk
	BatchDelete(ctx context.Context, ids []string) error

	// BatchDeleteByNas deletes sessions by NAS
	BatchDeleteByNas(ctx context.Context, nasAddr, nasId string) error
}

// AccountingRepository defines accounting record operations
type AccountingRepository interface {
	// Create Create accounting record
	Create(ctx context.Context, accounting *domain.RadiusAccounting) error

	// UpdateStop updates stop time and traffic counters
	UpdateStop(ctx context.Context, sessionId string, accounting *domain.RadiusAccounting) error
}

// NasRepository manages NAS devices
type NasRepository interface {
	// GetByIP finds a NAS by IP
	GetByIP(ctx context.Context, ip string) (*domain.NetNas, error)

	// GetByIdentifier finds a NAS by identifier
	GetByIdentifier(ctx context.Context, identifier string) (*domain.NetNas, error)

	// GetByIPOrIdentifier finds a NAS by IP or identifier
	GetByIPOrIdentifier(ctx context.Context, ip, identifier string) (*domain.NetNas, error)
}
