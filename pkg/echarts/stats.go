package echarts

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/spf13/cast"
)

type MetricLineItem struct {
	Id    int     `json:"id"`
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

func AvgMetricLine(src []MetricLineItem) (target []MetricLineItem) {
	df := dataframe.LoadStructs(src)
	groups := df.GroupBy("Time")
	aggre := groups.Aggregation([]dataframe.AggregationType{dataframe.Aggregation_MEAN}, []string{"Value"})
	sorted := aggre.Arrange(
		dataframe.Sort("Time"), // Sort in ascending order
	)
	var nitems []MetricLineItem
	for i, vals := range sorted.Records() {
		if i == 0 {
			continue
		}
		nitems = append(nitems, MetricLineItem{
			Id:    i,
			Time:  vals[0],
			Value: float64(int64(cast.ToFloat64(vals[1]))),
		})
	}
	return nitems
}

func SumMetricLine(src []MetricLineItem) (target []MetricLineItem) {
	df := dataframe.LoadStructs(src)
	groups := df.GroupBy("Time")
	aggre := groups.Aggregation([]dataframe.AggregationType{dataframe.Aggregation_SUM}, []string{"Value"})
	sorted := aggre.Arrange(
		dataframe.Sort("Time"), // Sort in ascending order
	)
	var nitems []MetricLineItem
	for i, vals := range sorted.Records() {
		if i == 0 {
			continue
		}
		nitems = append(nitems, MetricLineItem{
			Id:    i,
			Time:  vals[0],
			Value: float64(int64(cast.ToFloat64(vals[1]))),
		})
	}
	return nitems
}
