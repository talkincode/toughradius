package logs

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
)

// 操作日志查询

func initOplogRouter() {

	webserver.GET("/admin/oplog", func(c echo.Context) error {
		return c.Render(http.StatusOK, "oplog", map[string]interface{}{})
	})

	webserver.GET("/admin/oplog/query", func(c echo.Context) error {
		var count, start int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40)
		var data []models.SysOprLog
		prequery := web.NewPreQuery(c).
			DefaultOrderBy("opt_time desc").
			DateRange2("starttime", "endtime", "opt_time", time.Now().Add(-time.Hour*8), time.Now()).
			KeyFields("opr_name", "opt_action", "opr_ip", "opt_desc")

		var total int64
		common.Must(prequery.Query(app.GDB().Model(&models.SysOprLog{})).Count(&total).Error)

		query := prequery.Query(app.GDB().Debug().Model(&models.SysOprLog{})).Offset(start).Limit(count)
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, &web.PageResult{TotalCount: total, Pos: int64(start), Data: data})
	})

}
