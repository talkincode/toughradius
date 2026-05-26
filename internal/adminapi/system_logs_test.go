package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	webresp "github.com/talkincode/toughradius/v9/pkg/web"
)

func TestQuerySystemLogsAcceptsIssueFilters(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)
	logPath := filepath.Join(t.TempDir(), "toughradius.log")
	logLine := `{"level":"info","ts":1757396220,"msg":"radius service ready","namespace":"radius","job":"toughradius"}` + "\n"
	require.NoError(t, os.WriteFile(logPath, []byte(logLine), 0600))
	appCtx.Config().Logger.Filename = logPath

	req := httptest.NewRequest(http.MethodGet, "/admin/loki/query", nil)
	q := req.URL.Query()
	q.Set("starttime", "2025-09-08 13:37:00")
	q.Set("endtime", "2025-09-11 13:37:00")
	q.Set("job", "toughradius")
	q.Set("namespace", "radius")
	q.Set("level", "info")
	q.Set("keyword", "")
	q.Set("keyreg", "")
	q.Set("limit", "1000")
	req.URL.RawQuery = q.Encode()
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	err := querySystemLogs(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp webresp.WebRestResult
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), data["total"])
}
