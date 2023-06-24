package node

import (
	"fmt"
	"net/http"
	"strings"

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

	webserver.GET("/admin/node", func(c echo.Context) error {
		return c.Render(http.StatusOK, "node", map[string]interface{}{
			"oprlevel": webserver.GetCurrUserlevel(c),
		})
	})

	webserver.GET("/admin/node/options", func(c echo.Context) error {
		var data []models.NetNode
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

	webserver.GET("/admin/node/query", func(c echo.Context) error {
		var data []models.NetNode
		err := app.GDB().Find(&data).Error
		common.Must(err)
		return c.JSON(http.StatusOK, data)
	})

	webserver.POST("/admin/node/add", func(c echo.Context) error {
		form := new(models.NetNode)
		form.ID = common.UUIDint64()
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create node information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/node/update", func(c echo.Context) error {
		form := new(models.NetNode)
		common.Must(c.Bind(form))
		common.CheckEmpty("name", form.Name)
		common.Must(app.GDB().Save(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Update node information：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/node/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.NetNode{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Delete node information：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/node/import", func(c echo.Context) error {
		datas, err := webserver.ImportData(c, "node")
		common.Must(err)
		common.Must(app.GDB().Debug().Model(models.NetNode{}).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&datas).Error)
		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.GET("/admin/node/export", func(c echo.Context) error {
		var data []models.NetNode
		common.Must(app.GDB().Find(&data).Error)
		return webserver.ExportCsv(c, data, "node")
	})

}
