package app

import (
	"time"

	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/models"
)

func (a *Application) checkSuper() {
	var count int64
	a.gormDB.Model(&models.SysOpr{}).Where("username='admin' and level = ?", "super").Count(&count)
	if count == 0 {
		a.gormDB.Create(&models.SysOpr{
			ID:        common.UUIDint64(),
			Realname:  "administrator",
			Mobile:    "0000",
			Email:     "N/A",
			Username:  "admin",
			Password:  common.Sha256HashWithSalt("toughradius", common.SecretSalt),
			Level:     "super",
			Status:    "enabled",
			Remark:    "super",
			LastLogin: time.Now(),
		})
	}
}

func (a *Application) checkSettings() {
	var checkConfig = func(sortid int, stype, cname, value, remark string) {
		var count int64
		a.gormDB.Model(&models.SysConfig{}).Where("type = ? and name = ?", stype, cname).Count(&count)
		if count == 0 {
			a.gormDB.Create(&models.SysConfig{ID: 0, Sort: sortid, Type: stype, Name: cname, Value: value, Remark: remark})
		}
	}

	for sortid, name := range ConfigConstants {
		switch name {
		case ConfigSystemTitle:
			checkConfig(sortid, "system", ConfigSystemTitle, "Toughradius Management System", "System title")
		case ConfigSystemTheme:
			checkConfig(sortid, "system", ConfigSystemTheme, "light", "System theme")
		case ConfigSystemLoginRemark:
			checkConfig(sortid, "system", ConfigSystemLoginRemark, "Recommended browser: Chrome/Edge", "Login page description")
		case ConfigSystemLoginSubtitle:
			checkConfig(sortid, "system", ConfigSystemLoginSubtitle, "TeamsACS Community Edition", "Login form title")
		case ConfigCpeAutoRegister:
			checkConfig(sortid, "tr069", ConfigCpeAutoRegister, "enabled", "Auto register CPE device")
		case ConfigTR069AccessAddress:
			checkConfig(sortid, "tr069", ConfigTR069AccessAddress, "http://127.0.0.1:2999", "Toughradius TR069 access address, HTTP | https://domain:port")
		case ConfigTR069AccessPassword:
			checkConfig(sortid, "tr069", ConfigTR069AccessPassword, "touighradiustr069password", "Toughradius TR069 access password, It is provided to CPE to access TeamsACS")
		case ConfigCpeConnectionRequestPassword:
			checkConfig(sortid, "tr069", ConfigCpeConnectionRequestPassword, "toughradiuscpepassword", "CPE Connection authentication password, It is provided to TeamsACS to access CPE")
		case ConfigRadiusIgnorePwd:
			checkConfig(sortid, "radius", ConfigRadiusIgnorePwd, "disabled", "Radius ignore Passowrd check")
		case ConfigRadiusAccountingHistoryDays:
			checkConfig(sortid, "radius", ConfigRadiusAccountingHistoryDays, "disabled", "Radius accounting logging expire days")
		case ConfigRadiusAcctInterimInterval:
			checkConfig(sortid, "radius", ConfigRadiusAcctInterimInterval, "disabled", "Radius default Acctounting interim interval")
		case ConfigRadiusEapMethod:
			checkConfig(sortid, "radius", ConfigRadiusEapMethod, "eap-md5", "Radius eap method")

		}
	}
}
