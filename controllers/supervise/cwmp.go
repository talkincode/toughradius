package supervise

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/cwmp"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/events"
	"github.com/talkincode/toughradius/v8/models"
)

func getCwmpCmds(vendor string) []SuperviseAction {
	var cwmpCmds = []SuperviseAction{
		{Name: "Test cpe cwmp connection", Type: "cwmp", Level: "normal", Sid: "cwmpDeviceConnectTest"},
		{Name: "Get the list of RPC methods", Type: "cwmp", Level: "normal", Sid: "cwmpGetRPCMethods"},
		{Name: "Get list of parameter prefixes", Type: "cwmp", Level: "normal", Sid: "cwmpGetParameterNames"},
		{Name: "Get and update device information", Type: "cwmp", Level: "normal", Sid: "cwmpDeviceInfoUpdate"},
		{Name: "Configure device authentication information", Type: "cwmp", Level: "normal", Sid: "cwmpDeviceManagementAuthUpdate"},
		{Name: "Upload device logs", Type: "cwmp", Level: "normal", Sid: "cwmpDeviceUploadLog"},
		{Name: "Upload device backup (text)", Type: "cwmp", Level: "normal", Sid: "cwmpDeviceBackup"},
		{Name: "Factory reset", Type: "cwmp", Level: "major", Sid: "cwmpFactoryReset"},
		{Name: "Restart the device", Type: "cwmp", Level: "major", Sid: "cwmpReboot"},
	}
	switch vendor {
	case cwmp.VendorMikrotik:
		cwmpCmds = append(cwmpCmds, SuperviseAction{
			Name:  "Download factory configuration",
			Type:  "cwmp",
			Level: "major",
			Sid:   "cwmpMikrotikFactoryConfiguration"},
		)
	}
	return cwmpCmds
}

func connectDeviceAuth(session string, dev models.NetCpe) {
	if dev.CwmpUrl == "" {
		return
	}
	isok, err := cwmp.ConnectionRequestAuth(dev.Sn, app.GApp().GetTr069SettingsStringValue(app.ConfigCpeConnectionRequestPassword), dev.CwmpUrl)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error", fmt.Sprintf("TR069 connect device %s failure %s", dev.CwmpUrl, err.Error()))
	}

	if isok {
		events.PubSuperviseLog(dev.ID, session, "info", fmt.Sprintf("TR069 connect device %s success", dev.CwmpUrl))
	}
}

func execCwmp(c echo.Context, id string, deviceId int64, session string) error {
	var dev models.NetCpe
	common.Must(app.GDB().Where("id=?", deviceId).First(&dev).Error)
	if common.IsEmptyOrNA(dev.Sn) {
		return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("Device SN %s invalid", dev.Sn)))
	}

	switch id {
	case "cwmpMikrotikFactoryConfiguration":
		return execCwmpMikrotikFactoryConfiguration(c, id, deviceId, session)
	case "cwmpReboot":
		go cwmpDeviceReboot(id, dev, session)
	case "cwmpFactoryReset":
		go cwmpDeviceFactoryReset(id, dev, session)
	case "cwmpDeviceInfoUpdate":
		go cwmpDeviceInfoUpdate(id, dev, session)
	case "cwmpDeviceManagementAuthUpdate":
		go cwmpDeviceManagementAuthUpdate(id, dev, session)
	case "cwmpDeviceConnectTest":
		go cwmpDeviceConnectTest(id, dev, session)
	case "cwmpGetParameterNames":
		go cwmpGetParameterNames(id, dev, session)
	case "cwmpGetRPCMethods":
		go cwmpGetRPCMethods(id, dev, session)
	case "cwmpDeviceBackup":
		go cwmpDeviceBackup(id, dev, session)
	case "cwmpDeviceUploadLog":
		go cwmpDeviceUploadLog(id, dev, session)
	}
	return c.JSON(200, web.RestSucc("The instruction has been sent, please check the execution log later, please do not execute it repeatedly in a short time"))

}

func cwmpDeviceInfoUpdate(sid string, dev models.NetCpe, session string) {
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.GetParameterValues{
			ID:     session,
			Name:   "",
			NoMore: 0,
			ParameterNames: []string{
				"Device.DeviceInfo.",
				"Device.ManagementServer.",
			},
		},
	}, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error", fmt.Sprintf("TR069 Update device information push timeout %s", err.Error()))
	}
	go connectDeviceAuth(session, dev)

}

func cwmpDeviceManagementAuthUpdate(sid string, dev models.NetCpe, session string) {
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.UpdateManagementAuthInfo(session, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error", fmt.Sprintf("TR069 Update device management authentication information push timeout %s", err.Error()))
	} else {
		events.PubSuperviseStatus(dev.ID, session, fmt.Sprintf("TR069 The task of updating device management authentication information has been submitted， Please wait for the CPE connection to update"))
	}

}

func cwmpDeviceConnectTest(sid string, dev models.NetCpe, session string) {
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.GetParameterValues{
			ID:     session,
			Name:   "test connection",
			NoMore: 0,
			ParameterNames: []string{
				"Device.DeviceInfo.",
			},
		},
	}, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error",
			fmt.Sprintf("TR069 Update device message push timeout %s", err.Error()))
	}

	go connectDeviceAuth(session, dev)

}

func cwmpGetParameterNames(sid string, dev models.NetCpe, session string) {
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.GetParameterNames{
			ID:            session,
			Name:          "GetParameterNames",
			NoMore:        0,
			ParameterPath: "Device.",
			NextLevel:     "true",
		},
	}, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error",
			fmt.Sprintf("CWMP Update device message push timeout %s", err.Error()))
	}

	go connectDeviceAuth(session, dev)

}

func cwmpGetRPCMethods(sid string, dev models.NetCpe, session string) {
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.GetRPCMethods{
			ID:     session,
			Name:   "GetRPCMethods",
			NoMore: 0,
		},
	}, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error",
			fmt.Sprintf("CWMP GetRPCMethods message push timeout %s", err.Error()))
	}

	go connectDeviceAuth(session, dev)

}

func cwmpDeviceUploadLog(sid string, dev models.NetCpe, session string) {
	var token = common.Md5Hash(session + app.GConfig().Web.Secret + time.Now().Format("20060102"))
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.Upload{
			ID:         session,
			Name:       "Cwmp logupload Task",
			NoMore:     0,
			CommandKey: session,
			FileType:   "2 Vendor Log File",
			URL: fmt.Sprintf("%s/cwmpupload/%s/%s/%s.log",
				app.GApp().GetTr069SettingsStringValue(app.ConfigTR069AccessAddress),
				session, token, dev.Sn+"_"+time.Now().Format("20060102")),
			Username:     "",
			Password:     "",
			DelaySeconds: 5,
		},
	}, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error",
			fmt.Sprintf("CWMP Log upload message push timeout %s", err.Error()))
	}

	go connectDeviceAuth(session, dev)

}

func cwmpDeviceBackup(sid string, dev models.NetCpe, session string) {
	var token = common.Md5Hash(session + app.GConfig().Web.Secret + time.Now().Format("20060102"))
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.Upload{
			ID:         session,
			NoMore:     0,
			CommandKey: session,
			FileType:   "1 Vendor Configuration File",
			URL: fmt.Sprintf("%s/cwmpupload/%s/%s/%s.rsc",
				app.GApp().GetTr069SettingsStringValue(app.ConfigTR069AccessAddress),
				session, token, dev.Sn+"_"+time.Now().Format("20060102")),
			Username:     "",
			Password:     "",
			DelaySeconds: 5,
		},
	}, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error",
			fmt.Sprintf("CWMP Backup device message push timeout %s", err.Error()))
	}

	go connectDeviceAuth(session, dev)

}

func cwmpDeviceReboot(sid string, dev models.NetCpe, session string) {
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.Reboot{
			ID:         session,
			Name:       "Cwmp reboot Task",
			NoMore:     0,
			CommandKey: session,
		},
	}, 000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error",
			fmt.Sprintf("Sending CWMP Reboot device information push timed out %s", err.Error()))
	}

	go connectDeviceAuth(session, dev)

}

func cwmpDeviceFactoryReset(sid string, dev models.NetCpe, session string) {
	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	err := cpe.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      dev.Sn,
		Message: &cwmp.FactoryReset{
			ID:     session,
			Name:   "Cwmp FactoryReset Task",
			NoMore: 0,
		},
	}, 5000, true)
	if err != nil {
		events.PubSuperviseLog(dev.ID, session, "error",
			fmt.Sprintf("Sending CWMP FactoryReset device information push timed out %s", err.Error()))
	}

	go connectDeviceAuth(session, dev)

}
