package radius

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

func InitLogsRouter() {

	// 认证日志页面展示 assets/templates/radius_authlog.html
	webserver.GET("/admin/radius/authlog", func(c echo.Context) error {
		return c.Render(http.StatusOK, "radius_authlog", map[string]interface{}{})
	})

	// 记账日志页面展示 assets/templates/radius_accounting.html
	webserver.GET("/admin/radius/accounting", func(c echo.Context) error {
		return c.Render(http.StatusOK, "radius_accounting", nil)
	})

	// 记账日志查询
	webserver.GET("/admin/radius/accounting/query", func(c echo.Context) error {
		var count, start int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40)
		var data []models.RadiusAccounting
		prequery := web.NewPreQuery(c).
			DefaultOrderBy("acct_stop_time desc").
			DateRange2("starttime", "endtime", "acct_stop_time", time.Now().Add(-time.Hour*8), time.Now()).
			KeyFields("username", "framed_ipaddr", "mac_addr")

		var total int64
		common.Must(prequery.Query(app.GDB().Model(&models.RadiusAccounting{})).Count(&total).Error)

		query := prequery.Query(app.GDB().Debug().Model(&models.RadiusAccounting{})).Offset(start).Limit(count)
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, &web.PageResult{TotalCount: total, Pos: int64(start), Data: data})
	})

}
