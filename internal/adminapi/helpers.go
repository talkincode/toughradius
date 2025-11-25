package adminapi

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func parsePagination(c echo.Context) (int, int) {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.QueryParam("pageSize"))
	if err != nil || pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	return page, pageSize
}

func parseIDParam(c echo.Context, name string) (int64, error) {
	param := c.Param(name)
	if param == "" {
		param = c.Param("id")
	}
	if param == "" {
		return 0, errors.New("missing identifier")
	}
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func parseTimeInput(value string, fallback time.Time) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	layouts := []string{
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05", // ISO 8601 with seconds
		"2006-01-02T15:04",    // HTML5 datetime-local format (no seconds)
		"2006-01-02 15:04:05", // Common format with space
		"2006-01-02 15:04",    // Common format without seconds
		"2006-01-02",          // Date only
	}
	for _, layout := range layouts {
		if ts, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			// If only a date is provided, set it to the last second of that day
			if layout == "2006-01-02" {
				return ts.Add(23*time.Hour + 59*time.Minute + 59*time.Second), nil
			}
			return ts, nil
		}
	}
	return time.Time{}, errors.New("invalid time format")
}
