package app

import (
	"github.com/talkincode/toughradius/v9/pkg/metrics"
)

const (
	MetricsRadiusOline              = "radus_online"
	MetricsRadiusOffline            = "radus_offline"
	MetricsRadiusRejectNotExists    = "radus_reject_not_exists"
	MetricsRadiusRejectDisable      = "radus_reject_disabled"
	MetricsRadiusRejectExpire       = "radus_reject_expire"
	MetricsRadiusRejectOther        = "radus_reject_other"
	MetricsRadiusRejectLimit        = "radus_reject_limit"
	MetricsRadiusRejectBindError    = "radus_reject_bind_error"
	MetricsRadiusRejectLdapError    = "radus_reject_ldap_error"
	MetricsRadiusRejectPasswdError  = "radus_reject_passwd_error" //nolint:gosec // G101: this is a metric name, not a credential
	MetricsRadiusRejectUnauthorized = "radus_reject_unauthorized"
	MetricsRadiusAuthDrop           = "radus_auth_drop"
	MetricsRadiusAcctDrop           = "radus_acct_drop"
	MetricsRadiusAccept             = "radus_accept"
	MetricsRadiusAccounting         = "radus_accounting"
)

var metricsNames = []string{
	MetricsRadiusOline,
	MetricsRadiusOffline,
	MetricsRadiusRejectNotExists,
	MetricsRadiusRejectDisable,
	MetricsRadiusRejectExpire,
	MetricsRadiusRejectOther,
	MetricsRadiusRejectLimit,
	MetricsRadiusRejectBindError,
	MetricsRadiusRejectLdapError,
	MetricsRadiusRejectPasswdError,
	MetricsRadiusRejectUnauthorized,
	MetricsRadiusAuthDrop,
	MetricsRadiusAcctDrop,
	MetricsRadiusAccept,
	MetricsRadiusAccounting,
}

// GetRadiusMetrics retrieves the current counter value for a specific RADIUS metric.
// Metrics are thread-safe and use atomic operations for concurrent access.
//
// Common metrics include:
//   - MetricsRadiusAccept: Successful authentication count
//   - MetricsRadiusRejectPasswdError: Password mismatch rejections
//   - MetricsRadiusOline: Active online sessions
//   - MetricsRadiusAccounting: Accounting requests processed
//
// Parameters:
//   - name: Metric name (use constants like MetricsRadiusAccept)
//
// Returns:
//   - int64: Current counter value (0 if metric doesn't exist)
//
// Example:
//
//	onlineCount := app.GetRadiusMetrics(app.MetricsRadiusOline)
//	log.Printf("Active sessions: %d", onlineCount)
func GetRadiusMetrics(name string) int64 {
	return metrics.GetCounter(name)
}

// GetAllRadiusMetrics retrieves all RADIUS metrics as a map for monitoring dashboards.
// This is commonly used by the metrics endpoint to expose Prometheus-style metrics.
//
// Returns:
//   - map[string]int64: All metric names and their current values
//
// Example:
//
//	func MetricsHandler(c echo.Context) error {
//	    metrics := app.GetAllRadiusMetrics()
//	    return c.JSON(200, metrics)
//	}
func GetAllRadiusMetrics() map[string]int64 {
	result := make(map[string]int64)
	for _, name := range metricsNames {
		result[name] = GetRadiusMetrics(name)
	}
	return result
}

// IncRadiusMetric atomically increments a RADIUS metric counter by 1.
// This is called internally by RADIUS authentication and accounting handlers.
//
// Parameters:
//   - name: Metric name to increment (use MetricsRadius* constants)
//
// Side effects:
//   - Atomically increments the metric counter
//   - Creates the metric if it doesn't exist (initialized to 0)
//
// Example:
//
//	if authErr != nil {
//	    app.IncRadiusMetric(app.MetricsRadiusRejectPasswdError)
//	    return radius.CodeAccessReject
//	}
//	app.IncRadiusMetric(app.MetricsRadiusAccept)
func IncRadiusMetric(name string) {
	metrics.Inc(name)
}
