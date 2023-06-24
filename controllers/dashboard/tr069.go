package dashboard

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common/echarts"
	"github.com/talkincode/toughradius/v8/common/zaplog"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/webserver"
)

func initTr069MetricsRouter() {

	webserver.GET("/admin/metrics/tr069", func(c echo.Context) error {
		return c.Render(http.StatusOK, "tr069_metrics", map[string]string{})
	})

	webserver.GET("/admin/metrics/tr069/data", func(c echo.Context) error {
		type counterItem struct {
			Name  string      `json:"name"`
			Value interface{} `json:"value"`
			Icon  string      `json:"icon"`
		}

		var data []counterItem

		result := app.GetAllTr069Metrics()
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Total message", Value: result[app.MetricsTr069MessageTotal]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Inform message", Value: result[app.MetricsTr069Inform]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Download", Value: result[app.MetricsTr069Download]})

		var cpeCount int64
		app.GDB().Model(&models.NetCpe{}).Count(&cpeCount)
		data = append(data, counterItem{Icon: "mdi mdi-switch", Name: "CPE Total", Value: float64(cpeCount)})

		var deviceOnline int64
		app.GDB().Model(&models.NetCpe{}).
			Where("cwmp_status = 'success'").
			Count(&deviceOnline)
		data = append(data, counterItem{Icon: "mdi mdi-switch", Name: "Online CPE", Value: float64(deviceOnline)})

		var deviceOffline int64
		app.GDB().Model(&models.NetCpe{}).
			Where("cwmp_status <> 'success'").
			Count(&deviceOffline)
		data = append(data, counterItem{Icon: "mdi mdi-switch", Name: "Offline CPE", Value: float64(deviceOffline)})

		return c.JSON(http.StatusOK, data)
	})

	webserver.GET("/admin/metrics/tr069/line", func(c echo.Context) error {

		var onlineItems []echarts.MetricLineItem
		onlinePoints, err := zaplog.TSDB().Select(app.MetricsTr069MessageTotal, nil,
			time.Now().Add(-24*time.Hour).Unix(), time.Now().Unix())

		onlineSo := echarts.NewSeriesObject("line")
		if err == nil {
			for i, p := range onlinePoints {
				onlineItems = append(onlineItems, echarts.MetricLineItem{
					Id:    i + 1,
					Time:  time.Unix(p.Timestamp, 0).Format("2006-01-02 15"),
					Value: p.Value,
				})
			}

			result := echarts.SumMetricLine(onlineItems)
			onlineTsdata := echarts.NewTimeValues()
			for _, item := range result {
				timestamp, err := time.Parse("2006-01-02 15", item.Time)
				if err != nil {
					continue
				}
				onlineTsdata.AddData(timestamp.Unix()*1000, item.Value)
			}
			onlineSo.SetAttr("name", "Total Message")
			onlineSo.SetAttr("showSymbol", false)
			onlineSo.SetAttr("smooth", true)
			onlineSo.SetAttr("data", onlineTsdata)
		}

		var offlineItems []echarts.MetricLineItem
		offlinePoints, err := zaplog.TSDB().Select(app.MetricsTr069Download, nil,
			time.Now().Add(-24*time.Hour).Unix(), time.Now().Unix())

		offlineSo := echarts.NewSeriesObject("line")
		if err == nil {
			for i, p := range offlinePoints {
				offlineItems = append(offlineItems, echarts.MetricLineItem{
					Id:    i + 1,
					Time:  time.Unix(p.Timestamp, 0).Format("2006-01-02 15"),
					Value: p.Value,
				})
			}

			result := echarts.SumMetricLine(offlineItems)
			offlineTsdata := echarts.NewTimeValues()
			for _, item := range result {
				timestamp, err := time.Parse("2006-01-02 15", item.Time)
				if err != nil {
					continue
				}
				offlineTsdata.AddData(timestamp.Unix()*1000, item.Value)
			}
			offlineSo.SetAttr("name", "Tr069 Download")
			offlineSo.SetAttr("showSymbol", false)
			offlineSo.SetAttr("smooth", true)
			offlineSo.SetAttr("data", offlineTsdata)
		}

		return c.JSON(200, echarts.Series(onlineSo, offlineSo))
	})
}
