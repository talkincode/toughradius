package cpe

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

	webserver.GET("/admin/cpe/query", func(c echo.Context) error {
		var count, start int
		var nodeId string
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40).
			ReadString(&nodeId, "node_id")
		var data []models.NetCpe
		getQuery := func() *gorm.DB {
			query := app.GDB().Model(&models.NetCpe{})

			if len(web.ParseSortMap(c)) == 0 {
				query = query.Order("updated_at desc")
			} else {
				for name, stype := range web.ParseSortMap(c) {
					query = query.Order(fmt.Sprintf("%s %s", name, stype))
				}
			}

			if nodeId != "" {
				query = query.Where("node_id = ? ", nodeId)
			}

			for name, value := range web.ParseEqualMap(c) {
				query = query.Where(fmt.Sprintf("%s = ?", name), value)
			}

			for name, value := range web.ParseFilterMap(c) {
				if common.InSlice(name, []string{"node_id", "customer_id"}) {
					query = query.Where(fmt.Sprintf("%s = ?", name), value)
				} else {
					query = query.Where(fmt.Sprintf("%s like ?", name), "%"+value+"%")
				}
			}
			keyword := c.QueryParam("keyword")
			if keyword != "" {
				query = query.Where("name like ?", "%"+keyword+"%").
					Or("remark like ?", "%"+keyword+"%").
					Or("sn like ?", "%"+keyword+"%").
					Or("rd_ipaddr like ?", "%"+keyword+"%").
					Or("model like ?", "%"+keyword+"%")
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
		return webserver.ExportCsv(c, data, "cpe")
	})

}
