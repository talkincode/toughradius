package app

import (
	"time"

	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/models"
)

func (a *Application) initSettings() {
	a.gormDB.Where("1 = 1").Delete(&models.SysConfig{})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 1, Type: "system", Name: "SystemTitle", Value: "Toughradius Management System", Remark: "System title"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 1, Type: "system", Name: "SystemTheme", Value: "light", Remark: "System theme"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "system", Name: "SystemLoginRemark", Value: "Recommended browser: Chrome/Edge", Remark: "Login page description"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "system", Name: "SystemLoginSubtitle", Value: "Toughradius Community Edition", Remark: "Login form title"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "radius", Name: "AcctInterimInterval", Value: "120", Remark: "Default Acctounting interim interval"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "radius", Name: "AccountingHistoryDays", Value: "90", Remark: "Radius accounting logging expire days"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "radius", Name: "RadiusIgnorePwd", Value: "disabled", Remark: "Radius ignore Passowrd check"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "tr069", Name: "CpeAutoRegister", Value: "enabled", Remark: "Tr069 Cpe auto register"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "tr069", Name: "CwmpDownloadUrlPrefix", Value: "http://127.0.0.1:1819", Remark: "Tr069 Download Url Prefix"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "tr069", Name: "CpeConnectionRequestPassword", Value: "cwmppassword", Remark: "CPE Connection authentication password"})
}

func (a *Application) initOpr() {
	a.gormDB.Where("1 = 1").Delete(&models.SysOpr{})
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

func (a *Application) checkSuper() {
	var count int64
	a.gormDB.Model(&models.SysOpr{}).Where("username='admin', level = ?", "super").Count(&count)
	if count == 0 {
		a.initOpr()
	}
}

func (a *Application) checkSettings() {
	var existConfig = func(cname string) bool {
		var count int64
		a.gormDB.Model(&models.SysConfig{}).Where("name = ?", cname).Count(&count)
		return count > 0
	}
	switch {
	case !existConfig("SystemTitle"):
		a.gormDB.Create(&models.SysConfig{ID: 1000, Sort: 1, Type: "system", Name: "SystemTitle", Value: "Toughradius Management System", Remark: "System title"})
	case !existConfig("SystemTheme"):
		a.gormDB.Create(&models.SysConfig{ID: 1001, Sort: 2, Type: "system", Name: "SystemTheme", Value: "light", Remark: "System theme"})
	case !existConfig("SystemLoginRemark"):
		a.gormDB.Create(&models.SysConfig{ID: 1002, Sort: 3, Type: "system", Name: "SystemLoginRemark", Value: "Recommended browser: Chrome/Edge", Remark: "Login page description"})
	case !existConfig("SystemLoginSubtitle"):
		a.gormDB.Create(&models.SysConfig{ID: 1003, Sort: 4, Type: "system", Name: "SystemLoginSubtitle", Value: "Toughradius Community Edition", Remark: "Login form title"})
	case !existConfig("AcctInterimInterval"):
		a.gormDB.Create(&models.SysConfig{ID: 1004, Sort: 5, Type: "radius", Name: "AcctInterimInterval", Value: "120", Remark: "Default Acctounting interim interval"})
	case !existConfig("AccountingHistoryDays"):
		a.gormDB.Create(&models.SysConfig{ID: 1005, Sort: 6, Type: "radius", Name: "AccountingHistoryDays", Value: "90", Remark: "Radius accounting logging expire days"})
	case !existConfig("RadiusIgnorePwd"):
		a.gormDB.Create(&models.SysConfig{ID: 1006, Sort: 7, Type: "radius", Name: "RadiusIgnorePwd", Value: "disabled", Remark: "Radius ignore Passowrd check"})
	case !existConfig("CpeAutoRegister"):
		a.gormDB.Create(&models.SysConfig{ID: 1007, Sort: 8, Type: "tr069", Name: "CpeAutoRegister", Value: "enabled", Remark: "Tr069 Cpe auto register"})
	case !existConfig("CwmpDownloadUrlPrefix"):
		a.gormDB.Create(&models.SysConfig{ID: 1008, Sort: 9, Type: "tr069", Name: "CwmpDownloadUrlPrefix", Value: "http://127.0.0.1:1819", Remark: "Tr069 Download Url Prefix"})
	case !existConfig("CpeConnectionRequestPassword"):
		a.gormDB.Create(&models.SysConfig{ID: 1009, Sort: 10, Type: "tr069", Name: "CpeConnectionRequestPassword", Value: "cwmppassword", Remark: "CPE Connection authentication password"})
	}
}
