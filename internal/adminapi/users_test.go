package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// createTestUser 创建测试 User 数据
func createTestUser(db *gorm.DB, username string, profileID int64) *domain.RadiusUser {
	user := &domain.RadiusUser{
		Username:   username,
		Password:   "testpass123",
		ProfileId:  profileID,
		Realname:   "Test User",
		Mobile:     "13800138000",
		Status:     "enabled",
		NodeId:     1,
		ActiveNum:  1,
		UpRate:     1024,
		DownRate:   2048,
		AddrPool:   "192.168.1.0/24",
		IpAddr:     "",
		MacAddr:    "",
		Vlanid1:    0,
		Vlanid2:    0,
		BindMac:    0,
		BindVlan:   0,
		ExpireTime: time.Now().AddDate(1, 0, 0),
		Remark:     "Test user",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	db.Create(user)
	return user
}

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试 Profile
	profile := createTestProfile(db, "test-profile")

	// 创建测试数据
	createTestUser(db, "user1", profile.ID)
	createTestUser(db, "user2", profile.ID)
	createTestUser(db, "user3", profile.ID)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "获取所有 users - 默认分页",
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
			queryParams:    "?page=1&perPage=2",
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
			queryParams:    "?q=user1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				users := resp.Data.([]domain.RadiusUser)
				assert.Equal(t, "user1", users[0].Username)
				assert.Empty(t, users[0].Password) // 密码应被清空
			},
		},
		{
			name:           "按状态过滤",
			queryParams:    "?status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "按 profile_id 过滤",
			queryParams:    "?profile_id=" + string(rune(profile.ID)),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := listRadiusUsers(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// 将 data 转换为 user 数组
			dataBytes, _ := json.Marshal(response.Data)
			var users []domain.RadiusUser
			json.Unmarshal(dataBytes, &users)
			response.Data = users

			assert.Len(t, users, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	profile := createTestProfile(db, "test-profile")
	user := createTestUser(db, "testuser", profile.ID)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "获取存在的 user",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "获取不存在的 user",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			err := getRadiusUser(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultUser domain.RadiusUser
				json.Unmarshal(dataBytes, &resultUser)

				assert.Equal(t, user.Username, resultUser.Username)
				assert.Empty(t, resultUser.Password) // 密码应被清空
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试 Profile
	profile := createTestProfile(db, "test-profile")

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.RadiusUser)
	}{
		{
			name: "成功创建 user",
			requestBody: `{
				"username": "newuser",
				"password": "password123",
				"profile_id": "` + string(rune(profile.ID)) + `",
				"realname": "New User",
				"mobile": "13900139000",
				"status": "enabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, user *domain.RadiusUser) {
				assert.Equal(t, "newuser", user.Username)
				assert.Equal(t, "New User", user.Realname)
				assert.Equal(t, "enabled", user.Status)
				assert.Empty(t, user.Password) // 返回时密码应被清空
			},
		},
		{
			name: "创建时状态为空 - 使用默认值",
			requestBody: `{
				"username": "defaultuser",
				"password": "password123",
				"profile_id": "` + string(rune(profile.ID)) + `"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, user *domain.RadiusUser) {
				assert.Equal(t, "enabled", user.Status)
			},
		},
		{
			name:           "缺少必填字段 - 用户名",
			requestBody:    `{"password": "test", "profile_id": "1"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "缺少必填字段 - 密码",
			requestBody:    `{"username": "test", "profile_id": "1"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_PASSWORD",
		},
		{
			name:           "缺少必填字段 - profile_id",
			requestBody:    `{"username": "test", "password": "test123"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "用户名已存在",
			requestBody: `{
				"username": "duplicateuser",
				"password": "password123",
				"profile_id": "` + string(rune(profile.ID)) + `"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name: "关联的 profile 不存在",
			requestBody: `{
				"username": "testuser",
				"password": "password123",
				"profile_id": "999"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "PROFILE_NOT_FOUND",
		},
		{
			name:           "无效的 JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "前端格式 - boolean 类型的 status",
			requestBody: `{
				"username": "booluser",
				"password": "password123",
				"profile_id": "` + string(rune(profile.ID)) + `",
				"status": true
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, user *domain.RadiusUser) {
				assert.Equal(t, "enabled", user.Status) // true -> "enabled"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为重复用户名测试创建已存在的 user
			if tt.name == "用户名已存在" {
				createTestUser(db, "duplicateuser", profile.ID)
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := createRadiusUser(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var user domain.RadiusUser
				json.Unmarshal(dataBytes, &user)

				assert.NotZero(t, user.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &user)
				}
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	profile := createTestProfile(db, "test-profile")
	profile2 := createTestProfile(db, "another-profile")
	_ = createTestUser(db, "originaluser", profile.ID)
	createTestUser(db, "anotheruser", profile.ID)

	tests := []struct {
		name           string
		userID         string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.RadiusUser)
	}{
		{
			name:   "成功更新 user",
			userID: "1",
			requestBody: `{
				"realname": "Updated Name",
				"mobile": "13911139111"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, u *domain.RadiusUser) {
				assert.Equal(t, "Updated Name", u.Realname)
				assert.Equal(t, "13911139111", u.Mobile)
			},
		},
		{
			name:   "更新 profile - 应同步 profile 配置",
			userID: "1",
			requestBody: `{
				"profile_id": "` + string(rune(profile2.ID)) + `"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, u *domain.RadiusUser) {
				assert.Equal(t, profile2.ID, u.ProfileId)
				assert.Equal(t, profile2.ActiveNum, u.ActiveNum)
				assert.Equal(t, profile2.UpRate, u.UpRate)
			},
		},
		{
			name:   "部分更新 - 只更新状态",
			userID: "1",
			requestBody: `{
				"status": "disabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, u *domain.RadiusUser) {
				assert.Equal(t, "disabled", u.Status)
			},
		},
		{
			name:   "用户名冲突",
			userID: "1",
			requestBody: `{
				"username": "anotheruser"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name:           "user 不存在",
			userID:         "999",
			requestBody:    `{"realname": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			userID:         "invalid",
			requestBody:    `{"realname": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:   "更新关联的 profile 不存在",
			userID: "1",
			requestBody: `{
				"profile_id": "999"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "PROFILE_NOT_FOUND",
		},
		{
			name:   "前端格式 - boolean 类型更新",
			userID: "1",
			requestBody: `{
				"status": false,
				"bind_mac": true,
				"bind_vlan": true
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, u *domain.RadiusUser) {
				assert.Equal(t, "disabled", u.Status) // false -> "disabled"
				assert.Equal(t, 1, u.BindMac)         // true -> 1
				assert.Equal(t, 1, u.BindVlan)        // true -> 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+tt.userID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			err := updateRadiusUser(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedUser domain.RadiusUser
				json.Unmarshal(dataBytes, &updatedUser)

				assert.Empty(t, updatedUser.Password) // 密码应被清空
				if tt.checkResult != nil {
					tt.checkResult(t, &updatedUser)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	profile := createTestProfile(db, "test-profile")
	_ = createTestUser(db, "user-to-delete", profile.ID)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:           "成功删除 user",
			userID:         "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "user 不存在",
			userID:         "999",
			expectedStatus: http.StatusOK, // GORM Delete 不会返回错误
			checkDeleted:   false,
		},
		{
			name:           "无效的 ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			err := deleteRadiusUser(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.checkDeleted {
					// 验证 user 已被删除
					var count int64
					db.Model(&domain.RadiusUser{}).Where("id = ?", tt.userID).Count(&count)
					assert.Equal(t, int64(0), count)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestUserEdgeCases 测试边缘情况
func TestUserEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	profile := createTestProfile(db, "test-profile")

	t.Run("用户名自动去除空格", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "  spaceuser  ",
			"password": "password123",
			"profile_id": "` + string(rune(profile.ID)) + `"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createRadiusUser(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var user domain.RadiusUser
		json.Unmarshal(dataBytes, &user)

		assert.Equal(t, "spaceuser", user.Username)
	})

	t.Run("从 profile 继承配置", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "inherituser",
			"password": "password123",
			"profile_id": "` + string(rune(profile.ID)) + `"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createRadiusUser(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var user domain.RadiusUser
		json.Unmarshal(dataBytes, &user)

		// 应该继承 profile 的配置
		assert.Equal(t, profile.ActiveNum, user.ActiveNum)
		assert.Equal(t, profile.UpRate, user.UpRate)
		assert.Equal(t, profile.DownRate, user.DownRate)
		assert.Equal(t, profile.AddrPool, user.AddrPool)
	})

	t.Run("更新不存在的字段不应影响其他字段", func(t *testing.T) {
		user := createTestUser(db, "testuser", profile.ID)
		originalUsername := user.Username

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/1", strings.NewReader(`{"realname": "New Name"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := updateRadiusUser(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedUser domain.RadiusUser
		json.Unmarshal(dataBytes, &updatedUser)

		// 用户名不应该改变
		assert.Equal(t, originalUsername, updatedUser.Username)
		// realname 应该更新
		assert.Equal(t, "New Name", updatedUser.Realname)
	})

	t.Run("过期时间默认值", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "expireuser",
			"password": "password123",
			"profile_id": "` + string(rune(profile.ID)) + `"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createRadiusUser(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var user domain.RadiusUser
		json.Unmarshal(dataBytes, &user)

		// 默认过期时间应该是一年后
		expectedExpire := time.Now().AddDate(1, 0, 0)
		assert.WithinDuration(t, expectedExpire, user.ExpireTime, time.Hour*24)
	})
}
