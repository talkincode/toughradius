//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// apiClient is a thin HTTP client for the booted admin server that carries a
// JWT obtained via the real /auth/login endpoint, so every request flows
// through the production middleware chain (JWT auth, routing, binding).
type apiClient struct {
	base  string
	token string
	http  *http.Client
}

func newAPIClient(t *testing.T) *apiClient {
	t.Helper()
	c := &apiClient{base: h.webBaseURL, http: &http.Client{}}
	c.login(t, h.adminUser, h.adminPass)
	return c
}

func (c *apiClient) login(t *testing.T, username, password string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := c.http.Post(c.base+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equalf(t, http.StatusOK, resp.StatusCode, "login should succeed")

	var payload struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NotEmpty(t, payload.Data.Token, "login must return a token")
	c.token = payload.Data.Token
}

// getJSON performs an authenticated GET and returns the raw response body.
func (c *apiClient) get(t *testing.T, path string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

// postMultipart uploads a single file field ("upload") and returns the response.
func (c *apiClient) postMultipart(t *testing.T, path, filename string, content []byte) (int, []byte) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("upload", filename)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, c.base+path, &buf)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.http.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

// unwrapData decodes the {"data": ...} envelope into v.
func unwrapData(t *testing.T, body []byte, v interface{}) {
	t.Helper()
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoErrorf(t, json.Unmarshal(body, &env), "response not JSON: %s", string(body))
	require.NoError(t, json.Unmarshal(env.Data, v))
}

func csvRow(values ...string) string {
	out := ""
	for i, v := range values {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out + "\n"
}
