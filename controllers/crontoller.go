package controllers

import (
	"github.com/talkincode/toughradius/controllers/apitoken"
	"github.com/talkincode/toughradius/controllers/charts"
	"github.com/talkincode/toughradius/controllers/cpe"
	"github.com/talkincode/toughradius/controllers/cwmpconfig"
	"github.com/talkincode/toughradius/controllers/cwmppreset"
	"github.com/talkincode/toughradius/controllers/dashboard"
	"github.com/talkincode/toughradius/controllers/factoryreset"
	"github.com/talkincode/toughradius/controllers/firmwareconfig"
	"github.com/talkincode/toughradius/controllers/index"
	"github.com/talkincode/toughradius/controllers/logs"
	"github.com/talkincode/toughradius/controllers/metrics"
	"github.com/talkincode/toughradius/controllers/node"
	"github.com/talkincode/toughradius/controllers/opr"
	"github.com/talkincode/toughradius/controllers/radius"
	"github.com/talkincode/toughradius/controllers/settings"
	"github.com/talkincode/toughradius/controllers/supervise"
	"github.com/talkincode/toughradius/controllers/translate"
	"github.com/talkincode/toughradius/controllers/vpe"
)

// Init web 控制器初始化
func Init() {
	index.InitRouter()
	opr.InitRouter()
	settings.InitRouter()
	dashboard.InitRouter()
	charts.InitRouter()
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
