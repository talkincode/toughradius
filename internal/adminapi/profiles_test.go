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

	"gorm.io/gorm"
)

// createTestProfile creates test profile data
func createTestProfile(db *gorm.DB, name string) *domain.RadiusProfile {
	profile := &domain.RadiusProfile{
		Name:           name,
		Status:         "enabled",
		AddrPool:       "192.168.1.0/24",
		ActiveNum:      1,
		UpRate:         10240,
		DownRate:       20480,
		Domain:         "test.com",
		IPv6PrefixPool: "2001:db8::/64",
		BindMac:        0,
		BindVlan:       0,
		Remark:         "Test profile",
		NodeId:         1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	db.Create(profile)
	return profile
}

func TestListProfiles(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create test data
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
			name:           "List all profiles - default pagination",
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
			name:           "Paginated query - page 2",
			queryParams:    "?page=2&perPage=2",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				assert.Equal(t, int64(3), resp.Meta.Total)
				assert.Equal(t, 2, resp.Meta.Page)
			},
		},
		{
			name:           "Search by name",
			queryParams:    "?name=profile1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				profiles := resp.Data.([]domain.RadiusProfile) //nolint:errcheck // type assertion is safe in test
				assert.Equal(t, "profile1", profiles[0].Name)
			},
		},
		{
			name:           "Filter by status",
			queryParams:    "?status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Search by name - case insensitive",
			queryParams:    "?name=PROFILE1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				profiles := resp.Data.([]domain.RadiusProfile) //nolint:errcheck // type assertion is safe in test
				assert.Equal(t, "profile1", profiles[0].Name)
			},
		},
		{
			name:           "Filter by addr_pool",
			queryParams:    "?addr_pool=192.168",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Filter by domain",
			queryParams:    "?domain=test",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Sorting test - ASC",
			queryParams:    "?sort=name&order=ASC",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			checkResponse: func(t *testing.T, resp *Response) {
				profiles := resp.Data.([]domain.RadiusProfile) //nolint:errcheck // type assertion is safe in test
				assert.Equal(t, "profile1", profiles[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := ListProfiles(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// Convert the response data to a slice of profiles
			dataBytes, _ := json.Marshal(response.Data)
			var profiles []domain.RadiusProfile
			_ = json.Unmarshal(dataBytes, &profiles)
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
	appCtx := setupTestApp(t, db)

	// Create test data
	profile := createTestProfile(db, "test-profile")

	tests := []struct {
		name           string
		profileID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get existing profile",
			profileID:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get missing profile",
			profileID:      "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "Invalid ID",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
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
				_ = json.Unmarshal(dataBytes, &resultProfile)

				assert.Equal(t, profile.Name, resultProfile.Name)
				assert.Equal(t, profile.Status, resultProfile.Status)
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateProfile(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.RadiusProfile)
	}{
		{
			name: "Successfully create profile",
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
			name: "Missing status on create - use default",
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
			name:           "Missing required field - name",
			requestBody:    `{"status": "enabled"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "Name already exists",
			requestBody: `{
				"name": "duplicate-profile",
				"status": "enabled"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NAME_EXISTS",
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "Frontend format - boolean status and bind fields",
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
			// Create an existing profile to test duplicate names
			if tt.name == "Name already exists" {
				createTestProfile(db, "duplicate-profile")
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/radius-profiles", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := CreateProfile(c)
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
				var profile domain.RadiusProfile
				_ = json.Unmarshal(dataBytes, &profile)

				assert.NotZero(t, profile.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &profile)
				}
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create test data
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
			name:      "Successfully update profile",
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
			name:      "Partial update - status only",
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
			name:      "Name conflict",
			profileID: "1",
			requestBody: `{
				"name": "another-profile"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "NAME_EXISTS",
		},
		{
			name:           "Profile not found",
			profileID:      "999",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name:           "Invalid ID",
			profileID:      "invalid",
			requestBody:    `{"name": "test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_ID",
		},
		{
			name:      "Frontend format - boolean update",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			err := UpdateProfile(c)
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
				var updatedProfile domain.RadiusProfile
				_ = json.Unmarshal(dataBytes, &updatedProfile)

				if tt.checkResult != nil {
					tt.checkResult(t, &updatedProfile)
				}
			} else {
				var errResponse ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestDeleteProfile(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Create test data
	_ = createTestProfile(db, "profile-to-delete")
	profile2 := createTestProfile(db, "profile-in-use")

	// Create a user that is using profile2
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
			name:           "Successfully delete unused profile",
			profileID:      "1",
			expectedStatus: http.StatusOK,
			checkDeleted:   true,
		},
		{
			name:           "Cannot delete profile in use",
			profileID:      "2",
			expectedStatus: http.StatusConflict,
			expectedError:  "IN_USE",
		},
		{
			name:           "Profile not found",
			profileID:      "999",
			expectedStatus: http.StatusOK, // GORM Delete does not return error
			checkDeleted:   false,
		},
		{
			name:           "Invalid ID",
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
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(tt.profileID)

			err := DeleteProfile(c)
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
					// Validate the profile has been deleted
					var count int64
					db.Model(&domain.RadiusProfile{}).Where("id = ?", tt.profileID).Count(&count)
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

// TestProfileEdgeCases Test edge cases
func TestProfileEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	t.Run("Large pagination parameters", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles?perPage=1000", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListProfiles(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		// perPage should be limited to 100
		assert.LessOrEqual(t, response.Meta.PageSize, 100)
	})

	t.Run("Negative pagination parameters", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles?page=-1&perPage=-10", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListProfiles(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		// Should fall back to default values
		assert.Equal(t, 1, response.Meta.Page)
		assert.Equal(t, 10, response.Meta.PageSize)
	})

	t.Run("Invalid sort direction", func(t *testing.T) {
		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/radius-profiles?order=INVALID", nil)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := ListProfiles(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Updating non-existent fields should not affect others", func(t *testing.T) {
		profile := createTestProfile(db, "test-profile")
		originalName := profile.Name

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/radius-profiles/%d", profile.ID), strings.NewReader(`{"up_rate": 30720}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", profile.ID))

		err := UpdateProfile(c)
		require.NoError(t, err)

		var response Response
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedProfile domain.RadiusProfile
		_ = json.Unmarshal(dataBytes, &updatedProfile)

		// Name should remain unchanged
		assert.Equal(t, originalName, updatedProfile.Name)
		// up_rate should be updated
		assert.Equal(t, 30720, updatedProfile.UpRate)
	})
}
