package app

import (
	"time"

	istats "github.com/montanaflynn/stats"
	"github.com/talkincode/toughradius/v8/common/zaplog"
)

const (
	MetricsTr069MessageTotal = "tr069_message_total"
	MetricsTr069Inform       = "tr069_inform"
	MetricsTr069Download     = "tr069_download"
)

var tr069MetricsNames = []string{
	MetricsTr069MessageTotal,
	MetricsTr069Inform,
	MetricsTr069Download,
}

func GetH24Metrics(name string) int64 {
	var value float64 = 0
	vals := make([]float64, 0)
	points, err := zaplog.TSDB().Select(name, nil,
		time.Now().Add(-86400*time.Second).Unix(), time.Now().Unix())
	if err != nil {
		return 0
	}
	for _, p := range points {
		vals = append(vals, p.Value)
	}
	value, _ = istats.Sum(vals)
	if value < 0 {
		value = 0
	}
	return int64(value)
}

func GetAllTr069Metrics() map[string]int64 {
	var result = make(map[string]int64)
	for _, name := range tr069MetricsNames {
		result[name] = GetH24Metrics(name)
	}
	return result
}
