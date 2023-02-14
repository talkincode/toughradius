package settings

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/assets"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
)

func InitRouter() {

	// settings page
	webserver.GET("/admin/settings", func(c echo.Context) error {
		return c.Render(http.StatusOK, "settings", nil)
	})

	// query settings
	webserver.GET("/admin/settings/:type/query", func(c echo.Context) error {
		ctype := c.Param("type")
		var result = make(map[string]interface{})
		var data []models.SysConfig
		if err := app.GDB().Where("type", ctype).Order("sort").Find(&data).Error; err != nil {
			log.Error(err)
			return c.JSON(http.StatusOK, result)
		}
		for _, datum := range data {
			result[datum.Name] = datum.Value
		}
		return c.JSON(http.StatusOK, result)
	})

	webserver.GET("/admin/settings/configlist", func(c echo.Context) error {
		type item struct {
			Name  string `json:"name"`
			Title string `json:"title"`
			Icon  string `json:"icon"`
		}
		var data []item
		data = append(data, item{Name: "system", Title: "System config", Icon: "mdi mdi-cogs"})
		data = append(data, item{Name: "radius", Title: "Radius config", Icon: "mdi mdi-radius"})
		data = append(data, item{Name: "tr069", Title: "Tr069 config", Icon: "mdi mdi-switch"})
		return c.JSON(http.StatusOK, data)
	})

	// update settings
	webserver.POST("/admin/settings/save", func(c echo.Context) error {
		var op, id, value string
		web.NewParamReader(c).
			ReadRequiedString(&op, "webix_operation").
			ReadRequiedString(&id, "id").
			ReadRequiedString(&value, "value")
		switch op {
		case "update":
			app.GDB().Model(&models.SysConfig{}).Where("id=?", id).Updates(map[string]interface{}{
				"value": value,
			})
			return c.JSON(http.StatusOK, map[string]interface{}{"status": "updated"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{})
	})

	webserver.POST("/admin/settings/add", func(c echo.Context) error {
		form := new(models.SysConfig)
		form.ID = common.UUIDint64()
		form.CreatedAt = time.Now()
		form.UpdatedAt = time.Now()
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.CheckEmpty("sort", form.Sort)
		common.CheckEmpty("type", form.Type)

		var count int64 = 0
		app.GDB().Model(models.SysConfig{}).Where("type=? and name = ?", form.Type, form.Name).Count(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, web.RestError("configuration name already exists"))
		}

		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create settings information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/settings/update", func(c echo.Context) error {
		values, err := c.FormParams()
		common.Must(err)
		ctype := c.FormValue("ctype")
		for k, _ := range values {
			if common.InSlice(k, []string{"submit", "ctype"}) {
				continue
			}
			app.GDB().Debug().Model(models.SysConfig{}).Where("type=? and name = ?", ctype, k).Update("value", c.FormValue(k))
		}
		webserver.PubOpLog(c, fmt.Sprintf("Update settings information：%v", values))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/settings/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.SysConfig{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete setting information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/settings/tr069/quickset", func(c echo.Context) error {
		return c.Render(http.StatusOK, "cwmp_quickset", nil)
	})

	webserver.GET("/admin/settings/tr069/quickset/mikrotik_cpe_setup_tr069.rsc", func(c echo.Context) error {
		ret := app.GApp().InjectCwmpConfigVars("", assets.Tr069Mikrotik, map[string]string{
			"CacrtContent": app.GApp().GetCacrtContent(),
		})
		c.Response().Header().Set("Content-Disposition", "attachment;filename=mikrotik_cpe_setup_tr069.rsc")
		return c.String(http.StatusOK, ret)
	})

}
