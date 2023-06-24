package freeradius

import (
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/models"
	"go.uber.org/zap"
)

// freeradius api service

func getOnlineCount(username string) (int64, error) {
	var count int64
	err := app.GDB().Model(&models.RadiusOnline{}).
		Where("username = ?", username).Count(&count).Error
	return count, err
}

func getOnlineCountBySessionid(acctSessionId string) (int64, error) {
	var count int64
	err := app.GDB().Model(&models.RadiusOnline{}).
		Where("acct_session_id = ?", acctSessionId).Count(&count).Error
	return count, err
}

func BatchClearRadiusOnlineDataByNas(nasip, nasid string) error {
	return app.GDB().Where("nas_addr = ? or nas_id = ?", nasip, nasid).Delete(models.RadiusOnline{}).Error
}

func getAcctStartTime(sessionTime string) time.Time {
	m, _ := time.ParseDuration("-" + sessionTime + "s")
	return time.Now().Add(m)
}

func getInputTotal(form *web.WebForm) int64 {
	var acctInputOctets = form.GetInt64Val("acctInputOctets", 0)
	var acctInputGigawords = form.GetInt64Val("acctInputGigaword", 0)
	return acctInputOctets + acctInputGigawords*4*1024*1024*1024
}

func getOutputTotal(form *web.WebForm) int64 {
	var acctOutputOctets = form.GetInt64Val("acctOutputOctets", 0)
	var acctOutputGigawords = form.GetInt64Val("acctOutputGigawords", 0)
	return acctOutputOctets + acctOutputGigawords*4*1024*1024*1024
}

// updateRadiusOnline 更新记账信息
func updateRadiusOnline(form *web.WebForm) error {
	var user models.RadiusUser
	err := app.GDB().Where("username=?", form.GetVal("username")).First(&user).Error
	if err != nil {
		return fmt.Errorf("user %s not exists", form.GetVal("username"))
	}
	var sessionId = form.GetVal2("acctSessionId", "")
	var statusType = form.GetVal2("acctStatusType", "")
	radOnline := models.RadiusOnline{
		ID:                common.UUIDint64(),
		Username:          form.GetVal("username"),
		NasId:             form.GetVal("nasid"),
		NasAddr:           form.GetVal("nasip"),
		NasPaddr:          form.GetVal("nasip"),
		SessionTimeout:    form.GetIntVal("sessionTimeout", 0),
		FramedIpaddr:      form.GetVal2("framedIPAddress", "0.0.0.0"),
		FramedNetmask:     form.GetVal2("framedIPNetmask", common.NA),
		MacAddr:           form.GetVal2("macAddr", common.NA),
		NasPort:           0,
		NasClass:          common.NA,
		NasPortId:         form.GetVal2("nasPortId", common.NA),
		NasPortType:       0,
		ServiceType:       0,
		AcctSessionId:     sessionId,
		AcctSessionTime:   form.GetIntVal("acctSessionTime", 0),
		AcctInputTotal:    getInputTotal(form),
		AcctOutputTotal:   getOutputTotal(form),
		AcctInputPackets:  form.GetIntVal("acctInputPackets", 0),
		AcctOutputPackets: form.GetIntVal("acctOutputPackets", 0),
		AcctStartTime:     getAcctStartTime(form.GetVal2("acctSessionTime", "0")),
		LastUpdate:        time.Now(),
	}

	switch statusType {
	case "Start", "Update", "Alive", "Interim-Update":
		if statusType == "Start" {
			log.Info2("add radius online",
				zap.String("namespace", "freeradius"),
				zap.String("metrics", app.MetricsRadiusOline),
			)
			app.GDB().Model(models.RadiusUser{}).
				Where("username=?", user.Username).
				Update("last_online", time.Now())
		} else {
			app.GDB().Model(models.RadiusUser{}).
				Where("username=? and last_online is null", user.Username).
				Update("last_online", radOnline.AcctStartTime)
		}

		ocount, _ := getOnlineCountBySessionid(sessionId)
		if ocount == 0 {
			log.Info2("Add radius online",
				zap.String("namespace", "radius"),
				zap.String("username", radOnline.Username))
			return app.GDB().Create(&radOnline).Error
		} else {
			log.Info2("Update radius online",
				zap.String("namespace", "radius"),
				zap.String("username", radOnline.Username))
			return app.GDB().Model(models.RadiusOnline{}).
				Where("acct_session_id=?", sessionId).
				Updates(radOnline).Error
		}
	case "Stop":
		log.Info2("stop radius online",
			zap.String("namespace", "freeradius"),
			zap.String("metrics", app.MetricsRadiusOffline),
		)
		app.GDB().Create(&models.RadiusAccounting{
			ID:                radOnline.ID,
			Username:          radOnline.Username,
			NasId:             radOnline.NasId,
			NasAddr:           radOnline.NasAddr,
			NasPaddr:          radOnline.NasPaddr,
			SessionTimeout:    radOnline.SessionTimeout,
			FramedIpaddr:      radOnline.FramedIpaddr,
			FramedNetmask:     radOnline.FramedNetmask,
			MacAddr:           radOnline.MacAddr,
			NasPort:           radOnline.NasPort,
			NasClass:          radOnline.NasClass,
			NasPortId:         radOnline.NasPortId,
			NasPortType:       radOnline.NasPortType,
			ServiceType:       radOnline.ServiceType,
			AcctSessionId:     radOnline.AcctSessionId,
			AcctSessionTime:   radOnline.AcctSessionTime,
			AcctInputTotal:    radOnline.AcctInputTotal,
			AcctOutputTotal:   radOnline.AcctOutputTotal,
			AcctInputPackets:  radOnline.AcctInputPackets,
			AcctOutputPackets: radOnline.AcctOutputPackets,
			AcctStartTime:     radOnline.AcctStartTime,
			LastUpdate:        time.Now(),
			AcctStopTime:      time.Now(),
		})
		return app.GDB().Where("acct_session_id=?", sessionId).
			Delete(&models.RadiusOnline{}).Error
	case "Accounting-On", "Accounting-Off":
		// app.GDB().Raw("truncate table radius_online")
	}

	return nil
}
