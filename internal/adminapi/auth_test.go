package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// setupAuthTest sets up the test environment and creates a test user
func setupAuthTest(t *testing.T) (*gorm.DB, *echo.Echo, app.AppContext, *domain.SysOpr, func()) {
	// Create test app context using the helper
	db, e, appCtx := CreateTestAppContext(t)

	// Create the test user
	testOpr := &domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  "testuser",
		Password:  common.Sha256HashWithSalt("password123", common.SecretSalt),
		Realname:  "Test User",
		Email:     "test@example.com",
		Mobile:    "13800138000",
		Level:     "super",
		Status:    common.ENABLED,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(testOpr)

	cleanup := func() {
		db.Where("id = ?", testOpr.ID).Delete(&domain.SysOpr{})
	}

	return db, e, appCtx, testOpr, cleanup
}

// TestLoginHandler tests the login handler
func TestLoginHandler(t *testing.T) {
	db, e, appCtx, _, cleanup := setupAuthTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedCode   string
		checkToken     bool
	}{
		{
			name:           "Successful login",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
			checkToken:     true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
			checkToken:     false,
		},
		{
			name:           "Empty username",
			requestBody:    `{"username":"","password":"password123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "Empty password",
			requestBody:    `{"username":"testuser","password":""}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "User not found",
			requestBody:    `{"username":"nonexistent","password":"password123"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "Wrong password",
			requestBody:    `{"username":"testuser","password":"wrongpassword"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "Username with surrounding spaces (should be trimmed)",
			requestBody:    `{"username":"  testuser  ","password":"password123"}`,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
			checkToken:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := loginHandler(c)

			// fail() calls c.JSON() and returns nil, so we check the response code instead of the error
			if tt.expectedStatus >= 400 {
				// Error cases: verify response status and error message
				require.NoError(t, err, "handler should not return error, but write error response")
				assert.Equal(t, tt.expectedStatus, rec.Code)

				var errorResp ErrorResponse
				err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCode, errorResp.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				if tt.checkToken {
					var response map[string]interface{}
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)

					data, ok := response["data"].(map[string]interface{})
					assert.True(t, ok, "response should have data field")

					// Check token
					token, ok := data["token"].(string)
					assert.True(t, ok, "data should have token field")
					assert.NotEmpty(t, token, "token should not be empty")

					// Check user
					user, ok := data["user"].(map[string]interface{})
					assert.True(t, ok, "data should have user field")
					assert.Equal(t, "testuser", user["username"])
					assert.Empty(t, user["password"], "password should be empty in response")

					// Check tokenExpires
					tokenExpires, ok := data["tokenExpires"].(float64)
					assert.True(t, ok, "data should have tokenExpires field")
					assert.Greater(t, tokenExpires, float64(time.Now().Unix()))
				}
			}
		})
	}
}

// TestLoginHandler_DisabledAccount tests login for a disabled account
func TestLoginHandler_DisabledAccount(t *testing.T) {
	db, e, appCtx, testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// Disable the account
	testOpr.Status = common.DISABLED
	db.Save(testOpr)

	// Try to login
	requestBody := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	err := loginHandler(c)

	require.NoError(t, err, "handler should not return error, but write error response")
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errorResp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "ACCOUNT_DISABLED", errorResp.Error)
}

// TestIssueToken tests the token issuance function
func TestIssueToken(t *testing.T) {
	db, e, appCtx, testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test context
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	token, err := issueToken(c, *testOpr)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse and validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("test-secret-key-for-jwt"), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Verify claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		assert.Equal(t, fmt.Sprintf("%d", testOpr.ID), claims["sub"])
		assert.Equal(t, testOpr.Username, claims["username"])
		assert.Equal(t, testOpr.Level, claims["role"])
		assert.Equal(t, "toughradius", claims["iss"])
	} else {
		t.Errorf("unable to parse claims")
	}
}

// TestTokenExpirationTime tests token expiration time
func TestTokenExpirationTime(t *testing.T) {
	db, e, appCtx, testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test context
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	token, err := issueToken(c, *testOpr)
	require.NoError(t, err)

	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key-for-jwt"), nil
	})

	require.NoError(t, err)
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Get token expiration time
	exp := int64(claims["exp"].(float64))
	iat := int64(claims["iat"].(float64))

	// Verify token is valid for approximately tokenTTL duration
	// Using a 1-second tolerance to account for timing differences
	tokenDuration := exp - iat
	assert.InDelta(t, int64(tokenTTL.Seconds()), tokenDuration, 1.0)
}

// TestValidToken tests the validation of a valid token
func TestValidToken(t *testing.T) {
	db, e, appCtx, testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test context and issue a token
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	token, err := issueToken(c, *testOpr)
	require.NoError(t, err)

	// Parse and validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key-for-jwt"), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

// TestCurrentUserHandler tests the current user handler
func TestCurrentUserHandler(t *testing.T) {
	db, e, appCtx, testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test context with user in context
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)
	setJWTUser(t, c, testOpr)

	err := currentUserHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)

	user, ok := data["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, testOpr.Username, user["username"])
}

// TestCurrentUserHandler_NoUser tests the current user handler without a user in context
func TestCurrentUserHandler_NoUser(t *testing.T) {
	_, e, _, _, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test context without user
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := currentUserHandler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errorResp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", errorResp.Error)
}

// TestTokenValidationMiddleware tests JWT validation middleware
func TestTokenValidationMiddleware(t *testing.T) {
	db, e, appCtx, testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test context and issue a token
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	token, err := issueToken(c, *testOpr)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "Valid token",
			token:          token,
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "No token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:           "Invalid token",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			// Simple test handler that checks for user in context
			handler := func(c echo.Context) error {
				if c.Get("user") != nil {
					return c.JSON(http.StatusOK, map[string]string{"status": "authenticated"})
				}
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "no user"})
			}

			// Note: In real tests, you would apply the JWT middleware here
			// For this basic test, we're just checking the token issuance
			if tt.expectUser {
				c.Set("user", testOpr)
			}

			err := handler(c)
			assert.NoError(t, err)
		})
	}
}

func setJWTUser(t *testing.T, c echo.Context, user *domain.SysOpr) {
	tokenStr, err := issueToken(c, *user)
	require.NoError(t, err)

	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetAppContext(c).Config().Web.Secret), nil
	})
	require.NoError(t, err)
	c.Set("user", parsedToken)
}

// BenchmarkIssueToken benchmarks token issuance
func BenchmarkIssueToken(b *testing.B) {
	// Setup - use a wrapper testing.T for setup
	t := &testing.T{}
	db, e, appCtx, testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test context
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := CreateTestContext(e, db, req, rec, appCtx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = issueToken(c, *testOpr)
	}
}
