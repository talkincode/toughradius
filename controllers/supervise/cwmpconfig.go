package supervise

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/cwmp"
	"github.com/talkincode/toughradius/v8/common/timeutil"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/events"
	"github.com/talkincode/toughradius/v8/models"
)

func execCwmpConfig(c echo.Context, id string, deviceId int64, session string) error {
	var dev models.NetCpe
	common.Must(app.GDB().Where("id=?", deviceId).First(&dev).Error)
	if common.IsEmptyOrNA(dev.Sn) {
		return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("Device SN %s invalid", dev.Sn)))
	}

	var script models.CwmpConfig
	err := app.GDB().Where("id=? ", id).First(&script).Error
	if err != nil {
		return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("TR069 configuration does not exist %s", err.Error())))
	}

	// 并发检查
	var scount int64
	app.GDB().Model(models.CwmpConfigSession{}).
		Where("device_id = ?  and exec_status = ? and exec_time < ?", dev.ID, "initialize",
			time.Now().Add(time.Second*time.Duration(script.Timeout))).Count(&scount)
	if scount > 0 {
		return c.JSON(http.StatusOK, web.RestError("The current device already has a task running, please wait for the execution to complete"))
	}

	cpe := app.GetCwmpCpe(dev.Sn)

	if !app.MatchDevice(dev, script.Oui, script.ProductClass, script.SoftwareVersion) {
		return c.JSON(http.StatusOK,
			web.RestError(fmt.Sprintf("device version %s mismatch %s", dev.SoftwareVersion, script.SoftwareVersion)))
	}

	if !app.MatchTaskTags(dev.TaskTags, script.TaskTags) {
		return c.JSON(http.StatusOK,
			web.RestError(fmt.Sprintf("Device Task tags %s mismatch %s", dev.TaskTags, script.TaskTags)))
	}

	go func() {
		// 创建脚本下发记录
		var scontent = script.Content

		scriptSession := &models.CwmpConfigSession{
			ID:              common.UUIDint64(),
			ConfigId:        script.ID,
			CpeId:           dev.ID,
			Session:         session,
			Name:            script.Name,
			Level:           script.Level,
			SoftwareVersion: script.SoftwareVersion,
			ProductClass:    script.ProductClass,
			Oui:             script.Oui,
			TaskTags:        script.TaskTags,
			Content:         script.Content,
			ExecStatus:      "initialize",
			LastError:       "",
			Timeout:         script.Timeout,
			ExecTime:        time.Now(),
			RespTime:        timeutil.EmptyTime,
			CreatedAt:       time.Time{},
			UpdatedAt:       time.Time{},
		}
		common.Must(app.GDB().Create(scriptSession).Error)

		// 文件下载 token 当日有效
		var token = common.Md5Hash(session + app.GConfig().Web.Secret + time.Now().Format("20060102"))

		err = cpe.SendCwmpEventData(models.CwmpEventData{
			Session: session,
			Sn:      dev.Sn,
			Message: &cwmp.Download{
				ID:         session,
				Name:       "Cwmp VenderConfiguration Task",
				NoMore:     0,
				CommandKey: session,
				FileType:   "3 Vendor Configuration File",
				URL: fmt.Sprintf("%s/cwmpfiles/%s/%s/latest.alter",
					app.GApp().GetTr069SettingsStringValue(app.ConfigTR069AccessAddress), session, token),
				Username:       "",
				Password:       "",
				FileSize:       len([]byte(scontent)),
				TargetFileName: common.IfEmptyStr(script.TargetFilename, session+".alter"),
				DelaySeconds:   5,
				SuccessURL:     "",
				FailureURL:     "",
			},
		}, 5000, true)
		if err != nil {
			events.PubSuperviseLog(dev.ID, session, "error", fmt.Sprintf("TR069 Push config timed out %s", err.Error()))
		}

		go connectDeviceAuth(session, dev)

	}()

	return c.JSON(200, web.RestSucc("The instruction has been sent, please check the execution log later, please do not execute it repeatedly in a short time"))

}
