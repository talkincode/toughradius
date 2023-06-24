package zaplog

import (
	"encoding/json"
	"errors"
	slog "log"
	"strconv"
	"time"

	"github.com/guonaihong/gout"
	"github.com/talkincode/toughradius/v8/common"
)

// loki
type lokiWriter struct {
	client *LokiClient
}

func (c lokiWriter) Write(p []byte) (int, error) {
	type logInfo struct {
		Level  string `json:"level"`  // 日志级别
		Ts     string `json:"ts"`     // 格式化后的时间(在zap那边配置的)
		Caller string `json:"caller"` // 日志输出的文件名和行号
		Msg    string `json:"msg"`    // 日志内容
	}
	var li logInfo
	err := json.Unmarshal(p, &li)
	if err != nil {
		return 0, err
	}

	c.client.logAppend(li.Level, string(p))

	return 0, nil
}

type LokiClient struct {
	Job         string
	ApiUrl      string
	ApiUser     string
	ApiPwd      string
	Debug       bool
	BuffSize    int
	logChl      chan string
	stopProcess chan struct{}
}

func NewLokiClient(job, apiUrl, user, pwd string, buff int) *LokiClient {
	c := &LokiClient{Job: job, ApiUrl: apiUrl, ApiUser: user, ApiPwd: pwd, BuffSize: buff}
	c.logChl = make(chan string, c.BuffSize)
	c.stopProcess = make(chan struct{})
	return c
}

type Labels map[string]string

type StreamObject struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type Streams struct {
	Streams []*StreamObject `json:"streams"`
}

func NewStreams(streams ...*StreamObject) *Streams {
	return &Streams{Streams: streams}
}

func NewStreamObject(labels Labels) *StreamObject {
	return &StreamObject{
		Stream: labels,
		Values: make([][]string, 0),
	}
}

func (s *StreamObject) AddLogItem(log ...string) {
	for _, v := range log {
		s.Values = append(s.Values, []string{strconv.FormatInt(time.Now().UnixNano(), 10), v})
	}
}

var emptyLog = ""
var readLogtimeout = errors.New("read channel time out")
var writeLogtimeout = errors.New("write channel time out")

func readWithTimeout(ch <-chan string) (log string, err error) {
	select {
	case log = <-ch:
		return log, nil
	case <-time.After(time.Millisecond * 100):
		return emptyLog, readLogtimeout
	}
}

func writeWithTimeout(ch chan<- string, log string) (err error) {
	select {
	case ch <- log:
		return nil
	case <-time.After(time.Millisecond * 50):
		return writeLogtimeout
	}
}

func (c *LokiClient) push(labels Labels, log ...string) error {
	var so = NewStreamObject(labels)
	so.AddLogItem(log...)
	var req = NewStreams(so)
	var resp string
	err := gout.
		POST(common.UrlJoin2(c.ApiUrl, "/loki/api/v1/push")).
		Debug(c.Debug).
		SetTimeout(time.Second*15).
		SetBasicAuth(c.ApiUser, c.ApiPwd).
		SetJSON(req).
		BindBody(&resp).
		Do()
	if err != nil {
		slog.Println("post loki logs error ", err.Error())
		return err
	}

	return nil
}

func (c *LokiClient) logAppend(level, msg string) {
	defer func() {
		if err := recover(); err != nil {
			err2, ok := err.(error)
			if ok {
				slog.Println(err2.Error())
			}
		}
	}()

	err := writeWithTimeout(c.logChl, msg)
	if err != nil {
		slog.Println(err.Error())
	}
}

func (c *LokiClient) Info(msg string) {
	c.logAppend("info", msg)
}

func (c *LokiClient) Warn(msg string) {
	c.logAppend("warn", msg)
}

func (c *LokiClient) Error(msg string) {
	c.logAppend("error", msg)
}

func (c *LokiClient) Start() {
	timeC := time.NewTicker(time.Millisecond * 1000)

	processLog := func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Println(err)
			}
		}()
		queue := make([]string, 0)
		for {
			clog, err := readWithTimeout(c.logChl)
			if err == readLogtimeout {
				break
			}
			if clog == "" {
				continue
			}
			queue = append(queue, clog)
		}

		if len(queue) > 0 {
			_ = c.push(Labels{"job": c.Job}, queue...)
		}
	}

	for {
		select {
		case <-c.stopProcess:
			return
		case <-timeC.C:
			go processLog()
		}
	}
}

func (c *LokiClient) Stop() {
	close(c.stopProcess)
	close(c.logChl)
}
