package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// TestMigrateDelegatedIPv6Columns verifies that AutoMigrate provisions the
// Delegated-IPv6-Prefix columns (M3.2) on both the user and profile tables. The
// in-memory backend exercises the SQLite path; the PostgreSQL path is covered by
// the integration harness, which runs the same AutoMigrate at startup.
func TestMigrateDelegatedIPv6Columns(t *testing.T) {
	app := newTestApplication(t)
	m := app.gormDB.Migrator()

	assert.True(t, m.HasColumn(&domain.RadiusUser{}, "DelegatedIpv6Prefix"),
		"radius_user.delegated_ipv6_prefix should exist after migration")
	assert.True(t, m.HasColumn(&domain.RadiusUser{}, "DelegatedIpv6PrefixPool"),
		"radius_user.delegated_ipv6_prefix_pool should exist after migration")
	assert.True(t, m.HasColumn(&domain.RadiusProfile{}, "DelegatedIpv6PrefixPool"),
		"radius_profile.delegated_ipv6_prefix_pool should exist after migration")
}

// TestDelegatedIPv6RoundTrip confirms the new columns persist and reload intact,
// proving the migration produces usable storage rather than just a schema entry.
func TestDelegatedIPv6RoundTrip(t *testing.T) {
	app := newTestApplication(t)

	user := &domain.RadiusUser{
		ID:                      common.UUIDint64(),
		Username:                "ipv6-delegated-user",
		DelegatedIpv6Prefix:     "2001:db8:1234::/48",
		DelegatedIpv6PrefixPool: "pd-pool-a",
	}
	require.NoError(t, app.gormDB.Create(user).Error)

	var got domain.RadiusUser
	require.NoError(t, app.gormDB.Where("username = ?", "ipv6-delegated-user").First(&got).Error)
	assert.Equal(t, "2001:db8:1234::/48", got.DelegatedIpv6Prefix)
	assert.Equal(t, "pd-pool-a", got.DelegatedIpv6PrefixPool)
}
