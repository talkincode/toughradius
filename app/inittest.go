package app

import (
	"time"

	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/models"
)

func (a *Application) InitTest() {
	a.initTestSettings()
	a.initTestOpr()
	a.initTestNode()
	a.initTestVpe()
	a.initTestRadiusProfile()
	a.initTestRadiusAccount()

}

func (a *Application) initTestSettings() {
	a.gormDB.Where("1 = 1").Delete(&models.SysConfig{})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 1, Type: "system", Name: "SystemTitle", Value: "ToughRADIUS management system", Remark: "System title"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 1, Type: "system", Name: "SystemTheme", Value: "light", Remark: "System theme"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "system", Name: "SystemLoginRemark", Value: "Recommended browser: Chrome/Edge", Remark: "Login page description"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "system", Name: "SystemLoginSubtitle", Value: "ToughRADIUS community edition", Remark: "Login form title"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "radius", Name: "AcctInterimInterval", Value: "120", Remark: "Default Acctounting interim interval"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "radius", Name: "AccountingHistoryDays", Value: "90", Remark: "Radius accounting logging expire days"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "radius", Name: "RadiusIgnorePwd", Value: "disabled", Remark: "Radius ignore Passowrd check"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "tr069", Name: "CwmpDownloadUrlPrefix", Value: "http://127.0.0.1:1819", Remark: "Tr069 Download Url Prefix"})
	a.gormDB.Create(&models.SysConfig{ID: 0, Sort: 3, Type: "tr069", Name: "CpeConnectionRequestPassword", Value: "cwmppassword", Remark: "CPE Connection authentication password"})
}

func (a *Application) initTestOpr() {
	a.gormDB.Where("1 = 1").Delete(&models.SysOpr{})
	a.gormDB.Create(&models.SysOpr{
		ID:        common.UUIDint64(),
		Realname:  "管理员",
		Mobile:    "0000",
		Email:     "N/A",
		Username:  "admin",
		Password:  common.Sha256HashWithSalt("toughradius", common.SecretSalt),
		Level:     "super",
		Status:    "enabled",
		Remark:    "super",
		LastLogin: time.Now(),
	})
	a.gormDB.Create(&models.SysOpr{
		ID:        common.UUIDint64(),
		Realname:  "API用户",
		Mobile:    "0000",
		Email:     "N/A",
		Username:  "apiuser",
		Password:  common.Sha256HashWithSalt("Api_189", common.SecretSalt),
		Level:     "api",
		Status:    "enabled",
		Remark:    "API-only",
		LastLogin: time.Now(),
	})
}

func (a *Application) initTestNode() {
	a.gormDB.Where("1 = 1").Delete(&models.NetNode{})
	a.gormDB.Create(&models.NetNode{
		ID:        9999,
		Name:      "testnode",
		Remark:    "Test Node",
		Tags:      "test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
}

func (a *Application) initTestVpe() {
	a.gormDB.Where("1 = 1").Delete(&models.NetVpe{})
	a.gormDB.Create(&models.NetVpe{
		ID:         9999,
		NodeId:     9999,
		LdapId:     0,
		Name:       "test vope",
		Identifier: "tradtest",
		Hostname:   "",
		Ipaddr:     "127.0.0.2",
		Secret:     "secret",
		CoaPort:    0,
		Model:      "",
		VendorCode: "14988",
		Status:     "enabled",
		Tags:       "",
		Remark:     "test vpe",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
}

func (a *Application) initTestRadiusProfile() {
	a.gormDB.Where("1 = 1").Delete(&models.RadiusProfile{})
	a.gormDB.Create(&models.RadiusProfile{
		ID:        9999,
		NodeId:    9999,
		Name:      "testprofile",
		Status:    "enabled",
		AddrPool:  "",
		ActiveNum: 1,
		UpRate:    100000,
		DownRate:  100000,
		Remark:    "Test Profile",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
}

func (a *Application) initTestRadiusAccount() {
	a.gormDB.Where("1 = 1").Delete(&models.RadiusUser{})
	expire, _ := time.Parse("2006-01-02 15:04:05", "2024-12-31 23:59:59")
	a.gormDB.Create(&models.RadiusUser{
		ID:         common.UUIDint64(),
		NodeId:     9999,
		ProfileId:  9999,
		Realname:   "test user",
		Mobile:     "1360000000",
		Username:   "test01",
		Password:   "111111",
		AddrPool:   "",
		ActiveNum:  1,
		UpRate:     100000,
		DownRate:   100000,
		Vlanid1:    0,
		Vlanid2:    0,
		IpAddr:     "",
		MacAddr:    "",
		BindVlan:   0,
		BindMac:    0,
		ExpireTime: expire,
		Status:     "enabled",
		Remark:     "test user",
		LastOnline: time.Time{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})

}
