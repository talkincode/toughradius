package echarts

import (
	"encoding/json"
	"sync/atomic"
)

type Dict map[string]interface{}

type SeriesObject struct {
	Type  string
	Data  interface{}
	attrs map[string]interface{}
}

func NewSeriesObject(Type string) *SeriesObject {
	so := &SeriesObject{Type: Type}
	so.attrs = make(map[string]interface{})
	return so
}

func (d *SeriesObject) SetAttr(key string, value interface{}) {
	d.attrs[key] = value
}

func (d *SeriesObject) MarshalJSON() ([]byte, error) {
	jo := make(map[string]interface{})
	jo["type"] = d.Type
	jo["data"] = d.Data
	for k, v := range d.attrs {
		jo[k] = v
	}
	return json.Marshal(jo)
}

func Series(s ...*SeriesObject) map[string][]*SeriesObject {
	return map[string][]*SeriesObject{
		"series": s,
	}
}

type NameValuePair struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func NewNameValuePair(name string, value int64) *NameValuePair {
	return &NameValuePair{Name: name, Value: value}
}

func (d *NameValuePair) Incr() {
	atomic.AddInt64(&d.Value, 1)
}

type TimeValues struct {
	value [][]interface{}
}

func NewTimeValues() *TimeValues {
	return &TimeValues{}
}

func (tv *TimeValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(tv.value)
}

func (tv *TimeValues) AddData(time int64, value interface{}) {
	tv.value = append(tv.value, []interface{}{time, value})
}
