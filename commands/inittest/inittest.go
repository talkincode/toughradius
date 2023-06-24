package main

import (
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/config"
)

func main() {
	app.InitGlobalApplication(config.LoadConfig("../toughradius.yml"))
	app.GApp().InitTest()
}
