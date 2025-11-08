package webserver

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gocarina/gocsv"
	_ "github.com/gocarina/gocsv"
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"github.com/talkincode/toughradius/v9/pkg/excel"
	"github.com/talkincode/toughradius/v9/pkg/web"
	webui "github.com/talkincode/toughradius/v9/web"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	elog "github.com/labstack/gommon/log"
	"github.com/pkg/errors"
)

const apiBasePath = "/api/v1"

var JwtSkipPrefix = []string{
	"/ready",
	"/realip",
	apiBasePath + "/auth/login",
	apiBasePath + "/auth/refresh",
}

var server *AdminServer

type AdminServer struct {
	root      *echo.Echo
	api       *echo.Group
	jwtConfig echojwt.Config
}

func Init() {
	server = NewAdminServer()
}

func Listen() error {
	return server.Start()
}

// NewAdminServer 创建管理系统服务器
func NewAdminServer() *AdminServer {
	appconfig := app.GConfig()
	s := &AdminServer{}
	s.root = echo.New()
	s.root.Pre(middleware.RemoveTrailingSlash())
	s.root.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/metrics")
		},
		Level: 1,
	}))
	// 失败恢复处理中间件
	s.root.Use(ServerRecover(appconfig.System.Debug))
	// 日志处理中间件
	s.root.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: appconfig.System.Appid + " ${time_rfc3339} ${remote_ip} ${method} ${uri} ${protocol} ${status} ${id} ${user_agent} ${latency} ${bytes_in} ${bytes_out} ${error}\n",
		Output: os.Stdout,
	}))
	// p := prometheus.NewPrometheus("toughradius", nil)
	// p.Use(s.root)

	// React Admin Web UI 静态文件服务
	s.setupReactAdminStatic()

	s.root.HideBanner = true
	// 设置日志级别
	s.root.Logger.SetLevel(common.If(appconfig.System.Debug, elog.DEBUG, elog.INFO).(elog.Lvl))
	s.root.Debug = appconfig.System.Debug

	// 根路径重定向到 /admin
	s.root.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/admin")
	})

	s.root.GET("/ready", func(c echo.Context) error {
		return c.JSON(200, web.RestSucc("OK"))
	})

	s.root.GET("/realip", func(c echo.Context) error {
		return c.String(200, c.RealIP())
	})

	s.root.GET("/.well-known/appspecific/com.chrome.devtools.json", chromeDevtoolsManifest)

	// Chrome DevTools 配置文件请求处理
	s.root.GET("/.well-known/appspecific/com.chrome.devtools.json", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"applications": []map[string]interface{}{
				{
					"name":    "ToughRADIUS",
					"version": "9.0",
					"url":     "/admin",
				},
			},
		})
	})


	// JWT 中间件
	s.jwtConfig = echojwt.Config{
		SigningKey:    []byte(appconfig.Web.Secret),
		SigningMethod: echojwt.AlgorithmHS256,
		Skipper:       jwtSkipFunc(),
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusBadRequest, web.RestError("Resource access is limited "+err.Error()))
		},
	}

	// init api -------------------------------
	s.api = s.root.Group(apiBasePath)
	s.api.Use(echojwt.WithConfig(s.jwtConfig))

	return s
}

// setupReactAdminStatic 设置 React Admin 静态文件服务
func (s *AdminServer) setupReactAdminStatic() {
	// 尝试从嵌入的文件系统加载
	webStaticFS, err := getWebStaticFS()
	if err != nil {
		zap.S().Warnf("Failed to load embedded web static files: %v, using development mode", err)
		// 开发模式：从本地文件系统读取
		s.root.Static("/admin", "web/dist/admin")
		return
	}

	// 生产模式：从嵌入的文件系统读取
	// 处理 /admin 路径下的静态资源
	staticHandler := http.FileServer(webStaticFS)

	renderIndex := func(c echo.Context) error {
		indexFile, err := webStaticFS.Open("/index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Web UI not found")
		}
		defer indexFile.Close()

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.Stream(http.StatusOK, "text/html", indexFile)
	}

	// 注册 /admin/assets/* 路由用于静态资源
	s.root.GET("/admin/assets/*", echo.WrapHandler(http.StripPrefix("/admin", staticHandler)))

	// 处理精确的 /admin 路径（不带尾随斜杠）
	s.root.GET("/admin", renderIndex)

	// 注册 /admin/* 路由用于 SPA，所有路由返回 index.html
	s.root.GET("/admin/*", func(c echo.Context) error {
		// 尝试获取请求的文件
		path := strings.TrimPrefix(c.Request().URL.Path, "/admin")
		if path == "" || path == "/" {
			path = "/index.html"
		}

		file, err := webStaticFS.Open(path)
		if err == nil {
			file.Close()
			return echo.WrapHandler(http.StripPrefix("/admin", staticHandler))(c)
		}

		// 如果文件不存在，返回 index.html（用于 SPA 路由）
		indexFile, err := webStaticFS.Open("/index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Web UI not found")
		}
		defer indexFile.Close()

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.Stream(http.StatusOK, "text/html", indexFile)
	})

	zap.S().Info("React Admin static files loaded successfully")
}

// getWebStaticFS 获取嵌入的 Web 静态文件系统
func getWebStaticFS() (http.FileSystem, error) {
	return webui.GetAdminFileSystem()
}

func chromeDevtoolsManifest(c echo.Context) error {
	scheme := "http"
	if c.IsTLS() {
		scheme = "https"
	}
	host := c.Request().Host
	if host == "" {
		host = "localhost"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)
	payload := map[string]interface{}{
		"description":          "ToughRADIUS exposes no remote debugging targets; this manifest satisfies Chrome DevTools well-known checks.",
		"devtoolsFrontendUrl":  fmt.Sprintf("%s/admin", baseURL),
		"documentation":        "https://developer.chrome.com/docs/devtools/",
		"status":               "not_available",
		"websocketDebuggerUrl": "",
		"supportedEndpoints":   []string{"/api/v1"},
		"generatedAt":          time.Now().UTC(),
	}
	return c.JSON(http.StatusOK, payload)
}

// Start Admin Server
func (s *AdminServer) Start() error {
	appconfig := app.GConfig()
	go func() {
		zap.S().Infof("Prepare to start the TLS management port %s:%d", appconfig.Web.Host, appconfig.Web.TlsPort)
		err := s.root.StartTLS(fmt.Sprintf("%s:%d", appconfig.Web.Host, appconfig.Web.TlsPort),
			path.Join(appconfig.GetPrivateDir(), "toughradius.tls.crt"), path.Join(appconfig.GetPrivateDir(), "toughradius.tls.key"))
		if err != nil {
			zap.S().Errorf("Error starting TLS management port %s", err.Error())
		}
	}()
	zap.S().Infof("Start the management server %s:%d", appconfig.Web.Host, appconfig.Web.Port)
	err := s.root.Start(fmt.Sprintf("%s:%d", appconfig.Web.Host, appconfig.Web.Port))
	if err != nil {
		zap.S().Errorf("Error starting management server %s", err.Error())
	}
	return err
}

// ParseJwtToken 解析 Jwt Token
func (s *AdminServer) ParseJwtToken(tokenstr string) (jwt.MapClaims, error) {
	config := s.jwtConfig
	token, err := jwt.Parse(tokenstr, func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		if len(config.SigningKeys) > 0 {
			if kid, ok := t.Header["kid"].(string); ok {
				if key, ok := config.SigningKeys[kid]; ok {
					return key, nil
				}
			}
			return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
		}
		return config.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims := token.Claims.(jwt.MapClaims)
	return claims, err
}

// ServerRecover Web 服务恢复处理中间件
func ServerRecover(debug bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					if debug {
						zap.S().Errorf("%+v", errors.WithStack(err))
					}
					c.Error(echo.NewHTTPError(http.StatusInternalServerError, err.Error()))
				}
			}()
			return next(c)
		}
	}
}

// skipFUnc Web 请求过滤中间件
func jwtSkipFunc() func(c echo.Context) bool {
	return func(c echo.Context) bool {
		if os.Getenv("TOUGHRADIUS_DEVMODE") == "true" {
			return true
		}

		for _, prefix := range JwtSkipPrefix {
			if strings.HasPrefix(c.Path(), prefix) {
				return true
			}
		}
		return false
	}
}

// ImportData Import the file contents
func ImportData(c echo.Context, sheet string) ([]map[string]interface{}, error) {
	file, err := c.FormFile("upload")
	if err != nil {
		return nil, err
	}
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	var items []map[string]interface{}
	if strings.HasSuffix(file.Filename, "json") {
		items, err = web.ReadImportJsonData(src)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(file.Filename, "csv") {
		items, err = web.ReadImportCsvData(src)
		if err != nil {
			return nil, err
		}
	} else {
		items, err = web.ReadImportExcelData(src, sheet)
		if err != nil {
			return nil, err
		}
	}
	var datas = make([]map[string]interface{}, 0)
	for _, item := range items {
		_id, ok := item["id"]
		if !ok || common.IsEmptyOrNA(cast.ToString(_id)) {
			_id, ok = item["ID"]
		}
		if !ok || common.IsEmptyOrNA(cast.ToString(_id)) {
			_id, ok = item["Id"]
		}
		if !ok || common.IsEmptyOrNA(cast.ToString(_id)) {
			item["id"] = strconv.FormatInt(common.UUIDint64(), 10)
		}
		datas = append(datas, item)
	}
	return datas, nil
}

func ExportData(c echo.Context, data []map[string]interface{}, sheet string) error {
	filename := fmt.Sprintf("%s-%d.xlsx", sheet, common.UUIDint64())
	filepath := path.Join(app.GConfig().GetDataDir(), filename)
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheet)
	names := make([]string, 0)
	for i, item := range data {
		if i == 0 {
			for k, _ := range item {
				names = append(names, k)
			}
			sort.Slice(names, func(i, j int) bool {
				return names[i] == "id"
			})
			for j, name := range names {
				xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", excel.COLNAMES[j], 1), name)
			}
		}
		for j, name := range names {
			_value := cast.ToString(item[name])
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", excel.COLNAMES[j], i+2), _value)
		}
	}
	xlsx.SetActiveSheet(index)
	err := xlsx.SaveAs(filepath)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.xlsx", sheet))
	return c.File(filepath)
}

func ExportCsv(c echo.Context, v interface{}, name string) error {
	filename := fmt.Sprintf("%s-%d.csv", name, common.UUIDint64())
	filepath := path.Join(app.GConfig().GetDataDir(), filename)
	nfs, err := os.Create(filepath)
	defer nfs.Close()
	if err != nil {
		return err
	}
	err = gocsv.Marshal(v, nfs)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.csv", name))
	return c.File(filepath)
}

func ExportJson(c echo.Context, v interface{}, name string) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.json", name))
	return c.JSONBlob(http.StatusOK, bs)
}

func GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add GET Router %s", path)
	return server.root.GET(path, h, m...)
}

func POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add POST Router %s", path)
	return server.root.POST(path, h, m...)
}

func PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add PUT Router %s", path)
	return server.root.PUT(path, h, m...)
}

func DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add DELETE Router %s", path)
	return server.root.DELETE(path, h, m...)
}

func ApiGET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add API GET Router %s%s", apiBasePath, path)
	return server.api.GET(path, h, m...)
}

func ApiDELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add API DELETE Router %s%s", apiBasePath, path)
	return server.api.DELETE(path, h, m...)
}

func ApiPOST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add API POST Router %s%s", apiBasePath, path)
	return server.api.POST(path, h, m...)
}

func ApiPUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	zap.S().Debugf("Add API PUT Router %s%s", apiBasePath, path)
	return server.api.PUT(path, h, m...)
}

func ApiANY(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) []*echo.Route {
	zap.S().Debugf("Add API ANY Router %s%s", apiBasePath, path)
	return server.api.Any(path, h, m...)
}
