package adminapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// UserRequest Used to handle user data sent from frontend
type UserRequest struct {
	NodeID     interface{} `json:"node_id"`                                     // Can be int64 or string
	ProfileID  interface{} `json:"profile_id" validate:"required"`              // Can be int64 or string
	Realname   string      `json:"realname" validate:"omitempty,max=100"`       // Real name
	Email      string      `json:"email" validate:"omitempty,email,max=100"`    // Email
	Mobile     string      `json:"mobile" validate:"omitempty,max=20"`          // Mobile number (optional, max 20 characters)
	Address    string      `json:"address" validate:"omitempty,max=255"`        // addresses
	Username   string      `json:"username" validate:"required,min=3,max=50"`   // Username
	Password   string      `json:"password" validate:"omitempty,min=6,max=128"` // Password
	AddrPool   string      `json:"addr_pool" validate:"omitempty,max=50"`       // Address pool
	Vlanid1    int         `json:"vlanid1" validate:"gte=0,lte=4096"`           // VLAN ID 1
	Vlanid2    int         `json:"vlanid2" validate:"gte=0,lte=4096"`           // VLAN ID 2
	IpAddr     string      `json:"ip_addr" validate:"omitempty,ipv4"`           // IPv4addresses
	Ipv6Addr   string      `json:"ipv6_addr" validate:"omitempty"`              // IPv6addresses
	MacAddr    string      `json:"mac_addr" validate:"omitempty,mac"`           // MACaddresses
	BindVlan   interface{} `json:"bind_vlan"`                                   // Can be int or boolean
	BindMac    interface{} `json:"bind_mac"`                                    // Can be int or boolean
	ExpireTime string      `json:"expire_time" validate:"omitempty"`            // Expiration time
	Status     interface{} `json:"status"`                                      // Can be string or boolean
	Remark     string      `json:"remark" validate:"omitempty,max=500"`         // Remark
}

// toRadiusUser Convert UserRequest Convert to RadiusUser
func (ur *UserRequest) toRadiusUser() *domain.RadiusUser {
	user := &domain.RadiusUser{
		Realname: ur.Realname,
		Mobile:   ur.Mobile,
		Username: strings.TrimSpace(ur.Username),
		Password: ur.Password,
		AddrPool: ur.AddrPool,
		Vlanid1:  ur.Vlanid1,
		Vlanid2:  ur.Vlanid2,
		IpAddr:   ur.IpAddr,
		MacAddr:  ur.MacAddr,
		Remark:   ur.Remark,
	}

	// Handle profile_id
	switch v := ur.ProfileID.(type) {
	case float64:
		user.ProfileId = int64(v)
	case string:
		if v != "" {
			profileId, _ := strconv.ParseInt(v, 10, 64)
			user.ProfileId = profileId
		}
	}

	// Handle node_id
	switch v := ur.NodeID.(type) {
	case float64:
		user.NodeId = int64(v)
	case string:
		if v != "" {
			nodeId, _ := strconv.ParseInt(v, 10, 64)
			user.NodeId = nodeId
		}
	}

	// Handle status field：boolean true -> "enabled", false -> "disabled"
	switch v := ur.Status.(type) {
	case bool:
		if v {
			user.Status = common.ENABLED
		} else {
			user.Status = common.DISABLED
		}
	case string:
		user.Status = strings.ToLower(v)
	}

	// Handle bind_mac field
	switch v := ur.BindMac.(type) {
	case bool:
		if v {
			user.BindMac = 1
		} else {
			user.BindMac = 0
		}
	case float64:
		user.BindMac = int(v)
	}

	// Handle bind_vlan field
	switch v := ur.BindVlan.(type) {
	case bool:
		if v {
			user.BindVlan = 1
		} else {
			user.BindVlan = 0
		}
	case float64:
		user.BindVlan = int(v)
	}

	return user
}

func registerUserRoutes() {
	webserver.ApiGET("/users", listRadiusUsers)
	webserver.ApiGET("/users/:id", getRadiusUser)
	webserver.ApiPOST("/users", createRadiusUser)
	webserver.ApiPUT("/users/:id", updateRadiusUser)
	webserver.ApiDELETE("/users/:id", deleteRadiusUser)
}

func listRadiusUsers(c echo.Context) error {
	page, pageSize := parsePagination(c)

	base := app.GDB().Model(&domain.RadiusUser{}).
		Select("radius_user.*, COALESCE(ro.count, 0) AS online_count").
		Joins("LEFT JOIN (SELECT username, COUNT(1) AS count FROM radius_online GROUP BY username) ro ON radius_user.username = ro.username")

	base = applyUserFilters(base, c)

	var total int64
	countQuery := base.Session(&gorm.Session{NewDB: true})
	if err := countQuery.Model(&domain.RadiusUser{}).Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询用户失败", err.Error())
	}

	var users []domain.RadiusUser
	if err := base.
		Order("radius_user.username ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询用户失败", err.Error())
	}

	for i := range users {
		users[i].Password = ""
	}

	return paged(c, users, total, page, pageSize)
}

func getRadiusUser(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的用户 ID", nil)
	}
	var user domain.RadiusUser
	if err := app.GDB().Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "用户不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询用户失败", err.Error())
	}
	user.Password = ""
	return ok(c, user)
}

func createRadiusUser(c echo.Context) error {
	var req UserRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析用户参数", err.Error())
	}

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	// Convert to RadiusUser
	user := req.toRadiusUser()

	// Additional business logic validation
	if user.Username == "" {
		return fail(c, http.StatusBadRequest, "MISSING_USERNAME", "用户名不能为空", nil)
	}
	if req.Password == "" {
		return fail(c, http.StatusBadRequest, "MISSING_PASSWORD", "密码不能为空", nil)
	}
	if user.ProfileId == 0 {
		return fail(c, http.StatusBadRequest, "MISSING_PROFILE_ID", "计费策略不能为空", nil)
	}

	// CheckUsernamealready exists
	var exists int64
	app.GDB().Model(&domain.RadiusUser{}).Where("username = ?", user.Username).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "USERNAME_EXISTS", "用户名已存在", nil)
	}

	// Validate if accounting profile exists
	var profile domain.RadiusProfile
	if err := app.GDB().Where("id = ?", user.ProfileId).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fail(c, http.StatusBadRequest, "PROFILE_NOT_FOUND", "关联的计费策略不存在", nil)
		}
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询计费策略失败", err.Error())
	}

	// ParseExpiration time
	expire, err := parseTimeInput(req.ExpireTime, time.Now().AddDate(1, 0, 0))
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_EXPIRE_TIME", "过期时间格式不正确", nil)
	}

	// Set default values and inherit from profile inherited values
	user.ID = common.UUIDint64()
	user.AddrPool = common.If(user.AddrPool != "", user.AddrPool, profile.AddrPool).(string)
	user.ActiveNum = profile.ActiveNum
	user.UpRate = profile.UpRate
	user.DownRate = profile.DownRate
	user.ExpireTime = expire
	if user.Status == "" {
		user.Status = common.ENABLED
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := app.GDB().Create(&user).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "创建用户失败", err.Error())
	}

	user.Password = ""
	return ok(c, user)
}

func updateRadiusUser(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的用户 ID", nil)
	}

	var req UserRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析用户参数", err.Error())
	}

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	var user domain.RadiusUser
	if err := app.GDB().Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "用户不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询用户失败", err.Error())
	}

	updateData := req.toRadiusUser()

	// Validate username uniqueness (if username modified)
	if updateData.Username != "" && updateData.Username != user.Username {
		var count int64
		app.GDB().Model(&domain.RadiusUser{}).Where("username = ? AND id != ?", updateData.Username, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "USERNAME_EXISTS", "用户名已存在", nil)
		}
	}

	// If updated ProfileID，need toValidateand sync Profile configuration
	if updateData.ProfileId != 0 && updateData.ProfileId != user.ProfileId {
		var profile domain.RadiusProfile
		if err := app.GDB().Where("id = ?", updateData.ProfileId).First(&profile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fail(c, http.StatusBadRequest, "PROFILE_NOT_FOUND", "关联的计费策略不存在", nil)
			}
			return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询计费策略失败", err.Error())
		}
		user.ProfileId = updateData.ProfileId
		user.ActiveNum = profile.ActiveNum
		user.UpRate = profile.UpRate
		user.DownRate = profile.DownRate
		user.AddrPool = profile.AddrPool
	}

	// Update other fields
	updates := map[string]interface{}{}
	if updateData.NodeId > 0 {
		updates["node_id"] = updateData.NodeId
	}
	if updateData.Realname != "" {
		updates["realname"] = updateData.Realname
	}
	if updateData.Mobile != "" {
		updates["mobile"] = updateData.Mobile
	}
	if updateData.Username != "" {
		updates["username"] = updateData.Username
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if updateData.AddrPool != "" {
		updates["addr_pool"] = updateData.AddrPool
	}
	if updateData.Vlanid1 > 0 {
		updates["vlanid1"] = updateData.Vlanid1
	}
	if updateData.Vlanid2 > 0 {
		updates["vlanid2"] = updateData.Vlanid2
	}
	if updateData.IpAddr != "" {
		updates["ip_addr"] = updateData.IpAddr
	}
	if updateData.MacAddr != "" {
		updates["mac_addr"] = updateData.MacAddr
	}
	if updateData.BindVlan >= 0 {
		updates["bind_vlan"] = updateData.BindVlan
	}
	if updateData.BindMac >= 0 {
		updates["bind_mac"] = updateData.BindMac
	}
	if updateData.Remark != "" {
		updates["remark"] = updateData.Remark
	}
	if updateData.Status != "" {
		updates["status"] = updateData.Status
	}
	if req.ExpireTime != "" {
		expire, err := parseTimeInput(req.ExpireTime, user.ExpireTime)
		if err != nil {
			return fail(c, http.StatusBadRequest, "INVALID_EXPIRE_TIME", "过期时间格式不正确", nil)
		}
		updates["expire_time"] = expire
	}

	updates["updated_at"] = time.Now()

	if err := app.GDB().Model(&user).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "更新用户失败", err.Error())
	}

	// Re-query latest data
	app.GDB().Where("id = ?", id).First(&user)
	user.Password = ""
	return ok(c, user)
}

func deleteRadiusUser(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "无效的用户 ID", nil)
	}
	if err := app.GDB().Where("id = ?", id).Delete(&domain.RadiusUser{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "删除用户失败", err.Error())
	}
	return ok(c, map[string]interface{}{
		"id": id,
	})
}

func applyUserFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		db = db.Where("radius_user.status = ?", strings.ToLower(status))
	}

	if q := strings.TrimSpace(c.QueryParam("q")); q != "" {
		like := "%" + q + "%"
		db = db.Where("(radius_user.username ILIKE ? OR radius_user.realname ILIKE ? OR radius_user.mobile ILIKE ?)", like, like, like)
	}

	if profileID := firstNotEmpty(c.QueryParam("profileId"), c.QueryParam("profile_id")); profileID != "" {
		if id, err := strconv.ParseInt(profileID, 10, 64); err == nil {
			db = db.Where("radius_user.profile_id = ?", id)
		}
	}

	if nodeID := firstNotEmpty(c.QueryParam("nodeId"), c.QueryParam("node_id")); nodeID != "" {
		if id, err := strconv.ParseInt(nodeID, 10, 64); err == nil {
			db = db.Where("radius_user.node_id = ?", id)
		}
	}

	if expireBefore := c.QueryParam("expireBefore"); expireBefore != "" {
		if ts, err := parseTimeInput(expireBefore, time.Time{}); err == nil && !ts.IsZero() {
			db = db.Where("radius_user.expire_time <= ?", ts)
		}
	}
	return db
}

func firstNotEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
