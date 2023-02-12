package lokiquery

import (
	"testing"
	"time"

	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/config"
)

func TestLokiQuery_QueryString(t *testing.T) {
	lq := NewLokiQueryForm("", "", "")
	t.Log(lq.QueryString())
	lq.AddLabel("job", "teamsacs")
	lq.AddLabel("level", "info")
	lq.AddLineContains("hello")
	t.Log(lq.QueryString())
}

func TestLokiCountOverTime(t *testing.T) {
	app.InitGlobalApplication(config.LoadConfig("../../toughradius.yml"))
	lq := NewLokiQueryForm(app.GConfig().Logger.LokiApi, app.GConfig().Logger.LokiUser, app.GConfig().Logger.LokiPwd)
	lq.Step = "5m"
	lq.Limit = 10
	lq.Debug = true
	lq.Start = time.Now().Add(-time.Hour * 24).UnixNano()
	lq.End = time.Now().UnixNano()
	lq.AddLabel("job", "teamsacs_master")
	lq.AddLabel("level", "info")
	t.Log(lq.QueryString())
	r, err := LokiCountOverTime(lq)
	if err != nil {
		t.Log(err)
	}
	t.Log(r)
}

func TestLokiSumRate(t *testing.T) {
	app.InitGlobalApplication(config.LoadConfig("../../toughradius.yml"))
	lq := NewLokiQueryForm(app.GConfig().Logger.LokiApi, app.GConfig().Logger.LokiUser, app.GConfig().Logger.LokiPwd)
	lq.Step = "5m"
	lq.Limit = 1000
	lq.Debug = true
	lq.Start = time.Now().Add(-time.Hour * 24).UnixNano()
	lq.End = time.Now().UnixNano()
	lq.AddLabel("job", "teamsacs_master")
	lq.AddLabel("level", "info")
	t.Log(lq.QueryString())
	r, err := LokiSumRate(lq)
	if err != nil {
		t.Log(err)
	}
	t.Log(r)
}
