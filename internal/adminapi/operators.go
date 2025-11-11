package adminapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"github.com/talkincode/toughradius/v9/pkg/validutil"
)

// Operator request structure
type operatorPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Realname string `json:"realname"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	Level    string `json:"level"`
	Status   string `json:"status"`
	Remark   string `json:"remark"`
}

// Register operator management routes
func registerOperatorsRoutes() {
	// Personal account settings routes - must be before :id routes registered
	webserver.ApiGET("/system/operators/me", getCurrentOperator)
	webserver.ApiPUT("/system/operators/me", updateCurrentOperator)

	// Operator management routes
	webserver.ApiGET("/system/operators", listOperators)
	webserver.ApiGET("/system/operators/:id", getOperator)
	webserver.ApiPOST("/system/operators", createOperator)
	webserver.ApiPUT("/system/operators/:id", updateOperator)
	webserver.ApiDELETE("/system/operators/:id", deleteOperator)
}

// Get current logged-in operator info
func getCurrentOperator(c echo.Context) error {
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}
	return ok(c, currentOpr)
} // Update current logged-in operator info（excluding permissions and status）
func updateCurrentOperator(c echo.Context) error {
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	var payload operatorPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse operator parameters", nil)
	}

	// Update allowed fields (excluding level and status)
	if payload.Username != "" {
		username := strings.TrimSpace(payload.Username)
		if len(username) < 3 || len(username) > 30 {
			return fail(c, http.StatusBadRequest, "INVALID_USERNAME", "Username length must be between 3 and 30 characters", nil)
		}
		// Checkusername already used by other account
		if username != currentOpr.Username {
			var exists int64
			GetDB(c).Model(&domain.SysOpr{}).Where("username = ? AND id != ?", username, currentOpr.ID).Count(&exists)
			if exists > 0 {
				return fail(c, http.StatusConflict, "USERNAME_EXISTS", "Username already exists", nil)
			}
		}
		currentOpr.Username = username
	}
	if payload.Password != "" {
		password := strings.TrimSpace(payload.Password)
		if len(password) < 6 || len(password) > 50 {
			return fail(c, http.StatusBadRequest, "INVALID_PASSWORD", "Password length must be between 6 and 50 characters", nil)
		}
		if !validutil.CheckPassword(password) {
			return fail(c, http.StatusBadRequest, "WEAK_PASSWORD", "Password must contain letters and numbers", nil)
		}
		currentOpr.Password = common.Sha256HashWithSalt(password, common.SecretSalt)
	}
	if payload.Realname != "" {
		currentOpr.Realname = payload.Realname
	}
	if payload.Mobile != "" {
		if !validutil.IsCnMobile(payload.Mobile) {
			return fail(c, http.StatusBadRequest, "INVALID_MOBILE", "Invalid mobile number format", nil)
		}
		currentOpr.Mobile = payload.Mobile
	}
	if payload.Email != "" {
		if !validutil.IsEmail(payload.Email) {
			return fail(c, http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format", nil)
		}
		currentOpr.Email = payload.Email
	}
	if payload.Remark != "" {
		currentOpr.Remark = payload.Remark
	}
	currentOpr.UpdatedAt = time.Now()

	if err := GetDB(c).Save(&currentOpr).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update operator", err.Error())
	}

	currentOpr.Password = ""
	return ok(c, currentOpr)
}

// List operators（Only super admin and admin can access）
func listOperators(c echo.Context) error {
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	// Only super admin and admin can view operator list
	if currentOpr.Level != "super" && currentOpr.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "No permission to access operator list", nil)
	}

	page, pageSize := parsePagination(c)

	base := GetDB(c).Model(&domain.SysOpr{})
	base = applyOperatorFilters(base, c)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query operators", err.Error())
	}

	var operators []domain.SysOpr
	if err := base.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&operators).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query operators", err.Error())
	}

	// Mask password
	for i := range operators {
		operators[i].Password = ""
	}

	return paged(c, operators, total, page, pageSize)
}

// Get a single operator (only super admins and admins can access)
func getOperator(c echo.Context) error {
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	// Only super admins and admins can view operator details
	if currentOpr.Level != "super" && currentOpr.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "No permission to access operator details", nil)
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid operator ID", nil)
	}

	var operator domain.SysOpr
	if err := GetDB(c).Where("id = ?", id).First(&operator).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "OPERATOR_NOT_FOUND", "Operator not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query operators", err.Error())
	}

	// Mask password
	operator.Password = ""
	return ok(c, operator)
}

// CreateOperator（Only super admin can operate）
func createOperator(c echo.Context) error {
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	// Only super admin can create operators
	if currentOpr.Level != "super" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only super admins can create operators", nil)
	}

	var payload operatorPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse operator parameters", nil)
	}

	payload.Username = strings.TrimSpace(payload.Username)
	payload.Password = strings.TrimSpace(payload.Password)

	// Validaterequired fields
	if payload.Username == "" {
		return fail(c, http.StatusBadRequest, "MISSING_USERNAME", "Username is required", nil)
	}
	if payload.Password == "" {
		return fail(c, http.StatusBadRequest, "MISSING_PASSWORD", "Password is required", nil)
	}
	if payload.Realname == "" {
		return fail(c, http.StatusBadRequest, "MISSING_REALNAME", "Real name is required", nil)
	}

	// ValidateUsernameformat（3-30characters，letters、digits、underscore）
	if len(payload.Username) < 3 || len(payload.Username) > 30 {
		return fail(c, http.StatusBadRequest, "INVALID_USERNAME", "Username length must be between 3 and 30 characters", nil)
	}

	// ValidatePasswordlength
	if len(payload.Password) < 6 || len(payload.Password) > 50 {
		return fail(c, http.StatusBadRequest, "INVALID_PASSWORD", "Password length must be between 6 and 50 characters", nil)
	}

	// Validate password strength (at least contains letters and digits)
	if !validutil.CheckPassword(payload.Password) {
		return fail(c, http.StatusBadRequest, "WEAK_PASSWORD", "Password must contain letters and numbers", nil)
	}

	// ValidateEmailformat（if provided）
	if payload.Email != "" && !validutil.IsEmail(payload.Email) {
		return fail(c, http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format", nil)
	}

	// ValidateMobile numberformat（if provided）
	if payload.Mobile != "" && !validutil.IsCnMobile(payload.Mobile) {
		return fail(c, http.StatusBadRequest, "INVALID_MOBILE", "Invalid mobile number format", nil)
	}

	// Validatepermission level
	payload.Level = strings.ToLower(strings.TrimSpace(payload.Level))
	if payload.Level == "" {
		payload.Level = "operator"
	}
	if payload.Level != "super" && payload.Level != "admin" && payload.Level != "operator" {
		return fail(c, http.StatusBadRequest, "INVALID_LEVEL", "Permission level must be super, admin, or operator", nil)
	}

	// CheckUsernamealready exists
	var exists int64
	GetDB(c).Model(&domain.SysOpr{}).Where("username = ?", payload.Username).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "USERNAME_EXISTS", "Username already exists", nil)
	}

	// PasswordEncrypt（Using SHA256 + Salt，consistent with login validation）
	hashedPassword := common.Sha256HashWithSalt(payload.Password, common.SecretSalt)

	// StatusHandle
	status := strings.ToLower(payload.Status)
	if status != common.ENABLED && status != common.DISABLED {
		status = common.ENABLED
	}

	operator := domain.SysOpr{
		ID:        common.UUIDint64(),
		Username:  payload.Username,
		Password:  hashedPassword,
		Realname:  payload.Realname,
		Mobile:    payload.Mobile,
		Email:     payload.Email,
		Level:     payload.Level,
		Status:    status,
		Remark:    payload.Remark,
		LastLogin: time.Time{}, // Initialize to zero value
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := GetDB(c).Create(&operator).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create operator", err.Error())
	}

	// Mask password
	operator.Password = ""
	return ok(c, operator)
}

// Update an operator
func updateOperator(c echo.Context) error {
	// Permission check: get the currently logged-in operator
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	// Only super admin and admin can update operators
	if currentOpr.Level != "super" && currentOpr.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only super admins and admins can update operators", nil)
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid operator ID", nil)
	}

	var payload operatorPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse operator parameters", nil)
	}

	// Check if modifying self
	isEditingSelf := currentOpr.ID == id

	// If modifying self，not allowed to modify permissions and status
	if isEditingSelf && (payload.Level != "" || payload.Status != "") {
		return fail(c, http.StatusForbidden, "CANNOT_MODIFY_SELF_PERMISSION", "Cannot modify your own permissions or status", nil)
	}

	var operator domain.SysOpr
	if err := GetDB(c).Where("id = ?", id).First(&operator).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "OPERATOR_NOT_FOUND", "Operator not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query operators", err.Error())
	}

	// Update fields
	if payload.Username != "" {
		username := strings.TrimSpace(payload.Username)
		// ValidateUsernameformat
		if len(username) < 3 || len(username) > 30 {
			return fail(c, http.StatusBadRequest, "INVALID_USERNAME", "Username length must be between 3 and 30 characters", nil)
		}
		// Checkusername already used by other account
		if username != operator.Username {
			var exists int64
			GetDB(c).Model(&domain.SysOpr{}).Where("username = ? AND id != ?", username, id).Count(&exists)
			if exists > 0 {
				return fail(c, http.StatusConflict, "USERNAME_EXISTS", "Username already exists", nil)
			}
		}
		operator.Username = username
	}
	if payload.Password != "" {
		// If new password provided，encrypt and update
		password := strings.TrimSpace(payload.Password)
		// ValidatePasswordlength
		if len(password) < 6 || len(password) > 50 {
			return fail(c, http.StatusBadRequest, "INVALID_PASSWORD", "Password length must be between 6 and 50 characters", nil)
		}
		// ValidatePasswordstrength
		if !validutil.CheckPassword(password) {
			return fail(c, http.StatusBadRequest, "WEAK_PASSWORD", "Password must contain letters and numbers", nil)
		}
		operator.Password = common.Sha256HashWithSalt(password, common.SecretSalt)
	}
	if payload.Realname != "" {
		operator.Realname = payload.Realname
	}
	if payload.Mobile != "" {
		// ValidateMobile numberformat
		if !validutil.IsCnMobile(payload.Mobile) {
			return fail(c, http.StatusBadRequest, "INVALID_MOBILE", "Invalid mobile number format", nil)
		}
		operator.Mobile = payload.Mobile
	}
	if payload.Email != "" {
		// ValidateEmailformat
		if !validutil.IsEmail(payload.Email) {
			return fail(c, http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format", nil)
		}
		operator.Email = payload.Email
	}
	if payload.Level != "" {
		level := strings.ToLower(strings.TrimSpace(payload.Level))
		if level == "super" || level == "admin" || level == "operator" {
			operator.Level = level
		}
	}
	if payload.Status != "" {
		status := strings.ToLower(payload.Status)
		if status == common.ENABLED || status == common.DISABLED {
			operator.Status = status
		}
	}
	if payload.Remark != "" {
		operator.Remark = payload.Remark
	}
	operator.UpdatedAt = time.Now()

	if err := GetDB(c).Save(&operator).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update operator", err.Error())
	}

	// Mask password
	operator.Password = ""
	return ok(c, operator)
}

// DeleteOperator
func deleteOperator(c echo.Context) error {
	// Permission check: get the currently logged-in operator
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unable to retrieve current user information", nil)
	}

	// Only super admins and admins can delete an operator
	if currentOpr.Level != "super" && currentOpr.Level != "admin" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only super admins and admins can delete operators", nil)
	}

	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid operator ID", nil)
	}

	// Cannot delete oneself
	if currentOpr.ID == id {
		return fail(c, http.StatusForbidden, "CANNOT_DELETE_SELF", "Cannot delete your own account", nil)
	}

	// Only super admins can delete other admins
	var targetOpr domain.SysOpr
	if err := GetDB(c).Where("id = ?", id).First(&targetOpr).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "OPERATOR_NOT_FOUND", "Operator not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query operators", err.Error())
	}

	if targetOpr.Level == "super" && currentOpr.Level != "super" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "Only super admins can delete another super admin account", nil)
	}

	if err := GetDB(c).Where("id = ?", id).Delete(&domain.SysOpr{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete operator", err.Error())
	}

	return ok(c, map[string]interface{}{
		"id": id,
	})
}

// Filter conditions
func applyOperatorFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if username := strings.TrimSpace(c.QueryParam("username")); username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}

	if realname := strings.TrimSpace(c.QueryParam("realname")); realname != "" {
		db = db.Where("realname LIKE ?", "%"+realname+"%")
	}

	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		db = db.Where("status = ?", strings.ToLower(status))
	}

	if level := strings.TrimSpace(c.QueryParam("level")); level != "" {
		db = db.Where("level = ?", strings.ToLower(level))
	}

	return db
}
