package web

import (
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
	"os/exec"
)

var (
	sseEent  = []byte("event: ")
	sseBegin = []byte("data: ")
	sseEnd   = []byte("\n\n")
)

type SSE struct {
	EchoContext echo.Context
	context.Context
}

type SSEMessage struct {
	Id     string      `json:"id"`
	Action string      `json:"action"`
	Error  string      `json:"error,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

func (sse *SSE) Write(data []byte) (n int, err error) {
	if err = sse.Err(); err != nil {
		return
	}
	_, err = sse.EchoContext.Response().Writer.Write(sseBegin)
	n, err = sse.EchoContext.Response().Writer.Write(data)
	_, err = sse.EchoContext.Response().Writer.Write(sseEnd)
	if err != nil {
		return
	}
	sse.EchoContext.Response().Writer.(http.Flusher).Flush()
	return
}

func (sse *SSE) WriteEvent(event string, data []byte) (err error) {
	if err = sse.Err(); err != nil {
		return
	}
	_, err = sse.EchoContext.Response().Writer.Write(sseEent)
	_, err = sse.EchoContext.Response().Writer.Write([]byte(event))
	_, err = sse.EchoContext.Response().Writer.Write([]byte("\n"))
	_, err = sse.Write(data)
	return
}

func NewSSE(ectx echo.Context) *SSE {
	header := ectx.Response().Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	header.Set("Access-Control-Allow-Origin", "*")
	return &SSE{
		ectx,
		ectx.Request().Context(),
	}
}

func (sse *SSE) WriteJSON(data interface{}) (err error) {
	var jsonData []byte
	if jsonData, err = json.Marshal(data); err == nil {
		if _, err = sse.Write(jsonData); err != nil {
			return
		}
		return
	}
	return
}

func (sse *SSE) WriteMessage(msg SSEMessage) (err error) {
	var jsonData []byte
	if jsonData, err = json.Marshal(&msg); err == nil {
		if _, err = sse.Write(jsonData); err != nil {
			return
		}
		return
	}
	return
}

func (sse *SSE) WriteText(msg string) (err error) {
	if _, err = sse.Write([]byte(msg)); err != nil {
		return
	}
	return
}

func (sse *SSE) WriteExec(cmd *exec.Cmd) error {
	cmd.Stderr = sse
	cmd.Stdout = sse
	return cmd.Run()
}
