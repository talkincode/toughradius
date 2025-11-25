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
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"golang.org/x/sync/errgroup"

	// Import vendor parsers for auto-registration via init()
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins"
	_ "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers/parsers"
)

var g errgroup.Group

// Build information, injected via ldflags at compile time
// Example: go build -ldflags "-X main.version=1.0.0 -X main.buildTime=2024-01-01T00:00:00Z -X main.gitCommit=abc123"
var (
	version   = "develop"
	buildTime = "unknown"
	gitCommit = "unknown"
)

var (
	h        = flag.Bool("h", false, "help usage")
	showVer  = flag.Bool("v", false, "show version")
	conffile = flag.String("c", "", "config yaml file")
	initdb   = flag.Bool("initdb", false, "run initdb")
	printcfg = flag.Bool("printcfg", false, "print config")
)

func PrintVersion() {
	fmt.Fprintf(os.Stdout, "ToughRADIUS %s\n", version)
	fmt.Fprintf(os.Stdout, "Build Time: %s\n", buildTime)
	fmt.Fprintf(os.Stdout, "Git Commit: %s\n", gitCommit)
	fmt.Fprintf(os.Stdout, "Go Version: %s\n", runtime.Version())
	fmt.Fprintf(os.Stdout, "OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
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

	// Create and initialize application context
	application := app.NewApplication(_config)
	application.Init(_config)

	if *initdb {
		application.InitDb()
		return
	}
	defer application.Release()

	// Initialize web server and admin API with dependency injection
	g.Go(func() error {
		webserver.Init(application)
		adminapi.Init(application)
		return webserver.Listen(application)
	})

	// Initialize RADIUS service with dependency injection
	radiusService := radiusd.NewRadiusService(application)
	defer radiusService.Release()

	// Initialize plugin system after RadiusService is created
	plugins.InitPlugins(application, radiusService.SessionRepo, radiusService.AccountingRepo)

	// Start RADIUS Auth server
	g.Go(func() error {
		return radiusd.ListenRadiusAuthServer(application, radiusd.NewAuthService(radiusService))
	})

	// Start RADIUS Acct server
	g.Go(func() error {
		return radiusd.ListenRadiusAcctServer(application, radiusd.NewAcctService(radiusService))
	})

	// Start RadSec server
	g.Go(func() error {
		radsec := radiusd.NewRadsecService(
			radiusd.NewAuthService(radiusService),
			radiusd.NewAcctService(radiusService),
		)
		return radiusd.ListenRadsecServer(application, radsec)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
