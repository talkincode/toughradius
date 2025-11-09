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

type radiusUserPayload struct {
	NodeID     int64  `json:"node_id,string"`
	ProfileID  int64  `json:"profile_id,string"`
	Realname   string `json:"realname"`
	Mobile     string `json:"mobile"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	AddrPool   string `json:"addr_pool"`
	Vlanid1    int    `json:"vlanid1"`
	Vlanid2    int    `json:"vlanid2"`
	IpAddr     string `json:"ip_addr"`
	MacAddr    string `json:"mac_addr"`
	BindVlan   int    `json:"bind_vlan"`
	BindMac    int    `json:"bind_mac"`
	ExpireTime string `json:"expire_time"`
	Status     string `json:"status"`
	Remark     string `json:"remark"`
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
	var payload radiusUserPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析用户参数", nil)
	}
	payload.Username = strings.TrimSpace(payload.Username)
	if payload.Username == "" || payload.Password == "" || payload.ProfileID == 0 {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "username、password、profile_id 不能为空", nil)
	}

	var exists int64
	app.GDB().Model(&domain.RadiusUser{}).Where("username = ?", payload.Username).Count(&exists)
	if exists > 0 {
		return fail(c, http.StatusConflict, "USERNAME_EXISTS", "账号已存在", nil)
	}

	var profile domain.RadiusProfile
	if err := app.GDB().Where("id = ?", payload.ProfileID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fail(c, http.StatusBadRequest, "PROFILE_NOT_FOUND", "关联的计费策略不存在", nil)
		}
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询计费策略失败", err.Error())
	}

	expire, err := parseTimeInput(payload.ExpireTime, time.Now().AddDate(1, 0, 0))
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_EXPIRE_TIME", "expire_time 格式不正确", nil)
	}

	user := domain.RadiusUser{
		ID:         common.UUIDint64(),
		NodeId:     payload.NodeID,
		ProfileId:  payload.ProfileID,
		Realname:   payload.Realname,
		Mobile:     payload.Mobile,
		Username:   payload.Username,
		Password:   payload.Password,
		AddrPool:   common.If(payload.AddrPool != "", payload.AddrPool, profile.AddrPool).(string),
		ActiveNum:  profile.ActiveNum,
		UpRate:     profile.UpRate,
		DownRate:   profile.DownRate,
		Vlanid1:    payload.Vlanid1,
		Vlanid2:    payload.Vlanid2,
		IpAddr:     payload.IpAddr,
		MacAddr:    payload.MacAddr,
		BindVlan:   payload.BindVlan,
		BindMac:    payload.BindMac,
		ExpireTime: expire,
		Status:     strings.ToLower(common.If(payload.Status == "", common.ENABLED, payload.Status).(string)),
		Remark:     payload.Remark,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

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

	var payload radiusUserPayload
	if err := c.Bind(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析用户参数", nil)
	}

	var user domain.RadiusUser
	if err := app.GDB().Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusNotFound, "USER_NOT_FOUND", "用户不存在", nil)
	} else if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询用户失败", err.Error())
	}

	if payload.ProfileID != 0 && payload.ProfileID != user.ProfileId {
		var profile domain.RadiusProfile
		if err := app.GDB().Where("id = ?", payload.ProfileID).First(&profile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fail(c, http.StatusBadRequest, "PROFILE_NOT_FOUND", "关联的计费策略不存在", nil)
			}
			return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询计费策略失败", err.Error())
		}
		user.ProfileId = payload.ProfileID
		user.ActiveNum = profile.ActiveNum
		user.UpRate = profile.UpRate
		user.DownRate = profile.DownRate
		user.AddrPool = profile.AddrPool
	}

	if payload.NodeID != 0 {
		user.NodeId = payload.NodeID
	}
	if payload.Realname != "" {
		user.Realname = payload.Realname
	}
	if payload.Mobile != "" {
		user.Mobile = payload.Mobile
	}
	if payload.Username != "" {
		user.Username = payload.Username
	}
	if payload.Password != "" {
		user.Password = payload.Password
	}
	if payload.AddrPool != "" {
		user.AddrPool = payload.AddrPool
	}
	if payload.Vlanid1 != 0 {
		user.Vlanid1 = payload.Vlanid1
	}
	if payload.Vlanid2 != 0 {
		user.Vlanid2 = payload.Vlanid2
	}
	if payload.IpAddr != "" {
		user.IpAddr = payload.IpAddr
	}
	if payload.MacAddr != "" {
		user.MacAddr = payload.MacAddr
	}
	if payload.BindVlan != 0 {
		user.BindVlan = payload.BindVlan
	}
	if payload.BindMac != 0 {
		user.BindMac = payload.BindMac
	}
	if payload.Remark != "" {
		user.Remark = payload.Remark
	}
	if payload.Status != "" {
		user.Status = strings.ToLower(payload.Status)
	}
	if payload.ExpireTime != "" {
		expire, err := parseTimeInput(payload.ExpireTime, user.ExpireTime)
		if err != nil {
			return fail(c, http.StatusBadRequest, "INVALID_EXPIRE_TIME", "expire_time 格式不正确", nil)
		}
		user.ExpireTime = expire
	}
	user.UpdatedAt = time.Now()

	if err := app.GDB().Save(&user).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "更新用户失败", err.Error())
	}

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
