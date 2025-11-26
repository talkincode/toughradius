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

// GetRadiusMetrics returns the counter value for a specific metric
func GetRadiusMetrics(name string) int64 {
	return metrics.GetCounter(name)
}

// GetAllRadiusMetrics returns all RADIUS metrics as a map
func GetAllRadiusMetrics() map[string]int64 {
	result := make(map[string]int64)
	for _, name := range metricsNames {
		result[name] = GetRadiusMetrics(name)
	}
	return result
}

// IncRadiusMetric increments a RADIUS metric counter
func IncRadiusMetric(name string) {
	metrics.Inc(name)
}
