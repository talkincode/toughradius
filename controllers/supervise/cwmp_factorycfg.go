package supervise

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/cwmp"
	"github.com/talkincode/toughradius/v8/common/timeutil"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/events"
	"github.com/talkincode/toughradius/v8/models"
)

func execCwmpMikrotikFactoryConfiguration(c echo.Context, id string, deviceId int64, session string) error {
	var dev models.NetCpe
	common.Must(app.GDB().Where("id=?", deviceId).First(&dev).Error)
	if common.IsEmptyOrNA(dev.Sn) {
		return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("Device SN %s invalid", dev.Sn)))
	}

	var factscript models.CwmpFactoryReset
	err := app.GDB().Where("id=?", dev.FactoryresetId).First(&factscript).Error
	if err != nil {
		return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("Device %s Factoryreset configuration not set", dev.Sn)))
	}

	if factscript.Content == "" {
		return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("Device %s The factoryreset configuration content is empty", dev.Sn)))
	}

	cpe := app.GApp().CwmpTable().GetCwmpCpe(dev.Sn)
	if !cpe.MatchDevice(factscript.Oui, factscript.ProductClass, factscript.SoftwareVersion) {
		return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf("Device %s Does not match CwmpFactoryReset", dev.Sn)))
	}

	// CPE Vars Replace
	var scontent = factscript.Content

	go func() {
		// 创建脚本下发记录
		scriptSession := &models.CwmpConfigSession{
			ID:              common.UUIDint64(),
			ConfigId:        cast.ToString(factscript.ID),
			CpeId:           dev.ID,
			Session:         session,
			Name:            id,
			Level:           "major",
			SoftwareVersion: factscript.SoftwareVersion,
			ProductClass:    factscript.ProductClass,
			Oui:             factscript.Oui,
			Content:         factscript.Content,
			ExecStatus:      "initialize",
			LastError:       "",
			Timeout:         120,
			ExecTime:        time.Now(),
			RespTime:        timeutil.EmptyTime,
			CreatedAt:       time.Time{},
			UpdatedAt:       time.Time{},
		}
		common.Must(app.GDB().Create(scriptSession).Error)

		// 文件下载 token 当日有效
		var token = common.Md5Hash(session + app.GConfig().Web.Secret + time.Now().Format("20060102"))
		err := cpe.SendCwmpEventData(models.CwmpEventData{
			Session: session,
			Sn:      dev.Sn,
			Message: &cwmp.Download{
				ID:         session,
				Name:       "Cwmp FactoryConfiguration Task",
				NoMore:     0,
				CommandKey: session,
				FileType:   "X MIKROTIK Factory Configuration File",
				URL: fmt.Sprintf("%s/cwmpfiles/%s/%s/latest.alter",
					app.GApp().GetTr069SettingsStringValue(app.ConfigTR069AccessAddress), session, token),
				Username:       "",
				Password:       "",
				FileSize:       len([]byte(scontent)),
				TargetFileName: session + ".alter",
				DelaySeconds:   5,
				SuccessURL:     "",
				FailureURL:     "",
			},
		}, 5000, true)
		if err != nil {
			events.PubSuperviseLog(dev.ID, session, "error", fmt.Sprintf("TR069 Push factoryreet configuration timeout %s", err.Error()))
		}

		go connectDeviceAuth(session, dev)

	}()

	return c.JSON(200, web.RestSucc("The next factory configuration command has been sent, please check the execution log later, please do not execute it repeatedly in a short time"))

}
