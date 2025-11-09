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
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	customValidator "github.com/talkincode/toughradius/v9/internal/pkg/validator"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestEcho 创建带验证器的 Echo 实例
func setupTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = customValidator.NewValidator()
	return e
}

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(
		&domain.RadiusProfile{},
		&domain.RadiusUser{},
		&domain.NetNode{},
		&domain.NetNas{},
	)
	require.NoError(t, err)

	return db
}

// setupTestApp 初始化测试应用
func setupTestApp(t *testing.T, db *gorm.DB) {
	cfg := &config.AppConfig{
		System: config.SysConfig{
			Appid:    "TestApp",
			Location: "Asia/Shanghai",
			Workdir:  "/tmp/test",
			Debug:    true,
		},
	}
	testApp := app.NewApplication(cfg)

	// 使用反射设置私有字段 db
	// 注意: 这里需要访问 app 包的内部结构
	// 在实际测试中，可能需要在 app 包中提供 SetDB 方法
	app.SetGApp(testApp)
	app.SetGDB(db)
}

// createTestProfile 创建测试 Profile 数据
func createTestProfile(db *gorm.DB, name string) *domain.RadiusProfile {
	profile := &domain.RadiusProfile{
		Name:       name,
		Status:     "enabled",
		AddrPool:   "192.168.1.0/24",
		ActiveNum:  1,
		UpRate:     10240,
		DownRate:   20480,
		Domain:     "test.com",
		IPv6Prefix: "2001:db8::/64",
		BindMac:    0,
		BindVlan:   0,
		Remark:     "Test profile",
		NodeId:     1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	db.Create(profile)
	return profile
}

func TestListProfiles(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	createTestProfile(db, "profile1")
	createTestProfile(db, "profile2")
	createTestProfile(db, "profile3")

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "获取所有 profiles - 默认分页",
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
			name:           "分页查询 - 第2页",
			queryParams:    "?page=2&perPage=2",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 2, resp.Meta.Page)
			},
		},
		{
			name:           "按名称搜索",
			queryParams:    "?name=profile1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				profiles := resp.Data.([]domain.RadiusProfile)
				assert.Equal(t, "profile1", profiles[0].Name)
			},
		},
		{
			name:           "按状态过滤",
			queryParams:    "?status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "排序测试 - ASC",
			queryParams:    "?sort=name&order=ASC",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkResponse: func(t *testing.T, resp *Response) {
				profiles := resp.Data.([]domain.RadiusProfile)
				assert.Equal(t, "profile1", profiles[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := ListProfiles(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// 将 data 转换为 profile 数组
			dataBytes, _ := json.Marshal(response.Data)
			var profiles []domain.RadiusProfile
			json.Unmarshal(dataBytes, &profiles)
			response.Data = profiles

			assert.Len(t, profiles, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	profile := createTestProfile(db, "test-profile")

	tests := []struct {
		name           string
		profileID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "获取存在的 profile",
			profileID:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "获取不存在的 profile",
			profileID:      "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			profileID:      "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles/"+tt.profileID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			err := GetProfile(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultProfile domain.RadiusProfile
				json.Unmarshal(dataBytes, &resultProfile)

				assert.Equal(t, profile.Name, resultProfile.Name)
				assert.Equal(t, profile.Status, resultProfile.Status)
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateProfile(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.RadiusProfile)
	}{
		{
			name: "成功创建 profile",
			requestBody: `{
				"name": "new-profile",
				"status": "enabled",
				"addr_pool": "192.168.1.0/24",
				"active_num": 1,
				"up_rate": 10240,
				"down_rate": 20480,
				"domain": "test.com",
				"node_id": "1"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, profile *domain.RadiusProfile) {
				assert.Equal(t, "new-profile", profile.Name)
				assert.Equal(t, "enabled", profile.Status)
				assert.Equal(t, "192.168.1.0/24", profile.AddrPool)
				assert.Equal(t, 10240, profile.UpRate)
			},
		},
		{
			name: "创建时状态为空 - 使用默认值",
			requestBody: `{
				"name": "default-status-profile",
				"addr_pool": "10.0.0.0/24"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, profile *domain.RadiusProfile) {
				assert.Equal(t, "enabled", profile.Status)
			},
		},
		{
			name:           "缺少必填字段 - 名称",
			requestBody:    `{"status": "enabled"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_NAME",
		},
		{
			name: "名称已存在",
			requestBody: `{
				"name": "duplicate-profile",
				"status": "enabled"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NAME_EXISTS",
		},
		{
			name:           "无效的 JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "前端格式 - boolean 类型的 status 和 bind 字段",
			requestBody: `{
				"status": true,
				"active_num": 1,
				"up_rate": 1024,
				"down_rate": 1024,
				"bind_mac": false,
				"bind_vlan": false,
				"name": "default"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, profile *domain.RadiusProfile) {
				assert.Equal(t, "default", profile.Name)
				assert.Equal(t, "enabled", profile.Status) // true -> "enabled"
				assert.Equal(t, 1024, profile.UpRate)
				assert.Equal(t, 1024, profile.DownRate)
				assert.Equal(t, 0, profile.BindMac)  // false -> 0
				assert.Equal(t, 0, profile.BindVlan) // false -> 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为重复名称测试创建已存在的 profile
			if tt.name == "名称已存在" {
				createTestProfile(db, "duplicate-profile")
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/radius-profiles", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := CreateProfile(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var profile domain.RadiusProfile
				json.Unmarshal(dataBytes, &profile)

				assert.NotZero(t, profile.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &profile)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	_ = createTestProfile(db, "original-profile")
	createTestProfile(db, "another-profile")

	tests := []struct {
		name           string
		profileID      string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.RadiusProfile)
	}{
		{
			name:      "成功更新 profile",
			profileID: "1",
			requestBody: `{
				"name": "updated-profile",
				"up_rate": 20480,
				"down_rate": 40960
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, p *domain.RadiusProfile) {
				assert.Equal(t, "updated-profile", p.Name)
				assert.Equal(t, 20480, p.UpRate)
				assert.Equal(t, 40960, p.DownRate)
			},
		},
		{
			name:      "部分更新 - 只更新状态",
			profileID: "1",
			requestBody: `{
				"status": "disabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, p *domain.RadiusProfile) {
				assert.Equal(t, "disabled", p.Status)
			},
		},
		{
			name:      "名称冲突",
			profileID: "1",
			requestBody: `{
				"name": "another-profile"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NAME_EXISTS",
		},
		{
			name:           "profile 不存在",
			profileID:      "999",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			profileID:      "invalid",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:      "前端格式 - boolean 类型更新",
			profileID: "1",
			requestBody: `{
				"status": false,
				"bind_mac": true,
				"bind_vlan": true
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, p *domain.RadiusProfile) {
				assert.Equal(t, "disabled", p.Status) // false -> "disabled"
				assert.Equal(t, 1, p.BindMac)         // true -> 1
				assert.Equal(t, 1, p.BindVlan)        // true -> 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/radius-profiles/"+tt.profileID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			err := UpdateProfile(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedProfile domain.RadiusProfile
				json.Unmarshal(dataBytes, &updatedProfile)

				if tt.checkResult != nil {
					tt.checkResult(t, &updatedProfile)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteProfile(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	_ = createTestProfile(db, "profile-to-delete")
	profile2 := createTestProfile(db, "profile-in-use")

	// 创建一个使用 profile2 的用户
	user := &domain.RadiusUser{
		Username:  "testuser",
		Password:  "testpass",
		ProfileId: profile2.ID,
		NodeId:    1,
		ActiveNum: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(user)

	tests := []struct {
		name           string
		profileID      string
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:           "成功删除未使用的 profile",
			profileID:      "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "无法删除正在使用的 profile",
			profileID:      "2",
			expectedStatus: http.StatusConflict,
			expectedError:  "IN_USE",
		},
		{
			name:           "profile 不存在",
			profileID:      "999",
			expectedStatus: http.StatusOK, // GORM Delete 不会返回错误
			checkDeleted:   false,
		},
		{
			name:           "无效的 ID",
			profileID:      "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/radius-profiles/"+tt.profileID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			err := DeleteProfile(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.checkDeleted {
					// 验证 profile 已被删除
					var count int64
					db.Model(&domain.RadiusProfile{}).Where("id = ?", tt.profileID).Count(&count)
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

// TestProfileEdgeCases 测试边缘情况
func TestProfileEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	t.Run("超大分页参数", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles?perPage=1000", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListProfiles(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		// perPage 应该被限制为 100
		assert.LessOrEqual(t, response.Meta.PageSize, 100)
	})

	t.Run("负数分页参数", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles?page=-1&perPage=-10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListProfiles(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		// 应该使用默认值
		assert.Equal(t, 1, response.Meta.Page)
		assert.Equal(t, 10, response.Meta.PageSize)
	})

	t.Run("无效的排序方向", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles?order=INVALID", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListProfiles(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("更新不存在的字段不应影响其他字段", func(t *testing.T) {
		profile := createTestProfile(db, "test-profile")
		originalName := profile.Name

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/radius-profiles/1", strings.NewReader(`{"up_rate": 30720}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := UpdateProfile(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedProfile domain.RadiusProfile
		json.Unmarshal(dataBytes, &updatedProfile)

		// 名称不应该改变
		assert.Equal(t, originalName, updatedProfile.Name)
		// up_rate 应该更新
		assert.Equal(t, 30720, updatedProfile.UpRate)
	})
}
