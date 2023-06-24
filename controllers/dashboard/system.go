package dashboard

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/echarts"
	"github.com/talkincode/toughradius/v8/common/zaplog"
	"github.com/talkincode/toughradius/v8/webserver"
)

func initSystemMetricsRouter() {

	webserver.GET("/admin/metrics/cpuuse/line", func(c echo.Context) error {
		var items []echarts.MetricLineItem

		points, err := zaplog.TSDB().Select("toughradius_cpuuse", nil,
			time.Now().Add(-24*time.Hour).Unix(), time.Now().Unix())
		if err != nil {
			return c.JSON(200, common.EmptyList)
		}
		for i, p := range points {
			items = append(items, echarts.MetricLineItem{
				Id:    i + 1,
				Time:  time.Unix(p.Timestamp, 0).Format("2006-01-02 15:04"),
				Value: p.Value,
			})
		}

		result := echarts.AvgMetricLine(items)
		tsdata := echarts.NewTimeValues()
		for _, item := range result {
			timestamp, err := time.Parse("2006-01-02 15:04", item.Time)
			if err != nil {
				continue
			}
			tsdata.AddData(timestamp.Unix()*1000, item.Value)
		}
		so := echarts.NewSeriesObject("line")
		so.SetAttr("showSymbol", false)
		so.SetAttr("smooth", true)
		so.SetAttr("areaStyle", echarts.Dict{})
		so.SetAttr("data", tsdata)

		return c.JSON(200, echarts.Series(so))
	})

	webserver.GET("/admin/metrics/memuse/line", func(c echo.Context) error {

		var items []echarts.MetricLineItem

		points, err := zaplog.TSDB().Select("toughradius_memuse", nil,
			time.Now().Add(-24*time.Hour).Unix(), time.Now().Unix())
		if err != nil {
			return c.JSON(200, common.EmptyList)
		}
		for i, p := range points {
			items = append(items, echarts.MetricLineItem{
				Id:    i + 1,
				Time:  time.Unix(p.Timestamp, 0).Format("2006-01-02 15:04"),
				Value: p.Value,
			})
		}

		result := echarts.AvgMetricLine(items)
		tsdata := echarts.NewTimeValues()
		for _, item := range result {
			timestamp, err := time.Parse("2006-01-02 15:04", item.Time)
			if err != nil {
				continue
			}
			tsdata.AddData(timestamp.Unix()*1000, item.Value)
		}
		so := echarts.NewSeriesObject("line")
		so.SetAttr("showSymbol", false)
		so.SetAttr("smooth", true)
		so.SetAttr("areaStyle", echarts.Dict{})
		so.SetAttr("data", tsdata)
		return c.JSON(200, echarts.Series(so))
	})
}
