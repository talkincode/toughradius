package opr

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/webserver"
	"gorm.io/gorm"
)

// 设备数据管理

func InitRouter() {

	webserver.GET("/admin/opr", func(c echo.Context) error {
		return c.Render(http.StatusOK, "opr", nil)
	})

	webserver.GET("/admin/opr/options", func(c echo.Context) error {
		var data []models.SysOpr
		common.Must(app.GDB().Find(&data).Error)
		var options = make([]web.JsonOptions, 0)
		for _, d := range data {
			options = append(options, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Username,
			})
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.GET("/admin/opr/get", func(c echo.Context) error {
		var id string
		web.NewParamReader(c).ReadRequiedString(&id, "id")
		var data models.SysOpr
		err := app.GDB().Where("id=?", id).First(&data).Error
		if err != nil {
			return c.JSON(http.StatusOK, common.EmptyData)
		}
		return c.JSON(http.StatusOK, data)
	})

	webserver.GET("/admin/opr/current", func(c echo.Context) error {
		var opr = webserver.GetCurrUser(c)
		return c.JSON(http.StatusOK, opr)
	})

	webserver.GET("/admin/opr/query", func(c echo.Context) error {
		sess, _ := session.Get(webserver.UserSession, c)
		var data []models.SysOpr
		getQuery := func() *gorm.DB {
			query := app.GDB().Model(&models.SysOpr{})
			for name, stype := range web.ParseSortMap(c) {
				query = query.Order(fmt.Sprintf("%s %s", name, stype))
			}
			for name, value := range web.ParseFilterMap(c) {
				query = query.Where(fmt.Sprintf("%s = ?", name), value)
			}
			return query
		}
		clevel := sess.Values[webserver.UserSessionLevel]
		query := getQuery()
		if clevel != "super" {
			query = query.Where("level <> 'super'")
		}
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, data)
	})

	webserver.POST("/admin/opr/add", func(c echo.Context) error {
		form := new(models.SysOpr)
		common.Must(c.Bind(form))
		common.MustNotEmpty("username", form.Username)
		common.MustNotEmpty("password", form.Password)
		form.Password = common.Sha256HashWithSalt(form.Password, common.SecretSalt)
		if common.IsEmptyOrNA(form.Status) {
			form.Status = common.ENABLED
		}
		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create operator information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/opr/update", func(c echo.Context) error {
		form := new(models.SysOpr)
		common.Must(c.Bind(form))
		common.MustNotEmpty("username", form.Username)
		if !common.IsEmptyOrNA(form.Password) {
			form.Password = common.Sha256HashWithSalt(form.Password, common.SecretSalt)
		}
		if common.IsEmptyOrNA(form.Status) {
			form.Status = common.ENABLED
		}
		common.Must(app.GDB().Save(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Update operator information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/opr/uppassword", func(c echo.Context) error {
		oldpassword := c.FormValue("oldpassword")
		cpassword := c.FormValue("cpassword")
		password := c.FormValue("password")

		cuser := webserver.GetCurrUser(c)

		if password == "" {
			return c.JSON(http.StatusOK, web.RestError("password can not be blank"))
		}

		if password != cpassword {
			return c.JSON(http.StatusOK, web.RestError("Confirm passwords do not match"))
		}

		if common.Sha256HashWithSalt(oldpassword, common.SecretSalt) != cuser.Password {
			return c.JSON(http.StatusOK, web.RestError("Old password does not match"))
		}

		newPasswdEnc := common.Sha256HashWithSalt(password, common.SecretSalt)
		app.GDB().Model(&models.SysOpr{}).Where("id=?", cuser.ID).Update("password", newPasswdEnc)
		webserver.PubOpLog(c, fmt.Sprintf("Update operator password for %v", cuser.Username))
		return c.JSON(http.StatusOK, web.RestSucc("password has been updated"))
	})

	webserver.GET("/admin/opr/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Where("level <> 'super'").Delete(models.SysOpr{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete operator information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

}
