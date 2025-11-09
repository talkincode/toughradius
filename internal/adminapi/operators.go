package adminapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// 操作员请求结构
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

// 注册操作员管理路由
func registerOperatorsRoutes() {
	webserver.ApiGET("/system/operators", listOperators)
	webserver.ApiGET("/system/operators/:id", getOperator)
	webserver.ApiPOST("/system/operators", createOperator)
	webserver.ApiPUT("/system/operators/:id", updateOperator)
	webserver.ApiDELETE("/system/operators/:id", deleteOperator)
}

// 获取操作员列表
func listOperators(c echo.Context) error {
	page, pageSize := parsePagination(c)

	base := app.GDB().Model(&domain.SysOpr{})
	base = applyOperatorFilters(base, c)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询操作员失败", err.Error())
	}

	var operators []domain.SysOpr
	if err := base.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&operators).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询操作员失败", err.Error())
	}

	// 密码脱敏
	for i := range operators {
		operators[i].Password = ""
	}

	return paged(c, operators, total, page, pageSize)
}

// 获取单个操作员
func getOperator(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的操作员 ID", nil)
	}

	var operator domain.SysOpr
	if err := app.GDB().Where("id = ?", id).First(&operator).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "OPERATOR_NOT_FOUND", "操作员不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询操作员失败", err.Error())
	}

	// 密码脱敏
	operator.Password = ""
	return ok(c, operator)
}

// 创建操作员
func createOperator(c echo.Context) error {
	var payload operatorPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析操作员参数", nil)
	}

	payload.Username = strings.TrimSpace(payload.Username)
	payload.Password = strings.TrimSpace(payload.Password)

	if payload.Username == "" || payload.Password == "" || payload.Realname == "" {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "username、password、realname 不能为空", nil)
	}

	// 验证权限级别
	payload.Level = strings.ToLower(strings.TrimSpace(payload.Level))
	if payload.Level == "" {
		payload.Level = "operator"
	}
	if payload.Level != "super" && payload.Level != "admin" && payload.Level != "operator" {
		return fail(c, http.StatusBadRequest, "INVALID_LEVEL", "权限级别必须是 super、admin 或 operator", nil)
	}

	// 检查用户名是否已存在
	var exists int64
	app.GDB().Model(&domain.SysOpr{}).Where("username = ?", payload.Username).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "USERNAME_EXISTS", "用户名已存在", nil)
	}

	// 密码加密（使用 SHA256 + Salt，与登录验证保持一致）
	hashedPassword := common.Sha256HashWithSalt(payload.Password, common.SecretSalt)

	// 状态处理
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
		LastLogin: time.Time{}, // 初始化为零值
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := app.GDB().Create(&operator).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "创建操作员失败", err.Error())
	}

	// 密码脱敏
	operator.Password = ""
	return ok(c, operator)
}

// 更新操作员
func updateOperator(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的操作员 ID", nil)
	}

	var payload operatorPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析操作员参数", nil)
	}

	var operator domain.SysOpr
	if err := app.GDB().Where("id = ?", id).First(&operator).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "OPERATOR_NOT_FOUND", "操作员不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询操作员失败", err.Error())
	}

	// 更新字段
	if payload.Username != "" {
		operator.Username = strings.TrimSpace(payload.Username)
	}
	if payload.Password != "" {
		// 如果提供了新密码，则加密后更新
		operator.Password = common.Sha256HashWithSalt(strings.TrimSpace(payload.Password), common.SecretSalt)
	}
	if payload.Realname != "" {
		operator.Realname = payload.Realname
	}
	if payload.Mobile != "" {
		operator.Mobile = payload.Mobile
	}
	if payload.Email != "" {
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

	if err := app.GDB().Save(&operator).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "更新操作员失败", err.Error())
	}

	// 密码脱敏
	operator.Password = ""
	return ok(c, operator)
}

// 删除操作员
func deleteOperator(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的操作员 ID", nil)
	}

	// 权限检查：获取当前登录的操作员
	currentOpr, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "无法获取当前用户信息", nil)
	}

	// 不能删除自己
	if currentOpr.ID == id {
		return fail(c, http.StatusForbidden, "CANNOT_DELETE_SELF", "不能删除自己的账号", nil)
	}

	// 只有超级管理员才能删除其他管理员
	var targetOpr domain.SysOpr
	if err := app.GDB().Where("id = ?", id).First(&targetOpr).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "OPERATOR_NOT_FOUND", "操作员不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询操作员失败", err.Error())
	}

	if targetOpr.Level == "super" && currentOpr.Level != "super" {
		return fail(c, http.StatusForbidden, "PERMISSION_DENIED", "只有超级管理员才能删除超级管理员账号", nil)
	}

	if err := app.GDB().Where("id = ?", id).Delete(&domain.SysOpr{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "删除操作员失败", err.Error())
	}

	return ok(c, map[string]interface{}{
		"id": id,
	})
}

// 筛选条件
func applyOperatorFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if username := strings.TrimSpace(c.QueryParam("username")); username != "" {
		db = db.Where("username ILIKE ?", "%"+username+"%")
	}

	if realname := strings.TrimSpace(c.QueryParam("realname")); realname != "" {
		db = db.Where("realname ILIKE ?", "%"+realname+"%")
	}

	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		db = db.Where("status = ?", strings.ToLower(status))
	}

	if level := strings.TrimSpace(c.QueryParam("level")); level != "" {
		db = db.Where("level = ?", strings.ToLower(level))
	}

	return db
}
