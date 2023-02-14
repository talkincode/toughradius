package charts

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/echarts"
	"github.com/talkincode/toughradius/models"
	"github.com/talkincode/toughradius/webserver"
)

func InitRouter() {

	webserver.GET("/admin/charts/cpe/:type/pie", func(c echo.Context) error {
		stype := c.Param("type")
		statname := stype
		switch stype {
		case "model":
			statname = "TeamsBox 型号统计"
		case "version":
			statname = "TeamsBox 版本统计"
		}

		return c.Render(http.StatusOK, "cpe_stat_pie", map[string]string{
			"stattype": stype,
			"statname": statname,
		})
	})

	webserver.GET("/admin/charts/cpe/:type/pie/data", func(c echo.Context) error {
		stype := c.Param("type")
		var cpes []models.NetCpe
		common.Must(app.GDB().Find(&cpes).Error)
		var statdata = map[string]*echarts.NameValuePair{
			"unknow": {Value: 0, Name: "未知"},
		}
		for _, dev := range cpes {
			var name string
			switch stype {
			case "model":
				name = dev.Model
			case "version":
				name = dev.SoftwareVersion
			default:
				continue
			}
			if name == "" {
				continue
			}
			if _, ok := statdata[name]; !ok {
				statdata[name] = &echarts.NameValuePair{Name: name, Value: 1}
			} else {
				statdata[name].Incr()
			}
		}

		result := make([]*echarts.NameValuePair, 0)
		for _, pair := range statdata {
			result = append(result, pair)
		}
		return c.JSON(http.StatusOK, result)
	})

}
