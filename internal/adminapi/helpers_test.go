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
