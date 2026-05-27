package web

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/gorm"
)

// PreQuery is a fluent query builder for constructing GORM queries from HTTP request parameters.
// It provides a chainable API for building complex database queries with:
//   - Date range filtering
//   - Field equality matching
//   - Keyword search across multiple fields
//   - Sorting and pagination
//   - Custom parameter mapping
//
// This is commonly used in REST API list endpoints to translate query parameters
// into SQL WHERE clauses and ORDER BY statements.
//
// Example:
//
//	prequery := NewPreQuery(c).
//	    DefaultOrderBy("created_at DESC").
//	    DateRange("dateRange", "created_at", startTime, endTime).
//	    KeyFields("username", "email").
//	    EqualFields("status")
//
//	var users []User
//	query := prequery.Query(db.Model(&User{}))
//	query.Find(&users)
type PreQuery struct {
	context          echo.Context
	defaultOrderby   string
	dateRange        DateRange
	timeField        string
	equalFieldds     []string
	keyfilterFieldds []string
	params           map[string]interface{}
	form             *WebForm
}

// NewPreQuery creates a new PreQuery instance from an Echo context.
// This initializes the query builder with request parameters from the HTTP context.
//
// Parameters:
//   - c: Echo context containing query/form parameters
//
// Returns:
//   - *PreQuery: Chainable query builder instance
//
// Example:
//
//	func ListUsers(c echo.Context) error {
//	    prequery := web.NewPreQuery(c)
//	    // Chain query methods...
//	}
func NewPreQuery(c echo.Context) *PreQuery {
	return &PreQuery{
		context: c,
		form:    NewWebForm(c),
		params:  make(map[string]interface{}),
	}
}

// DefaultOrderBy sets the default sort order when no sort parameter is provided in the request.
// The sort field should be a valid database column name with optional direction.
//
// Parameters:
//   - fd: Default ORDER BY clause (e.g., "id DESC", "created_at ASC")
//
// Returns:
//   - *PreQuery: Self for method chaining
//
// Example:
//
//	prequery.DefaultOrderBy("created_at DESC")
func (p *PreQuery) DefaultOrderBy(fd string) *PreQuery {
	p.defaultOrderby = fd
	return p
}

// DateRange adds a date range filter to the query from a JSON-encoded query parameter.
// The query parameter should contain a DateRange JSON object with "start" and "end" fields.
//
// Parameters:
//   - queryfd: Name of query parameter containing JSON date range (e.g., "dateRange")
//   - timefd: Database column name for date/time filtering (e.g., "created_at")
//   - defaltStart: Default start time if not provided in request
//   - defaultEnd: Default end time if not provided in request
//
// Returns:
//   - *PreQuery: Self for method chaining
//
// Example:
//
//	// Request: ?dateRange={"start":"2024-01-01 00:00:00","end":"2024-12-31 23:59:59"}
//	prequery.DateRange("dateRange", "created_at", time.Now().AddDate(0, -1, 0), time.Now())
func (p *PreQuery) DateRange(queryfd, timefd string, defaltStart time.Time, defaultEnd time.Time) *PreQuery {
	daterange, err := p.form.GetDateRange(queryfd)
	if err == nil {
		p.dateRange = daterange
		if p.dateRange.Start == "" {
			p.dateRange.Start = defaltStart.Format("2006-01-02 15:04:05")
		}
		if p.dateRange.End == "" {
			p.dateRange.End = defaultEnd.Format("2006-01-02 15:04:05")
		}
		p.timeField = timefd
	}
	return p
}

// DateRange2 adds a date range filter using separate start/end query parameters.
// This is an alternative to DateRange when the client sends separate parameters instead of JSON.
//
// Parameters:
//   - startfd: Query parameter name for start date (e.g., "start_time")
//   - endfd: Query parameter name for end date (e.g., "end_time")
//   - timefd: Database column name for date/time filtering
//   - defaltStart: Default start time if startfd is missing
//   - defaultEnd: Default end time if endfd is missing
//
// Returns:
//   - *PreQuery: Self for method chaining
//
// Example:
//
//	// Request: ?start_time=2024-01-01&end_time=2024-12-31
//	prequery.DateRange2("start_time", "end_time", "created_at", yesterday, now)
func (p *PreQuery) DateRange2(startfd, endfd, timefd string, defaltStart time.Time, defaultEnd time.Time) *PreQuery {
	p.dateRange = DateRange{
		Start: p.form.GetVal2(startfd, defaltStart.Format("2006-01-02 15:04:05")),
		End:   p.form.GetVal2(endfd, defaultEnd.Format("2006-01-02 15:04:05")),
	}
	p.timeField = timefd
	return p
}

// EqualFields specifies which filter parameters should use exact matching (=) instead of LIKE.
// By default, filter parameters use LIKE with wildcard matching.
//
// Parameters:
//   - fd: Field names that require exact equality matching
//
// Returns:
//   - *PreQuery: Self for method chaining
//
// Example:
//
//	// status and user_id will use "=" while other fields use "LIKE"
//	prequery.EqualFields("status", "user_id")
func (p *PreQuery) EqualFields(fd ...string) *PreQuery {
	p.equalFieldds = fd
	return p
}

// KeyFields specifies which database columns should be searched when a "keyword" query parameter is present.
// The keyword will be matched against all specified fields using LIKE with OR logic.
//
// Parameters:
//   - fd: Database column names to search (e.g., "username", "email", "phone")
//
// Returns:
//   - *PreQuery: Self for method chaining
//
// Example:
//
//	// Request: ?keyword=john
//	// SQL: WHERE username LIKE '%john%' OR email LIKE '%john%'
//	prequery.KeyFields("username", "email")
func (p *PreQuery) KeyFields(fd ...string) *PreQuery {
	p.keyfilterFieldds = fd
	return p
}

// QueryField maps a specific query parameter to a database column for filtering.
// This allows custom parameter-to-column mapping beyond the automatic filtering.
//
// Parameters:
//   - column: Database column name (e.g., "user_id")
//   - qfield: Query parameter name (e.g., "userId")
//
// Returns:
//   - *PreQuery: Self for method chaining
//
// Example:
//
//	// Request: ?node_id=123
//	// SQL: WHERE nas_node_id = '123'
//	prequery.QueryField("nas_node_id", "node_id")
func (p *PreQuery) QueryField(column, qfield string) *PreQuery {
	value := p.form.GetVal(qfield)
	if value != "" {
		p.params[column] = value
	}
	return p
}

// SetParam manually sets a filter parameter that will be applied as an equality condition.
// This is useful for programmatically adding filters beyond HTTP request parameters.
//
// Parameters:
//   - key: Database column name
//   - value: Value to match (will be used in WHERE key = value)
//
// Returns:
//   - *PreQuery: Self for method chaining
//
// Example:
//
//	// Force filter by current user's node
//	prequery.SetParam("node_id", currentUser.NodeID)
func (p *PreQuery) SetParam(key string, value interface{}) *PreQuery {
	p.params[key] = value
	return p
}

// Query applies all configured filters to a GORM query and returns the modified query.
// This is the final step in the builder chain that constructs the actual SQL WHERE and ORDER BY clauses.
//
// The method processes:
//  1. Sorting from "sort" query parameter (or default sort if not provided)
//  2. Date range filtering (if configured via DateRange/DateRange2)
//  3. Exact match filters from "equal" query parameters
//  4. Custom parameters from SetParam and QueryField
//  5. Wildcard filters from "filter" query parameters
//  6. Keyword search across KeyFields (if "keyword" parameter present)
//
// Parameters:
//   - query: Base GORM query to modify
//
// Returns:
//   - *gorm.DB: Modified query with WHERE and ORDER BY clauses applied
//
// Example:
//
//	prequery := NewPreQuery(c).
//	    DefaultOrderBy("id DESC").
//	    KeyFields("username", "email")
//
//	var users []User
//	db := prequery.Query(app.GDB().Model(&User{}))
//	db.Find(&users)
func (p *PreQuery) Query(query *gorm.DB) *gorm.DB {
	if len(ParseSortMap(p.context)) == 0 {
		query = query.Order(p.defaultOrderby)
	} else {
		for name, stype := range ParseSortMap(p.context) {
			query = query.Order(fmt.Sprintf("%s %s", name, stype))
		}
	}

	if p.dateRange.Start != "" {
		query = query.Where(p.timeField+" >= ? ", p.dateRange.Start)
	}

	if p.dateRange.End != "" {
		query = query.Where(p.timeField+" <= ?", p.dateRange.End)
	}

	for name, value := range ParseEqualMap(p.context) {
		if common.IsEmptyOrNA(value) {
			continue
		}
		if _, ok := p.params[name]; ok {
			continue
		}
		query = query.Where(fmt.Sprintf("%s = ?", name), value)
	}

	for name, value := range p.params {
		query = query.Where(fmt.Sprintf("%s = ?", name), value)
	}

	for name, value := range ParseFilterMap(p.context) {
		if common.IsEmptyOrNA(value) {
			continue
		}
		if _, ok := p.params[name]; ok {
			continue
		}
		if common.InSlice(name, p.equalFieldds) {
			query = query.Where(fmt.Sprintf("%s = ?", name), value)
		} else {
			query = query.Where(fmt.Sprintf("%s like ?", name), "%"+value+"%")
		}
	}

	keyword := p.context.QueryParam("keyword")
	if keyword != "" {
		for i, keyfd := range p.keyfilterFieldds {
			if i == 0 {
				query = query.Where(keyfd+" like ?", "%"+keyword+"%")
			} else {
				query = query.Or(keyfd+" like ?", "%"+keyword+"%")
			}
		}
	}

	return query
}
