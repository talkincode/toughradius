package adminapi

import (
	"encoding/json"
	"fmt"
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

// createTestNas creates test NAS data
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
	result := db.Create(nas)
	if result.Error != nil {
		panic(fmt.Sprintf("Failed to create test NAS: %v", result.Error))
	}
	return nas
}

func TestListNAS(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)

	// Create test data
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
			name:           "List all NAS - default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(3), resp["total"])
			},
		},
		{
			name:           "Paginated query - page 1",
			queryParams:    "?page=1&perPage=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, float64(3), resp["total"])
			},
		},
		{
			name:           "Search by name",
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
			name:           "Search by IP address (prefix match)",
			queryParams:    "?ipaddr=192.168.1",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Search by IP address (exact match)",
			queryParams:    "?ipaddr=192.168.1.1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by status",
			queryParams:    "?status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Search by name - case insensitive",
			queryParams:    "?name=NAS1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				nasData := data[0].(map[string]interface{})
				assert.Equal(t, "nas1", nasData["name"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/nas"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := ListNAS(c)
			require.NoError(t, err)

			// Debug: print response body if status is not 200
			if rec.Code != http.StatusOK {
				t.Logf("Response body: %s", rec.Body.String())
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			if data, ok := response["data"].([]interface{}); ok {
				assert.Len(t, data, tt.expectedCount)
			} else {
				t.Logf("Response: %+v", response)
				t.Fatal("response data is not a slice")
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestGetNAS(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)

	// Create test data
	nas := createTestNas(db, "test-nas", "192.168.1.100")

	tests := []struct {
		name           string
		nasID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get existing NAS",
			nasID:          "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get missing NAS",
			nasID:          "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "Invalid ID",
			nasID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/nas/"+tt.nasID, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
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
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateNAS(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.NetNas)
	}{
		{
			name: "Successfully create NAS",
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
			name: "Create NAS with minimal parameters",
			requestBody: `{
				"name": "minimal-nas",
				"ipaddr": "10.0.0.2",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, nas *domain.NetNas) {
				assert.Equal(t, "minimal-nas", nas.Name)
				assert.Equal(t, "enabled", nas.Status) // Default status
				assert.Equal(t, 3799, nas.CoaPort)     // Default CoA port
			},
		},
		{
			name:           "Missing required field - name",
			requestBody:    `{"ipaddr": "10.0.0.3", "secret": "test"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing required field - IP address",
			requestBody:    `{"name": "test", "secret": "test123"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing required field - secret",
			requestBody:    `{"name": "test", "ipaddr": "10.0.0.4"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Secret too short (<6 characters)",
			requestBody: `{
				"name": "test",
				"ipaddr": "10.0.0.5",
				"secret": "short"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "IP address already exists",
			requestBody: `{
				"name": "duplicate-nas",
				"ipaddr": "192.168.100.100",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "IPADDR_EXISTS",
		},
		{
			name: "Invalid IP address format",
			requestBody: `{
				"name": "test",
				"ipaddr": "invalid-ip",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid CoA port (>65535)",
			requestBody: `{
				"name": "test",
				"ipaddr": "10.0.0.6",
				"secret": "secret123",
				"coa_port": 70000
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid status value",
			requestBody: `{
				"name": "test",
				"ipaddr": "10.0.0.7",
				"secret": "secret123",
				"status": "invalid"
			}`,
			expectedStatus: http.StatusBadRequest,
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
				"name": "` + strings.Repeat("a", 101) + `",
				"ipaddr": "10.0.0.8",
				"secret": "secret123"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an existing NAS to test duplicate IP handling
			if tt.name == "IP address already exists" {
				createTestNas(db, "existing-nas", "192.168.100.100")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

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
	db, e, appCtx := CreateTestAppContext(t)

	// Create test data
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
			name:  "Successfully update NAS",
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
			name:  "Partial update - status only",
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
			name:  "Update IP address",
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
			name:  "IP address conflict",
			nasID: "1",
			requestBody: `{
				"ipaddr": "192.168.2.2"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "IPADDR_EXISTS",
		},
		{
			name:           "NAS not found",
			nasID:          "999",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "Invalid ID",
			nasID:          "invalid",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:  "Invalid IP format",
			nasID: "1",
			requestBody: `{
				"ipaddr": "not-an-ip"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:  "Invalid port",
			nasID: "1",
			requestBody: `{
				"coa_port": 0
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:  "Secret too short",
			nasID: "1",
			requestBody: `{
				"secret": "short"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/v1/nas/"+tt.nasID, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
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
	db, e, appCtx := CreateTestAppContext(t)

	// Create test data
	_ = createTestNas(db, "nas-to-delete", "192.168.4.1")

	tests := []struct {
		name           string
		nasID          string
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:           "Successfully delete NAS",
			nasID:          "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "NAS not found",
			nasID:          "999",
			expectedStatus: http.StatusOK, // GORM Delete does not return error
			checkDeleted:   false,
		},
		{
			name:           "Invalid ID",
			nasID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/nas/"+tt.nasID, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
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
					// Validate the NAS has been deleted
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

// TestNASEdgeCases Test edge cases
func TestNASEdgeCases(t *testing.T) {
	db, e, appCtx := CreateTestAppContext(t)

	t.Run("Updating non-existent fields should not affect others", func(t *testing.T) {
		nas := createTestNas(db, "test-nas", "192.168.5.1")
		originalName := nas.Name
		originalIpaddr := nas.Ipaddr

		req := httptest.NewRequest(http.MethodPut, "/api/v1/nas/1", strings.NewReader(`{"remark": "New remark"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := UpdateNAS(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedNas domain.NetNas
		json.Unmarshal(dataBytes, &updatedNas)

		// Name and IP should remain unchanged
		assert.Equal(t, originalName, updatedNas.Name)
		assert.Equal(t, originalIpaddr, updatedNas.Ipaddr)
		// Remark should be updated
		assert.Equal(t, "New remark", updatedNas.Remark)
	})

	t.Run("Defaults should be set correctly", func(t *testing.T) {
		requestBody := `{
			"name": "default-test",
			"ipaddr": "10.0.1.1",
			"secret": "secret123"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := CreateNAS(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var nas domain.NetNas
		json.Unmarshal(dataBytes, &nas)

		// Default status should be enabled
		assert.Equal(t, "enabled", nas.Status)
		// Default CoA port should be 3799
		assert.Equal(t, 3799, nas.CoaPort)
	})

	t.Run("IPv4 address validation", func(t *testing.T) {
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
			requestBody := `{
				"name": "ip-test",
				"ipaddr": "` + tt.ip + `",
				"secret": "secret123"
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := CreateNAS(c)
			require.NoError(t, err)

			if tt.isValid {
				assert.Equal(t, http.StatusOK, rec.Code, "IP %s should be valid", tt.ip)
			} else {
				assert.Equal(t, http.StatusBadRequest, rec.Code, "IP %s should be invalid", tt.ip)
			}
		}
	})

	t.Run("Port range validation", func(t *testing.T) {
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

		for idx, tt := range tests {
			ip := fmt.Sprintf("10.0.%d.%d", 2+(idx/200), 10+(idx%200))
			requestBody := `{
				"name": "port-test",
				"ipaddr": "` + ip + `",
				"secret": "secret123",
				"coa_port": ` + strconv.Itoa(tt.port) + `
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/nas", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

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
