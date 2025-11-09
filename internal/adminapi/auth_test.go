package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupAuthTestDB 创建测试数据库
func setupAuthTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移操作员表
	err = db.AutoMigrate(&domain.SysOpr{})
	require.NoError(t, err)

	return db
}

// setupAuthTestApp 初始化测试应用
func setupAuthTestApp(t *testing.T, db *gorm.DB) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    true,
		},
		Web: config.WebConfig{
			Secret: "test-secret-key-for-jwt",
		},
	}
	testApp := app.NewApplication(cfg)
	app.SetGApp(testApp)
	app.SetGDB(db)
}

// setupAuthTest 初始化测试环境并创建测试用户
func setupAuthTest(t *testing.T) (*domain.SysOpr, func()) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	// 创建测试用户
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
	app.GDB().Create(testOpr)

	cleanup := func() {
		app.GDB().Where("id = ?", testOpr.ID).Delete(&domain.SysOpr{})
	}

	return testOpr, cleanup
}

// TestLoginHandler 测试登录处理函数
func TestLoginHandler(t *testing.T) {
	_, cleanup := setupAuthTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedCode   string
		checkToken     bool
	}{
		{
			name:           "成功登录",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
			checkToken:     true,
		},
		{
			name:           "无效的JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
			checkToken:     false,
		},
		{
			name:           "空用户名",
			requestBody:    `{"username":"","password":"password123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "空密码",
			requestBody:    `{"username":"testuser","password":""}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "用户不存在",
			requestBody:    `{"username":"nonexistent","password":"password123"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "密码错误",
			requestBody:    `{"username":"testuser","password":"wrongpassword"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "INVALID_CREDENTIALS",
			checkToken:     false,
		},
		{
			name:           "用户名前后有空格（应被trim）",
			requestBody:    `{"username":"  testuser  ","password":"password123"}`,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
			checkToken:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := loginHandler(c)

			// fail() 函数调用 c.JSON() 并返回nil，所以我们检查响应码而不是错误
			if tt.expectedStatus >= 400 {
				// 错误情况：检查响应状态码和错误信息
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

					// 检查 token
					token, ok := data["token"].(string)
					assert.True(t, ok, "data should have token field")
					assert.NotEmpty(t, token, "token should not be empty")

					// 检查 user
					user, ok := data["user"].(map[string]interface{})
					assert.True(t, ok, "data should have user field")
					assert.Equal(t, "testuser", user["username"])
					assert.Empty(t, user["password"], "password should be empty in response")

					// 检查 tokenExpires
					tokenExpires, ok := data["tokenExpires"].(float64)
					assert.True(t, ok, "data should have tokenExpires field")
					assert.Greater(t, tokenExpires, float64(time.Now().Unix()))
				}
			}
		})
	}
}

// TestLoginHandler_DisabledAccount 测试禁用账号登录
func TestLoginHandler_DisabledAccount(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	// 创建禁用的测试用户
	disabledOpr := &domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  "disableduser",
		Password:  common.Sha256HashWithSalt("password123", common.SecretSalt),
		Realname:  "Disabled User",
		Email:     "disabled@example.com",
		Level:     "operator",
		Status:    common.DISABLED,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	app.GDB().Create(disabledOpr)
	defer app.GDB().Where("id = ?", disabledOpr.ID).Delete(&domain.SysOpr{})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login",
		strings.NewReader(`{"username":"disableduser","password":"password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := loginHandler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errorResp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "ACCOUNT_DISABLED", errorResp.Error)
}

// TestIssueToken 测试 JWT token 生成
func TestIssueToken(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	testOpr := domain.SysOpr{
		ID:       12345,
		Username: "testuser",
		Level:    "super",
	}

	token, err := issueToken(testOpr)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 解析 token 验证内容
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(app.GConfig().Web.Secret), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	// 验证 claims
	assert.Equal(t, "12345", claims["sub"])
	assert.Equal(t, "testuser", claims["username"])
	assert.Equal(t, "super", claims["role"])
	assert.Equal(t, "toughradius", claims["iss"])

	// 验证时间字段
	exp, ok := claims["exp"].(float64)
	assert.True(t, ok)
	assert.Greater(t, exp, float64(time.Now().Unix()))

	iat, ok := claims["iat"].(float64)
	assert.True(t, ok)
	assert.LessOrEqual(t, iat, float64(time.Now().Unix()))
}

// TestCurrentUserHandler 测试获取当前用户信息
func TestCurrentUserHandler(t *testing.T) {
	testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// 生成有效 token
	token, err := issueToken(*testOpr)
	assert.NoError(t, err)

	// 解析 token
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(app.GConfig().Web.Secret), nil
	})
	assert.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 模拟 JWT 中间件设置的 context
	c.Set("user", parsedToken)

	err = currentUserHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)

	user, ok := data["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testuser", user["username"])
	assert.Empty(t, user["password"], "password should be empty")
}

// TestCurrentUserHandler_NoToken 测试无 token 情况
func TestCurrentUserHandler_NoToken(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 不设置 user context

	err := currentUserHandler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errorResp ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", errorResp.Error)
}

// TestResolveOperatorFromContext 测试从 context 解析操作员
func TestResolveOperatorFromContext(t *testing.T) {
	testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// 生成有效 token
	token, err := issueToken(*testOpr)
	assert.NoError(t, err)

	// 解析 token
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(app.GConfig().Web.Secret), nil
	})
	assert.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", parsedToken)

	operator, err := resolveOperatorFromContext(c)
	assert.NoError(t, err)
	assert.NotNil(t, operator)
	assert.Equal(t, testOpr.ID, operator.ID)
	assert.Equal(t, testOpr.Username, operator.Username)
	assert.Empty(t, operator.Password, "password should be empty")
}

// TestResolveOperatorFromContext_NoUser 测试无用户上下文
func TestResolveOperatorFromContext_NoUser(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	operator, err := resolveOperatorFromContext(c)
	assert.Error(t, err)
	assert.Nil(t, operator)
	assert.Contains(t, err.Error(), "no user in context")
}

// TestResolveOperatorFromContext_InvalidTokenType 测试无效的 token 类型
func TestResolveOperatorFromContext_InvalidTokenType(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", "not a token") // 设置错误类型

	operator, err := resolveOperatorFromContext(c)
	assert.Error(t, err)
	assert.Nil(t, operator)
	assert.Contains(t, err.Error(), "invalid token type")
}

// TestResolveOperatorFromContext_InvalidClaims 测试无效的 claims
func TestResolveOperatorFromContext_InvalidClaims(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	// 创建一个没有 sub claim 的 token
	now := time.Now()
	claims := jwt.MapClaims{
		"username": "testuser",
		"exp":      now.Add(tokenTTL).Unix(),
		"iat":      now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", token)

	operator, err := resolveOperatorFromContext(c)
	assert.Error(t, err)
	assert.Nil(t, operator)
	assert.Contains(t, err.Error(), "invalid token subject")
}

// TestResolveOperatorFromContext_UserNotFound 测试用户不存在
func TestResolveOperatorFromContext_UserNotFound(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	// 创建一个指向不存在用户的 token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      "999999999", // 不存在的用户 ID
		"username": "nonexistent",
		"role":     "operator",
		"exp":      now.Add(tokenTTL).Unix(),
		"iat":      now.Unix(),
		"iss":      "toughradius",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", token)

	operator, err := resolveOperatorFromContext(c)
	assert.Error(t, err)
	assert.Nil(t, operator)
}

// TestResolveOperatorFromContext_InvalidSubFormat 测试无效的 sub 格式
func TestResolveOperatorFromContext_InvalidSubFormat(t *testing.T) {
	db := setupAuthTestDB(t)
	setupAuthTestApp(t, db)

	// 创建一个 sub 不是数字的 token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      "not-a-number",
		"username": "testuser",
		"exp":      now.Add(tokenTTL).Unix(),
		"iat":      now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", token)

	operator, err := resolveOperatorFromContext(c)
	assert.Error(t, err)
	assert.Nil(t, operator)
	assert.Contains(t, err.Error(), "invalid token id")
}

// TestLoginHandler_LastLoginUpdate 测试登录成功后更新最后登录时间
func TestLoginHandler_LastLoginUpdate(t *testing.T) {
	testOpr, cleanup := setupAuthTest(t)
	defer cleanup()

	// 记录登录前的时间
	beforeLogin := time.Now()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login",
		strings.NewReader(`{"username":"testuser","password":"password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := loginHandler(c)
	assert.NoError(t, err)

	// 等待 goroutine 完成更新
	time.Sleep(100 * time.Millisecond)

	// 重新查询用户，检查 last_login
	var updatedOpr domain.SysOpr
	app.GDB().Where("id = ?", testOpr.ID).First(&updatedOpr)

	// last_login 应该在登录之后
	assert.True(t, updatedOpr.LastLogin.After(beforeLogin) || updatedOpr.LastLogin.Equal(beforeLogin))
}

// TestTokenTTL 测试 token 过期时间
func TestTokenTTL(t *testing.T) {
	assert.Equal(t, 12*time.Hour, tokenTTL, "token TTL should be 12 hours")
}

// BenchmarkLoginHandler 登录处理性能基准测试
func BenchmarkLoginHandler(b *testing.B) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	if err := db.AutoMigrate(&domain.SysOpr{}); err != nil {
		b.Fatal(err)
	}

	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    false,
		},
		Web: config.WebConfig{
			Secret: "test-secret-key-for-jwt",
		},
	}
	testApp := app.NewApplication(cfg)
	app.SetGApp(testApp)
	app.SetGDB(db)

	testOpr := &domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  "benchuser",
		Password:  common.Sha256HashWithSalt("password123", common.SecretSalt),
		Realname:  "Bench User",
		Level:     "operator",
		Status:    common.ENABLED,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	app.GDB().Create(testOpr)
	defer app.GDB().Where("id = ?", testOpr.ID).Delete(&domain.SysOpr{})

	requestBody := `{"username":"benchuser","password":"password123"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = loginHandler(c)
	}
}

// BenchmarkIssueToken token 生成性能基准测试
func BenchmarkIssueToken(b *testing.B) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    false,
		},
		Web: config.WebConfig{
			Secret: "test-secret-key-for-jwt",
		},
	}
	testApp := app.NewApplication(cfg)
	app.SetGApp(testApp)

	testOpr := domain.SysOpr{
		ID:       12345,
		Username: "benchuser",
		Level:    "operator",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = issueToken(testOpr)
	}
}
