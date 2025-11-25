package web

type Metrics struct {
	Icon  string
	Value interface{}
	Title string
}

func NewMetrics(icon string, value interface{}, title string) *Metrics {
	return &Metrics{Icon: icon, Value: value, Title: title}
}
