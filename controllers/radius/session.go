package radius

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
	"gorm.io/gorm"
)

func InitSessionRouter() {

	// RADIUS Online user page display assets/templates/radius_session.html
	webserver.GET("/admin/radius/session", func(c echo.Context) error {
		return c.Render(http.StatusOK, "radius_session", map[string]interface{}{})
	})

	// Radius online quantity
	webserver.GET("/admin/radius/online/count/:username", func(c echo.Context) error {
		username := c.Param("username")
		var count int64
		app.GDB().Model(models.RadiusOnline{}).Where("username = ?", username).Count(&count)
		return c.String(200, fmt.Sprintf("%d", count))
	})

	webserver.GET("/admin/radius/session/query", func(c echo.Context) error {
		var count, start int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40)
		var data []models.RadiusOnline
		getQuery := func() *gorm.DB {
			query := app.GDB().Model(&models.RadiusOnline{})
			for name, stype := range web.ParseSortMap(c) {
				query = query.Order(fmt.Sprintf("%s %s", name, stype))
			}
			for name, value := range web.ParseFilterMap(c) {
				query = query.Where(fmt.Sprintf("%s like ?", name), "%"+value+"%")
			}
			keyword := c.QueryParam("keyword")
			if keyword != "" {
				query = query.Where("username like ?", "%"+keyword+"%").
					Or("nas_addr like ?", "%"+keyword+"%").
					Or("framed_ipaddr like ?", "%"+keyword+"%").
					Or("nas_paddr like ?", "%"+keyword+"%")
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

	webserver.GET("/admin/radius/session/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.RadiusOnline{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("delete RADIUS online：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

}
