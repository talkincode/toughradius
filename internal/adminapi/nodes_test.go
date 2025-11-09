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

// handleTestError 处理测试中的 Echo HTTP 错误
func handleTestError(rec *httptest.ResponseRecorder, err error) {
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			rec.Code = he.Code
			rec.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			json.NewEncoder(rec).Encode(he.Message)
		}
	} else if rec.Code == 0 {
		rec.Code = http.StatusOK
	}
}

// createTestNode 创建测试 Node 数据
func createTestNode(db *gorm.DB, name string) *domain.NetNode {
	node := &domain.NetNode{
		Name:      name,
		Tags:      "test,node",
		Remark:    "Test node",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(node)
	return node
}

func TestListNodes(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	createTestNode(db, "node1")
	createTestNode(db, "node2")
	createTestNode(db, "node3")

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "获取所有 nodes - 默认分页",
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
			name:           "分页查询 - 第2页",
			queryParams:    "?page=2&pageSize=2",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 2, resp.Meta.Page)
			},
		},
		{
			name:           "按名称搜索",
			queryParams:    "?name=node1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				nodes := resp.Data.([]domain.NetNode)
				assert.Equal(t, "node1", nodes[0].Name)
			},
		},
		{
			name:           "按名称模糊搜索",
			queryParams:    "?name=node",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/network/nodes"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := listNodes(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// 将 data 转换为 node 数组
			dataBytes, _ := json.Marshal(response.Data)
			var nodes []domain.NetNode
			json.Unmarshal(dataBytes, &nodes)
			response.Data = nodes

			assert.Len(t, nodes, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetNode(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	node := createTestNode(db, "testnode")

	tests := []struct {
		name           string
		nodeID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "获取存在的 node",
			nodeID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "获取不存在的 node",
			nodeID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NODE_NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			nodeID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/network/nodes/"+tt.nodeID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.nodeID)

			err := getNode(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultNode domain.NetNode
				json.Unmarshal(dataBytes, &resultNode)

				assert.Equal(t, node.Name, resultNode.Name)
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateNode(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.NetNode)
	}{
		{
			name: "成功创建 node",
			requestBody: `{
				"name": "new-node",
				"tags": "production,main",
				"remark": "This is a new node"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, node *domain.NetNode) {
				assert.Equal(t, "new-node", node.Name)
				assert.Equal(t, "production,main", node.Tags)
				assert.Equal(t, "This is a new node", node.Remark)
				assert.NotZero(t, node.ID)
			},
		},
		{
			name: "创建 node - 最小参数",
			requestBody: `{
				"name": "minimal-node"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, node *domain.NetNode) {
				assert.Equal(t, "minimal-node", node.Name)
				assert.Empty(t, node.Tags)
				assert.Empty(t, node.Remark)
			},
		},
		{
			name:           "缺少必填字段 - 名称",
			requestBody:    `{"tags": "test"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "名称为空字符串",
			requestBody:    `{"name": ""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "名称已存在",
			requestBody: `{
				"name": "duplicate-node"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NODE_EXISTS",
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
				"name": "` + strings.Repeat("a", 101) + `"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "备注超长 (>500字符)",
			requestBody: `{
				"name": "test-node",
				"remark": "` + strings.Repeat("x", 501) + `"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "名称自动去除空格",
			requestBody: `{
				"name": "  spaced-node  "
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, node *domain.NetNode) {
				assert.Equal(t, "spaced-node", node.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为重复名称测试创建已存在的 node
			if tt.name == "名称已存在" {
				createTestNode(db, "duplicate-node")
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/network/nodes", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// 调用处理函数并处理可能的错误
			handleTestError(rec, createNode(c))

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var node domain.NetNode
				json.Unmarshal(dataBytes, &node)

				assert.NotZero(t, node.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &node)
				}
			} else if tt.expectedError != "" {
				// 对于非验证错误，检查我们的自定义错误响应
				var errResponse ErrorResponse
				if json.Unmarshal(rec.Body.Bytes(), &errResponse) == nil {
					if errResponse.Error != "" {
						assert.Equal(t, tt.expectedError, errResponse.Error)
					}
				}
			}
		})
	}
}

func TestUpdateNode(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	_ = createTestNode(db, "original-node")
	createTestNode(db, "another-node")

	tests := []struct {
		name           string
		nodeID         string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.NetNode)
	}{
		{
			name:   "成功更新 node",
			nodeID: "1",
			requestBody: `{
				"name": "updated-node",
				"tags": "updated,tags",
				"remark": "Updated remark"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, n *domain.NetNode) {
				assert.Equal(t, "updated-node", n.Name)
				assert.Equal(t, "updated,tags", n.Tags)
				assert.Equal(t, "Updated remark", n.Remark)
			},
		},
		{
			name:   "部分更新 - 只更新备注",
			nodeID: "1",
			requestBody: `{
				"remark": "Only remark updated"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, n *domain.NetNode) {
				assert.Equal(t, "Only remark updated", n.Remark)
			},
		},
		{
			name:   "名称冲突",
			nodeID: "1",
			requestBody: `{
				"name": "another-node"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NODE_EXISTS",
		},
		{
			name:           "node 不存在",
			nodeID:         "999",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "NODE_NOT_FOUND",
		},
		{
			name:           "无效的 ID",
			nodeID:         "invalid",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:   "名称超长",
			nodeID: "1",
			requestBody: `{
				"name": "` + strings.Repeat("a", 101) + `"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "更新为空名称应该失败",
			nodeID: "1",
			requestBody: `{
				"name": ""
			}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/network/nodes/"+tt.nodeID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.nodeID)

			err := updateNode(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedNode domain.NetNode
				json.Unmarshal(dataBytes, &updatedNode)

				if tt.checkResult != nil {
					tt.checkResult(t, &updatedNode)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteNode(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	// 创建测试数据
	_ = createTestNode(db, "node-to-delete")
	node2 := createTestNode(db, "node-in-use")

	// 创建一个关联 node2 的 NAS 设备
	nas := &domain.NetNas{
		NodeId:    node2.ID,
		Name:      "test-nas",
		Ipaddr:    "192.168.1.1",
		Secret:    "secret",
		CoaPort:   3799,
		Status:    "enabled",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(nas)

	tests := []struct {
		name           string
		nodeID         string
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:           "成功删除未使用的 node",
			nodeID:         "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "无法删除正在使用的 node",
			nodeID:         "2",
			expectedStatus: http.StatusConflict,
			expectedError:  "NODE_IN_USE",
		},
		{
			name:           "node 不存在",
			nodeID:         "999",
			expectedStatus: http.StatusOK, // GORM Delete 不会返回错误
			checkDeleted:   false,
		},
		{
			name:           "无效的 ID",
			nodeID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/network/nodes/"+tt.nodeID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.nodeID)

			err := deleteNode(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.checkDeleted {
					// 验证 node 已被删除
					var count int64
					db.Model(&domain.NetNode{}).Where("id = ?", tt.nodeID).Count(&count)
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

// TestNodeEdgeCases 测试边缘情况
func TestNodeEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	setupTestApp(t, db)

	t.Run("更新不存在的字段不应影响其他字段", func(t *testing.T) {
		node := createTestNode(db, "test-node")
		originalName := node.Name
		originalTags := node.Tags

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/network/nodes/1", strings.NewReader(`{"remark": "New remark"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := updateNode(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedNode domain.NetNode
		json.Unmarshal(dataBytes, &updatedNode)

		// 名称和标签不应该改变
		assert.Equal(t, originalName, updatedNode.Name)
		assert.Equal(t, originalTags, updatedNode.Tags)
		// 备注应该更新
		assert.Equal(t, "New remark", updatedNode.Remark)
	})

	t.Run("创建时间和更新时间自动设置", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{"name": "time-test-node"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network/nodes", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := createNode(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var node domain.NetNode
		json.Unmarshal(dataBytes, &node)

		assert.NotZero(t, node.CreatedAt)
		assert.NotZero(t, node.UpdatedAt)
		assert.WithinDuration(t, time.Now(), node.CreatedAt, time.Second*2)
	})

	t.Run("更新时更新时间应该变化", func(t *testing.T) {
		node := createTestNode(db, "update-time-test")
		originalUpdateTime := node.UpdatedAt

		// 等待一小段时间确保时间不同
		time.Sleep(time.Millisecond * 100)

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/network/nodes/1", strings.NewReader(`{"remark": "Updated"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := updateNode(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedNode domain.NetNode
		json.Unmarshal(dataBytes, &updatedNode)

		assert.True(t, updatedNode.UpdatedAt.After(originalUpdateTime))
	})
}
