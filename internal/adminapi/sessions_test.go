package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// createTestOnlineSession 创建测试在线会话数据
func createTestOnlineSession(db *gorm.DB, username, nasAddr, framedIp string) *domain.RadiusOnline {
	session := &domain.RadiusOnline{
		Username:          username,
		NasId:             "test-nas-id",
		NasAddr:           nasAddr,
		NasPaddr:          "192.168.100.1",
		SessionTimeout:    3600,
		FramedIpaddr:      framedIp,
		FramedNetmask:     "255.255.255.0",
		MacAddr:           "00:11:22:33:44:55",
		NasPort:           1,
		NasClass:          "test-class",
		NasPortId:         "port-1",
		NasPortType:       15, // Ethernet
		ServiceType:       2,  // Framed
		AcctSessionId:     "session-" + username,
		AcctSessionTime:   1800,
		AcctInputTotal:    1024000,
		AcctOutputTotal:   2048000,
		AcctInputPackets:  1000,
		AcctOutputPackets: 2000,
		AcctStartTime:     time.Now().Add(-30 * time.Minute),
		LastUpdate:        time.Now(),
	}
	db.Create(session)
	return session
}

func TestListOnlineSessions(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移在线会话表
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	// 创建测试数据
	createTestOnlineSession(db, "user1", "192.168.1.1", "10.0.0.1")
	createTestOnlineSession(db, "user2", "192.168.1.1", "10.0.0.2")
	createTestOnlineSession(db, "user3", "192.168.1.2", "10.0.0.3")
	createTestOnlineSession(db, "testuser", "192.168.1.2", "10.0.0.4")

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "获取所有在线会话 - 默认分页",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(4), resp.Meta.Total)
			},
		},
		{
			name:           "分页查询 - 第1页",
			queryParams:    "?page=1&perPage=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(4), resp.Meta.Total)
				assert.Equal(t, 1, resp.Meta.Page)
				assert.Equal(t, 2, resp.Meta.PageSize)
			},
		},
		{
			name:           "分页查询 - 第2页",
			queryParams:    "?page=2&perPage=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(4), resp.Meta.Total)
				assert.Equal(t, 2, resp.Meta.Page)
			},
		},
		{
			name:           "按用户名搜索 - 精确匹配",
			queryParams:    "?username=user1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				data := resp.Data.([]interface{})
				sessionData := data[0].(map[string]interface{})
				assert.Equal(t, "user1", sessionData["username"])
			},
		},
		{
			name:           "按用户名搜索 - 模糊匹配",
			queryParams:    "?username=user",
			expectedStatus: http.StatusOK,
			expectedCount:  4, // user1, user2, user3, testuser
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(4), resp.Meta.Total)
			},
		},
		{
			name:           "按 NAS 地址过滤",
			queryParams:    "?nas_addr=192.168.1.1",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // user1, user2
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(2), resp.Meta.Total)
			},
		},
		{
			name:           "按 IP 地址过滤",
			queryParams:    "?framed_ipaddr=10.0.0.1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				data := resp.Data.([]interface{})
				sessionData := data[0].(map[string]interface{})
				assert.Equal(t, "10.0.0.1", sessionData["framed_ipaddr"])
			},
		},
		{
			name:           "多条件过滤 - 用户名和NAS地址",
			queryParams:    "?username=user&nas_addr=192.168.1.2",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // user3, testuser
		},
		{
			name:           "排序 - 按用户名升序",
			queryParams:    "?sort=username&order=ASC",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			checkResponse: func(t *testing.T, resp *Response) {
				data := resp.Data.([]interface{})
				first := data[0].(map[string]interface{})
				assert.Equal(t, "testuser", first["username"])
			},
		},
		{
			name:           "排序 - 按开始时间降序",
			queryParams:    "?sort=acct_start_time&order=DESC",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "无效的排序方向 - 使用默认值",
			queryParams:    "?sort=username&order=INVALID",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "查询不存在的用户",
			queryParams:    "?username=nonexistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(0), resp.Meta.Total)
			},
		},
		{
			name:           "查询不存在的NAS",
			queryParams:    "?nas_addr=10.10.10.10",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "无效的页码 - 使用默认值",
			queryParams:    "?page=0&perPage=10",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "无效的每页数量 - 使用默认值",
			queryParams:    "?page=1&perPage=0",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "超大每页数量 - 限制到最大值",
			queryParams:    "?page=1&perPage=200",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := ListOnlineSessions(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			data := response.Data.([]interface{})
			assert.Len(t, data, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetOnlineSession(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移在线会话表
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	// 创建测试数据
	session := createTestOnlineSession(db, "test-user", "192.168.1.100", "10.0.1.1")

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, *domain.RadiusOnline)
	}{
		{
			name:           "获取存在的会话",
			sessionID:      "1",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, s *domain.RadiusOnline) {
				assert.Equal(t, session.Username, s.Username)
				assert.Equal(t, session.NasAddr, s.NasAddr)
				assert.Equal(t, session.FramedIpaddr, s.FramedIpaddr)
				assert.Equal(t, session.AcctSessionId, s.AcctSessionId)
			},
		},
		{
			name:           "获取不存在的会话",
			sessionID:      "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "无效的 ID - 非数字",
			sessionID:      "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:           "无效的 ID - 负数",
			sessionID:      "-1",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "无效的 ID - 零",
			sessionID:      "0",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+tt.sessionID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)

			err := GetOnlineSession(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultSession domain.RadiusOnline
				json.Unmarshal(dataBytes, &resultSession)

				assert.NotZero(t, resultSession.ID)
				if tt.checkResponse != nil {
					tt.checkResponse(t, &resultSession)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteOnlineSession(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移在线会话表
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	tests := []struct {
		name           string
		sessionID      string
		setupData      func() *domain.RadiusOnline
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:      "成功删除会话",
			sessionID: "1",
			setupData: func() *domain.RadiusOnline {
				return createTestOnlineSession(db, "user-to-delete", "192.168.2.1", "10.0.2.1")
			},
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:      "删除不存在的会话",
			sessionID: "999",
			setupData: func() *domain.RadiusOnline {
				return nil
			},
			expectedStatus: http.StatusOK, // GORM Delete 不会返回错误
			checkDeleted:   false,
		},
		{
			name:      "无效的 ID - 非数字",
			sessionID: "invalid",
			setupData: func() *domain.RadiusOnline {
				return nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:      "无效的 ID - 负数",
			sessionID: "-1",
			setupData: func() *domain.RadiusOnline {
				return nil
			},
			expectedStatus: http.StatusOK, // 负数也能解析，只是查不到记录而已
			checkDeleted:   false,
		},
		{
			name:      "无效的 ID - 空字符串",
			sessionID: "",
			setupData: func() *domain.RadiusOnline {
				return nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清空表以确保测试独立性
			db.Exec("DELETE FROM radius_online")

			// 设置测试数据
			var session *domain.RadiusOnline
			if tt.setupData != nil {
				session = tt.setupData()
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/"+tt.sessionID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)

			err := DeleteOnlineSession(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				// 验证返回消息在 Data 中
				responseData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok, "响应数据应该是 map")

				message, ok := responseData["message"]
				assert.True(t, ok, "应该包含 message 字段")
				assert.Equal(t, "用户已强制下线", message)

				if tt.checkDeleted && session != nil {
					// 验证会话已被删除
					var count int64
					db.Model(&domain.RadiusOnline{}).Where("id = ?", session.ID).Count(&count)
					assert.Equal(t, int64(0), count, "会话应该已被删除")
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestSessionsEdgeCases 测试边缘情况
func TestSessionsEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移在线会话表
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	t.Run("空数据库查询", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListOnlineSessions(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NotNil(t, response.Meta)
		assert.Equal(t, int64(0), response.Meta.Total)
		data := response.Data.([]interface{})
		assert.Len(t, data, 0)
	})

	t.Run("大量数据分页性能", func(t *testing.T) {
		// 创建多条测试数据
		for i := 0; i < 50; i++ {
			username := "perftest" + string(rune(i))
			createTestOnlineSession(db, username, "192.168.10.1", "10.1.0."+string(rune(i)))
		}

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?page=3&perPage=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListOnlineSessions(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		assert.Len(t, data, 10, "第3页应该有10条数据")
	})

	t.Run("特殊字符在用户名中", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		createTestOnlineSession(db, "user@domain.com", "192.168.20.1", "10.2.0.1")
		createTestOnlineSession(db, "user-with-dash", "192.168.20.1", "10.2.0.2")
		createTestOnlineSession(db, "user_with_underscore", "192.168.20.1", "10.2.0.3")

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?username=@", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		assert.Len(t, data, 1, "应该找到包含@的用户")
	})

	t.Run("IP 地址边界值", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		createTestOnlineSession(db, "test1", "0.0.0.0", "0.0.0.0")
		createTestOnlineSession(db, "test2", "255.255.255.255", "255.255.255.255")
		createTestOnlineSession(db, "test3", "127.0.0.1", "127.0.0.1")

		// 测试查询 0.0.0.0
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?nas_addr=0.0.0.0", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		assert.Len(t, data, 1)
	})

	t.Run("会话数据完整性检查", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		session := createTestOnlineSession(db, "integrity-test", "192.168.30.1", "10.3.0.1")

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+fmt.Sprint(session.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(session.ID))

		err := GetOnlineSession(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		// 验证响应数据是否包含会话信息
		require.NotNil(t, response.Data, "响应数据不应该为空")

		// 将响应转换为 map 以便验证字段
		responseMap, ok := response.Data.(map[string]interface{})
		require.True(t, ok, "响应数据应该是 map")

		// 验证关键字段存在且正确
		assert.Equal(t, session.Username, responseMap["username"])
		assert.Equal(t, session.NasAddr, responseMap["nas_addr"])
		assert.Equal(t, session.FramedIpaddr, responseMap["framed_ipaddr"])
		assert.Equal(t, session.MacAddr, responseMap["mac_addr"])
		assert.Equal(t, session.AcctSessionId, responseMap["acct_session_id"])
	})

	t.Run("删除后再次查询", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		createTestOnlineSession(db, "delete-test", "192.168.40.1", "10.4.0.1")

		// 先删除
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := DeleteOnlineSession(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// 再次查询应该返回 404
		e = setupTestEcho()
		req = httptest.NewRequest(http.MethodGet, "/api/v1/sessions/1", nil)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err = GetOnlineSession(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("并发会话相同用户名", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		// 同一用户多个会话（多终端登录）
		createTestOnlineSession(db, "multilogin", "192.168.50.1", "10.5.0.1")
		createTestOnlineSession(db, "multilogin", "192.168.50.2", "10.5.0.2")
		createTestOnlineSession(db, "multilogin", "192.168.50.3", "10.5.0.3")

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?username=multilogin", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NotNil(t, response.Meta)
		assert.Equal(t, int64(3), response.Meta.Total, "同一用户应该有3个并发会话")
	})

	t.Run("会话时间测试", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		// 创建不同时间的会话
		session1 := createTestOnlineSession(db, "time1", "192.168.60.1", "10.6.0.1")
		session1.AcctStartTime = time.Now().Add(-2 * time.Hour)
		db.Save(session1)

		session2 := createTestOnlineSession(db, "time2", "192.168.60.2", "10.6.0.2")
		session2.AcctStartTime = time.Now().Add(-1 * time.Hour)
		db.Save(session2)

		// 按时间降序排列
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?sort=acct_start_time&order=DESC", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		// 最新的会话应该在前面
		firstSession := data[0].(map[string]interface{})
		assert.Equal(t, "time2", firstSession["username"])
	})
}

// TestSessionsFilterCombinations 测试各种过滤条件组合
func TestSessionsFilterCombinations(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移在线会话表
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	// 创建多样化的测试数据
	createTestOnlineSession(db, "alice", "192.168.1.1", "10.0.1.1")
	createTestOnlineSession(db, "bob", "192.168.1.1", "10.0.1.2")
	createTestOnlineSession(db, "charlie", "192.168.1.2", "10.0.2.1")
	createTestOnlineSession(db, "dave", "192.168.1.2", "10.0.2.2")
	createTestOnlineSession(db, "alice2", "192.168.1.3", "10.0.3.1")

	tests := []struct {
		name          string
		queryParams   string
		expectedCount int
		description   string
	}{
		{
			name:          "用户名+NAS地址组合",
			queryParams:   "?username=alice&nas_addr=192.168.1.1",
			expectedCount: 1,
			description:   "应该只返回 alice 在 192.168.1.1 的会话",
		},
		{
			name:          "用户名模糊+IP精确",
			queryParams:   "?username=alice&framed_ipaddr=10.0.1.1",
			expectedCount: 1,
			description:   "应该返回 alice 的特定IP会话",
		},
		{
			name:          "NAS地址+分页",
			queryParams:   "?nas_addr=192.168.1.1&page=1&perPage=1",
			expectedCount: 1,
			description:   "分页应该限制返回数量",
		},
		{
			name:          "三个条件组合",
			queryParams:   "?username=alice&nas_addr=192.168.1.1&framed_ipaddr=10.0.1.1",
			expectedCount: 1,
			description:   "所有条件都匹配",
		},
		{
			name:          "条件不匹配",
			queryParams:   "?username=alice&nas_addr=192.168.1.2",
			expectedCount: 0,
			description:   "alice 不在 192.168.1.2 上",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := ListOnlineSessions(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var response Response
			json.Unmarshal(rec.Body.Bytes(), &response)
			data := response.Data.([]interface{})
			assert.Len(t, data, tt.expectedCount, tt.description)
		})
	}
}
