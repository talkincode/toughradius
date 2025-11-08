package app

const (
	ConfigSystemTitle         = "SystemTitle"
	ConfigSystemTheme         = "SystemTheme"
	ConfigSystemLoginRemark   = "SystemLoginRemark"
	ConfigSystemLoginSubtitle = "SystemLoginSubtitle"

	ConfigRadiusIgnorePwd             = "RadiusIgnorePwd"
	ConfigRadiusAccountingHistoryDays = "AccountingHistoryDays"
	ConfigRadiusAcctInterimInterval   = "AcctInterimInterval"
	ConfigRadiusEapMethod             = "RadiusEapMethod"
)

const (
	RadiusVendorMikrotik = "14988"
	RadiusVendorIkuai    = "10055"
	RadiusVendorHuawei   = "2011"
	RadiusVendorZte      = "3902"
	RadiusVendorH3c      = "25506"
	RadiusVendorRadback  = "2352"
	RadiusVendorCisco    = "9"
	RadiusVendorStandard = "0"
)

var ConfigConstants = []string{
	ConfigSystemTitle,
	ConfigSystemTheme,
	ConfigSystemLoginRemark,
	ConfigSystemLoginSubtitle,
	ConfigRadiusIgnorePwd,
	ConfigRadiusAccountingHistoryDays,
	ConfigRadiusAcctInterimInterval,
	ConfigRadiusEapMethod,
}
