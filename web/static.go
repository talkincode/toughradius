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

// GetFileSystem returns the root dist directory
func GetFileSystem() (http.FileSystem, error) {
	return getSubFS("dist")
}

// GetAdminFileSystem returns the dist/admin directory
func GetAdminFileSystem() (http.FileSystem, error) {
	return getSubFS("dist/admin")
}

// RegisterStaticRoutes registers static file routes
func RegisterStaticRoutes(e *echo.Echo) error {
	fsys, err := GetFileSystem()
	if err != nil {
		return err
	}

	// Static file handling
	staticHandler := http.FileServer(fsys)

	// Register static asset routes
	e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/", staticHandler)))

	// SPA routing - all non-API routes return index.html
	e.GET("/*", func(c echo.Context) error {
		path := c.Request().URL.Path

		// Skip API routes
		if len(path) >= 4 && path[:4] == "/api" {
			return echo.ErrNotFound
		}

		// Attempt to open the requested file
		file, err := fsys.Open(path)
		if err == nil {
			_ = file.Close() //nolint:errcheck
			return echo.WrapHandler(staticHandler)(c)
		}

		// If the file does not exist, return index.html (SPA routing)
		indexFile, err := fsys.Open("index.html")
		if err != nil {
			return err
		}
		defer func() { _ = indexFile.Close() }() //nolint:errcheck

		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.Stream(http.StatusOK, "text/html", indexFile)
	})

	return nil
}
