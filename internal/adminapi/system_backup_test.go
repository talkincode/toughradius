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

// restoreRequest builds a multipart restore request body for the given payload.
func restoreRequest(t *testing.T, payload []byte) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload", "backup.json")
	require.NoError(t, err)
	_, err = part.Write(payload)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, httptest.NewRecorder()
}

func TestRestoreSystem_IncompatibleVersion(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	req, rec := restoreRequest(t, []byte(`{"version":"8.5"}`))
	c := CreateTestContext(setupTestEcho(), db, req, rec, appCtx)

	require.NoError(t, restoreSystem(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "INVALID_BACKUP", resp.Error)
}

func TestRestoreSystem_InvalidOperatorLevel(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Operator with an unrecognized privilege level must be rejected outright.
	payload := []byte(`{"version":"9.0","operators":[{"id":"42","username":"evil","level":"root","status":"enabled"}]}`)
	req, rec := restoreRequest(t, payload)
	c := CreateTestContext(setupTestEcho(), db, req, rec, appCtx)

	require.NoError(t, restoreSystem(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var count int64
	db.Model(&domain.SysOpr{}).Count(&count)
	assert.Equal(t, int64(0), count, "no operator should have been written")
}

func TestRestoreSystem_OperatorsRequireSuper(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// A well-formed operator payload, but the caller is only an admin (not super):
	// restoring the operators table must be forbidden to prevent escalation.
	payload := []byte(`{"version":"9.0","operators":[{"id":"42","username":"newadmin","level":"super","status":"enabled"}]}`)
	req, rec := restoreRequest(t, payload)
	c := CreateTestContext(setupTestEcho(), db, req, rec, appCtx)
	c.Set("current_operator", &domain.SysOpr{ID: 7, Username: "admin", Level: LevelAdmin, Status: "enabled"})

	require.NoError(t, restoreSystem(c))
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var count int64
	db.Model(&domain.SysOpr{}).Count(&count)
	assert.Equal(t, int64(0), count, "operators must not be written for non-super callers")
}

func TestRestoreSystem_OperatorsAllowedForSuper(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	payload := []byte(`{"version":"9.0","operators":[{"id":"42","username":"newadmin","level":"admin","status":"enabled"}]}`)
	req, rec := restoreRequest(t, payload)
	// CreateTestContext injects a super operator by default.
	c := CreateTestContext(setupTestEcho(), db, req, rec, appCtx)

	require.NoError(t, restoreSystem(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var count int64
	db.Model(&domain.SysOpr{}).Where("username = ?", "newadmin").Count(&count)
	assert.Equal(t, int64(1), count)
}
