package supervise

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/snmp"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/events"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/webserver"
)

func InitRouter() {

	webserver.GET("/admin/supervise", func(c echo.Context) error {
		return c.Render(http.StatusOK, "supervise", nil)
	})

	webserver.GET("/admin/supervise/type/options", func(c echo.Context) error {
		var opts = make([]web.JsonOptions, 0)
		opts = append(opts, web.JsonOptions{Id: "cwmp", Value: "TR069 Preset"})
		opts = append(opts, web.JsonOptions{Id: "cwmpconfig", Value: "TR069 Config"})
		return c.JSON(http.StatusOK, opts)
	})

	webserver.GET("/admin/supervise/action/query", func(c echo.Context) error {
		var devid, ctype string
		common.Must(web.NewParamReader(c).
			ReadRequiedString(&devid, "devid").
			ReadRequiedString(&ctype, "ctype").
			LastError)

		var dev models.NetCpe
		common.Must(app.GDB().Where("id=?", devid).First(&dev).Error)

		var actions []SuperviseAction

		switch ctype {
		case "cwmpconfig":
			if dev.VendorCode == snmp.VendorMikrotik {
				var data []models.CwmpConfig
				err := app.GDB().Find(&data).Error
				if err != nil {
					log.Error(err)
				}
				for _, sdata := range data {
					if sdata.SoftwareVersion != "" && sdata.SoftwareVersion != dev.SoftwareVersion {
						continue
					}
					actions = append(actions, SuperviseAction{
						Name:  sdata.Name,
						Type:  "cwmpconfig",
						Level: sdata.Level,
						Sid:   sdata.ID,
					})
				}
			}
		case "cwmp":
			actions = append(actions, getCwmpCmds(dev.VendorCode)...)
		}

		return c.JSON(http.StatusOK, actions)
	})

	webserver.POST("/admin/superviselog/firmware/update", func(c echo.Context) error {
		var devids, session, firmwareid string
		common.Must(web.NewParamReader(c).
			ReadRequiedString(&devids, "devids").
			ReadRequiedString(&session, "session").
			ReadRequiedString(&firmwareid, "firmwareid").LastError)
		return execCwmpUpdateFirmware(c, strings.Split(devids, ","), firmwareid, session)
	})

	webserver.GET("/admin/supervise/action/execute", func(c echo.Context) error {
		var id, stype, session string
		var deviceId int64
		common.Must(web.NewParamReader(c).
			ReadRequiedString(&session, "session").
			ReadRequiedString(&id, "id").
			ReadRequiedString(&stype, "type").
			ReadInt64(&deviceId, "devid", 0).LastError)
		switch stype {
		case "cwmp":
			return execCwmp(c, id, deviceId, session)
		case "cwmpconfig":
			return execCwmpConfig(c, id, deviceId, session)
		}
		return c.JSON(200, web.RestError("unsupported action type "+stype))
	})

	webserver.GET("/admin/supervise/action/listen", func(c echo.Context) error {
		sse := web.NewSSE(c)
		var session string
		common.Must(web.NewParamReader(c).
			ReadRequiedString(&session, "session").LastError)

		var writeMessage = func(session, level, message string) {
			if session == session {
				if level != "" {
					sse.WriteText(fmt.Sprintf("%s :: ", strings.ToUpper(level)))
				}
				for _, s := range strings.Split(message, "\n") {
					sse.WriteText(s)
				}
			}
		}

		var listenFunc = func(devid int64, session, level, message string) {
			writeMessage(session, level, message)
		}
		var listenFunc2 = func(devid int64, session, message string) {
			writeMessage(session, "", message)
		}
		var listenFunc3 = func(sn, session, level, message string) {
			writeMessage(session, level, message)
		}

		events.Supervisor.SubscribeAsync(events.EventSuperviseLog, listenFunc, false)
		events.Supervisor.SubscribeAsync(events.EventSuperviseStatus, listenFunc2, false)
		events.Supervisor.SubscribeAsync(events.EventCwmpSuperviseStatus, listenFunc3, false)
		var unsubscribe = func() {
			events.Supervisor.Unsubscribe(events.EventSuperviseLog, listenFunc)
			events.Supervisor.Unsubscribe(events.EventSuperviseStatus, listenFunc2)
			events.Supervisor.Unsubscribe(events.EventCwmpSuperviseStatus, listenFunc3)
		}
		for {
			select {
			case <-c.Request().Context().Done():
				unsubscribe()
				return nil
			case <-time.After(time.Second * 120):
				unsubscribe()
				return nil
			}
		}
	})

}
