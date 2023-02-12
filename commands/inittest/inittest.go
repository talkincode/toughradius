package main

import (
	"github.com/talkincode/toughradius/app"
	"github.com/talkincode/toughradius/config"
)

func main() {
	app.InitGlobalApplication(config.LoadConfig("../toughradius.yml"))
	app.GApp().InitTest()
}
