//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// uniqueSuffix returns a per-test suffix so tests stay isolated on the shared
// database without truncation or parallel-unsafe global resets.
func uniqueSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func seedProfile(t *testing.T, name string) int64 {
	t.Helper()
	p := &domain.RadiusProfile{
		ID:        common.UUIDint64(),
		Name:      name,
		Status:    common.ENABLED,
		AddrPool:  "it-pool",
		ActiveNum: 1,
		UpRate:    1000,
		DownRate:  2000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(p).Error)
	return p.ID
}

// TestSystemBackupRestoreRoundTrip exercises the real /system/backup and
// /system/restore HTTP endpoints against PostgreSQL and proves the security
// behaviour we rely on: backups carry plaintext RADIUS passwords (unlike the
// list API, which strips them) and restore reinstates a deleted user with the
// password intact via an ON CONFLICT upsert on real Postgres.
func TestSystemBackupRestoreRoundTrip(t *testing.T) {
	c := newAPIClient(t)
	suffix := uniqueSuffix()
	profileID := seedProfile(t, "it-profile-"+suffix)

	username := "it-backup-" + suffix
	const password = "secret-Pw-123"
	user := &domain.RadiusUser{
		ID:         common.UUIDint64(),
		ProfileId:  profileID,
		Username:   username,
		Password:   password,
		Status:     common.ENABLED,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, h.appCtx.DB().Create(user).Error)

	// 1) Download a backup over HTTP and confirm the plaintext password is present.
	// The endpoint streams the bare SystemBackup JSON (no {"data":...} envelope).
	status, backupBytes := c.get(t, "/api/v1/system/backup")
	require.Equalf(t, http.StatusOK, status, "backup body: %s", string(backupBytes))

	var backup struct {
		Version string              `json:"version"`
		Users   []domain.RadiusUser `json:"users"`
	}
	require.NoErrorf(t, json.Unmarshal(backupBytes, &backup), "backup not JSON: %s", string(backupBytes))
	require.Equal(t, "9.0", backup.Version)
	require.True(t, containsUserWithPassword(backup.Users, username, password),
		"backup must contain %s with its plaintext password", username)

	// 2) Delete the user directly, simulating data loss.
	require.NoError(t, h.appCtx.DB().Where("username = ?", username).Delete(&domain.RadiusUser{}).Error)
	var count int64
	require.NoError(t, h.appCtx.DB().Model(&domain.RadiusUser{}).Where("username = ?", username).Count(&count).Error)
	require.Equal(t, int64(0), count)

	// 3) Restore the backup over HTTP (multipart upload of the bare backup JSON).
	status, body := c.postMultipart(t, "/api/v1/system/restore", "backup.json", backupBytes)
	require.Equalf(t, http.StatusOK, status, "restore body: %s", string(body))

	// 4) The user is back with its password preserved.
	var restored domain.RadiusUser
	require.NoError(t, h.appCtx.DB().Where("username = ?", username).First(&restored).Error)
	assert.Equal(t, password, restored.Password, "restore must preserve the plaintext password")
}

// TestSystemRestoreRejectsNonBackup ensures uploading a non-backup file (e.g. a
// CSV meant for user import) is rejected with a clear error, mirroring the
// real-world confusion between the restore and import features.
func TestSystemRestoreRejectsNonBackup(t *testing.T) {
	c := newAPIClient(t)
	csv := csvRow("username", "password", "profile_id") + csvRow("x", "y", "1")
	status, body := c.postMultipart(t, "/api/v1/system/restore", "users.csv", []byte(csv))
	require.Equalf(t, http.StatusBadRequest, status, "expected 400, body: %s", string(body))
	assert.Contains(t, string(body), "INVALID_BACKUP")
}

// TestUserImportCSV exercises the real /users/import endpoint with a multipart
// CSV upload against PostgreSQL and verifies the row lands in the database.
func TestUserImportCSV(t *testing.T) {
	c := newAPIClient(t)
	suffix := uniqueSuffix()
	profileID := seedProfile(t, "it-import-profile-"+suffix)

	username := "it-import-" + suffix
	csv := csvRow("username", "password", "profile_id") +
		csvRow(username, "import-Pw-123", fmt.Sprintf("%d", profileID))

	status, body := c.postMultipart(t, "/api/v1/users/import", "users.csv", []byte(csv))
	require.Equalf(t, http.StatusOK, status, "import body: %s", string(body))

	var result struct {
		Total   int `json:"total"`
		Success int `json:"success"`
		Failed  int `json:"failed"`
	}
	unwrapData(t, body, &result)
	assert.Equal(t, 1, result.Total)
	assert.Equalf(t, 1, result.Success, "import should succeed, body: %s", string(body))

	var created domain.RadiusUser
	require.NoError(t, h.appCtx.DB().Where("username = ?", username).First(&created).Error)
	assert.Equal(t, "import-Pw-123", created.Password)
	assert.Equal(t, profileID, created.ProfileId)
}

func containsUserWithPassword(users []domain.RadiusUser, username, password string) bool {
	for _, u := range users {
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}
