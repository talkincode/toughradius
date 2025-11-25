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
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// createTestOperator creates test operator data
func createTestOperator(db *gorm.DB, username string, level string) *domain.SysOpr {
	opr := &domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  username,
		Password:  common.Sha256HashWithSalt("Password123", common.GetSecretSalt()),
		Realname:  "Test Operator",
		Mobile:    "13800138000",
		Email:     "test@example.com",
		Level:     level,
		Status:    common.ENABLED,
		Remark:    "Test operator",
		LastLogin: time.Time{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(opr)
	return opr
}

func TestListOperators(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)

	// Automatically migrate the operator table
	db.AutoMigrate(&domain.SysOpr{})

	// Create test data
	createTestOperator(db, "admin1", "admin")
	createTestOperator(db, "operator1", "operator")
	createTestOperator(db, "superadmin", "super")

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		checkResponse  func(*testing.T, *Response)
	}{
		{
			name:           "List all operators - default pagination",
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
			queryParams:    "?username=admin1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				operators := resp.Data.([]domain.SysOpr)
				assert.Equal(t, "admin1", operators[0].Username)
				assert.Empty(t, operators[0].Password) // Password should be cleared
			},
		},
		{
			name:           "Search by real name",
			queryParams:    "?realname=Test",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Filter by permission level",
			queryParams:    "?level=super",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				operators := resp.Data.([]domain.SysOpr)
				assert.Equal(t, "super", operators[0].Level)
			},
		},
		{
			name:           "Filter by status",
			queryParams:    "?status=enabled",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Search by username - case insensitive",
			queryParams:    "?username=ADMIN1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResponse: func(t *testing.T, resp *Response) {
				operators := resp.Data.([]domain.SysOpr)
				assert.Equal(t, "admin1", operators[0].Username)
			},
		},
		{
			name:           "Search by realname - case insensitive",
			queryParams:    "?realname=TEST",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/system/operators"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := listOperators(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response Response
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// Convert the response data to a slice of operators
			dataBytes, _ := json.Marshal(response.Data)
			var operators []domain.SysOpr
			json.Unmarshal(dataBytes, &operators)
			response.Data = operators

			assert.Len(t, operators, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestGetOperator(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	// Create test data
	opr := createTestOperator(db, "testadmin", "admin")

	tests := []struct {
		name           string
		operatorID     int64
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Get existing operator",
			operatorID:     opr.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get missing operator",
			operatorID:     999999,
			expectedStatus: http.StatusNotFound,
			expectedError:  "OPERATOR_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/system/operators/"+strconv.FormatInt(tt.operatorID, 10), nil)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(tt.operatorID, 10))

			err := getOperator(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var resultOpr domain.SysOpr
				json.Unmarshal(dataBytes, &resultOpr)

				assert.Equal(t, opr.Username, resultOpr.Username)
				assert.Empty(t, resultOpr.Password) // Password should be cleared
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestCreateOperator(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.SysOpr)
	}{
		{
			name: "Successfully create operator",
			requestBody: `{
				"username": "newadmin",
				"password": "Password123",
				"realname": "New Admin",
				"mobile": "13900139000",
				"email": "newadmin@example.com",
				"level": "admin",
				"status": "enabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, opr *domain.SysOpr) {
				assert.Equal(t, "newadmin", opr.Username)
				assert.Equal(t, "New Admin", opr.Realname)
				assert.Equal(t, "admin", opr.Level)
				assert.Equal(t, "enabled", opr.Status)
				assert.Empty(t, opr.Password) // Password should be cleared in the response
			},
		},
		{
			name: "Use default permission level on create",
			requestBody: `{
				"username": "defaultoper",
				"password": "Password123",
				"realname": "Default Operator",
				"mobile": "13800138001"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, opr *domain.SysOpr) {
				assert.Equal(t, "operator", opr.Level) // Defaults to operator
				assert.Equal(t, "enabled", opr.Status) // Defaults to enabled
			},
		},
		{
			name: "Username too short",
			requestBody: `{
				"username": "ab",
				"password": "Password123",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USERNAME",
		},
		{
			name: "Username too long",
			requestBody: `{
				"username": "abcdefghijklmnopqrstuvwxyz12345",
				"password": "Password123",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USERNAME",
		},
		{
			name:           "Missing username",
			requestBody:    `{"password": "Password123", "realname": "Test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_USERNAME",
		},
		{
			name:           "Missing password",
			requestBody:    `{"username": "testuser", "realname": "Test"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_PASSWORD",
		},
		{
			name:           "Missing real name",
			requestBody:    `{"username": "testuser", "password": "Password123"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_REALNAME",
		},
		{
			name: "Password too short",
			requestBody: `{
				"username": "testuser",
				"password": "12345",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PASSWORD",
		},
		{
			name: "Password too long",
			requestBody: `{
				"username": "testuser",
				"password": "` + strings.Repeat("a", 51) + `",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PASSWORD",
		},
		{
			name: "Password strength insufficient - numbers only",
			requestBody: `{
				"username": "testuser",
				"password": "123456789",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "WEAK_PASSWORD",
		},
		{
			name: "Password strength insufficient - letters only",
			requestBody: `{
				"username": "testuser",
				"password": "abcdefgh",
				"realname": "Test"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "WEAK_PASSWORD",
		},
		{
			name: "Invalid email format",
			requestBody: `{
				"username": "testuser",
				"password": "Password123",
				"realname": "Test",
				"email": "invalid-email"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_EMAIL",
		},
		{
			name: "Invalid mobile format",
			requestBody: `{
				"username": "testuser",
				"password": "Password123",
				"realname": "Test",
				"mobile": "12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_MOBILE",
		},
		{
			name: "Invalid permission level",
			requestBody: `{
				"username": "testuser",
				"password": "Password123",
				"realname": "Test",
				"level": "invalid"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_LEVEL",
		},
		{
			name: "Username already exists",
			requestBody: `{
				"username": "duplicateuser",
				"password": "Password123",
				"realname": "Duplicate User"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "Username trims spaces automatically",
			requestBody: `{
				"username": "  spaceuser  ",
				"password": "Password123",
				"realname": "Space User"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, opr *domain.SysOpr) {
				assert.Equal(t, "spaceuser", opr.Username)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an operator with an existing username to test duplicate handling
			if tt.name == "Username already exists" {
				createTestOperator(db, "duplicateuser", "operator")
			}

			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := createOperator(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var opr domain.SysOpr
				json.Unmarshal(dataBytes, &opr)

				assert.NotZero(t, opr.ID)
				if tt.checkResult != nil {
					tt.checkResult(t, &opr)
				}
			} else if tt.expectedError != "" {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

func TestUpdateOperator(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	// Create test data
	opr := createTestOperator(db, "originaluser", "operator")
	createTestOperator(db, "anotheruser", "admin")

	tests := []struct {
		name           string
		operatorID     int64
		requestBody    string
		expectedStatus int
		expectedError  string
		checkResult    func(*testing.T, *domain.SysOpr)
	}{
		{
			name:       "Successfully update operator",
			operatorID: opr.ID,
			requestBody: `{
				"realname": "Updated Name",
				"mobile": "13911139111"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "Updated Name", o.Realname)
				assert.Equal(t, "13911139111", o.Mobile)
			},
		},
		{
			name:       "Update permission level",
			operatorID: opr.ID,
			requestBody: `{
				"level": "admin"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "admin", o.Level)
			},
		},
		{
			name:       "Update password",
			operatorID: opr.ID,
			requestBody: `{
				"password": "NewPass456"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Empty(t, o.Password) // Password should be cleared in the response
				// Validate the password was updated (need to re-query the database)
				var updatedOpr domain.SysOpr
				db.Where("id = ?", opr.ID).First(&updatedOpr)
				expectedHash := common.Sha256HashWithSalt("NewPass456", common.GetSecretSalt())
				assert.Equal(t, expectedHash, updatedOpr.Password)
			},
		},
		{
			name:       "Partial update - status only",
			operatorID: opr.ID,
			requestBody: `{
				"status": "disabled"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "disabled", o.Status)
			},
		},
		{
			name:       "Username conflict",
			operatorID: opr.ID,
			requestBody: `{
				"username": "anotheruser"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  "USERNAME_EXISTS",
		},
		{
			name:           "Operator not found",
			operatorID:     999999,
			requestBody:    `{"realname": "test"}`,
			expectedStatus: http.StatusNotFound,
			expectedError:  "OPERATOR_NOT_FOUND",
		},
		{
			name:       "Invalid username - too short",
			operatorID: opr.ID,
			requestBody: `{
				"username": "ab"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_USERNAME",
		},
		{
			name:       "Invalid password - too short",
			operatorID: opr.ID,
			requestBody: `{
				"password": "12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PASSWORD",
		},
		{
			name:       "Password strength insufficient",
			operatorID: opr.ID,
			requestBody: `{
				"password": "123456789"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "WEAK_PASSWORD",
		},
		{
			name:       "Invalid email format",
			operatorID: opr.ID,
			requestBody: `{
				"email": "invalid-email"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_EMAIL",
		},
		{
			name:       "Invalid mobile format",
			operatorID: opr.ID,
			requestBody: `{
				"mobile": "12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_MOBILE",
		},
		{
			name:       "Update email and mobile",
			operatorID: opr.ID,
			requestBody: `{
				"email": "newemail@example.com",
				"mobile": "13922229222"
			}`,
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, o *domain.SysOpr) {
				assert.Equal(t, "newemail@example.com", o.Email)
				assert.Equal(t, "13922229222", o.Mobile)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupTestEcho()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/system/operators/"+strconv.FormatInt(tt.operatorID, 10), strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)
			c.SetParamNames("id")
			c.SetParamValues(strconv.FormatInt(tt.operatorID, 10))

			err := updateOperator(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				dataBytes, _ := json.Marshal(response.Data)
				var updatedOpr domain.SysOpr
				json.Unmarshal(dataBytes, &updatedOpr)

				assert.Empty(t, updatedOpr.Password) // Password should be cleared
				if tt.checkResult != nil {
					tt.checkResult(t, &updatedOpr)
				}
			} else {
				var errResponse ErrorResponse
				json.Unmarshal(rec.Body.Bytes(), &errResponse)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestOperatorEdgeCases Test edge cases
func TestOperatorEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	appCtx := setupTestApp(t, db)
	db.AutoMigrate(&domain.SysOpr{})

	t.Run("Username trims spaces automatically", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "  spaceadmin  ",
			"password": "Password123",
			"realname": "Space Admin"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := createOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		assert.Equal(t, "spaceadmin", opr.Username)
	})

	t.Run("Password trims spaces", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "testadmin2",
			"password": "  Password123  ",
			"realname": "Test Admin 2"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := createOperator(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		// Validate the password is stored correctly (trimmed then hashed)
		var savedOpr domain.SysOpr
		db.Where("id = ?", opr.ID).First(&savedOpr)
		expectedHash := common.Sha256HashWithSalt("Password123", common.GetSecretSalt())
		assert.Equal(t, expectedHash, savedOpr.Password)
	})

	t.Run("Updating non-existent fields should not affect others", func(t *testing.T) {
		opr := createTestOperator(db, "testadmin3", "admin")
		originalUsername := opr.Username
		originalLevel := opr.Level

		e := setupTestEcho()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/system/operators/"+strconv.FormatInt(opr.ID, 10), strings.NewReader(`{"realname": "New Name"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(opr.ID, 10))

		err := updateOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var updatedOpr domain.SysOpr
		json.Unmarshal(dataBytes, &updatedOpr)

		// Username and permission level should not change
		assert.Equal(t, originalUsername, updatedOpr.Username)
		assert.Equal(t, originalLevel, updatedOpr.Level)
		// Realname should be updated
		assert.Equal(t, "New Name", updatedOpr.Realname)
	})

	t.Run("Permission level is case-insensitive", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "casetest",
			"password": "Password123",
			"realname": "Case Test",
			"level": "ADMIN"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := createOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		assert.Equal(t, "admin", opr.Level) // Should be converted to lowercase
	})

	t.Run("Status is case-insensitive", func(t *testing.T) {
		e := setupTestEcho()
		requestBody := `{
			"username": "statustest",
			"password": "Password123",
			"realname": "Status Test",
			"status": "DISABLED"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := CreateTestContext(e, db, req, rec, appCtx)

		err := createOperator(c)
		require.NoError(t, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)
		dataBytes, _ := json.Marshal(response.Data)
		var opr domain.SysOpr
		json.Unmarshal(dataBytes, &opr)

		assert.Equal(t, "disabled", opr.Status) // Should be converted to lowercase
	})

	t.Run("Validate valid mobile numbers", func(t *testing.T) {
		validMobiles := []string{
			"13800138000",
			"15912345678",
			"18612345678",
			"+8613800138000",
			"8613800138000",
		}

		for _, mobile := range validMobiles {
			e := setupTestEcho()
			requestBody := `{
				"username": "mobile` + mobile + `",
				"password": "Password123",
				"realname": "Mobile Test",
				"mobile": "` + mobile + `"
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := createOperator(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code, "Mobile number %s should be valid", mobile)
		}
	})

	t.Run("Validate valid emails", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"admin@test.com",
		}

		for i, email := range validEmails {
			e := setupTestEcho()
			username := "emailtest" + strconv.Itoa(i)
			requestBody := `{
				"username": "` + username + `",
				"password": "Password123",
				"realname": "Email Test",
				"email": "` + email + `"
			}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/operators", strings.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := CreateTestContext(e, db, req, rec, appCtx)

			err := createOperator(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code, "Email %s should be valid", email)
		}
	})
}
