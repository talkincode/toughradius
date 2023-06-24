package cpe

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/webserver"
	"gorm.io/gorm/clause"
)

func InitRouter() {

	webserver.GET("/admin/cpe", func(c echo.Context) error {
		return c.Render(http.StatusOK, "cpe", nil)
	})

	webserver.GET("/admin/cpe/options", func(c echo.Context) error {
		var ids = c.QueryParam("ids")
		var data []models.NetCpe
		query := app.GDB().Model(&models.NetCpe{})
		if ids != "" {
			query = query.Where("id in (?)", strings.Split(ids, ","))
		}
		common.Must(query.Find(&data).Error)
		var options = make([]web.JsonOptions, 0)
		for _, d := range data {
			options = append(options, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Name,
			})
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.GET("/admin/cpe/sn/options", func(c echo.Context) error {
		var snlist = c.QueryParam("snlist")
		var data []models.NetCpe
		query := app.GDB().Model(&models.NetCpe{})
		if snlist != "" {
			query = query.Where("sn in (?)", strings.Split(snlist, ","))
		}
		common.Must(query.Find(&data).Error)
		var options = make([]web.JsonOptions, 0)
		for _, d := range data {
			options = append(options, web.JsonOptions{
				Id:    d.Sn,
				Value: d.Name,
			})
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.ApiGET("/cpe/query", queryCpe)

	webserver.GET("/admin/cpe/query", queryCpe)

	webserver.GET("/admin/cpe/params", func(c echo.Context) error {
		var sn string
		web.NewParamReader(c).ReadRequiedString(&sn, "sn")
		var data []models.NetCpeParam
		err := app.GDB().Where("sn=?", sn).Order("name asc").Find(&data).Error
		if err != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, data)
	})

	webserver.GET("/admin/cpe/get", func(c echo.Context) error {
		var id string
		web.NewParamReader(c).
			ReadRequiedString(&id, "id")
		var data models.NetCpe
		common.Must(app.GDB().Where("id=?", id).First(&data).Error)
		return c.JSON(http.StatusOK, data)
	})

	webserver.POST("/admin/cpe/add", func(c echo.Context) error {
		form := new(models.NetCpe)
		form.ID = common.UUIDint64()
		form.CreatedAt = time.Now()
		form.UpdatedAt = time.Now()
		common.Must(c.Bind(form))
		common.CheckEmpty("sn", form.Sn)
		common.CheckEmpty("name", form.Name)

		var count int64 = 0
		app.GDB().Model(models.NetCpe{}).Where("sn=?", form.Sn).Count(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, web.RestError("SN 已经存在"))
		}

		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create CPE information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/cpe/update", func(c echo.Context) error {
		form := new(models.NetCpe)
		common.Must(c.Bind(form))
		common.CheckEmpty("sn", form.Sn)
		common.CheckEmpty("name", form.Name)
		app.GDB().Where("id=?", form.ID).Updates(form)
		app.GApp().CwmpTable().ClearCwmpCpeCache(form.Sn)
		webserver.PubOpLog(c, fmt.Sprintf("Update CPE information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/cpe/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		var sns []string
		app.GDB().Raw(`select sn from net_cpe where id in ?`, strings.Split(ids, ",")).Scan(&sns)
		for _, sn := range sns {
			app.GApp().CwmpTable().ClearCwmpCpe(sn)
		}
		common.Must(app.GDB().Delete(models.NetCpe{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete CPE information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/cpe/import", func(c echo.Context) error {
		datas, err := webserver.ImportData(c, "cpe")
		common.Must(err)
		common.Must(app.GDB().Model(models.NetCpe{}).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(datas).Error)
		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.GET("/admin/cpe/export", func(c echo.Context) error {
		var data []models.NetCpe
		common.Must(app.GDB().Find(&data).Error)
		switch c.QueryParam("fmt") {
		case "csv":
			return webserver.ExportCsv(c, data, "cpe")
		case "json":
			return webserver.ExportJson(c, data, "cpe")
		default:
			return webserver.ExportCsv(c, data, "cpe")
		}
	})

}

// @Summary		Query CPE list
// @Description	Query cpe list
// @Tags			CPE
// @Accept			json
// @Produce		json
// @Param			node_id		query	string	false	"node_id"
// @Param			customer_id	query	string	false	"customer_id"
// @Param			keyword		query	string	false	"keyword"
// @Security		BearerAuth
// @Success		200	{array}	models.NetCpe
// @Router			/api/cpe/query [get]
func queryCpe(c echo.Context) error {
	prequery := web.NewPreQuery(c).
		DefaultOrderBy("name asc").
		QueryField("node_id", "node_id").
		QueryField("customer_id", "customer_id").
		QueryField("sn", "sn").
		KeyFields("sn", "name", "remark", "model")

	result, err := web.QueryPageResult[models.NetCpe](c, app.GDB(), prequery)
	if err != nil {
		return c.JSON(http.StatusOK, common.EmptyList)
	}
	return c.JSON(http.StatusOK, result)
}
