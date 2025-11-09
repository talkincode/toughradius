package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// createTestOperator 创建测试操作员数据
func createTestOperator(db *gorm.DB, username string, level string) *domain.SysOpr {
	opr := &domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  username,
		Password:  common.Sha256HashWithSalt("Password123", common.SecretSalt),
		Realname:  "Test Operator",
		Mobile:    "13800138000",
		Email:     "test@example.com",
		Level:     level,
		Status:    common.ENABLED,
		Remark:    "Test operator",
		LastLogin: time.Time{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(opr)
	return opr
}

func TestListOperators(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 自动迁移操作员表
	db.AutoMigrate(&domain.SysOpr{})

	// 创建测试数据
	createTestOperator(db, "admin1", "admin")
	createTestOperator(db, "operator1", "operator")
	createTestOperator(db, "superadmin", "super")

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "获取所有操作员 - 默认分页",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 1, resp.Meta.Page)
			},
		},
		{
			name:           "分页查询 - 第1页",
			queryParams:    "?page=1&pageSize=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 1, resp.Meta.Page)
				assert.Equal(t, 2, resp.Meta.PageSize)
			},
		},
		{
			name:           "按用户名搜索",
			queryParams:    "?username=admin1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				operators := resp.Data.([]domain.SysOpr)
				assert.Equal(t, "admin1", operators[0].Username)
				assert.Empty(t, operators[0].Password) // 密码应被清空
			},
		},
		{
			name:           "按真实姓名搜索",
			queryParams:    "?realname=Test",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "按权限级别过滤",
			queryParams:    "?level=super",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				operators := resp.Data.([]domain.SysOpr)
				assert.Equal(t, "super", operators[0].Level)
			},
		},
		{
			name:           "按状态过滤",
			queryParams:    "?status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/system/operators"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := listOperators(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// 将 data 转换为 operator 数组
			dataBytes, _ := json.Marshal(response.Data)
			var operators []domain.SysOpr
			json.Unmarshal(dataBytes, &operators)
			response.Data = operators

			assert.Len(t, operators, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetOperator(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	// 创建测试数据
	opr := createTestOperator(db, "testadmin", "admin")

	tests := []struct {
		name           string
		operatorID     int64
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "获取存在的操作员",
			operatorID:     opr.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "获取不存在的操作员",
			operatorID:     999999,
			expectedStatus: http.StatusNotFound,
			expectedError:  "OPERATOR_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/system/operators/"+strconv.FormatInt(tt.operatorID, 10), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(tt.operatorID, 10))

			err := getOperator(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultOpr domain.SysOpr
				json.Unmarshal(dataBytes, &resultOpr)

				assert.Equal(t, opr.Username, resultOpr.Username)
				assert.Empty(t, resultOpr.Password) // 密码应被清空
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateOperator(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.SysOpr)
	}{
		{
			name: "成功创建操作员",
			requestBody: `{
				"username": "newadmin",
				"password": "Password123",
				"realname": "New Admin",
				"mobile": "13900139000",
				"email": "newadmin@example.com",
				"level": "admin",
				"status": "enabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, opr *domain.SysOpr) {
				assert.Equal(t, "newadmin", opr.Username)
				assert.Equal(t, "New Admin", opr.Realname)
				assert.Equal(t, "admin", opr.Level)
				assert.Equal(t, "enabled", opr.Status)
				assert.Empty(t, opr.Password) // 返回时密码应被清空
			},
		},
		{
			name: "创建时使用默认权限级别",
			requestBody: `{
				"username": "defaultoper",
				"password": "Password123",
				"realname": "Default Operator",
				"mobile": "13800138001"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, opr *domain.SysOpr) {
				assert.Equal(t, "operator", opr.Level) // 默认为 operator
				assert.Equal(t, "enabled", opr.Status) // 默认为 enabled
			},
		},
		{
			name: "用户名太短",
			requestBody: `{
				"username": "ab",
				"password": "Password123",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USERNAME",
		},
		{
			name: "用户名太长",
			requestBody: `{
				"username": "abcdefghijklmnopqrstuvwxyz12345",
				"password": "Password123",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USERNAME",
		},
		{
			name:           "缺少用户名",
			requestBody:    `{"password": "Password123", "realname": "Test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_USERNAME",
		},
		{
			name:           "缺少密码",
			requestBody:    `{"username": "testuser", "realname": "Test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_PASSWORD",
		},
		{
			name:           "缺少真实姓名",
			requestBody:    `{"username": "testuser", "password": "Password123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_REALNAME",
		},
		{
			name: "密码太短",
			requestBody: `{
				"username": "testuser",
				"password": "12345",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PASSWORD",
		},
		{
			name: "密码太长",
			requestBody: `{
				"username": "testuser",
				"password": "` + strings.Repeat("a", 51) + `",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PASSWORD",
		},
		{
			name: "密码强度不足 - 只有数字",
			requestBody: `{
				"username": "testuser",
				"password": "123456789",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "WEAK_PASSWORD",
		},
		{
			name: "密码强度不足 - 只有字母",
			requestBody: `{
				"username": "testuser",
				"password": "abcdefgh",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "WEAK_PASSWORD",
		},
		{
			name: "无效的邮箱格式",
			requestBody: `{
				"username": "testuser",
				"password": "Password123",
				"realname": "Test",
				"email": "invalid-email"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_EMAIL",
		},
		{
			name: "无效的手机号格式",
			requestBody: `{
				"username": "testuser",
				"password": "Password123",
				"realname": "Test",
				"mobile": "12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_MOBILE",
		},
		{
			name: "无效的权限级别",
			requestBody: `{
				"username": "testuser",
				"password": "Password123",
				"realname": "Test",
				"level": "invalid"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_LEVEL",
		},
		{
			name: "用户名已存在",
			requestBody: `{
				"username": "duplicateuser",
				"password": "Password123",
				"realname": "Duplicate User"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name:           "无效的 JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "用户名自动去除空格",
			requestBody: `{
				"username": "  spaceuser  ",
				"password": "Password123",
				"realname": "Space User"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, opr *domain.SysOpr) {
				assert.Equal(t, "spaceuser", opr.Username)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为用户名已存在测试创建已存在的操作员
			if tt.name == "用户名已存在" {
				createTestOperator(db, "duplicateuser", "operator")
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := createOperator(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var opr domain.SysOpr
				json.Unmarshal(dataBytes, &opr)

				assert.NotZero(t, opr.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &opr)
				}
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateOperator(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	// 创建测试数据
	opr := createTestOperator(db, "originaluser", "operator")
	createTestOperator(db, "anotheruser", "admin")

	tests := []struct {
		name           string
		operatorID     int64
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.SysOpr)
	}{
		{
			name:       "成功更新操作员",
			operatorID: opr.ID,
			requestBody: `{
				"realname": "Updated Name",
				"mobile": "13911139111"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "Updated Name", o.Realname)
				assert.Equal(t, "13911139111", o.Mobile)
			},
		},
		{
			name:       "更新权限级别",
			operatorID: opr.ID,
			requestBody: `{
				"level": "admin"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "admin", o.Level)
			},
		},
		{
			name:       "更新密码",
			operatorID: opr.ID,
			requestBody: `{
				"password": "NewPass456"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Empty(t, o.Password) // 返回时应被清空
				// 验证密码已更新（需要从数据库重新查询）
				var updatedOpr domain.SysOpr
				db.Where("id = ?", opr.ID).First(&updatedOpr)
				expectedHash := common.Sha256HashWithSalt("NewPass456", common.SecretSalt)
				assert.Equal(t, expectedHash, updatedOpr.Password)
			},
		},
		{
			name:       "部分更新 - 只更新状态",
			operatorID: opr.ID,
			requestBody: `{
				"status": "disabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "disabled", o.Status)
			},
		},
		{
			name:       "用户名冲突",
			operatorID: opr.ID,
			requestBody: `{
				"username": "anotheruser"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name:           "操作员不存在",
			operatorID:     999999,
			requestBody:    `{"realname": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "OPERATOR_NOT_FOUND",
		},
		{
			name:       "无效的用户名 - 太短",
			operatorID: opr.ID,
			requestBody: `{
				"username": "ab"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USERNAME",
		},
		{
			name:       "无效的密码 - 太短",
			operatorID: opr.ID,
			requestBody: `{
				"password": "12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PASSWORD",
		},
		{
			name:       "密码强度不足",
			operatorID: opr.ID,
			requestBody: `{
				"password": "123456789"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "WEAK_PASSWORD",
		},
		{
			name:       "无效的邮箱格式",
			operatorID: opr.ID,
			requestBody: `{
				"email": "invalid-email"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_EMAIL",
		},
		{
			name:       "无效的手机号格式",
			operatorID: opr.ID,
			requestBody: `{
				"mobile": "12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_MOBILE",
		},
		{
			name:       "更新邮箱和手机号",
			operatorID: opr.ID,
			requestBody: `{
				"email": "newemail@example.com",
				"mobile": "13922229222"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "newemail@example.com", o.Email)
				assert.Equal(t, "13922229222", o.Mobile)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/system/operators/"+strconv.FormatInt(tt.operatorID, 10), strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(tt.operatorID, 10))

			err := updateOperator(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedOpr domain.SysOpr
				json.Unmarshal(dataBytes, &updatedOpr)

				assert.Empty(t, updatedOpr.Password) // 密码应被清空
				if tt.checkResult != nil {
					tt.checkResult(t, &updatedOpr)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestOperatorEdgeCases 测试边缘情况
func TestOperatorEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	t.Run("用户名自动去除空格", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "  spaceadmin  ",
			"password": "Password123",
			"realname": "Space Admin"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		assert.Equal(t, "spaceadmin", opr.Username)
	})

	t.Run("密码去除空格", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "testadmin2",
			"password": "  Password123  ",
			"realname": "Test Admin 2"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createOperator(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		// 验证密码已正确存储（去除空格后加密）
		var savedOpr domain.SysOpr
		db.Where("id = ?", opr.ID).First(&savedOpr)
		expectedHash := common.Sha256HashWithSalt("Password123", common.SecretSalt)
		assert.Equal(t, expectedHash, savedOpr.Password)
	})

	t.Run("更新不存在的字段不应影响其他字段", func(t *testing.T) {
		opr := createTestOperator(db, "testadmin3", "admin")
		originalUsername := opr.Username
		originalLevel := opr.Level

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/system/operators/"+strconv.FormatInt(opr.ID, 10), strings.NewReader(`{"realname": "New Name"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(opr.ID, 10))

		err := updateOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedOpr domain.SysOpr
		json.Unmarshal(dataBytes, &updatedOpr)

		// 用户名和权限级别不应该改变
		assert.Equal(t, originalUsername, updatedOpr.Username)
		assert.Equal(t, originalLevel, updatedOpr.Level)
		// realname 应该更新
		assert.Equal(t, "New Name", updatedOpr.Realname)
	})

	t.Run("权限级别大小写不敏感", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "casetest",
			"password": "Password123",
			"realname": "Case Test",
			"level": "ADMIN"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		assert.Equal(t, "admin", opr.Level) // 应该转换为小写
	})

	t.Run("状态大小写不敏感", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "statustest",
			"password": "Password123",
			"realname": "Status Test",
			"status": "DISABLED"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		assert.Equal(t, "disabled", opr.Status) // 应该转换为小写
	})

	t.Run("验证有效的中国手机号", func(t *testing.T) {
		validMobiles := []string{
			"13800138000",
			"15912345678",
			"18612345678",
			"+8613800138000",
			"8613800138000",
		}

		for _, mobile := range validMobiles {
			e := setupTestEcho()
			requestBody := `{
				"username": "mobile` + mobile + `",
				"password": "Password123",
				"realname": "Mobile Test",
				"mobile": "` + mobile + `"
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := createOperator(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code, "手机号 %s 应该是有效的", mobile)
		}
	})

	t.Run("验证有效的邮箱", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"admin@test.com",
		}

		for i, email := range validEmails {
			e := setupTestEcho()
			username := "emailtest" + strconv.Itoa(i)
			requestBody := `{
				"username": "` + username + `",
				"password": "Password123",
				"realname": "Email Test",
				"email": "` + email + `"
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := createOperator(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code, "邮箱 %s 应该是有效的", email)
		}
	})
}
