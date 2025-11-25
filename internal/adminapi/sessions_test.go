package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/gorm"
)

// createTestOnlineSession Create test online session data
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
	appCtx := setupTestApp(t, db)

	// Migrate online session table
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	// Create test data
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
			name:           "List online sessions - default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(4), resp.Meta.Total)
			},
		},
		{
			name:           "Paginated query - page 1",
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
			name:           "Paginated query - page 2",
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
			name:           "Search by username - exact match",
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
			name:           "Search by username - fuzzy match",
			queryParams:    "?username=user",
			expectedStatus: http.StatusOK,
			expectedCount:  4, // user1, user2, user3, testuser
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(4), resp.Meta.Total)
			},
		},
		{
			name:           "Filter by NAS address",
			queryParams:    "?nas_addr=192.168.1.1",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // user1, user2
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(2), resp.Meta.Total)
			},
		},
		{
			name:           "Filter by IP address",
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
			name:           "Multi-filter - username and NAS address",
			queryParams:    "?username=user&nas_addr=192.168.1.2",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // user3, testuser
		},
		{
			name:           "Sort - username ascending",
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
			name:           "Sort - start time descending",
			queryParams:    "?sort=acct_start_time&order=DESC",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "Invalid sort direction - default fallback",
			queryParams:    "?sort=username&order=INVALID",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "Query non-existent user",
			queryParams:    "?username=nonexistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, int64(0), resp.Meta.Total)
			},
		},
		{
			name:           "Query non-existent NAS",
			queryParams:    "?nas_addr=10.10.10.10",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "Invalid page number - default fallback",
			queryParams:    "?page=0&perPage=10",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "Invalid per-page size - default fallback",
			queryParams:    "?page=1&perPage=0",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
		},
		{
			name:           "Large per-page size - limited to max",
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
			c := CreateTestContext(e, db, req, rec, appCtx)

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
	appCtx := setupTestApp(t, db)

	// Migrate online session table
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	// Create test data
	session := createTestOnlineSession(db, "test-user", "192.168.1.100", "10.0.1.1")

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, *domain.RadiusOnline)
	}{
		{
			name:           "Get existing session",
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
			name:           "Get missing session",
			sessionID:      "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "Invalid ID - non-numeric",
			sessionID:      "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:           "Invalid ID - negative",
			sessionID:      "-1",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "Invalid ID - zero",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
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
	appCtx := setupTestApp(t, db)

	// Migrate online session table
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
			name:      "Successfully delete session",
			sessionID: "1",
			setupData: func() *domain.RadiusOnline {
				return createTestOnlineSession(db, "user-to-delete", "192.168.2.1", "10.0.2.1")
			},
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:      "Delete missing session",
			sessionID: "999",
			setupData: func() *domain.RadiusOnline {
				return nil
			},
			expectedStatus: http.StatusNotFound, // API checks existence before delete
			expectedError:  "NOT_FOUND",
			checkDeleted:   false,
		},
		{
			name:      "Invalid ID - non-numeric",
			sessionID: "invalid",
			setupData: func() *domain.RadiusOnline {
				return nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:      "Invalid ID - negative",
			sessionID: "-1",
			setupData: func() *domain.RadiusOnline {
				return nil
			},
			expectedStatus: http.StatusNotFound, // Negative numbers can be parsed, but no record found
			expectedError:  "NOT_FOUND",
			checkDeleted:   false,
		},
		{
			name:      "Invalid ID - empty string",
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
			// Clear table to ensure test independence
			db.Exec("DELETE FROM radius_online")

			// Setup test data
			var session *domain.RadiusOnline
			if tt.setupData != nil {
				session = tt.setupData()
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/"+tt.sessionID, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.sessionID)

			err := DeleteOnlineSession(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				// ValidateResponse message in Data in
				responseData, ok := response.Data.(map[string]interface{})
				assert.True(t, ok, "Response data should be a map")

				message, ok := responseData["message"]
				assert.True(t, ok, "Should contain message field")
				assert.Equal(t, "User has been forced offline", message)

				if tt.checkDeleted && session != nil {
					// ValidateSession deleted
					var count int64
					db.Model(&domain.RadiusOnline{}).Where("id = ?", session.ID).Count(&count)
					assert.Equal(t, int64(0), count, "Session should be deleted")
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestSessionsEdgeCases Test edge cases
func TestSessionsEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Migrate online session table
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	t.Run("Empty database query", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

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

	t.Run("Large dataset pagination performance", func(t *testing.T) {
		// Create multiple test data
		for i := 0; i < 50; i++ {
			username := "perftest" + string(rune(i))
			createTestOnlineSession(db, username, "192.168.10.1", "10.1.0."+string(rune(i)))
		}

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?page=3&perPage=10", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListOnlineSessions(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		assert.Len(t, data, 10, "Page 3 should have 10 entries")
	})

	t.Run("Special characters in username", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		createTestOnlineSession(db, "user@domain.com", "192.168.20.1", "10.2.0.1")
		createTestOnlineSession(db, "user-with-dash", "192.168.20.1", "10.2.0.2")
		createTestOnlineSession(db, "user_with_underscore", "192.168.20.1", "10.2.0.3")

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?username=@", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		assert.Len(t, data, 1, "Should find user containing @")
	})

	t.Run("IP address boundary values", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		createTestOnlineSession(db, "test1", "0.0.0.0", "0.0.0.0")
		createTestOnlineSession(db, "test2", "255.255.255.255", "255.255.255.255")
		createTestOnlineSession(db, "test3", "127.0.0.1", "127.0.0.1")

		// Test query 0.0.0.0
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?nas_addr=0.0.0.0", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		assert.Len(t, data, 1)
	})

	t.Run("Session data integrity check", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		session := createTestOnlineSession(db, "integrity-test", "192.168.30.1", "10.3.0.1")

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+fmt.Sprint(session.ID), nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(session.ID))

		err := GetOnlineSession(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		// ValidateResponse data contains session info
		require.NotNil(t, response.Data, "Response data should not be empty")

		// ConvertresponseConvert to map to verify fields
		responseMap, ok := response.Data.(map[string]interface{})
		require.True(t, ok, "Response data should be a map")

		// ValidateKey fields exist and correct
		assert.Equal(t, session.Username, responseMap["username"])
		assert.Equal(t, session.NasAddr, responseMap["nas_addr"])
		assert.Equal(t, session.FramedIpaddr, responseMap["framed_ipaddr"])
		assert.Equal(t, session.MacAddr, responseMap["mac_addr"])
		assert.Equal(t, session.AcctSessionId, responseMap["acct_session_id"])
	})

	t.Run("Query after deletion", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		session := createTestOnlineSession(db, "delete-test", "192.168.40.1", "10.4.0.1")
		sessionID := strconv.FormatInt(int64(session.ID), 10)

		// Delete first
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/"+sessionID, nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(sessionID)

		err := DeleteOnlineSession(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Query again should return 404
		e = setupTestEcho()
		req = httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID, nil)
		rec = httptest.NewRecorder()
		c = CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(sessionID)

		err = GetOnlineSession(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Concurrent sessions same username", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		// Same user multiple sessions（multiple device login）
		createTestOnlineSession(db, "multilogin", "192.168.50.1", "10.5.0.1")
		createTestOnlineSession(db, "multilogin", "192.168.50.2", "10.5.0.2")
		createTestOnlineSession(db, "multilogin", "192.168.50.3", "10.5.0.3")

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?username=multilogin", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NotNil(t, response.Meta)
		assert.Equal(t, int64(3), response.Meta.Total, "Same user should have 3 concurrent sessions")
	})

	t.Run("Session time tests", func(t *testing.T) {
		db.Exec("DELETE FROM radius_online")
		// Create sessions with different times
		session1 := createTestOnlineSession(db, "time1", "192.168.60.1", "10.6.0.1")
		session1.AcctStartTime = time.Now().Add(-2 * time.Hour)
		db.Save(session1)

		session2 := createTestOnlineSession(db, "time2", "192.168.60.2", "10.6.0.2")
		session2.AcctStartTime = time.Now().Add(-1 * time.Hour)
		db.Save(session2)

		// Sort by time descending
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?sort=acct_start_time&order=DESC", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListOnlineSessions(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		data := response.Data.([]interface{})
		// Latest session should be first
		firstSession := data[0].(map[string]interface{})
		assert.Equal(t, "time2", firstSession["username"])
	})
}

// TestSessionsFilterCombinations Test various filter combinations
func TestSessionsFilterCombinations(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Migrate online session table
	err := db.AutoMigrate(&domain.RadiusOnline{})
	require.NoError(t, err)

	// Create diverse test data
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
			name:          "Username+NAS address combo",
			queryParams:   "?username=alice&nas_addr=192.168.1.1",
			expectedCount: 1,
			description:   "Should only return alice sessions on 192.168.1.1",
		},
		{
			name:          "Username fuzzy+IP exact",
			queryParams:   "?username=alice&framed_ipaddr=10.0.1.1",
			expectedCount: 1,
			description:   "Should return alice session for specific IP",
		},
		{
			name:          "NAS address + pagination",
			queryParams:   "?nas_addr=192.168.1.1&page=1&perPage=1",
			expectedCount: 1,
			description:   "Pagination should limit returned entries",
		},
		{
			name:          "Three condition combo",
			queryParams:   "?username=alice&nas_addr=192.168.1.1&framed_ipaddr=10.0.1.1",
			expectedCount: 1,
			description:   "All conditions should match",
		},
		{
			name:          "Conditions do not match",
			queryParams:   "?username=alice&nas_addr=192.168.1.2",
			expectedCount: 0,
			description:   "Alice is not on 192.168.1.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

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
