package adminapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func sortContext(t *testing.T, sort, order string) echo.Context {
	t.Helper()
	q := url.Values{}
	if sort != "" {
		q.Set("sort", sort)
	}
	if order != "" {
		q.Set("order", order)
	}
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	return echo.New().NewContext(req, rec)
}

func paginationContext(t *testing.T, rawQuery string) echo.Context {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/?"+rawQuery, nil)
	rec := httptest.NewRecorder()
	return echo.New().NewContext(req, rec)
}

// TestParsePagination covers the React Admin / system-config contract (#583):
// accept both pageSize and perPage, and clamp oversized values instead of
// silently falling back to the default page size of 20.
func TestParsePagination(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		wantPage     int
		wantPageSize int
	}{
		{"defaults when empty", "", 1, 20},
		{"pageSize accepted", "page=2&pageSize=50", 2, 50},
		{"perPage accepted (react-admin)", "page=1&perPage=1000", 1, 1000},
		{"pageSize preferred over perPage", "page=1&pageSize=30&perPage=100", 1, 30},
		{"oversized clamped to max", "page=1&perPage=5000", 1, 1000},
		{"invalid page falls back", "page=0&perPage=10", 1, 10},
		{"invalid size falls back", "page=1&perPage=abc", 1, 20},
		{"negative size falls back", "page=1&pageSize=-5", 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := paginationContext(t, tt.query)
			page, pageSize := parsePagination(c)
			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantPageSize, pageSize)
		})
	}
}

func TestParseSort(t *testing.T) {
	allowed := map[string]bool{"id": true, "name": true, "created_at": true}

	tests := []struct {
		name         string
		sort         string
		order        string
		defaultField string
		defaultOrder string
		wantField    string
		wantOrder    string
	}{
		{"valid field and order", "name", "ASC", "id", "DESC", "name", "ASC"},
		{"unknown field falls back to default", "password; DROP TABLE", "ASC", "id", "DESC", "id", "ASC"},
		{"empty field falls back to default", "", "DESC", "id", "DESC", "id", "DESC"},
		{"invalid order falls back to default", "name", "bogus", "id", "DESC", "name", "DESC"},
		{"lowercase order is normalized", "name", "desc", "id", "ASC", "name", "DESC"},
		{"invalid default order coerced to ASC", "x", "y", "id", "weird", "id", "ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sortContext(t, tt.sort, tt.order)
			field, order := parseSort(c, allowed, tt.defaultField, tt.defaultOrder)
			assert.Equal(t, tt.wantField, field)
			assert.Equal(t, tt.wantOrder, order)
		})
	}
}
