//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
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

// adminTokenOnce guards a single admin login shared by the whole test process.
var (
	adminTokenOnce sync.Once
	adminTokenVal  string
	adminTokenErr  error
)

func newAPIClient(t *testing.T) *apiClient {
	t.Helper()
	return &apiClient{base: h.webBaseURL, http: &http.Client{}, token: sharedAdminToken(t)}
}

// sharedAdminToken logs in once per test process and caches the JWT for reuse by
// every apiClient. The token's TTL (server-side tokenTTL is 12h) dwarfs any
// suite run, so reuse is safe and mirrors how a real client behaves. It also
// keeps the suite within the login endpoint's production rate limit (burst 5,
// then one token every 3s per client IP): logging in afresh for each test would
// trip HTTP 429 from 127.0.0.1 once the suite grew past the initial burst.
func sharedAdminToken(t *testing.T) string {
	t.Helper()
	adminTokenOnce.Do(func() {
		adminTokenVal, adminTokenErr = loginToken(h.webBaseURL, h.adminUser, h.adminPass)
	})
	require.NoError(t, adminTokenErr, "shared admin login should succeed")
	require.NotEmpty(t, adminTokenVal, "shared admin login must return a token")
	return adminTokenVal
}

// loginToken performs the real /auth/login round-trip and returns the bearer
// token, so authenticated requests still flow through the production JWT
// middleware chain.
func loginToken(base, username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post(base+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login status %d: %s", resp.StatusCode, string(msg))
	}
	var payload struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Data.Token == "" {
		return "", fmt.Errorf("login returned an empty token")
	}
	return payload.Data.Token, nil
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

// post performs an authenticated POST with an optional JSON body (nil for none)
// and returns the status and raw response body.
func (c *apiClient) post(t *testing.T, path string, body []byte) (int, []byte) {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPost, c.base+path, reqBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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
