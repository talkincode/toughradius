package cwmppreset

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/assets"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
	"gopkg.in/yaml.v2"
)

func InitRouter() {

	webserver.GET("/admin/cwmp/preset", func(c echo.Context) error {
		return c.Render(http.StatusOK, "cwmp_preset", map[string]interface{}{
			"oprlevel": webserver.GetCurrUserlevel(c),
		})
	})

	webserver.GET("/admin/cwmp/preset/template", func(c echo.Context) error {
		return c.String(http.StatusOK, assets.Tr069PresetTemplate)
	})

	webserver.GET("/admin/cwmp/preset/task", func(c echo.Context) error {
		return c.Render(http.StatusOK, "cwmp_preset_task", nil)
	})

	webserver.GET("/admin/cwmp/preset/options", func(c echo.Context) error {
		var data []models.CwmpPreset
		common.Must(app.GDB().Find(&data).Error)
		var opts = make([]web.JsonOptions, 0)
		for _, d := range data {
			opts = append(opts, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Name,
			})
		}
		return c.JSON(http.StatusOK, opts)
	})

	webserver.GET("/admin/cwmp/preset/sched/options", func(c echo.Context) error {
		var opts = make([]web.JsonOptions, 0)
		opts = append(opts, web.JsonOptions{Id: "5m", Value: "every 5 minutes"})
		opts = append(opts, web.JsonOptions{Id: "10m", Value: "every 10 minutes"})
		opts = append(opts, web.JsonOptions{Id: "30m", Value: "every 30 minutes"})
		opts = append(opts, web.JsonOptions{Id: "1h", Value: "every 1 hours"})
		opts = append(opts, web.JsonOptions{Id: "4h", Value: "every 4 hours"})
		opts = append(opts, web.JsonOptions{Id: "8h", Value: "every 8 hours"})
		opts = append(opts, web.JsonOptions{Id: "12h", Value: "every 12 hours"})
		opts = append(opts, web.JsonOptions{Id: "daily@h0", Value: "0:00 a.m. every day"})
		for i := 1; i < 24; i++ {
			key := fmt.Sprintf("daily@h%d", i)
			kdesc := fmt.Sprintf("%d o'clock every day", i)
			opts = append(opts, web.JsonOptions{Id: key, Value: kdesc})
		}
		return c.JSON(http.StatusOK, opts)
	})

	webserver.GET("/admin/cwmp/preset/query", func(c echo.Context) error {
		prequery := web.NewPreQuery(c).
			DefaultOrderBy("updated_at desc").
			KeyFields("name", "event", "task_tags", "content")

		result, err := web.QueryPageResult[models.CwmpPreset](c, app.GDB(), prequery)
		if err != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, result)
	})

	webserver.GET("/admin/cwmp/preset/execute", func(c echo.Context) error {
		id := c.QueryParam("pid")
		if id == "" {
			return c.JSON(http.StatusOK, web.RestError("ID is empty"))
		}

		var count int64
		app.GDB().Model(&models.CwmpPresetTask{}).
			Where("preset_id = ? and status = ?", id, "pending").Count(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, web.RestError(app.Trans("cwmp",
				"The task is already in progress, please try again later")))
		}
		var snarray = make([]string, 0)
		snlist := c.QueryParam("snlist")
		if snlist != "" {
			snarray = strings.Split(snlist, ",")
		}
		err := app.GApp().CreateCwmpPresetTaskById(id, snarray)
		if err != nil {
			return c.JSON(http.StatusOK, web.RestError(err.Error()))
		}
		return c.JSON(http.StatusOK,
			web.RestSucc(app.Trans("cwmp",
				"The task has been triggered, please do not trigger it again in a short time")))
	})

	webserver.GET("/admin/cwmp/presettask/query", queryCwmpPresetTask)

	webserver.ApiGET("/api/cwmp/presettask/query", queryCwmpPresetTask)

	webserver.POST("/admin/cwmp/preset/add", func(c echo.Context) error {
		form := new(models.CwmpPreset)
		form.ID = common.UUIDint64()
		common.Must(c.Bind(form))
		common.MustNotEmpty("Event", form.Event)
		common.MustNotEmpty("Content", form.Content)
		var data models.CwmpPresetContent
		err := yaml.Unmarshal([]byte(form.Content), &data)
		if err != nil {
			return c.JSON(http.StatusOK, web.RestError("Yaml format error"))
		}
		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create Cwmp Preset Information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/cwmp/preset/update", func(c echo.Context) error {
		form := new(models.CwmpPreset)
		common.Must(c.Bind(form))
		common.MustNotEmpty("Event", form.Event)
		common.MustNotEmpty("Content", form.Content)
		var data models.CwmpPresetContent
		err := yaml.Unmarshal([]byte(form.Content), &data)
		if err != nil {
			return c.JSON(http.StatusOK, web.RestError("Yaml format error"))
		}
		common.Must(app.GDB().Save(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Modify CwmpPreset information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/cwmp/preset/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.CwmpPreset{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete Cwmp Preset information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/cwmp/presettask/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.CwmpPresetTask{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete CwmpPresetTask information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

}

//	@Summary		Query cwmp preset task
//	@Description	Query cwmp preset task
//	@Tags			TR069
//	@Accept			json
//	@Produce		json
//	@Param			cpe_id	query	string	false	"cpe_id"
//	@Param			keyword	query	string	false	"keyword"
//	@Security		BearerAuth
//	@Success		200	{array}	models.CwmpPresetTask
//	@Router			/api/cwmp/preset/task/query [get]
func queryCwmpPresetTask(c echo.Context) error {
	prequery := web.NewPreQuery(c).
		DefaultOrderBy("created_at desc").
		DateRange2("starttime", "endtime", "created_at", time.Now().Add(-time.Hour*24), time.Now()).
		QueryField("sn", "sn").
		KeyFields("sn", "name", "batch", "event")

	result, err := web.QueryPageResult[models.CwmpPresetTask](c, app.GDB(), prequery)
	if err != nil {
		return c.JSON(http.StatusOK, common.EmptyList)
	}
	return c.JSON(http.StatusOK, result)
}
