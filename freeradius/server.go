package freeradius

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	elog "github.com/labstack/gommon/log"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/zaplog/log"
	"go.uber.org/zap"
)

var server *FreeradiusServer

type FreeradiusServer struct {
	root *echo.Echo
}

func Listen() error {
	server = NewFreeRADIUSServer()
	server.initRouter()
	return server.Start()
}

func NewFreeRADIUSServer() *FreeradiusServer {
	s := new(FreeradiusServer)
	s.root = echo.New()
	s.root.Pre(middleware.RemoveTrailingSlash())
	s.root.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					c.Error(echo.NewHTTPError(http.StatusInternalServerError, err.Error()))
				}
			}()
			return next(c)
		}
	})
	s.root.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "freeradius ${time_rfc3339} ${remote_ip} ${method} ${uri} ${protocol} ${status} ${id} ${user_agent} ${latency} ${bytes_in} ${bytes_out} ${error}\n",
		Output: os.Stdout,
	}))
	s.root.HideBanner = true
	s.root.Logger.SetOutput(zap.NewStdLog(zap.L()).Writer())
	s.root.Logger.SetLevel(common.If(app.GConfig().Freeradius.Debug, elog.DEBUG, elog.INFO).(elog.Lvl))
	s.root.Debug = app.GConfig().Freeradius.Debug
	return s
}

// Start 启动服务器
func (s *FreeradiusServer) Start() error {
	log.Infof("Starting Freeradius API server %s:%d", app.GConfig().Freeradius.Host, app.GConfig().Freeradius.Port)
	err := s.root.Start(fmt.Sprintf("%s:%d", app.GConfig().Freeradius.Host, app.GConfig().Freeradius.Port))
	if err != nil {
		log.Errorf("Starting Freeradius API server error %s", err.Error())
	}
	return err
}
