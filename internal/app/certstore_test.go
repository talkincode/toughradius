package app

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

func newCertStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.SysCert{}))
	return db
}

func TestCertStoreServerKeyPair(t *testing.T) {
	db := newCertStoreTestDB(t)
	require.NoError(t, db.Create(&domain.SysCert{
		ID:         1,
		Name:       "srv",
		CertType:   "server",
		Cert:       "CERTPEM",
		PrivateKey: "KEYPEM",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}).Error)

	store := NewCertStore(db)

	t.Run("returns cert and key", func(t *testing.T) {
		certPEM, keyPEM, err := store.ServerKeyPair("srv")
		require.NoError(t, err)
		assert.Equal(t, "CERTPEM", string(certPEM))
		assert.Equal(t, "KEYPEM", string(keyPEM))
	})

	t.Run("unknown name errors", func(t *testing.T) {
		_, _, err := store.ServerKeyPair("missing")
		assert.Error(t, err)
	})

	t.Run("missing private key errors", func(t *testing.T) {
		require.NoError(t, db.Create(&domain.SysCert{
			ID: 2, Name: "nokey", CertType: "ca", Cert: "CA", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}).Error)
		_, _, err := store.ServerKeyPair("nokey")
		assert.Error(t, err)
	})
}

func TestCertStoreCABundle(t *testing.T) {
	db := newCertStoreTestDB(t)
	require.NoError(t, db.Create(&domain.SysCert{
		ID: 1, Name: "ca", CertType: "ca", Cert: "CABUNDLE", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	store := NewCertStore(db)

	t.Run("returns ca bundle", func(t *testing.T) {
		caPEM, err := store.CABundle("ca")
		require.NoError(t, err)
		assert.Equal(t, "CABUNDLE", string(caPEM))
	})

	t.Run("unknown name errors", func(t *testing.T) {
		_, err := store.CABundle("missing")
		assert.Error(t, err)
	})
}

func TestCertStoreNilGuards(t *testing.T) {
	var store *CertStore
	_, _, err := store.ServerKeyPair("x")
	assert.Error(t, err)
	_, err = store.CABundle("x")
	assert.Error(t, err)
}
