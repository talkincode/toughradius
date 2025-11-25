package adminapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

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

// UserUpdateRequest Used to handle user update data
type UserUpdateRequest struct {
	NodeID          interface{} `json:"node_id"`                                       // Can be int64 or string
	ProfileID       interface{} `json:"profile_id"`                                    // Can be int64 or string
	Realname        string      `json:"realname" validate:"omitempty,max=100"`         // Real name
	Email           string      `json:"email" validate:"omitempty,email,max=100"`      // Email
	Mobile          string      `json:"mobile" validate:"omitempty,max=20"`            // Mobile number (optional, max 20 characters)
	Address         string      `json:"address" validate:"omitempty,max=255"`          // addresses
	Username        string      `json:"username" validate:"omitempty,min=3,max=50"`    // Username
	Password        string      `json:"password" validate:"omitempty,min=6,max=128"`   // Password
	AddrPool        string      `json:"addr_pool" validate:"omitempty,max=50"`         // Address pool
	Vlanid1         int         `json:"vlanid1" validate:"gte=0,lte=4096"`             // VLAN ID 1
	Vlanid2         int         `json:"vlanid2" validate:"gte=0,lte=4096"`             // VLAN ID 2
	IpAddr          string      `json:"ip_addr" validate:"omitempty,ipv4"`             // IPv4addresses
	Ipv6Addr        string      `json:"ipv6_addr" validate:"omitempty"`                // IPv6addresses
	MacAddr         string      `json:"mac_addr" validate:"omitempty,mac"`             // MACaddresses
	BindVlan        interface{} `json:"bind_vlan"`                                     // Can be int or boolean
	BindMac         interface{} `json:"bind_mac"`                                      // Can be int or boolean
	ExpireTime      string      `json:"expire_time" validate:"omitempty"`              // Expiration time
	Status          interface{} `json:"status"`                                        // Can be string or boolean
	Remark          string      `json:"remark" validate:"omitempty,max=500"`           // Remark
	IPv6PrefixPool  string      `json:"ipv6_prefix_pool" validate:"omitempty,max=100"` // IPv6 prefix pool name
	Domain          string      `json:"domain" validate:"omitempty,max=100"`           // User domain
	ProfileLinkMode int         `json:"profile_link_mode" validate:"gte=0,lte=1"`      // Profile link mode (0=static, 1=dynamic)
}

// toRadiusUser Convert UserUpdateRequest Convert to RadiusUser
func (ur *UserUpdateRequest) toRadiusUser() *domain.RadiusUser {
	user := &domain.RadiusUser{
		Realname:        ur.Realname,
		Mobile:          ur.Mobile,
		Username:        strings.TrimSpace(ur.Username),
		Password:        ur.Password,
		AddrPool:        ur.AddrPool,
		Vlanid1:         ur.Vlanid1,
		Vlanid2:         ur.Vlanid2,
		IpAddr:          ur.IpAddr,
		IpV6Addr:        ur.Ipv6Addr,
		IPv6PrefixPool:  ur.IPv6PrefixPool,
		MacAddr:         ur.MacAddr,
		Domain:          ur.Domain,
		ProfileLinkMode: ur.ProfileLinkMode,
		Remark:          ur.Remark,
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

	base := GetDB(c).Model(&domain.RadiusUser{}).
		Select("radius_user.*, COALESCE(ro.count, 0) AS online_count").
		Joins("LEFT JOIN (SELECT username, COUNT(1) AS count FROM radius_online GROUP BY username) ro ON radius_user.username = ro.username")

	base = applyUserFilters(base, c)

	var total int64
	countQuery := base.Session(&gorm.Session{NewDB: true})
	if err := countQuery.Model(&domain.RadiusUser{}).Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query users", err.Error())
	}

	var users []domain.RadiusUser
	if err := base.
		Order("radius_user.username ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query users", err.Error())
	}

	for i := range users {
		users[i].Password = ""
	}

	return paged(c, users, total, page, pageSize)
}

func getRadiusUser(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID", nil)
	}
	var user domain.RadiusUser
	if err := GetDB(c).Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query users", err.Error())
	}
	user.Password = ""
	return ok(c, user)
}

func createRadiusUser(c echo.Context) error {
	var req UserRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse user parameters", err.Error())
	}

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	// Convert to RadiusUser
	user := req.toRadiusUser()

	// Additional business logic validation
	if user.Username == "" {
		return fail(c, http.StatusBadRequest, "MISSING_USERNAME", "Username is required", nil)
	}
	if req.Password == "" {
		return fail(c, http.StatusBadRequest, "MISSING_PASSWORD", "Password is required", nil)
	}
	if user.ProfileId == 0 {
		return fail(c, http.StatusBadRequest, "MISSING_PROFILE_ID", "Billing profile is required", nil)
	}

	// CheckUsernamealready exists
	var exists int64
	GetDB(c).Model(&domain.RadiusUser{}).Where("username = ?", user.Username).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "USERNAME_EXISTS", "Username already exists", nil)
	}

	// Validate if accounting profile exists
	var profile domain.RadiusProfile
	if err := GetDB(c).Where("id = ?", user.ProfileId).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fail(c, http.StatusBadRequest, "PROFILE_NOT_FOUND", "Associated billing profile not found", nil)
		}
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query billing profile", err.Error())
	}

	// ParseExpiration time
	expire, err := parseTimeInput(req.ExpireTime, time.Now().AddDate(1, 0, 0))
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_EXPIRE_TIME", "Invalid expire time format", nil)
	}

	// Set default values and inherit from profile inherited values
	user.ID = common.UUIDint64()
	// Inherit all profile attributes (can be overridden by user-specific values)
	user.AddrPool = common.If(user.AddrPool != "", user.AddrPool, profile.AddrPool).(string)
	user.ActiveNum = profile.ActiveNum
	user.UpRate = profile.UpRate
	user.DownRate = profile.DownRate
	user.Domain = common.If(user.Domain != "", user.Domain, profile.Domain).(string)
	user.IPv6PrefixPool = common.If(user.IPv6PrefixPool != "", user.IPv6PrefixPool, profile.IPv6PrefixPool).(string)
	user.BindMac = common.If(user.BindMac > 0, user.BindMac, profile.BindMac).(int)
	user.BindVlan = common.If(user.BindVlan > 0, user.BindVlan, profile.BindVlan).(int)
	// Default to static mode (snapshot behavior)
	if user.ProfileLinkMode == 0 {
		user.ProfileLinkMode = domain.ProfileLinkModeStatic
	}
	user.ExpireTime = expire
	if user.Status == "" {
		user.Status = common.ENABLED
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := GetDB(c).Create(&user).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create user", err.Error())
	}

	user.Password = ""
	return ok(c, user)
}

func updateRadiusUser(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID", nil)
	}

	var req UserUpdateRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse user parameters", err.Error())
	}

	// Auto-validate request parameters
	if err := c.Validate(&req); err != nil {
		return err // Validation errors already formatted
	}

	var user domain.RadiusUser
	if err := GetDB(c).Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query users", err.Error())
	}

	updateData := req.toRadiusUser()

	// Validate username uniqueness (if username modified)
	if updateData.Username != "" && updateData.Username != user.Username {
		var count int64
		GetDB(c).Model(&domain.RadiusUser{}).Where("username = ? AND id != ?", updateData.Username, id).Count(&count)
		if count > 0 {
			return fail(c, http.StatusConflict, "USERNAME_EXISTS", "Username already exists", nil)
		}
	}

	// Update other fields
	updates := map[string]interface{}{}

	// If updated ProfileID，need toValidateand sync Profile configuration
	if updateData.ProfileId != 0 && updateData.ProfileId != user.ProfileId {
		var profile domain.RadiusProfile
		if err := GetDB(c).Where("id = ?", updateData.ProfileId).First(&profile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fail(c, http.StatusBadRequest, "PROFILE_NOT_FOUND", "Associated billing profile not found", nil)
			}
			return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query billing profile", err.Error())
		}
		updates["profile_id"] = updateData.ProfileId
		// Sync all profile attributes when switching profiles
		updates["active_num"] = profile.ActiveNum
		updates["up_rate"] = profile.UpRate
		updates["down_rate"] = profile.DownRate
		updates["addr_pool"] = profile.AddrPool
		updates["domain"] = profile.Domain
		updates["ipv6_prefix_pool"] = profile.IPv6PrefixPool
		updates["bind_mac"] = profile.BindMac
		updates["bind_vlan"] = profile.BindVlan

		// Invalidate profile cache to ensure fresh data
		GetAppContext(c).ProfileCache().Invalidate(updateData.ProfileId)
		GetAppContext(c).ProfileCache().Invalidate(user.ProfileId) // Also invalidate old profile
	}

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
	if updateData.IpV6Addr != "" {
		updates["ipv6_addr"] = updateData.IpV6Addr
	}
	if updateData.IPv6PrefixPool != "" {
		updates["ipv6_prefix_pool"] = updateData.IPv6PrefixPool
	}
	if updateData.MacAddr != "" {
		updates["mac_addr"] = updateData.MacAddr
	}
	if updateData.Domain != "" {
		updates["domain"] = updateData.Domain
	}
	if req.BindVlan != nil {
		updates["bind_vlan"] = updateData.BindVlan
	}
	if req.BindMac != nil {
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
			return fail(c, http.StatusBadRequest, "INVALID_EXPIRE_TIME", "Invalid expire time format", nil)
		}
		updates["expire_time"] = expire
	}

	// Handle ProfileLinkMode changes: switch from dynamic to static requires snapshot
	if updateData.ProfileLinkMode >= 0 && updateData.ProfileLinkMode != user.ProfileLinkMode {
		updates["profile_link_mode"] = updateData.ProfileLinkMode

		// When switching from dynamic (1) to static (0), snapshot current profile values
		if updateData.ProfileLinkMode == domain.ProfileLinkModeStatic && user.ProfileLinkMode == domain.ProfileLinkModeDynamic {
			var profile domain.RadiusProfile
			if err := GetDB(c).Where("id = ?", user.ProfileId).First(&profile).Error; err == nil {
				// Snapshot all profile attributes (unless user has specific overrides)
				if user.UpRate == 0 {
					updates["up_rate"] = profile.UpRate
				}
				if user.DownRate == 0 {
					updates["down_rate"] = profile.DownRate
				}
				if user.ActiveNum == 0 {
					updates["active_num"] = profile.ActiveNum
				}
				if user.AddrPool == "" || user.AddrPool == "NA" {
					updates["addr_pool"] = profile.AddrPool
				}
				if user.Domain == "" || user.Domain == "NA" {
					updates["domain"] = profile.Domain
				}
				if user.IPv6PrefixPool == "" || user.IPv6PrefixPool == "NA" {
					updates["ipv6_prefix_pool"] = profile.IPv6PrefixPool
				}
				if user.BindMac == 0 && user.MacAddr == "" {
					updates["bind_mac"] = profile.BindMac
				}
				if user.BindVlan == 0 && user.Vlanid1 == 0 && user.Vlanid2 == 0 {
					updates["bind_vlan"] = profile.BindVlan
				}
			}
		}
	}

	updates["updated_at"] = time.Now()

	if err := GetDB(c).Model(&user).Updates(updates).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update user", err.Error())
	}

	// Re-query latest data
	GetDB(c).Where("id = ?", id).First(&user)
	user.Password = ""
	return ok(c, user)
}

func deleteRadiusUser(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID", nil)
	}
	if err := GetDB(c).Where("id = ?", id).Delete(&domain.RadiusUser{}).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete user", err.Error())
	}
	return ok(c, map[string]interface{}{
		"id": id,
	})
}

func applyUserFilters(db *gorm.DB, c echo.Context) *gorm.DB {
	if status := strings.TrimSpace(c.QueryParam("status")); status != "" {
		db = db.Where("radius_user.status = ?", strings.ToLower(status))
	}

	// Handle username filter (partial match)
	if username := strings.TrimSpace(c.QueryParam("username")); username != "" {
		if strings.EqualFold(db.Dialector.Name(), "postgres") {
			db = db.Where("radius_user.username ILIKE ?", "%"+username+"%")
		} else {
			db = db.Where("LOWER(radius_user.username) LIKE ?", "%"+strings.ToLower(username)+"%")
		}
	}

	// Handle realname filter (partial match)
	if realname := strings.TrimSpace(c.QueryParam("realname")); realname != "" {
		if strings.EqualFold(db.Dialector.Name(), "postgres") {
			db = db.Where("radius_user.realname ILIKE ?", "%"+realname+"%")
		} else {
			db = db.Where("LOWER(radius_user.realname) LIKE ?", "%"+strings.ToLower(realname)+"%")
		}
	}

	// Handle email filter (partial match)
	if email := strings.TrimSpace(c.QueryParam("email")); email != "" {
		if strings.EqualFold(db.Dialector.Name(), "postgres") {
			db = db.Where("radius_user.email ILIKE ?", "%"+email+"%")
		} else {
			db = db.Where("LOWER(radius_user.email) LIKE ?", "%"+strings.ToLower(email)+"%")
		}
	}

	// Handle mobile filter (partial match)
	if mobile := strings.TrimSpace(c.QueryParam("mobile")); mobile != "" {
		db = db.Where("radius_user.mobile LIKE ?", "%"+mobile+"%")
	}

	// Global search across multiple fields
	if q := strings.TrimSpace(c.QueryParam("q")); q != "" {
		if strings.EqualFold(db.Dialector.Name(), "postgres") {
			like := "%" + q + "%"
			db = db.Where("(radius_user.username ILIKE ? OR radius_user.realname ILIKE ? OR radius_user.mobile ILIKE ?)", like, like, like)
		} else {
			qLower := strings.ToLower(q)
			like := "%" + qLower + "%"
			db = db.Where("(LOWER(radius_user.username) LIKE ? OR LOWER(radius_user.realname) LIKE ? OR LOWER(radius_user.mobile) LIKE ?)", like, like, like)
		}
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
