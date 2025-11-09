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
	"gorm.io/gorm"
)

// createTestNas 创建测试 NAS 数据
func createTestNas(db *gorm.DB, name, ipaddr string) *domain.NetNas {
	nas := &domain.NetNas{
		NodeId:     1,
		Name:       name,
		Identifier: "test-identifier",
		Hostname:   "test.example.com",
		Ipaddr:     ipaddr,
		Secret:     "testsecret123",
		CoaPort:    3799,
		Model:      "test-model",
		VendorCode: "9",
		Status:     "enabled",
		Tags:       "test,nas",
		Remark:     "Test NAS",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	db.Create(nas)
	return nas
}

func TestListNAS(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	createTestNas(db, "nas1", "192.168.1.1")
	createTestNas(db, "nas2", "192.168.1.2")
	createTestNas(db, "nas3", "192.168.1.3")

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "获取所有 NAS - 默认分页",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(3), resp["total"])
			},
		},
		{
			name:           "分页查询 - 第1页",
			queryParams:    "?page=1&perPage=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(3), resp["total"])
			},
		},
		{
			name:           "按名称搜索",
			queryParams:    "?name=nas1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				nasData := data[0].(map[string]interface{})
				assert.Equal(t, "nas1", nasData["name"])
			},
		},
		{
			name:           "按 IP 地址搜索",
			queryParams:    "?ipaddr=192.168.1.1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
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
			req := httptest.NewRequest(http.MethodGet, "/api/v1/nas"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := ListNAS(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			data := response["data"].([]interface{})
			assert.Len(t, data, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestGetNAS(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	nas := createTestNas(db, "test-nas", "192.168.1.100")

	tests := []struct {
		name           string
		nasID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "获取存在的 NAS",
			nasID:          "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "获取不存在的 NAS",
			nasID:          "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			nasID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/"+tt.nasID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.nasID)

			err := GetNAS(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultNas domain.NetNas
				json.Unmarshal(dataBytes, &resultNas)

				assert.Equal(t, nas.Name, resultNas.Name)
				assert.Equal(t, nas.Ipaddr, resultNas.Ipaddr)
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateNAS(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.NetNas)
	}{
		{
			name: "成功创建 NAS",
			requestBody: `{
				"name": "new-nas",
				"ipaddr": "10.0.0.1",
				"secret": "newsecret123",
				"node_id": "1",
				"hostname": "nas.example.com",
				"coa_port": 3799,
				"status": "enabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, nas *domain.NetNas) {
				assert.Equal(t, "new-nas", nas.Name)
				assert.Equal(t, "10.0.0.1", nas.Ipaddr)
				assert.Equal(t, "newsecret123", nas.Secret)
				assert.Equal(t, 3799, nas.CoaPort)
				assert.Equal(t, "enabled", nas.Status)
			},
		},
		{
			name: "创建 NAS - 最小参数",
			requestBody: `{
				"name": "minimal-nas",
				"ipaddr": "10.0.0.2",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, nas *domain.NetNas) {
				assert.Equal(t, "minimal-nas", nas.Name)
				assert.Equal(t, "enabled", nas.Status) // 默认值
				assert.Equal(t, 3799, nas.CoaPort)     // 默认值
			},
		},
		{
			name:           "缺少必填字段 - 名称",
			requestBody:    `{"ipaddr": "10.0.0.3", "secret": "test"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "缺少必填字段 - IP 地址",
			requestBody:    `{"name": "test", "secret": "test123"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "缺少必填字段 - Secret",
			requestBody:    `{"name": "test", "ipaddr": "10.0.0.4"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Secret 太短 (<6字符)",
			requestBody: `{
				"name": "test",
				"ipaddr": "10.0.0.5",
				"secret": "short"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "IP 地址已存在",
			requestBody: `{
				"name": "duplicate-nas",
				"ipaddr": "192.168.100.100",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "IPADDR_EXISTS",
		},
		{
			name: "无效的 IP 地址格式",
			requestBody: `{
				"name": "test",
				"ipaddr": "invalid-ip",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "无效的 CoA 端口 (>65535)",
			requestBody: `{
				"name": "test",
				"ipaddr": "10.0.0.6",
				"secret": "secret123",
				"coa_port": 70000
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "无效的状态值",
			requestBody: `{
				"name": "test",
				"ipaddr": "10.0.0.7",
				"secret": "secret123",
				"status": "invalid"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "无效的 JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "名称超长 (>100字符)",
			requestBody: `{
				"name": "` + strings.Repeat("a", 101) + `",
				"ipaddr": "10.0.0.8",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为重复 IP 测试创建已存在的 NAS
			if tt.name == "IP 地址已存在" {
				createTestNas(db, "existing-nas", "192.168.100.100")
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := CreateNAS(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var nas domain.NetNas
				json.Unmarshal(dataBytes, &nas)

				assert.NotZero(t, nas.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &nas)
				}
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateNAS(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	_ = createTestNas(db, "original-nas", "192.168.2.1")
	createTestNas(db, "another-nas", "192.168.2.2")

	tests := []struct {
		name           string
		nasID          string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.NetNas)
	}{
		{
			name:  "成功更新 NAS",
			nasID: "1",
			requestBody: `{
				"name": "updated-nas",
				"hostname": "updated.example.com",
				"coa_port": 3800
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, n *domain.NetNas) {
				assert.Equal(t, "updated-nas", n.Name)
				assert.Equal(t, "updated.example.com", n.Hostname)
				assert.Equal(t, 3800, n.CoaPort)
			},
		},
		{
			name:  "部分更新 - 只更新状态",
			nasID: "1",
			requestBody: `{
				"status": "disabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, n *domain.NetNas) {
				assert.Equal(t, "disabled", n.Status)
			},
		},
		{
			name:  "更新 IP 地址",
			nasID: "1",
			requestBody: `{
				"ipaddr": "192.168.3.1"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, n *domain.NetNas) {
				assert.Equal(t, "192.168.3.1", n.Ipaddr)
			},
		},
		{
			name:  "IP 地址冲突",
			nasID: "1",
			requestBody: `{
				"ipaddr": "192.168.2.2"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "IPADDR_EXISTS",
		},
		{
			name:           "NAS 不存在",
			nasID:          "999",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			nasID:          "invalid",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:  "无效的 IP 格式",
			nasID: "1",
			requestBody: `{
				"ipaddr": "not-an-ip"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "无效的端口",
			nasID: "1",
			requestBody: `{
				"coa_port": 0
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Secret 太短",
			nasID: "1",
			requestBody: `{
				"secret": "short"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/nas/"+tt.nasID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.nasID)

			err := UpdateNAS(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedNas domain.NetNas
				json.Unmarshal(dataBytes, &updatedNas)

				if tt.checkResult != nil {
					tt.checkResult(t, &updatedNas)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteNAS(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	_ = createTestNas(db, "nas-to-delete", "192.168.4.1")

	tests := []struct {
		name           string
		nasID          string
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:           "成功删除 NAS",
			nasID:          "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "NAS 不存在",
			nasID:          "999",
			expectedStatus: http.StatusOK, // GORM Delete 不会返回错误
			checkDeleted:   false,
		},
		{
			name:           "无效的 ID",
			nasID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/nas/"+tt.nasID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.nasID)

			err := DeleteNAS(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.checkDeleted {
					// 验证 NAS 已被删除
					var count int64
					db.Model(&domain.NetNas{}).Where("id = ?", tt.nasID).Count(&count)
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

// TestNASEdgeCases 测试边缘情况
func TestNASEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	t.Run("更新不存在的字段不应影响其他字段", func(t *testing.T) {
		nas := createTestNas(db, "test-nas", "192.168.5.1")
		originalName := nas.Name
		originalIpaddr := nas.Ipaddr

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/nas/1", strings.NewReader(`{"remark": "New remark"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := UpdateNAS(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedNas domain.NetNas
		json.Unmarshal(dataBytes, &updatedNas)

		// 名称和 IP 不应该改变
		assert.Equal(t, originalName, updatedNas.Name)
		assert.Equal(t, originalIpaddr, updatedNas.Ipaddr)
		// 备注应该更新
		assert.Equal(t, "New remark", updatedNas.Remark)
	})

	t.Run("默认值正确设置", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"name": "default-test",
			"ipaddr": "10.0.1.1",
			"secret": "secret123"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := CreateNAS(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var nas domain.NetNas
		json.Unmarshal(dataBytes, &nas)

		// 默认状态应该是 enabled
		assert.Equal(t, "enabled", nas.Status)
		// 默认 CoA 端口应该是 3799
		assert.Equal(t, 3799, nas.CoaPort)
	})

	t.Run("IPv4 地址验证", func(t *testing.T) {
		tests := []struct {
			ip      string
			isValid bool
		}{
			{"192.168.1.1", true},
			{"10.0.0.1", true},
			{"255.255.255.255", true},
			{"0.0.0.0", true},
			{"256.1.1.1", false},
			{"192.168.1", false},
			{"192.168.1.1.1", false},
			{"not-an-ip", false},
		}

		for _, tt := range tests {
			e := setupTestEcho()
			requestBody := `{
				"name": "ip-test",
				"ipaddr": "` + tt.ip + `",
				"secret": "secret123"
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := CreateNAS(c)
			require.NoError(t, err)

			if tt.isValid {
				assert.Equal(t, http.StatusOK, rec.Code, "IP %s should be valid", tt.ip)
			} else {
				assert.Equal(t, http.StatusBadRequest, rec.Code, "IP %s should be invalid", tt.ip)
			}
		}
	})

	t.Run("端口范围验证", func(t *testing.T) {
		tests := []struct {
			port    int
			isValid bool
		}{
			{1, true},
			{3799, true},
			{65535, true},
			{0, false},
			{65536, false},
			{-1, false},
		}

		for _, tt := range tests {
			e := setupTestEcho()
			requestBody := `{
				"name": "port-test",
				"ipaddr": "10.0.2.1",
				"secret": "secret123",
				"coa_port": ` + strconv.Itoa(tt.port) + `
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := CreateNAS(c)
			require.NoError(t, err)

			if tt.isValid {
				assert.Equal(t, http.StatusOK, rec.Code, "Port %d should be valid", tt.port)
			} else {
				assert.Equal(t, http.StatusBadRequest, rec.Code, "Port %d should be invalid", tt.port)
			}
		}
	})
}
