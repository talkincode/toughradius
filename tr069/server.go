package tr069

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	elog "github.com/labstack/gommon/log"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/assets"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/zaplog/log"
	"go.uber.org/zap"
)

var server *Tr069Server

const Tr069Session = "tr069_session"
const Tr069CookieName = "tr069_cookie"

type Tr069Server struct {
	root     *echo.Echo
	sesslock sync.Mutex
}

func Listen() error {
	server = NewTr069Server()
	server.initRouter()
	return server.Start()
}

func NewTr069Server() *Tr069Server {
	s := new(Tr069Server)
	s.root = echo.New()
	s.sesslock = sync.Mutex{}
	s.root.Pre(middleware.RemoveTrailingSlash())
	s.root.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					log.Error("Tr069 API server error %s", err.Error())
					c.Error(echo.NewHTTPError(http.StatusInternalServerError, err.Error()))
				}
			}()
			return next(c)
		}
	})
	s.root.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "cwmp-acs ${time_rfc3339} ${remote_ip} ${method} ${uri} ${protocol} ${status} ${id} ${user_agent} ${latency} ${bytes_in} ${bytes_out} ${error}\n",
		Output: os.Stdout,
	}))
	// s.root.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
	// 	log.Info(string(resBody))
	// 	log.Info(string(resBody))
	// }))
	s.root.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper: func(c echo.Context) bool {
			rpath := c.Request().RequestURI
			if strings.HasPrefix(rpath, "/cwmpfiles") ||
				strings.HasPrefix(rpath, "/cwmpupload") {
				return true
			} else {
				return false
			}
		},
		Validator: func(username, password string, c echo.Context) (bool, error) {
			if username == "" {
				return false, nil
			}
			return true, nil
		},
		Realm: "Restricted",
	}))
	s.root.Use(session.Middleware(sessions.NewCookieStore([]byte(app.GConfig().Web.Secret))))
	s.root.HideBanner = true
	s.root.Logger.SetOutput(zap.NewStdLog(zap.L()).Writer())
	s.root.Logger.SetLevel(common.If(app.GConfig().Tr069.Debug, elog.DEBUG, elog.INFO).(elog.Lvl))
	s.root.Debug = app.GConfig().Tr069.Debug
	return s
}

func checkCert() {

}

func (s *Tr069Server) startTlsServer() error {
	caCert := path.Join(app.GConfig().System.Workdir, "private/ca.crt")
	serverCert := path.Join(app.GConfig().System.Workdir, "private/cwmp.tls.crt")
	serverKey := path.Join(app.GConfig().System.Workdir, "private/cwmp.tls.key")
	if !common.FileExists(caCert) {
		os.WriteFile(caCert, assets.CaCrt, 0644)
	}
	if !common.FileExists(serverCert) {
		os.WriteFile(serverCert, assets.CwmpCert, 0644)
	}
	if !common.FileExists(serverKey) {
		os.WriteFile(serverKey, assets.CwmpKey, 0644)
	}

	address := fmt.Sprintf("%s:%d", app.GConfig().Tr069.Host, app.GConfig().Tr069.Port)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(assets.CaCrt)
	ss := &http.Server{
		Addr:    address,
		Handler: s.root,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.VerifyClientCertIfGiven,
		},
	}
	return ss.ListenAndServeTLS(serverCert, serverKey)
}

// Start 启动服务器
func (s *Tr069Server) Start() (err error) {
	log.Infof("Start Tr069 API server %s:%d", app.GConfig().Tr069.Host, app.GConfig().Tr069.Port)
	if app.GConfig().Tr069.Tls {
		err = s.startTlsServer()
	} else {
		err = s.root.Start(fmt.Sprintf("%s:%d", app.GConfig().Tr069.Host, app.GConfig().Tr069.Port))
	}
	if err != nil {
		log.Errorf("Error starting Tr069 API server %s", err.Error())
	}
	return err
}

func (s *Tr069Server) GetLatestCookieSn(c echo.Context) string {
	cookie, err := c.Cookie(Tr069CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *Tr069Server) SetLatestInformByCookie(c echo.Context, sn string) {
	cookie := new(http.Cookie)
	cookie.Name = Tr069CookieName
	cookie.Value = sn
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.SetCookie(cookie)
}
