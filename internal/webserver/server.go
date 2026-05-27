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
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"github.com/talkincode/toughradius/v9/pkg/excel"
	customValidator "github.com/talkincode/toughradius/v9/pkg/validator"
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
	appCtx    app.AppContext // Application context
}

func Init(appCtx app.AppContext) {
	server = NewAdminServer(appCtx)
}

func Listen(appCtx app.AppContext) error {
	return server.Start()
}

// NewAdminServer creates the admin system server
func NewAdminServer(appCtx app.AppContext) *AdminServer {
	appconfig := appCtx.Config()
	s := &AdminServer{appCtx: appCtx}
	s.root = echo.New()
	s.root.Pre(middleware.RemoveTrailingSlash())
	s.root.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/metrics")
		},
		Level: 1,
	}))

	// Register the custom validator
	s.root.Validator = customValidator.NewValidator()

	// Failure recovery middleware
	s.root.Use(ServerRecover(appconfig.System.Debug))
	// Logging middleware
	s.root.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: appconfig.System.Appid + " ${time_rfc3339} ${remote_ip} ${method} ${uri} ${protocol} ${status} ${id} ${user_agent} ${latency} ${bytes_in} ${bytes_out} ${error}\n",
		Output: os.Stdout,
	}))
	// p := prometheus.NewPrometheus("toughradius", nil)
	// p.Use(s.root)

	// Serve React Admin web UI static files
	s.setupReactAdminStatic()

	s.root.HideBanner = true
	// Set the log level
	s.root.Logger.SetLevel(common.If(appconfig.System.Debug, elog.DEBUG, elog.INFO).(elog.Lvl)) //nolint:errcheck // type assertion is safe
	s.root.Debug = appconfig.System.Debug

	// Redirect the root path to /admin
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

	// Chrome DevTools config filerequestHandle
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

	// JWT middleware
	s.jwtConfig = echojwt.Config{
		SigningKey:    []byte(appconfig.Web.Secret),
		SigningMethod: echojwt.AlgorithmHS256,
		Skipper:       jwtSkipFunc(),
		ErrorHandler: func(c echo.Context, err error) error {
			zap.S().Warnf("JWT validation failed: %v, Path: %s, Auth Header: %s",
				err, c.Path(), c.Request().Header.Get("Authorization"))
			return c.JSON(http.StatusUnauthorized, web.RestError("Authentication failed: "+err.Error()))
		},
	}

	// init api -------------------------------
	s.api = s.root.Group(apiBasePath)
	s.api.Use(echojwt.WithConfig(s.jwtConfig))

	// Add middleware to inject appCtx into each request context
	s.root.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("appCtx", s.appCtx)
			return next(c)
		}
	})

	return s
}

// setupReactAdminStatic sets up React Admin static file serving
func (s *AdminServer) setupReactAdminStatic() {
	// Try loading from the embedded filesystem
	webStaticFS, err := getWebStaticFS()
	if err != nil {
		zap.S().Warnf("Failed to load embedded web static files: %v, using development mode", err)
		// Development mode: read from the local filesystem
		s.root.Static("/admin", "web/dist/admin")
		return
	}

	// Production mode: read from the embedded filesystem
	// Serve static assets under the /admin path
	staticHandler := http.FileServer(webStaticFS)

	renderIndex := func(c echo.Context) error {
		indexFile, err := webStaticFS.Open("/index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Web UI not found")
		}
		defer func() { _ = indexFile.Close() }() //nolint:errcheck

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.Stream(http.StatusOK, "text/html", indexFile)
	}

	// Register /admin/assets/* for static assets
	s.root.GET("/admin/assets/*", echo.WrapHandler(http.StripPrefix("/admin", staticHandler)))

	// Handle the exact /admin path (without trailing slash)
	s.root.GET("/admin", renderIndex)

	// Register /admin/* for the SPA; all routes return index.html
	s.root.GET("/admin/*", func(c echo.Context) error {
		// Try to fetch the requested file
		path := strings.TrimPrefix(c.Request().URL.Path, "/admin")
		if path == "" || path == "/" {
			path = "/index.html"
		}

		file, err := webStaticFS.Open(path)
		if err == nil {
			_ = file.Close() //nolint:errcheck
			return echo.WrapHandler(http.StripPrefix("/admin", staticHandler))(c)
		}

		// e.g., if the file is missing, return index.html (for SPA routing)
		indexFile, err := webStaticFS.Open("/index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Web UI not found")
		}
		defer func() { _ = indexFile.Close() }() //nolint:errcheck

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.Stream(http.StatusOK, "text/html", indexFile)
	})

	zap.S().Info("React Admin static files loaded successfully")
}

// getWebStaticFS returns the embedded web static filesystem
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
	appconfig := s.appCtx.Config()
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

// ParseJwtToken Parse Jwt Token
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
	claims := token.Claims.(jwt.MapClaims) //nolint:errcheck // type assertion is safe for JWT claims
	return claims, err
}

// ServerRecover is the web recovery middleware
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

// skipFunc filters web requests in middleware
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
	defer func() { _ = src.Close() }() //nolint:errcheck
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
	appCtx := c.Get("appCtx").(app.AppContext) //nolint:errcheck // type assertion is safe for middleware-set context
	filename := fmt.Sprintf("%s-%d.xlsx", sheet, common.UUIDint64())
	filepath := path.Join(appCtx.Config().GetDataDir(), filename)
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheet)
	names := make([]string, 0)
	if len(data) > 0 {
		for k := range data[0] {
			names = append(names, k)
		}
		sort.Strings(names)
		for idx, name := range names {
			if strings.EqualFold(name, "id") {
				copy(names[1:idx+1], names[:idx])
				names[0] = name
				break
			}
		}
		for j, name := range names {
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", excel.COLNAMES[j], 1), name)
		}
	}
	for i, item := range data {
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
	appCtx := c.Get("appCtx").(app.AppContext) //nolint:errcheck // type assertion is safe for middleware-set context
	filename := fmt.Sprintf("%s-%d.csv", name, common.UUIDint64())
	filepath := path.Join(appCtx.Config().GetDataDir(), filename)
	nfs, err := os.Create(filepath) //nolint:gosec // G304: path is constructed from app data directory
	if err != nil {
		return err
	}
	defer func() { _ = nfs.Close() }() //nolint:errcheck
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
