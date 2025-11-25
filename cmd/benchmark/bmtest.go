// Package main provides a command-line tool for benchmarking RADIUS server performance.
//
// This tool tests RADIUS authentication and accounting services under various load conditions,
// providing detailed metrics on throughput, latency, and error rates.
//
// Usage:
//
//	bmtest -b -server 127.0.0.1 -n 10000 -c 100  # Run benchmark test
//	bmtest -auth -u testuser -p password          # Test single authentication
//	bmtest -acct-start -u testuser                 # Test accounting-start
//	bmtest -config bm.yml                   # Load config from YAML file
//
// The tool supports:
//   - Concurrent authentication and accounting testing
//   - Real-time statistics reporting
//   - CSV output for post-analysis
//   - YAML configuration files
//   - Single request testing for debugging
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	bm "github.com/talkincode/toughradius/v9/internal/benchmark"
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"layeh.com/radius"
	"layeh.com/radius/rfc2866"
)

// Command-line flags
var (
	// General flags
	h      = flag.Bool("h", false, "show help message")
	config = flag.String("config", "", "path to YAML configuration file")

	// Server configuration
	server   = flag.String("server", "127.0.0.1", "RADIUS server address")
	authport = flag.Int("authport", 1812, "authentication port")
	acctport = flag.Int("acctport", 1813, "accounting port")
	secret   = flag.String("s", "secret", "RADIUS shared secret")
	timeout  = flag.Int("t", 10, "request timeout in seconds")

	// NAS configuration
	nasip = flag.String("nasip", "127.0.0.1", "NAS-IP-Address")

	// User configuration
	datafile = flag.String("d", "bmdata.json", "test data file (JSON)")
	user     = flag.String("u", "test01", "test username")
	passwd   = flag.String("p", "111111", "test password")
	userip   = flag.String("ip", "127.0.0.1", "user IP address")
	usermac  = flag.String("mac", "11:11:11:11:11:11", "user MAC address")

	// Load testing configuration
	total       = flag.Int64("n", 100, "total number of transactions")
	concurrency = flag.Int64("c", 10, "concurrent workers")
	interval    = flag.Int("i", 5, "statistics reporting interval (seconds)")

	// Output configuration
	output = flag.String("o", "bm.csv", "CSV output file")

	// Test modes
	benchmarkMode = flag.Bool("b", false, "run benchmark test")
	auth          = flag.Bool("auth", false, "send single authentication request")
	acctstart     = flag.Bool("acct-start", false, "send accounting-start request")
	acctupdate    = flag.Bool("acct-update", false, "send accounting-update request")
	acctstop      = flag.Bool("acct-stop", false, "send accounting-stop request")
)

func main() {
	flag.Parse()

	if *h {
		printUsage()
		os.Exit(0)
	}

	// Print system information
	printSystemInfo()

	// Determine operation mode
	if *auth {
		runAuthTest()
	} else if *acctstart || *acctupdate || *acctstop {
		runAcctTest()
	} else if *benchmarkMode {
		runBenchmark()
	} else {
		fmt.Println("No operation mode specified. Use -h for help.")
		os.Exit(1)
	}
}

// printUsage displays help information.
func printUsage() {
	fmt.Fprintf(os.Stderr, "ToughRADIUS Benchmark Tool v2.0\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  bmtest [options]\n\n")
	fmt.Fprintf(os.Stderr, "Test Modes:\n")
	fmt.Fprintf(os.Stderr, "  -b              Run full benchmark test (auth + accounting cycle)\n")
	fmt.Fprintf(os.Stderr, "  -auth           Send a single Access-Request for testing\n")
	fmt.Fprintf(os.Stderr, "  -acct-start     Send a single Accounting-Start request\n")
	fmt.Fprintf(os.Stderr, "  -acct-update    Send a single Accounting-Update request\n")
	fmt.Fprintf(os.Stderr, "  -acct-stop      Send a single Accounting-Stop request\n\n")
	fmt.Fprintf(os.Stderr, "Configuration:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  bmtest -b -n 10000 -c 100 -server 192.168.1.100\n")
	fmt.Fprintf(os.Stderr, "  bmtest -auth -u testuser -p testpass\n")
	fmt.Fprintf(os.Stderr, "  bmtest -config bm.yml -b\n")
}

// printSystemInfo displays system resource information.
func printSystemInfo() {
	fmt.Println()

	if hinfo, err := host.Info(); err == nil {
		fmt.Printf("System: %s %s %s %s\n", hinfo.OS, hinfo.KernelArch, hinfo.KernelVersion, hinfo.Platform)
	}

	if meminfo, err := mem.VirtualMemory(); err == nil {
		fmt.Printf("Memory: Total %d MB, Available %d MB\n", meminfo.Total/1048576, meminfo.Available/1048576)
	}

	if cinfo, err := cpu.Info(); err == nil && len(cinfo) > 0 {
		c := cinfo[0]
		fmt.Printf("CPU: %s, Cores: %d, Cache: %d KB\n", c.ModelName, c.Cores, c.CacheSize)
	}

	fmt.Println()
}

// runBenchmark executes the full benchmark test.
func runBenchmark() {
	cfg := buildConfigFromFlags()

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	task, err := bm.NewTask(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create benchmark task: %v\n", err)
		os.Exit(1)
	}

	if err := task.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}
}

// runAuthTest sends a single authentication request.
func runAuthTest() {
	fmt.Printf("Sending Access-Request to %s:%d\n", *server, *authport)
	fmt.Printf("Username: %s, Password: %s, Secret: %s\n", *user, *passwd, *secret)

	pb := bm.NewPacketBuilder(*secret, "bmtest", *nasip)
	packet, err := pb.BuildAuthRequest(*user, *passwd, *usermac)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build auth request: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Request Packet ---")
	fmt.Println(radiusd.FmtPacket(packet))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", *server, *authport)
	resp, err := radius.Exchange(ctx, packet, addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Response Packet ---")
	fmt.Println(radiusd.FmtPacket(resp))
}

// runAcctTest sends a single accounting request.
func runAcctTest() {
	var acctType rfc2866.AcctStatusType
	var typeName string

	if *acctstart {
		acctType = rfc2866.AcctStatusType_Value_Start
		typeName = "Accounting-Start"
	} else if *acctupdate {
		acctType = rfc2866.AcctStatusType_Value_InterimUpdate
		typeName = "Accounting-Interim-Update"
	} else if *acctstop {
		acctType = rfc2866.AcctStatusType_Value_Stop
		typeName = "Accounting-Stop"
	}

	fmt.Printf("Sending %s to %s:%d\n", typeName, *server, *acctport)
	fmt.Printf("Username: %s, Secret: %s\n", *user, *secret)

	pb := bm.NewPacketBuilder(*secret, "bmtest", *nasip)
	packet, err := pb.BuildAcctRequest(*user, *userip, *usermac, "test-session-id", acctType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build acct request: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Request Packet ---")
	fmt.Println(radiusd.FmtPacket(packet))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", *server, *acctport)
	resp, err := radius.Exchange(ctx, packet, addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Response Packet ---")
	fmt.Println(radiusd.FmtPacket(resp))
}

// buildConfigFromFlags creates a benchmark configuration from command-line flags.
func buildConfigFromFlags() *bm.Config {
	var cfg *bm.Config

	// Load from file if specified
	if *config != "" {
		var err error
		cfg, err = bm.LoadFromFile(*config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config file: %v\n", err)
			fmt.Println("Using default configuration with command-line overrides...")
			cfg = bm.NewDefaultConfig()
		}
	} else {
		cfg = bm.NewDefaultConfig()
	}

	// Override with command-line flags (flags take precedence)
	if flag.Lookup("server").Value.String() != flag.Lookup("server").DefValue {
		cfg.Server.Address = *server
	}
	if flag.Lookup("authport").Value.String() != flag.Lookup("authport").DefValue {
		cfg.Server.AuthPort = *authport
	}
	if flag.Lookup("acctport").Value.String() != flag.Lookup("acctport").DefValue {
		cfg.Server.AcctPort = *acctport
	}
	if flag.Lookup("s").Value.String() != flag.Lookup("s").DefValue {
		cfg.Server.Secret = *secret
	}
	if flag.Lookup("t").Value.String() != flag.Lookup("t").DefValue {
		cfg.Server.Timeout = *timeout
	}
	if flag.Lookup("nasip").Value.String() != flag.Lookup("nasip").DefValue {
		cfg.NAS.IP = *nasip
	}
	if flag.Lookup("d").Value.String() != flag.Lookup("d").DefValue {
		cfg.User.DataFile = *datafile
	}
	if flag.Lookup("u").Value.String() != flag.Lookup("u").DefValue {
		cfg.User.Username = *user
	}
	if flag.Lookup("p").Value.String() != flag.Lookup("p").DefValue {
		cfg.User.Password = *passwd
	}
	if flag.Lookup("ip").Value.String() != flag.Lookup("ip").DefValue {
		cfg.User.IP = *userip
	}
	if flag.Lookup("mac").Value.String() != flag.Lookup("mac").DefValue {
		cfg.User.MAC = *usermac
	}
	if flag.Lookup("n").Value.String() != flag.Lookup("n").DefValue {
		cfg.Load.Total = *total
	}
	if flag.Lookup("c").Value.String() != flag.Lookup("c").DefValue {
		cfg.Load.Concurrency = *concurrency
	}
	if flag.Lookup("i").Value.String() != flag.Lookup("i").DefValue {
		cfg.Load.Interval = *interval
	}
	if flag.Lookup("o").Value.String() != flag.Lookup("o").DefValue {
		cfg.Output.CSVFile = *output
	}

	return cfg
}
