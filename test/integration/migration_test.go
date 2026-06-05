//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

// TestPostgresMigration asserts that the production schema migrates cleanly on a
// real PostgreSQL server. It calls AutoMigrate directly (rather than the app's
// MigrateDB, which swallows the error) so a migration regression fails loudly,
// then verifies every declared table physically exists.
func TestPostgresMigration(t *testing.T) {
	db := h.appCtx.DB()

	err := db.Migrator().AutoMigrate(domain.Tables...)
	require.NoError(t, err, "AutoMigrate against PostgreSQL must succeed")

	for _, model := range domain.Tables {
		assert.Truef(t, db.Migrator().HasTable(model), "expected table for %T to exist", model)
	}

	// Spot-check a representative table's columns to catch silent column drift.
	for _, col := range []string{"id", "username", "password", "profile_id", "status"} {
		assert.Truef(t, db.Migrator().HasColumn(&domain.RadiusUser{}, col),
			"radius_user.%s column should exist", col)
	}
}
