package domain

var Tables = []interface{}{
	// System
	&SysConfig{},
	&SysOpr{},
	&SysOprLog{},
	&SysCert{},
	// Network
	&NetNode{},
	&NetNas{},
	// Radius
	&RadiusAccounting{},
	&RadiusOnline{},
	&RadiusSessionActionAudit{},
	&RadiusProfile{},
	&RadiusUser{},
}
