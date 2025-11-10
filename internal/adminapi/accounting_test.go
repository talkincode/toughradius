package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
)

func TestListAccounting(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移表结构
	err := db.AutoMigrate(&domain.RadiusAccounting{})
	require.NoError(t, err)

	// 插入测试数据
	now := time.Now()
	testRecords := []domain.RadiusAccounting{
		{
			Username:        "user1@test.com",
			AcctSessionId:   "sess-001",
			NasAddr:         "192.168.1.1",
			FramedIpaddr:    "10.0.0.1",
			AcctSessionTime: 3600,
			AcctInputTotal:  1024000,
			AcctOutputTotal: 2048000,
			AcctStartTime:   now.Add(-1 * time.Hour),
			AcctStopTime:    now,
		},
		{
			Username:        "user2@test.com",
			AcctSessionId:   "sess-002",
			NasAddr:         "192.168.1.2",
			FramedIpaddr:    "10.0.0.2",
			AcctSessionTime: 1800,
			AcctInputTotal:  512000,
			AcctOutputTotal: 1024000,
			AcctStartTime:   now.Add(-2 * time.Hour),
			AcctStopTime:    now.Add(-1 * time.Hour),
		},
	}

	for i := range testRecords {
		err = app.GDB().Create(&testRecords[i]).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		expectedCount  int
		checkFunc      func(*testing.T, *Response)
	}{
		{
			name: "获取所有记录",
			queryParams: map[string]string{
				"page":    "1",
				"perPage": "10",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkFunc: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(2), resp.Meta.Total)
				assert.Equal(t, 1, resp.Meta.Page)
				assert.Equal(t, 10, resp.Meta.PageSize)
			},
		},
		{
			name: "分页测试",
			queryParams: map[string]string{
				"page":    "1",
				"perPage": "1",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkFunc: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(2), resp.Meta.Total)
				assert.Equal(t, 1, resp.Meta.PageSize)
			},
		},
		{
			name: "按用户名过滤",
			queryParams: map[string]string{
				"page":     "1",
				"perPage":  "10",
				"username": "user1",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkFunc: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(1), resp.Meta.Total)
			},
		},
		{
			name: "按NAS地址过滤",
			queryParams: map[string]string{
				"page":     "1",
				"perPage":  "10",
				"nas_addr": "192.168.1.1",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "排序测试",
			queryParams: map[string]string{
				"page":    "1",
				"perPage": "10",
				"sort":    "username",
				"order":   "ASC",
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/accounting", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := ListAccounting(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.NoError(t, err)

			// 检查数据数组
			dataArray, ok := resp.Data.([]interface{})
			assert.True(t, ok, "Data should be an array")
			assert.Equal(t, tt.expectedCount, len(dataArray))

			if tt.checkFunc != nil {
				tt.checkFunc(t, &resp)
			}
		})
	}
}

func TestGetAccounting(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移表结构
	err := db.AutoMigrate(&domain.RadiusAccounting{})
	require.NoError(t, err)

	// 插入测试数据
	now := time.Now()
	testRecord := domain.RadiusAccounting{
		Username:          "user1@test.com",
		AcctSessionId:     "sess-001",
		NasAddr:           "192.168.1.1",
		FramedIpaddr:      "10.0.0.1",
		MacAddr:           "00:11:22:33:44:55",
		AcctSessionTime:   3600,
		AcctInputTotal:    1024000,
		AcctOutputTotal:   2048000,
		AcctInputPackets:  1000,
		AcctOutputPackets: 2000,
		AcctStartTime:     now.Add(-1 * time.Hour),
		AcctStopTime:      now,
	}
	err = app.GDB().Create(&testRecord).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		checkFunc      func(*testing.T, *Response)
	}{
		{
			name:           "成功获取记录",
			id:             "1",
			expectedStatus: http.StatusOK,
			checkFunc: func(t *testing.T, resp *Response) {
				dataMap, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok, "Data should be a map")
				assert.Equal(t, "user1@test.com", dataMap["username"])
				assert.Equal(t, "sess-001", dataMap["acct_session_id"])
			},
		},
		{
			name:           "无效的ID",
			id:             "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "记录不存在",
			id:             "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/accounting/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err := GetAccounting(c)
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK && tt.checkFunc != nil {
				var resp Response
				err = json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				tt.checkFunc(t, &resp)
			}
		})
	}
}

func TestAccountingFilters(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 迁移表结构
	err := db.AutoMigrate(&domain.RadiusAccounting{})
	require.NoError(t, err)

	// 插入不同时间的测试数据
	now := time.Now()
	testRecords := []domain.RadiusAccounting{
		{
			Username:        "user1@test.com",
			AcctSessionId:   "sess-001",
			NasAddr:         "192.168.1.1",
			FramedIpaddr:    "10.0.0.1",
			AcctSessionTime: 3600,
			AcctStartTime:   now.Add(-3 * time.Hour),
			AcctStopTime:    now.Add(-2 * time.Hour),
		},
		{
			Username:        "user2@test.com",
			AcctSessionId:   "sess-002",
			NasAddr:         "192.168.1.2",
			FramedIpaddr:    "10.0.0.2",
			AcctSessionTime: 1800,
			AcctStartTime:   now.Add(-1 * time.Hour),
			AcctStopTime:    now,
		},
	}

	for i := range testRecords {
		err = app.GDB().Create(&testRecords[i]).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedCount int
	}{
		{
			name: "按会话ID过滤",
			queryParams: map[string]string{
				"page":            "1",
				"perPage":         "10",
				"acct_session_id": "sess-001",
			},
			expectedCount: 1,
		},
		{
			name: "时间范围过滤 - 开始时间",
			queryParams: map[string]string{
				"page":       "1",
				"perPage":    "10",
				"start_time": now.Add(-90 * time.Minute).Format(time.RFC3339),
			},
			expectedCount: 1, // 只有 user2 (1小时前) 满足条件
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/accounting", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := ListAccounting(c)
			assert.NoError(t, err)

			var resp Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.NoError(t, err)

			dataArray, ok := resp.Data.([]interface{})
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCount, len(dataArray))
		})
	}
}
