package logs

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/lokiquery"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/webserver"
)

func initLokiRouter() {

	type lokilog struct {
		Job       string `json:"job"`
		Level     string `json:"level"`
		Caller    string `json:"caller"`
		Msg       string `json:"msg"`
		Timestamp string `json:"timestamp"`
	}

	webserver.GET("/admin/loki", func(c echo.Context) error {
		return c.Render(http.StatusOK, "loki_logger", map[string]interface{}{})
	})

	webserver.GET("/admin/loki/query", func(c echo.Context) error {
		// var data []lokilog
		var limit int
		var job, namespace, starttime, endtime, level, keyword, keyreg string
		web.NewParamReader(c).
			ReadInt(&limit, "limit", limit).
			ReadString(&job, "job").
			ReadString(&starttime, "starttime").
			ReadString(&endtime, "endtime").
			ReadString(&namespace, "namespace").
			ReadString(&keyword, "keyword").
			ReadString(&keyreg, "keyreg").
			ReadString(&level, "level")

		var start = time.Now().Add(-time.Hour * 24).UnixNano()
		var end = time.Now().UnixNano()

		if _start, err := time.Parse("2006-01-02 15:04:05", starttime); err == nil {
			start = _start.UnixNano()
		}
		if _end, err := time.Parse("2006-01-02 15:04:05", endtime); err == nil {
			end = _end.UnixNano()
		}

		if job == "none" {
			job = ""
		}

		if namespace == "global" {
			namespace = ""
		}

		cfg := app.GConfig().Logger
		query := lokiquery.NewLokiQueryForm(cfg.LokiApi, cfg.LokiUser, cfg.LokiPwd)
		query.Limit = limit
		// query.Debug = cfg.Mode == zaplog.Dev
		query.Start = start
		query.End = end
		query = query.AddLabel("job", job).
			AddLineContains(keyword).
			AddLineNotContainsReg(keyreg)

		if level != "" {
			query = query.AddLineContains(fmt.Sprintf(`"level":"%s"`, strings.ToUpper(level)))
		}

		if namespace != "" {
			query = query.AddLineContains(fmt.Sprintf(`"namespace":"%s"`, namespace))
		}

		result, err := lokiquery.LokiQuery(query)
		if err != nil {
			return c.JSON(http.StatusBadRequest, common.EmptyList)
		}

		return c.JSON(http.StatusOK, result)
	})

}
