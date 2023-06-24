package dashboard

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/echarts"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/webserver"
)

func InitRouter() {
	webserver.GET("/admin/sysstatus", func(c echo.Context) error {
		return c.Render(http.StatusOK, "sysstatus", map[string]string{})
	})
	webserver.GET("/admin/overview", func(c echo.Context) error {
		return c.Render(http.StatusOK, "overview", map[string]string{})
	})

	webserver.GET("/admin/overview/cpe/:type/pie/data", func(c echo.Context) error {
		stype := c.Param("type")
		var cpes []models.NetCpe
		common.Must(app.GDB().Find(&cpes).Error)
		var statdata = map[string]*echarts.NameValuePair{}
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
		// return c.JSON(http.StatusOK, result)
		so := echarts.NewSeriesObject("pie")
		so.SetAttr("radius", "60%")
		so.SetAttr("itemStyle", echarts.Dict{"borderRadius": 7})
		so.SetAttr("data", result)
		return c.JSON(200, echarts.Series(so))
	})

	webserver.GET("/admin/overview/data", func(c echo.Context) error {
		type counterItem struct {
			Name  string      `json:"name"`
			Value interface{} `json:"value"`
			Icon  string      `json:"icon"`
		}

		var data []counterItem

		var accountCount int64
		app.GDB().Model(&models.RadiusUser{}).Count(&accountCount)
		data = append(data, counterItem{Icon: "mdi mdi-account", Name: "Account Total", Value: float64(accountCount)})

		var accountOnline int64
		app.GDB().Raw(`select count(1) from radius_user a, radius_online b where a.username = b.username`).Scan(&accountOnline)
		data = append(data, counterItem{Icon: "mdi mdi-account", Name: "Online Account", Value: float64(accountOnline)})

		var accountOffline int64
		app.GDB().
			Raw(`select count(1) - (select count(1) from radius_user a,  radius_online b where a.username = b.username) from radius_user`).
			Scan(&accountOffline)
		data = append(data, counterItem{Icon: "mdi mdi-account-outline", Name: "Offline Account", Value: float64(accountOffline)})

		var cpeCount int64
		app.GDB().Model(&models.NetCpe{}).Count(&cpeCount)
		data = append(data, counterItem{Icon: "mdi mdi-switch", Name: "CPE Total", Value: float64(cpeCount)})

		var deviceOnline int64
		app.GDB().Model(&models.NetCpe{}).
			Where("cwmp_status = 'online'").
			Count(&deviceOnline)
		data = append(data, counterItem{Icon: "mdi mdi-switch", Name: "Online CPE", Value: float64(deviceOnline)})

		var deviceOffline int64
		app.GDB().Model(&models.NetCpe{}).
			Where("cwmp_status = 'offline'").
			Count(&deviceOffline)
		data = append(data, counterItem{Icon: "mdi mdi-switch", Name: "Offline CPE", Value: float64(deviceOffline)})

		return c.JSON(http.StatusOK, data)
	})

	initSystemMetricsRouter()
	initRadiusMetricsRouter()
	initTr069MetricsRouter()
}
