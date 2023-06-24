package vpe

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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func InitRouter() {

	webserver.GET("/admin/vpe", func(c echo.Context) error {
		return c.Render(http.StatusOK, "vpe", nil)
	})

	webserver.GET("/admin/vpe/options", func(c echo.Context) error {
		var data []models.NetVpe
		common.Must(app.GDB().Find(&data).Error)
		var options = make([]web.JsonOptions, 0)
		for _, d := range data {
			options = append(options, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Name,
			})
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.GET("/admin/vpe/vendor/options", func(c echo.Context) error {
		options := []web.JsonOptions{
			web.JsonOptions{Id: app.RadiusVendorStandard, Value: "Standard"},
			web.JsonOptions{Id: app.RadiusVendorHuawei, Value: "Huawei"},
			web.JsonOptions{Id: app.RadiusVendorH3c, Value: "H3c"},
			web.JsonOptions{Id: app.RadiusVendorZte, Value: "Zte"},
			web.JsonOptions{Id: app.RadiusVendorRadback, Value: "Radback"},
			web.JsonOptions{Id: app.RadiusVendorCisco, Value: "Cisco"},
			web.JsonOptions{Id: app.RadiusVendorMikrotik, Value: "Mikrotik"},
			web.JsonOptions{Id: app.RadiusVendorIkuai, Value: "Ikuai"},
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.GET("/admin/vpe/query", func(c echo.Context) error {
		var count, start int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40)
		var data []models.NetVpe
		getQuery := func() *gorm.DB {
			query := app.GDB().Model(&models.NetVpe{})
			if len(web.ParseSortMap(c)) == 0 {
				query = query.Order("updated_at desc")
			} else {
				for name, stype := range web.ParseSortMap(c) {
					query = query.Order(fmt.Sprintf("%s %s", name, stype))
				}
			}

			for name, value := range web.ParseEqualMap(c) {
				query = query.Where(fmt.Sprintf("%s = ?", name), value)
			}

			for name, value := range web.ParseFilterMap(c) {
				if common.InSlice(name, []string{"pnode_id"}) {
					query = query.Where(fmt.Sprintf("%s = ?", name), value)
				} else {
					query = query.Where(fmt.Sprintf("%s like ?", name), "%"+value+"%")
				}
			}

			keyword := c.QueryParam("keyword")
			if keyword != "" {
				query = query.Where("name like ?", "%"+keyword+"%").
					Or("remark like ?", "%"+keyword+"%").
					Or("identifier like ?", "%"+keyword+"%").
					Or("tags like ?", "%"+keyword+"%")
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

	webserver.POST("/admin/vpe/add", func(c echo.Context) error {
		form := new(models.NetVpe)
		form.ID = common.UUIDint64()
		form.CreatedAt = time.Now()
		form.UpdatedAt = time.Now()
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.CheckEmpty("VendorCode", form.VendorCode)
		common.CheckEmpty("Identifier", form.Identifier)
		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create VPE information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/vpe/update", func(c echo.Context) error {
		form := new(models.NetVpe)
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.CheckEmpty("VendorCode", form.VendorCode)
		common.CheckEmpty("Identifier", form.Identifier)
		common.Must(app.GDB().Where("id=?", form.ID).Updates(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Update VPE information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/vpe/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.NetVpe{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete VPE information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/vpe/import", func(c echo.Context) error {
		datas, err := webserver.ImportData(c, "vpe")
		common.Must(err)
		common.Must(app.GDB().Model(models.NetVpe{}).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(datas).Error)
		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.GET("/admin/vpe/export", func(c echo.Context) error {
		var data []models.NetVpe
		common.Must(app.GDB().Find(&data).Error)
		return webserver.ExportCsv(c, data, "vpe")
	})

}
