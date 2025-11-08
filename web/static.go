package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed dist/*
var distFS embed.FS

func getSubFS(path string) (http.FileSystem, error) {
	fsys, err := fs.Sub(distFS, path)
	if err != nil {
		return nil, err
	}
	return http.FS(fsys), nil
}

// GetFileSystem 返回 dist 根目录
func GetFileSystem() (http.FileSystem, error) {
	return getSubFS("dist")
}

// GetAdminFileSystem 返回 dist/admin 子目录
func GetAdminFileSystem() (http.FileSystem, error) {
	return getSubFS("dist/admin")
}

// RegisterStaticRoutes 注册静态文件路由
func RegisterStaticRoutes(e *echo.Echo) error {
	fsys, err := GetFileSystem()
	if err != nil {
		return err
	}

	// 静态文件处理
	staticHandler := http.FileServer(fsys)

	// 注册静态资源路由
	e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/", staticHandler)))

	// SPA 路由处理 - 所有非 API 路由返回 index.html
	e.GET("/*", func(c echo.Context) error {
		path := c.Request().URL.Path

		// 跳过 API 路由
		if len(path) >= 4 && path[:4] == "/api" {
			return echo.ErrNotFound
		}

		// 尝试获取请求的文件
		file, err := fsys.Open(path)
		if err == nil {
			file.Close()
			return echo.WrapHandler(staticHandler)(c)
		}

		// 如果文件不存在，返回 index.html（用于 SPA 路由）
		indexFile, err := fsys.Open("index.html")
		if err != nil {
			return err
		}
		defer indexFile.Close()

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.Stream(http.StatusOK, "text/html", indexFile)
	})

	return nil
}
