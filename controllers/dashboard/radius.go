package dashboard

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common/echarts"
	"github.com/talkincode/toughradius/common/zaplog"
	"github.com/talkincode/toughradius/webserver"
)

func initRadiusMetricsRouter() {

	webserver.GET("/admin/metrics/radius", func(c echo.Context) error {
		return c.Render(http.StatusOK, "radius_metrics", map[string]string{})
	})

	webserver.GET("/admin/metrics/radius/data", func(c echo.Context) error {
		type counterItem struct {
			Name  string      `json:"name"`
			Value interface{} `json:"value"`
			Icon  string      `json:"icon"`
		}

		var data []counterItem

		result := app.GetAllRadiusMetrics()
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth succeeded", Value: result[app.MetricsRadiusAccept]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (disabled)", Value: result[app.MetricsRadiusRejectDisable]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (expired)", Value: result[app.MetricsRadiusRejectExpire]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (Limit)", Value: result[app.MetricsRadiusRejectLimit]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (not exists)", Value: result[app.MetricsRadiusRejectNotExists]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (pwd error)", Value: result[app.MetricsRadiusRejectPasswdError]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (bind error)", Value: result[app.MetricsRadiusRejectBindError]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (unknow)", Value: result[app.MetricsRadiusRejectOther]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Auth failed (drop)", Value: result[app.MetricsRadiusAuthDrop]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h Acct succeeded", Value: result[app.MetricsRadiusAccounting]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h User online", Value: result[app.MetricsRadiusOline]})
		data = append(data, counterItem{Icon: "mdi mdi-circle-slice-2", Name: "24h User ofline", Value: result[app.MetricsRadiusOffline]})

		return c.JSON(http.StatusOK, data)
	})

	webserver.GET("/admin/metrics/radius/line", func(c echo.Context) error {

		var onlineItems []echarts.MetricLineItem
		onlinePoints, err := zaplog.TSDB().Select(app.MetricsRadiusOline, nil,
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
			onlineSo.SetAttr("name", "User Online")
			onlineSo.SetAttr("showSymbol", false)
			onlineSo.SetAttr("smooth", true)
			onlineSo.SetAttr("data", onlineTsdata)
		}

		var offlineItems []echarts.MetricLineItem
		offlinePoints, err := zaplog.TSDB().Select(app.MetricsRadiusOffline, nil,
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
			offlineSo.SetAttr("name", "User Offline")
			offlineSo.SetAttr("showSymbol", false)
			offlineSo.SetAttr("smooth", true)
			offlineSo.SetAttr("data", offlineTsdata)
		}

		return c.JSON(200, echarts.Series(onlineSo, offlineSo))
	})
}
