package radius

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InitUserRouter Radius User additions, deletions, modification and query
func InitUserRouter() {

	webserver.GET("/admin/radius/users", func(c echo.Context) error {
		return c.Render(http.StatusOK, "radius_users", map[string]interface{}{})
	})

	webserver.GET("/admin/radius/users/options", func(c echo.Context) error {
		var data []models.RadiusUser
		query := app.GDB().Model(&models.RadiusUser{})
		cid := c.QueryParam("node_id")
		if cid != "" {
			query = query.Where("node_id = ?", cid)
		}
		ids := c.QueryParam("ids")
		if ids != "" {
			query = query.Where("id in (?)", strings.Split(ids, ","))
		}
		common.Must(query.Find(&data).Error)
		var options = make([]web.JsonOptions, 0)
		for _, d := range data {
			options = append(options, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Username,
			})
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.GET("/admin/radius/users/query", func(c echo.Context) error {
		var count, start, expireDays int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40).
			ReadInt(&expireDays, "expire_days", 0)
		var data []models.RadiusUser
		getQuery := func() *gorm.DB {
			// query := app.GDB().Model(&models.RadiusUser{})
			query := app.GDB().Model(&models.RadiusUser{}).Select("radius_user.*, coalesce(ro.count, 0) as online_count").
				Joins("left join (select username, count(1) as count from radius_online  group by username) ro on radius_user.username = ro.username")
			if len(web.ParseSortMap(c)) == 0 {
				query = query.Order("radius_user.updated_at desc")
			} else {
				mobj := models.RadiusUser{}
				for name, stype := range web.ParseSortMap(c) {
					if common.GetFieldType(mobj, name) == "string" {
						query = query.Order(fmt.Sprintf("convert_to(radius_user.%s,'UTF8')  %s ", name, stype))
					} else {
						if name == "online_count" {
							query = query.Order(fmt.Sprintf("%s %s ", name, stype))
						} else {
							query = query.Order(fmt.Sprintf("radius_user.%s %s ", name, stype))
						}
					}
				}
			}

			for name, value := range web.ParseEqualMap(c) {
				query = query.Where(fmt.Sprintf("radius_user.%s = ?", name), value)
			}

			for name, value := range web.ParseFilterMap(c) {
				if common.InSlice(name, []string{"profile_id", "pnode_id"}) {
					query = query.Where(fmt.Sprintf("radius_user.%s = ?", name), value)
				} else {
					query = query.Where(fmt.Sprintf("radius_user.%s like ?", name), "%"+value+"%")
				}
			}

			if expireDays > 1 {
				query = query.Where("radius_user.expire_time <=  ?", time.Now().Add(time.Hour*24*time.Duration(expireDays)))
			}

			keyword := c.QueryParam("keyword")
			if keyword != "" {
				query = query.Where("radius_user.username like ?", "%"+keyword+"%").
					Or("radius_user.remark like ?", "%"+keyword+"%").
					Or("radius_user.realname like ?", "%"+keyword+"%").
					Or("radius_user.tags like ?", "%"+keyword+"%").
					Or("radius_user.mobile like ?", "%"+keyword+"%")
			}
			return query
		}
		var total int64
		common.Must(getQuery().Count(&total).Error)

		query := getQuery().Offset(start).Limit(count)
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, &web.PageResult{TotalCount: total, Pos: int64(start), Data: data})
	})

	webserver.GET("/admin/radius/users/get", func(c echo.Context) error {
		var id string
		web.NewParamReader(c).
			ReadRequiedString(&id, "id")
		var data models.RadiusUser
		common.Must(app.GDB().Where("id=?", id).First(&data).Error)
		return c.JSON(http.StatusOK, data)
	})

	webserver.POST("/admin/radius/users/add", func(c echo.Context) error {
		form := new(models.RadiusUser)
		common.Must(c.Bind(form))
		form.ID = common.UUIDint64()
		form.CreatedAt = time.Now()
		form.UpdatedAt = time.Now()
		timestr := c.FormValue("expire_time")[:10] + " 23:59:59"
		form.ExpireTime, _ = time.Parse("2006-01-02 15:04:05", timestr)
		common.CheckEmpty("username", form.Username)
		common.CheckEmpty("password", form.Password)

		var count int64 = 0
		app.GDB().Model(models.RadiusUser{}).Where("username=?", form.Username).Count(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, web.RestError("Username already exists"))
		}

		var profile models.RadiusProfile
		common.Must(app.GDB().Where("id=?", form.ProfileId).First(&profile).Error)

		form.ActiveNum = profile.ActiveNum
		form.UpRate = profile.UpRate
		form.DownRate = profile.DownRate
		form.AddrPool = profile.AddrPool

		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create RADIUS user：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/users/update", func(c echo.Context) error {
		form := new(models.RadiusUser)
		common.Must(c.Bind(form))
		timestr := c.FormValue("expire_time")[:10] + " 23:59:59"
		form.ExpireTime, _ = time.Parse("2006-01-02 15:04:05", timestr)
		common.CheckEmpty("username", form.Username)
		common.CheckEmpty("password", form.Password)
		var profile models.RadiusProfile
		common.Must(app.GDB().Where("id=?", form.ProfileId).First(&profile).Error)
		form.ActiveNum = profile.ActiveNum
		form.UpRate = profile.UpRate
		form.DownRate = profile.DownRate
		form.AddrPool = profile.AddrPool

		common.Must(app.GDB().Save(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Update RADIUS users：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/users/batchupdate", func(c echo.Context) error {
		profileId := c.FormValue("profile_id")
		userIds := c.FormValue("user_ids")
		expireDay := c.FormValue("expire_time")
		status := c.FormValue("status")
		if userIds == "" {
			return c.JSON(200, web.RestError("No user account selected"))
		}
		expire, err := time.Parse("2006-01-02", expireDay[:10])
		if err != nil {
			return c.JSON(200, web.RestError("Wrong time format"))
		}
		var profileNone = false
		var profile models.RadiusProfile
		err = app.GDB().Where("id=?", profileId).First(&profile).Error
		if err != nil {
			profileNone = true
		}

		var succ, errs int
		for _, uid := range strings.Split(userIds, ",") {
			var user models.RadiusUser
			err = app.GDB().Where("id=?", uid).First(&user).Error
			if err != nil {
				errs++
				continue
			}
			var data = map[string]interface{}{
				"expire_time": expire,
				"updated_at":  time.Now(),
			}

			if !profileNone {
				data["profile_id"] = profileId
				data["active_num"] = profile.ActiveNum
				data["up_rate"] = profile.UpRate
				data["down_rate"] = profile.DownRate
				data["addr_pool"] = profile.AddrPool
			}
			if common.InSlice(status, []string{"enabled", "disabled"}) {
				data["status"] = status
			}

			r := app.GDB().Debug().Model(&models.RadiusUser{}).Where("id=?", strings.TrimSpace(uid)).Updates(&data)
			if r.Error != nil {
				errs += 1
			} else {
				if r.RowsAffected > 0 {
					succ += 1
				}
			}
		}
		webserver.PubOpLog(c, fmt.Sprintf("Update RADIUS users in batches：succ=%d, errs=%d", succ, errs))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/radius/users/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.RadiusUser{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("delete RADIUS user：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/users/import", func(c echo.Context) error {
		datas, err := webserver.ImportData(c, "radius_users")
		common.Must(err)
		common.Must(app.GDB().Model(models.RadiusUser{}).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(datas).Error)
		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.POST("/admin/radius/users/importadd", func(c echo.Context) error {
		datas, err := webserver.ImportData(c, "radius_users")
		common.Must(err)
		var unames []string
		var unames2 []string
		app.GDB().Model(models.RadiusUser{}).Pluck("username", &unames)
		var users []models.RadiusUser
		for row, data := range datas {
			username := strings.ToLower(strings.TrimSpace(cast.ToString(data["username"])))
			if common.InSlice(username, unames) {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("row %d username %s exists", row+1, username)))
			}
			if common.InSlice(username, unames2) {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("row %d username %s duplicate", row+1, username)))
			}
			unames2 = append(unames2, username)
			password := cast.ToString(data["password"])
			expire := cast.ToString(data["expire_time"])
			mobile := cast.ToString(data["mobile"])
			remark := cast.ToString(data["remark"])
			realname := strings.TrimSpace(cast.ToString(data["realname"]))
			if realname == "" {
				realname = username
			}
			common.CheckEmpty("username", username)
			common.CheckEmpty("password", password)
			common.CheckEmpty("expire_time", expire)
			profileName := cast.ToString(data["profile"])
			if profileName == "" {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("行 %d customer or profile is empty", row+1)))
			}
			var profile models.RadiusProfile
			err := app.GDB().Where("name=?", profileName).First(&profile).Error
			if err != nil {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("row %d  Profile：%s does not exist", row+1, profileName)))
			}
			expireTime, err := time.Parse("2006-01-02", expire)
			if err != nil {
				expireTime, err = time.Parse("2006/01/02", expire)
				if err != nil {
					return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("row %d Expiration time format error：%s", row+1, expire)))
				}
			}
			user := models.RadiusUser{
				ID:          common.UUIDint64(),
				NodeId:      profile.NodeId,
				ProfileId:   profile.ID,
				Realname:    realname,
				Mobile:      mobile,
				Username:    username,
				Password:    password,
				AddrPool:    profile.AddrPool,
				ActiveNum:   profile.ActiveNum,
				UpRate:      profile.UpRate,
				DownRate:    profile.DownRate,
				IpAddr:      "",
				ExpireTime:  expireTime,
				Status:      "enabled",
				Remark:      remark,
				OnlineCount: 0,
				LastOnline:  time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			users = append(users, user)
		}
		err = app.GDB().Save(&users).Error
		if err != nil {
			return c.JSON(http.StatusOK, web.RestError(err.Error()))
		}
		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.GET("/admin/radius/users/export", func(c echo.Context) error {
		var data []models.RadiusUser
		common.Must(app.GDB().Find(&data).Error)
		return webserver.ExportCsv(c, data, "radius_users")
	})
}
