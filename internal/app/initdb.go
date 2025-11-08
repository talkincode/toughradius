package app

import (
	"time"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

func (a *Application) checkSuper() {
	var count int64
	a.gormDB.Model(&domain.SysOpr{}).Where("username='admin' and level = ?", "super").Count(&count)
	if count == 0 {
		a.gormDB.Create(&domain.SysOpr{
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
		a.gormDB.Model(&domain.SysConfig{}).Where("type = ? and name = ?", stype, cname).Count(&count)
		if count == 0 {
			a.gormDB.Create(&domain.SysConfig{ID: 0, Sort: sortid, Type: stype, Name: cname, Value: value, Remark: remark})
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
