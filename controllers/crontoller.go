package controllers

import (
	"github.com/talkincode/toughradius/v8/controllers/apitoken"
	"github.com/talkincode/toughradius/v8/controllers/cpe"
	"github.com/talkincode/toughradius/v8/controllers/cwmpconfig"
	"github.com/talkincode/toughradius/v8/controllers/cwmppreset"
	"github.com/talkincode/toughradius/v8/controllers/dashboard"
	"github.com/talkincode/toughradius/v8/controllers/factoryreset"
	"github.com/talkincode/toughradius/v8/controllers/firmwareconfig"
	"github.com/talkincode/toughradius/v8/controllers/index"
	"github.com/talkincode/toughradius/v8/controllers/logs"
	"github.com/talkincode/toughradius/v8/controllers/metrics"
	"github.com/talkincode/toughradius/v8/controllers/node"
	"github.com/talkincode/toughradius/v8/controllers/opr"
	"github.com/talkincode/toughradius/v8/controllers/radius"
	"github.com/talkincode/toughradius/v8/controllers/settings"
	"github.com/talkincode/toughradius/v8/controllers/supervise"
	"github.com/talkincode/toughradius/v8/controllers/translate"
	"github.com/talkincode/toughradius/v8/controllers/vpe"
)

// Init web 控制器初始化
func Init() {
	index.InitRouter()
	opr.InitRouter()
	settings.InitRouter()
	dashboard.InitRouter()
	// charts.InitRouter()
	vpe.InitRouter()
	cpe.InitRouter()
	logs.InitRouter()
	radius.InitRouter()
	node.InitRouter()
	factoryreset.InitRouter()
	firmwareconfig.InitRouter()
	cwmpconfig.InitRouter()
	supervise.InitRouter()
	apitoken.InitRouter()
	cwmppreset.InitRouter()
	metrics.InitRouter()
	translate.InitRouter()
}
