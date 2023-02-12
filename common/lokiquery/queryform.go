package lokiquery

import (
	"fmt"
)

type LokiQueryForm struct {
	Limit              int
	Start              int64
	End                int64
	Step               string
	Interval           string
	LokiApi            string
	LokiUser           string
	LokiPwd            string
	labels             map[string]string
	lineContains       []string
	lineNotContains    []string
	lineContainsReg    []string
	lineNotContainsReg []string
	Debug              bool
}

type Lokilog struct {
	Job       string `json:"job"`
	Level     string `json:"level"`
	Caller    string `json:"caller"`
	Msg       string `json:"msg"`
	Timestamp string `json:"timestamp"`
	Namespace string `json:"namespace,omitempty"`
	Metrics   string `json:"metrics,omitempty"`
	Username  string `json:"username,omitempty"`
	Result    string `json:"result,omitempty"`
	Nasip     string `json:"nasip,omitempty"`
}

func NewLokiQueryForm(lapi, luser, lpwd string) *LokiQueryForm {
	return &LokiQueryForm{LokiApi: lapi, LokiUser: luser, LokiPwd: lpwd}
}

func (q *LokiQueryForm) AddLabel(key, value string) *LokiQueryForm {
	if value == "" {
		return q
	}
	if q.labels == nil {
		q.labels = make(map[string]string)
	}
	q.labels[key] = value
	return q
}

func (q *LokiQueryForm) AddLineContains(arg string) *LokiQueryForm {
	if arg == "" {
		return q
	}
	q.lineContains = append(q.lineContains, arg)
	return q
}

func (q *LokiQueryForm) AddLineNotContains(arg string) *LokiQueryForm {
	if arg == "" {
		return q
	}
	q.lineNotContains = append(q.lineNotContains, arg)
	return q
}

func (q *LokiQueryForm) AddLineContainsReg(arg string) *LokiQueryForm {
	if arg == "" {
		return q
	}
	q.lineContainsReg = append(q.lineContainsReg, arg)
	return q
}

func (q *LokiQueryForm) AddLineNotContainsReg(arg string) *LokiQueryForm {
	if arg == "" {
		return q
	}
	q.lineNotContainsReg = append(q.lineNotContainsReg, arg)
	return q
}

func (q *LokiQueryForm) QueryString() string {
	var query string

	for k, v := range q.labels {
		query += k + "=\"" + v + "\","
	}

	if len(query) > 0 {
		query = query[:len(query)-1]
		query = fmt.Sprintf("{%s}", query)
	}

	for _, v := range q.lineContains {
		query += " |= `" + v + "`"
	}
	for _, v := range q.lineNotContains {
		query += " != `" + v + "`"
	}
	for _, v := range q.lineContainsReg {
		query += " |~ `" + v + "`"
	}
	for _, v := range q.lineNotContainsReg {
		query += " !~ `" + v + "`"
	}
	return query
}
