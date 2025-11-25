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

// handleTestError handles Echo HTTP errors during tests
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

// createTestNode creates test node data
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
	appCtx := setupTestApp(t, db)

	// Create test data
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
			name:           "List all nodes - default pagination",
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
			name:           "Paginated query - page 1",
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
			name:           "Paginated query - page 2",
			queryParams:    "?page=2&pageSize=2",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 2, resp.Meta.Page)
			},
		},
		{
			name:           "Search by name",
			queryParams:    "?name=node1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				nodes := resp.Data.([]domain.NetNode)
				assert.Equal(t, "node1", nodes[0].Name)
			},
		},
		{
			name:           "Search by name (partial)",
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
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := listNodes(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// Convert the response data to a slice of nodes
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
	appCtx := setupTestApp(t, db)

	// Create test data
	node := createTestNode(db, "testnode")

	tests := []struct {
		name           string
		nodeID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get existing node",
			nodeID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get missing node",
			nodeID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NODE_NOT_FOUND",
		},
		{
			name:           "Invalid ID",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
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
	appCtx := setupTestApp(t, db)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.NetNode)
	}{
		{
			name: "Successfully create node",
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
			name: "Create node with minimal parameters",
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
			name:           "Missing required field - name",
			requestBody:    `{"tags": "test"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Name is empty string",
			requestBody:    `{"name": ""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Name already exists",
			requestBody: `{
				"name": "duplicate-node"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NODE_EXISTS",
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "Name too long (>100 characters)",
			requestBody: `{
				"name": "` + strings.Repeat("a", 101) + `"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Remark too long (>500 characters)",
			requestBody: `{
				"name": "test-node",
				"remark": "` + strings.Repeat("x", 501) + `"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Name trims spaces automatically",
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
			// Create an existing node to test duplicate names
			if tt.name == "Name already exists" {
				createTestNode(db, "duplicate-node")
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/network/nodes", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			// Call the handler and handle any potential errors
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
				// For non-validation errors, check our custom error response
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
	appCtx := setupTestApp(t, db)

	// Create test data
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
			name:   "Successfully update node",
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
			name:   "Partial update - remark only",
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
			name:   "Name conflict",
			nodeID: "1",
			requestBody: `{
				"name": "another-node"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NODE_EXISTS",
		},
		{
			name:           "Node not found",
			nodeID:         "999",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "NODE_NOT_FOUND",
		},
		{
			name:           "Invalid ID",
			nodeID:         "invalid",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:   "Name too long",
			nodeID: "1",
			requestBody: `{
				"name": "` + strings.Repeat("a", 101) + `"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:   "Updating to empty name should fail",
			nodeID: "1",
			requestBody: `{
				"name": ""
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/network/nodes/"+tt.nodeID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
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
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteNode(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create test data
	_ = createTestNode(db, "node-to-delete")
	node2 := createTestNode(db, "node-in-use")

	// Create a NAS device associated with node2
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
			name:           "Successfully delete unused node",
			nodeID:         "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "Cannot delete node in use",
			nodeID:         "2",
			expectedStatus: http.StatusConflict,
			expectedError:  "NODE_IN_USE",
		},
		{
			name:           "Node not found",
			nodeID:         "999",
			expectedStatus: http.StatusOK, // GORM Delete does not return error
			checkDeleted:   false,
		},
		{
			name:           "Invalid ID",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
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
					// Validate the node has been deleted
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

// TestNodeEdgeCases Test edge cases
func TestNodeEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	t.Run("Updating non-existent fields should not affect others", func(t *testing.T) {
		node := createTestNode(db, "test-node")
		originalName := node.Name
		originalTags := node.Tags

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/network/nodes/1", strings.NewReader(`{"remark": "New remark"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := updateNode(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedNode domain.NetNode
		json.Unmarshal(dataBytes, &updatedNode)

		// Name and tags should remain unchanged
		assert.Equal(t, originalName, updatedNode.Name)
		assert.Equal(t, originalTags, updatedNode.Tags)
		// Remark should be updated
		assert.Equal(t, "New remark", updatedNode.Remark)
	})

	t.Run("Created and updated timestamps set automatically", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{"name": "time-test-node"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network/nodes", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

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

	t.Run("Updated timestamp should change on update", func(t *testing.T) {
		node := createTestNode(db, "update-time-test")
		originalUpdateTime := node.UpdatedAt

		// Wait briefly to ensure the timestamps differ
		time.Sleep(time.Millisecond * 100)

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/network/nodes/1", strings.NewReader(`{"remark": "Updated"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
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
