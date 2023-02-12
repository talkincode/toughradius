package app

import (
	"time"

	istats "github.com/montanaflynn/stats"
	"github.com/talkincode/toughradius/common/zaplog"
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
	MetricsRadiusRejectPasswdError  = "radus_reject_passwd_error"
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

func GetRadiusMetrics(name string) int64 {
	var value float64 = 0
	vals := make([]float64, 0)
	points, err := zaplog.TSDB().Select(name, nil,
		time.Now().Add(-86400*time.Second).Unix(), time.Now().Unix())
	if err != nil {
		return 0
	}
	for _, p := range points {
		vals = append(vals, p.Value)
	}
	value, _ = istats.Sum(vals)
	if value < 0 {
		value = 0
	}
	return int64(value)
}

func GetAllRadiusMetrics() map[string]int64 {
	var result = make(map[string]int64)
	for _, name := range metricsNames {
		result[name] = GetRadiusMetrics(name)
	}
	return result
}
