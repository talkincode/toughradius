package main

import (
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/config"
)

// 开发环境初始化数据库
/**

CREATE USER teamsacs WITH PASSWORD 'teamsacs'

CREATE DATABASE teamsacs OWNER postgres;

GRANT ALL PRIVILEGES ON DATABASE teamsacs TO teamsacs;

*/

func main() {
	app.InitGlobalApplication(config.LoadConfig(""))
	app.GApp().InitDb()
}
