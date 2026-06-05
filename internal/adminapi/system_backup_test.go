package adminapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestBackupSystem(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	profile := createTestProfile(db, "backup-profile")
	createTestUser(db, "backup_user", profile.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/backup", nil)
	rec := httptest.NewRecorder()
	e := setupTestEcho()
	c := CreateTestContext(e, db, req, rec, appCtx)

	require.NoError(t, backupSystem(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "attachment")

	var backup SystemBackup
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &backup))
	assert.Equal(t, backupVersion, backup.Version)
	assert.Len(t, backup.Profiles, 1)
	assert.Len(t, backup.Users, 1)
	assert.Equal(t, "backup_user", backup.Users[0].Username)
}

func TestRestoreSystem(t *testing.T) {
	// Build a backup payload from a source database.
	srcDB := setupTestDB(t)
	srcAppCtx := setupTestApp(t, srcDB)
	profile := createTestProfile(srcDB, "restore-profile")
	createTestUser(srcDB, "restore_user", profile.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/backup", nil)
	rec := httptest.NewRecorder()
	e := setupTestEcho()
	c := CreateTestContext(e, srcDB, req, rec, srcAppCtx)
	require.NoError(t, backupSystem(c))
	payload := rec.Body.Bytes()

	// Restore into a fresh, empty database.
	dstDB := setupTestDB(t)
	dstAppCtx := setupTestApp(t, dstDB)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("upload", "backup.json")
	_, _ = part.Write(payload)
	_ = writer.Close()

	rreq := httptest.NewRequest(http.MethodPost, "/api/v1/system/restore", body)
	rreq.Header.Set("Content-Type", writer.FormDataContentType())
	rrec := httptest.NewRecorder()
	rc := CreateTestContext(setupTestEcho(), dstDB, rreq, rrec, dstAppCtx)

	require.NoError(t, restoreSystem(rc))
	assert.Equal(t, http.StatusOK, rrec.Code)

	var count int64
	dstDB.Model(&domain.RadiusUser{}).Count(&count)
	assert.Equal(t, int64(1), count)
	dstDB.Model(&domain.RadiusProfile{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Restoring again should be idempotent (upsert, not duplicate).
	body2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(body2)
	part2, _ := writer2.CreateFormFile("upload", "backup.json")
	_, _ = part2.Write(payload)
	_ = writer2.Close()
	rreq2 := httptest.NewRequest(http.MethodPost, "/api/v1/system/restore", body2)
	rreq2.Header.Set("Content-Type", writer2.FormDataContentType())
	rrec2 := httptest.NewRecorder()
	rc2 := CreateTestContext(setupTestEcho(), dstDB, rreq2, rrec2, dstAppCtx)
	require.NoError(t, restoreSystem(rc2))

	dstDB.Model(&domain.RadiusUser{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestRestoreSystem_InvalidFile(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("upload", "backup.json")
	_, _ = part.Write([]byte("not-json"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := CreateTestContext(setupTestEcho(), db, req, rec, appCtx)

	require.NoError(t, restoreSystem(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRestoreSystem_MissingVersion(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Valid JSON but without a version stamp should be rejected.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("upload", "backup.json")
	_, _ = part.Write([]byte(`{"nodes":[],"users":[]}`))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := CreateTestContext(setupTestEcho(), db, req, rec, appCtx)

	require.NoError(t, restoreSystem(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}