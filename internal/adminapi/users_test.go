package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// createTestUser creates test user data
func createTestUser(db *gorm.DB, username string, profileID int64) *domain.RadiusUser {
	user := &domain.RadiusUser{
		ID:         common.UUIDint64(),
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

// createTestUserWithDetails creates test user with custom details
func createTestUserWithDetails(db *gorm.DB, username, realname, email, mobile string, profileID int64) *domain.RadiusUser {
	user := &domain.RadiusUser{
		ID:         common.UUIDint64(),
		Username:   username,
		Password:   "testpass123",
		ProfileId:  profileID,
		Realname:   realname,
		Email:      email,
		Mobile:     mobile,
		Status:     "enabled",
		NodeId:     1,
		ActiveNum:  1,
		UpRate:     1024,
		DownRate:   2048,
		AddrPool:   "192.168.1.0/24",
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
	appCtx := setupTestApp(t, db)

	// CreateTest Profile
	profile := createTestProfile(db, "test-profile")

	// Create test data
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
			name:           "List all users - default pagination",
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
			name:           "Search by username",
			queryParams:    "?q=user1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				users := resp.Data.([]domain.RadiusUser)
				assert.Equal(t, "user1", users[0].Username)
				assert.Empty(t, users[0].Password) // Password should be cleared
			},
		},
		{
			name:           "Filter by status",
			queryParams:    "?status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Filter by profile_id",
			queryParams:    "?profile_id=" + fmt.Sprintf("%d", profile.ID),
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := listRadiusUsers(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// Convert the response data to a slice of users
			dataBytes, _ := json.Marshal(response.Data)
			var users []domain.RadiusUser
			_ = json.Unmarshal(dataBytes, &users)
			response.Data = users

			assert.Len(t, users, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

// TestListUsersWithFieldFilters tests specific field filters (username, realname, email, mobile)
func TestListUsersWithFieldFilters(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create Test Profile
	profile := createTestProfile(db, "test-profile")

	// Create test users with distinct details
	createTestUserWithDetails(db, "alice", "Alice Chen", "alice@example.com", "13800001111", profile.ID)
	createTestUserWithDetails(db, "bob", "Bob Wang", "bob@test.com", "13800002222", profile.ID)
	createTestUserWithDetails(db, "charlie", "Charlie Li", "charlie@example.com", "13900003333", profile.ID)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "Filter by username - exact partial match",
			queryParams:    "?username=alice",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				users := resp.Data.([]domain.RadiusUser)
				assert.Equal(t, "alice", users[0].Username)
			},
		},
		{
			name:           "Filter by username - partial match",
			queryParams:    "?username=li",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // alice, charlie both contain "li"
		},
		{
			name:           "Filter by realname",
			queryParams:    "?realname=Wang",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				users := resp.Data.([]domain.RadiusUser)
				assert.Equal(t, "Bob Wang", users[0].Realname)
			},
		},
		{
			name:           "Filter by email - partial match",
			queryParams:    "?email=example.com",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // alice and charlie
		},
		{
			name:           "Filter by mobile",
			queryParams:    "?mobile=138",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // alice and bob
		},
		{
			name:           "Filter by mobile - exact match",
			queryParams:    "?mobile=13900003333",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				users := resp.Data.([]domain.RadiusUser)
				assert.Equal(t, "charlie", users[0].Username)
			},
		},
		{
			name:           "Combined filters - username and status",
			queryParams:    "?username=bob&status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "No match filter",
			queryParams:    "?username=nonexistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := listRadiusUsers(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// Convert the response data to a slice of users
			dataBytes, _ := json.Marshal(response.Data)
			var users []domain.RadiusUser
			_ = json.Unmarshal(dataBytes, &users)
			response.Data = users

			assert.Len(t, users, tt.expectedCount, "Expected %d users for query %s", tt.expectedCount, tt.queryParams)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create test data
	profile := createTestProfile(db, "test-profile")
	user := createTestUser(db, "testuser", profile.ID)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get existing user",
			userID:         fmt.Sprintf("%d", user.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get missing user",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
		{
			name:           "Invalid ID",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			err := getRadiusUser(c)
			if tt.expectedStatus >= 400 {
				if err != nil {
					e.HTTPErrorHandler(err, c)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultUser domain.RadiusUser
				_ = json.Unmarshal(dataBytes, &resultUser)

				assert.Equal(t, user.Username, resultUser.Username)
				assert.Empty(t, resultUser.Password) // Password should be cleared
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// CreateTest Profile
	profile := createTestProfile(db, "test-profile")

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.RadiusUser)
	}{
		{
			name: "Successfully create user",
			requestBody: `{
				"username": "newuser",
				"password": "password123",
				"profile_id": "` + fmt.Sprintf("%d", profile.ID) + `",
				"realname": "New User",
				"mobile": "13900139000",
				"status": "enabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, user *domain.RadiusUser) {
				assert.Equal(t, "newuser", user.Username)
				assert.Equal(t, "New User", user.Realname)
				assert.Equal(t, "enabled", user.Status)
				assert.Empty(t, user.Password) // Password should be cleared in the response
			},
		},
		{
			name: "Missing status on create - use default",
			requestBody: `{
				"username": "defaultuser",
				"password": "password123",
				"profile_id": "` + fmt.Sprintf("%d", profile.ID) + `"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, user *domain.RadiusUser) {
				assert.Equal(t, "enabled", user.Status)
			},
		},
		{
			name:           "Missing required field - username",
			requestBody:    `{"password": "test", "profile_id": "1"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing required field - password",
			requestBody:    `{"username": "test", "profile_id": "1"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_PASSWORD",
		},
		{
			name:           "Missing required field - profile_id",
			requestBody:    `{"username": "test", "password": "test123"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Username already exists",
			requestBody: `{
				"username": "duplicateuser",
				"password": "password123",
				"profile_id": "` + fmt.Sprintf("%d", profile.ID) + `"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name: "Associated profile not found",
			requestBody: `{
				"username": "testuser",
				"password": "password123",
				"profile_id": "999"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "PROFILE_NOT_FOUND",
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "Frontend format - boolean status",
			requestBody: `{
				"username": "booluser",
				"password": "password123",
				"profile_id": "` + fmt.Sprintf("%d", profile.ID) + `",
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
			// Create an existing user to test duplicate usernames
			if tt.name == "Username already exists" {
				createTestUser(db, "duplicateuser", profile.ID)
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := createRadiusUser(c)
			if tt.expectedStatus >= 400 {
				if err != nil {
					e.HTTPErrorHandler(err, c)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var user domain.RadiusUser
				_ = json.Unmarshal(dataBytes, &user)

				assert.NotZero(t, user.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &user)
				}
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create test data
	profile := createTestProfile(db, "test-profile")
	profile2 := createTestProfile(db, "another-profile")
	user := createTestUser(db, "originaluser", profile.ID)
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
			name:   "Successfully update user",
			userID: fmt.Sprintf("%d", user.ID),
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
			name:   "Update profile - should sync profile config",
			userID: fmt.Sprintf("%d", user.ID),
			requestBody: `{
				"profile_id": "` + fmt.Sprintf("%d", profile2.ID) + `"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, u *domain.RadiusUser) {
				assert.Equal(t, profile2.ID, u.ProfileId)
				assert.Equal(t, profile2.ActiveNum, u.ActiveNum)
				assert.Equal(t, profile2.UpRate, u.UpRate)
			},
		},
		{
			name:   "Partial update - status only",
			userID: fmt.Sprintf("%d", user.ID),
			requestBody: `{
				"status": "disabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, u *domain.RadiusUser) {
				assert.Equal(t, "disabled", u.Status)
			},
		},
		{
			name:   "Username conflict",
			userID: fmt.Sprintf("%d", user.ID),
			requestBody: `{
				"username": "anotheruser"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name:           "User not found",
			userID:         "999",
			requestBody:    `{"realname": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
		{
			name:           "Invalid ID",
			userID:         "invalid",
			requestBody:    `{"realname": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:   "Update with missing associated profile",
			userID: fmt.Sprintf("%d", user.ID),
			requestBody: `{
				"profile_id": "999"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "PROFILE_NOT_FOUND",
		},
		{
			name:   "Frontend format - boolean update",
			userID: fmt.Sprintf("%d", user.ID),
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
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			err := updateRadiusUser(c)
			if tt.expectedStatus >= 400 {
				if err != nil {
					e.HTTPErrorHandler(err, c)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedUser domain.RadiusUser
				_ = json.Unmarshal(dataBytes, &updatedUser)

				assert.Empty(t, updatedUser.Password) // Password should be cleared
				if tt.checkResult != nil {
					tt.checkResult(t, &updatedUser)
				}
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create test data
	profile := createTestProfile(db, "test-profile")
	user := createTestUser(db, "user-to-delete", profile.ID)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  string
		checkDeleted   bool
	}{
		{
			name:           "Successfully delete user",
			userID:         fmt.Sprintf("%d", user.ID),
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "User not found",
			userID:         "999",
			expectedStatus: http.StatusOK, // GORM Delete does not return error
			checkDeleted:   false,
		},
		{
			name:           "Invalid ID",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.userID)

			err := deleteRadiusUser(c)
			if tt.expectedStatus >= 400 {
				if err != nil {
					e.HTTPErrorHandler(err, c)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.checkDeleted {
					// Validate the user has been deleted
					var count int64
					db.Model(&domain.RadiusUser{}).Where("id = ?", tt.userID).Count(&count)
					assert.Equal(t, int64(0), count)
				}
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestUserEdgeCases Test edge cases
func TestUserEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	profile := createTestProfile(db, "test-profile")

	t.Run("Username trims spaces automatically", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "  spaceuser  ",
			"password": "password123",
			"profile_id": "` + fmt.Sprintf("%d", profile.ID) + `"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := createRadiusUser(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var user domain.RadiusUser
		_ = json.Unmarshal(dataBytes, &user)

		assert.Equal(t, "spaceuser", user.Username)
	})

	t.Run("Inherit configuration from profile", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "inherituser",
			"password": "password123",
			"profile_id": "` + fmt.Sprintf("%d", profile.ID) + `"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := createRadiusUser(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var user domain.RadiusUser
		_ = json.Unmarshal(dataBytes, &user)

		// Should inherit profile configuration
		assert.Equal(t, profile.ActiveNum, user.ActiveNum)
		assert.Equal(t, profile.UpRate, user.UpRate)
		assert.Equal(t, profile.DownRate, user.DownRate)
		assert.Equal(t, profile.AddrPool, user.AddrPool)
	})

	t.Run("Updating non-existent fields should not affect others", func(t *testing.T) {
		user := createTestUser(db, "testuser", profile.ID)
		originalUsername := user.Username

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%d", user.ID), strings.NewReader(`{"realname": "New Name"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", user.ID))

		err := updateRadiusUser(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedUser domain.RadiusUser
		_ = json.Unmarshal(dataBytes, &updatedUser)

		// Username should remain unchanged
		assert.Equal(t, originalUsername, updatedUser.Username)
		// Realname should be updated
		assert.Equal(t, "New Name", updatedUser.Realname)
	})

	t.Run("Default expire time", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "expireuser",
			"password": "password123",
			"profile_id": "` + fmt.Sprintf("%d", profile.ID) + `"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := createRadiusUser(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var user domain.RadiusUser
		_ = json.Unmarshal(dataBytes, &user)

		// Default expiration time should be one year later
		expectedExpire := time.Now().AddDate(1, 0, 0)
		assert.WithinDuration(t, expectedExpire, user.ExpireTime, time.Hour*24)
	})
}
