package cwmppreset

import (
	"fmt"
	"net/http"
	"strings"

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
		var data []models.CwmpPreset
		err := app.GDB().Find(&data).Error
		common.Must(err)
		return c.JSON(http.StatusOK, data)
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

	webserver.GET("/admin/cwmp/presettask/query", func(c echo.Context) error {
		var count, start int
		var sn string
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40).
			ReadString(&sn, "sn")
		var data []models.CwmpPresetTask
		prequery := web.NewPreQuery(c).
			DefaultOrderBy("updated_at desc").
			KeyFields("name")

		if sn != "" {
			prequery = prequery.SetParam("sn", sn)
		}

		var total int64
		common.Must(prequery.Query(app.GDB().Model(&models.CwmpPresetTask{})).Count(&total).Error)

		query := prequery.Query(app.GDB().Model(&models.CwmpPresetTask{})).Offset(start).Limit(count)
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, &web.PageResult{TotalCount: total, Pos: int64(start), Data: data})
	})

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
