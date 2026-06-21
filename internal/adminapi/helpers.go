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

// parseSort extracts and validates the `sort` and `order` query parameters
// against an allowlist of sortable columns, returning values that are safe to
// interpolate into an ORDER BY clause. Unknown or empty inputs fall back to
// defaultField/defaultOrder. Centralizing this guard ensures every list
// endpoint applies the same SQL-injection-safe column allowlist instead of
// re-implementing (and potentially forgetting) it.
//
// The returned field is re-derived from the matching allowlist key (a
// compile-time constant) and the returned order is one of two string literals,
// so neither value carries the raw query string into the SQL ORDER BY clause.
func parseSort(c echo.Context, allowed map[string]bool, defaultField, defaultOrder string) (field, order string) {
	field = defaultField
	if requested := c.QueryParam("sort"); requested != "" {
		for col := range allowed {
			if col == requested {
				field = col
				break
			}
		}
	}

	order = "ASC"
	switch strings.ToUpper(strings.TrimSpace(c.QueryParam("order"))) {
	case "DESC":
		order = "DESC"
	case "ASC":
		order = "ASC"
	default:
		if strings.ToUpper(strings.TrimSpace(defaultOrder)) == "DESC" {
			order = "DESC"
		}
	}
	return field, order
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
