package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	_ "time/tzdata"

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/adminapi"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/freeradius"
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group
var Version = "develop"

var (
	h        = flag.Bool("h", false, "help usage")
	showVer  = flag.Bool("v", false, "show version")
	conffile = flag.String("c", "", "config yaml file")
	initdb   = flag.Bool("initdb", false, "run initdb")
	printcfg = flag.Bool("printcfg", false, "print config")
)

func PrintVersion() {
	fmt.Fprintf(os.Stdout, "version: %s\n", Version)
}

func printHelp() {
	if *h {
		flag.PrintDefaults()
		os.Exit(0)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if *showVer {
		PrintVersion()
		os.Exit(0)
	}

	printHelp()

	_config := config.LoadConfig(*conffile)

	if *printcfg {
		fmt.Printf("%+v\n", common.ToJson(_config))
		return
	}

	if *initdb {
		app.InitGlobalApplication(_config)
		app.GApp().InitDb()
		return
	}

	app.InitGlobalApplication(_config)
	app.GApp().MigrateDB(false)
	defer app.Release()

	g.Go(func() error {
		webserver.Init()
		adminapi.Init()
		return webserver.Listen()
	})

	g.Go(func() error {
		return freeradius.Listen()
	})

	radiusService := radiusd.NewRadiusService()
	defer radiusService.Release()

	g.Go(func() error {
		return radiusd.ListenRadiusAuthServer(radiusd.NewAuthService(radiusService))
	})

	g.Go(func() error {
		return radiusd.ListenRadiusAcctServer(radiusd.NewAcctService(radiusService))
	})

	g.Go(func() error {
		radsec := radiusd.NewRadsecService(
			radiusd.NewAuthService(radiusService),
			radiusd.NewAcctService(radiusService),
		)
		return radiusd.ListenRadsecServer(radsec)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
