package models

var Tables = []interface{}{
	// System
	&SysConfig{},
	&SysOpr{},
	&SysOprLog{},
	&SysApiToken{},
	// Network
	&NetNode{},
	&NetCpe{},
	&NetVpe{},
	&NetCpeParam{},
	&NetLdapServer{},
	// Radius
	&RadiusAccounting{},
	&RadiusOnline{},
	&RadiusProfile{},
	&RadiusUser{},
	// Cwmp
	&CwmpConfigSession{},
	&CwmpConfig{},
	&CwmpFactoryReset{},
	&CwmpFirmwareConfig{},
	&CwmpPreset{},
	&CwmpPresetTask{},
}
