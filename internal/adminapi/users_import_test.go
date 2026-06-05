package adminapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

// newUploadRequest builds a multipart request carrying a single "upload" file.
func newUploadRequest(filename string, content []byte) (*http.Request, *httptest.ResponseRecorder) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("upload", filename)
	_, _ = part.Write(content)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	return req, rec
}

func TestImportRadiusUsers_JSON(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	profile := createTestProfile(db, "import-profile")

	lines := []map[string]interface{}{
		{"username": "imp_user1", "password": "secret123", "profile_id": profile.ID, "realname": "User One"},
		{"username": "imp_user2", "password": "secret123", "profile_id": profile.ID, "email": "u2@example.com"},
	}
	var content bytes.Buffer
	for _, l := range lines {
		bs, _ := json.Marshal(l)
		content.Write(bs)
		content.WriteByte('\n')
	}

	req, rec := newUploadRequest("users.json", content.Bytes())
	e := setupTestEcho()
	c := CreateTestContext(e, db, req, rec, appCtx)

	require.NoError(t, importRadiusUsers(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, _ := json.Marshal(resp.Data)
	var result ImportUserResult
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Success)
	assert.Equal(t, 0, result.Failed)

	var count int64
	db.Model(&domain.RadiusUser{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestImportRadiusUsers_CSV(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	profile := createTestProfile(db, "import-profile")

	csv := "username,password,profile_id,mobile\n" +
		"csv_user1,secret123," + strconv.FormatInt(profile.ID, 10) + ",13800138000\n" +
		"csv_user2,secret123," + strconv.FormatInt(profile.ID, 10) + ",13800138001\n"

	req, rec := newUploadRequest("users.csv", []byte(csv))
	e := setupTestEcho()
	c := CreateTestContext(e, db, req, rec, appCtx)

	require.NoError(t, importRadiusUsers(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, _ := json.Marshal(resp.Data)
	var result ImportUserResult
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, 2, result.Success)
}

func TestImportRadiusUsers_Errors(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	profile := createTestProfile(db, "import-profile")
	createTestUser(db, "dup_user", profile.ID)

	lines := []map[string]interface{}{
		{"username": "", "password": "secret123", "profile_id": profile.ID},         // missing username
		{"username": "no_pass", "profile_id": profile.ID},                           // missing password
		{"username": "bad_profile", "password": "secret123", "profile_id": 999999},  // missing profile
		{"username": "dup_user", "password": "secret123", "profile_id": profile.ID}, // duplicate
		{"username": "ok_user", "password": "secret123", "profile_id": profile.ID},  // valid
	}
	var content bytes.Buffer
	for _, l := range lines {
		bs, _ := json.Marshal(l)
		content.Write(bs)
		content.WriteByte('\n')
	}

	req, rec := newUploadRequest("users.json", content.Bytes())
	e := setupTestEcho()
	c := CreateTestContext(e, db, req, rec, appCtx)

	require.NoError(t, importRadiusUsers(c))

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, _ := json.Marshal(resp.Data)
	var result ImportUserResult
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, 5, result.Total)
	assert.Equal(t, 1, result.Success)
	assert.Equal(t, 4, result.Failed)
	assert.Len(t, result.Errors, 4)
}
