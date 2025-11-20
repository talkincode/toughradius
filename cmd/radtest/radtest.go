package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

const (
	defaultServer          = "127.0.0.1"
	defaultSecret          = "testing123"
	defaultUsername        = "test1"
	defaultPassword        = "111111"
	defaultCallingStation  = "AA-BB-CC-DD-EE-FF"
	defaultFramedIP        = "172.16.0.10"
	defaultNASIdentifier   = "radtest"
	defaultNASIP           = "127.0.0.1"
	defaultFlowDelay       = 1500 * time.Millisecond
	defaultTimeout         = 3 * time.Second
	defaultAcctSessionTime = 60
)

type command string

const (
	cmdAuth command = "auth"
	cmdAcct command = "acct"
	cmdFlow command = "flow"
)

type options struct {
	server          string
	authPort        int
	acctPort        int
	secret          string
	username        string
	password        string
	callingStation  string
	framedIP        string
	nasIdentifier   string
	nasIP           string
	nasPort         uint
	nasPortType     uint
	sessionID       string
	timeout         time.Duration
	acctType        string
	acctSessionTime int
	inOctets        uint64
	outOctets       uint64
	inPackets       uint64
	outPackets      uint64
	flowDelay       time.Duration
}

func defaultOptions() *options {
	return &options{
		server:          defaultServer,
		authPort:        1812,
		acctPort:        1813,
		secret:          defaultSecret,
		username:        defaultUsername,
		password:        defaultPassword,
		callingStation:  defaultCallingStation,
		framedIP:        defaultFramedIP,
		nasIdentifier:   defaultNASIdentifier,
		nasIP:           defaultNASIP,
		nasPort:         0,
		nasPortType:     0,
		timeout:         defaultTimeout,
		acctType:        "start",
		acctSessionTime: defaultAcctSessionTime,
		flowDelay:       defaultFlowDelay,
	}
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	var err error
	switch command(subcommand) {
	case cmdAuth:
		var opts *options
		opts, err = parseAuthFlags(args)
		if err == nil {
			err = runAuth(opts)
		}
	case cmdAcct:
		var opts *options
		opts, err = parseAcctFlags(args)
		if err == nil {
			err = runAcct(opts)
		}
	case cmdFlow:
		var opts *options
		opts, err = parseFlowFlags(args)
		if err == nil {
			err = runFlow(opts)
		}
	case "-h", "--help", "help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("%s failed: %v", subcommand, err)
	}
}

func parseAuthFlags(args []string) (*options, error) {
	opts := defaultOptions()
	fs := flag.NewFlagSet(string(cmdAuth), flag.ContinueOnError)
	bindCommonFlags(fs, opts)
	fs.StringVar(&opts.password, "password", opts.password, "User password for authentication")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opts, nil
}

func parseAcctFlags(args []string) (*options, error) {
	opts := defaultOptions()
	fs := flag.NewFlagSet(string(cmdAcct), flag.ContinueOnError)
	bindCommonFlags(fs, opts)
	bindAccountingFlags(fs, opts)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opts, nil
}

func parseFlowFlags(args []string) (*options, error) {
	opts := defaultOptions()
	fs := flag.NewFlagSet(string(cmdFlow), flag.ContinueOnError)
	bindCommonFlags(fs, opts)
	bindAccountingFlags(fs, opts)
	fs.DurationVar(&opts.flowDelay, "flow-delay", opts.flowDelay, "Delay between accounting stages")
	fs.StringVar(&opts.password, "password", opts.password, "User password for authentication")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opts, nil
}

func bindCommonFlags(fs *flag.FlagSet, opts *options) {
	fs.StringVar(&opts.server, "server", opts.server, "RADIUS server host/IP")
	fs.IntVar(&opts.authPort, "auth-port", opts.authPort, "RADIUS authentication port")
	fs.IntVar(&opts.acctPort, "acct-port", opts.acctPort, "RADIUS accounting port")
	fs.StringVar(&opts.secret, "secret", opts.secret, "Shared secret")
	fs.StringVar(&opts.username, "username", opts.username, "User-Name attribute")
	fs.StringVar(&opts.callingStation, "calling-station", opts.callingStation, "Calling-Station-ID (e.g., MAC)")
	fs.StringVar(&opts.framedIP, "framed-ip", opts.framedIP, "Framed-IP-Address for accounting")
	fs.StringVar(&opts.nasIdentifier, "nas-id", opts.nasIdentifier, "NAS-Identifier attribute")
	fs.StringVar(&opts.nasIP, "nas-ip", opts.nasIP, "NAS-IP-Address attribute")
	fs.UintVar(&opts.nasPort, "nas-port", opts.nasPort, "NAS-Port value")
	fs.UintVar(&opts.nasPortType, "nas-port-type", opts.nasPortType, "NAS-Port-Type value")
	fs.StringVar(&opts.sessionID, "session-id", opts.sessionID, "Acct-Session-Id (auto if empty)")
	fs.DurationVar(&opts.timeout, "timeout", opts.timeout, "Request timeout")
}

func bindAccountingFlags(fs *flag.FlagSet, opts *options) {
	fs.StringVar(&opts.acctType, "acct-type", opts.acctType, "Accounting request type: start|stop|interim")
	fs.IntVar(&opts.acctSessionTime, "session-time", opts.acctSessionTime, "Acct-Session-Time (seconds)")
	fs.Uint64Var(&opts.inOctets, "in-octets", opts.inOctets, "Acct-Input-Octets")
	fs.Uint64Var(&opts.outOctets, "out-octets", opts.outOctets, "Acct-Output-Octets")
	fs.Uint64Var(&opts.inPackets, "in-packets", opts.inPackets, "Acct-Input-Packets")
	fs.Uint64Var(&opts.outPackets, "out-packets", opts.outPackets, "Acct-Output-Packets")
}

func runAuth(opts *options) error {
	pkt, err := buildAuthPacket(opts)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", opts.server, opts.authPort)
	log.Printf("Sending Access-Request to %s (user=%s)", addr, opts.username)
	logPacket("-> Access-Request", pkt)
	resp, err := sendPacket(pkt, addr, opts.timeout)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("no response from server")
	}
	logResponse(resp)
	if resp.Code != radius.CodeAccessAccept {
		return fmt.Errorf("authentication failed: %s", resp.Code)
	}
	return nil
}

func runAcct(opts *options) error {
	status, err := parseAcctStatus(opts.acctType)
	if err != nil {
		return err
	}
	sessionID := ensureSessionID(opts)
	pkt, err := buildAcctPacket(opts, status)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", opts.server, opts.acctPort)
	log.Printf("Sending Accounting-Request (%s) to %s (session=%s)", strings.ToUpper(opts.acctType), addr, sessionID)
	logPacket(fmt.Sprintf("-> Accounting-Request (%s)", strings.ToUpper(opts.acctType)), pkt)
	resp, err := sendPacket(pkt, addr, opts.timeout)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("no response from server")
	}
	logResponse(resp)
	if resp.Code != radius.CodeAccountingResponse {
		return fmt.Errorf("unexpected response code: %s", resp.Code)
	}
	return nil
}

func runFlow(opts *options) error {
	if err := runAuth(opts); err != nil {
		return fmt.Errorf("auth stage failed: %w", err)
	}
	sessionID := opts.sessionID
	if sessionID == "" {
		sessionID = uuid.NewString()
		log.Printf("Generated session id: %s", sessionID)
	}
	startOpts := *opts
	startOpts.acctType = "start"
	startOpts.sessionID = sessionID
	if err := runAcct(&startOpts); err != nil {
		return fmt.Errorf("accounting-start failed: %w", err)
	}
	if opts.flowDelay > 0 {
		log.Printf("Waiting %s before accounting-stop", opts.flowDelay)
		time.Sleep(opts.flowDelay)
	}
	stopOpts := *opts
	stopOpts.acctType = "stop"
	stopOpts.sessionID = sessionID
	stopOpts.inOctets = max(opts.inOctets, 1024)
	stopOpts.outOctets = max(opts.outOctets, 4096)
	stopOpts.inPackets = max(opts.inPackets, 32)
	stopOpts.outPackets = max(opts.outPackets, 64)
	stopOpts.acctSessionTime = maxInt(opts.acctSessionTime, defaultAcctSessionTime)
	return runAcct(&stopOpts)
}

func buildAuthPacket(opts *options) (*radius.Packet, error) {
	pkt := radius.New(radius.CodeAccessRequest, []byte(opts.secret))
	if err := rfc2865.UserName_SetString(pkt, opts.username); err != nil {
		return nil, err
	}
	if err := rfc2865.UserPassword_SetString(pkt, opts.password); err != nil {
		return nil, err
	}
	setCommonNasAttributes(pkt, opts)
	if err := rfc2869.NASPortID_Set(pkt, []byte("slot=0/port=1")); err != nil {
		return nil, err
	}
	if err := rfc2865.CalledStationID_SetString(pkt, "11:11:11:11:11:11"); err != nil {
		return nil, err
	}
	if err := rfc2865.CallingStationID_SetString(pkt, opts.callingStation); err != nil {
		return nil, err
	}
	return pkt, nil
}

func buildAcctPacket(opts *options, status rfc2866.AcctStatusType) (*radius.Packet, error) {
	sessionID := opts.sessionID
	if sessionID == "" {
		sessionID = uuid.NewString()
	}
	pkt := radius.New(radius.CodeAccountingRequest, []byte(opts.secret))
	if err := rfc2865.UserName_SetString(pkt, opts.username); err != nil {
		return nil, err
	}
	setCommonNasAttributes(pkt, opts)
	if err := rfc2869.NASPortID_Set(pkt, []byte("slot=0/port=1")); err != nil {
		return nil, err
	}
	if err := rfc2865.CalledStationID_SetString(pkt, "11:11:11:11:11:11"); err != nil {
		return nil, err
	}
	if err := rfc2865.CallingStationID_SetString(pkt, opts.callingStation); err != nil {
		return nil, err
	}
	if err := rfc2866.AcctSessionID_SetString(pkt, sessionID); err != nil {
		return nil, err
	}
	if err := rfc2866.AcctStatusType_Set(pkt, status); err != nil {
		return nil, err
	}
	if err := rfc2866.AcctSessionTime_Set(pkt, rfc2866.AcctSessionTime(toUint32FromInt(opts.acctSessionTime))); err != nil {
		return nil, err
	}
	if err := rfc2866.AcctInputOctets_Set(pkt, rfc2866.AcctInputOctets(clampUint32(opts.inOctets))); err != nil {
		return nil, err
	}
	if err := rfc2866.AcctOutputOctets_Set(pkt, rfc2866.AcctOutputOctets(clampUint32(opts.outOctets))); err != nil {
		return nil, err
	}
	if err := rfc2866.AcctInputPackets_Set(pkt, rfc2866.AcctInputPackets(clampUint32(opts.inPackets))); err != nil {
		return nil, err
	}
	if err := rfc2866.AcctOutputPackets_Set(pkt, rfc2866.AcctOutputPackets(clampUint32(opts.outPackets))); err != nil {
		return nil, err
	}
	if err := rfc2865.FramedIPAddress_Set(pkt, parseIP(opts.framedIP)); err != nil {
		return nil, err
	}
	return pkt, nil
}

func setCommonNasAttributes(pkt *radius.Packet, opts *options) {
	_ = rfc2865.NASIdentifier_Set(pkt, []byte(opts.nasIdentifier))
	_ = rfc2865.NASIPAddress_Set(pkt, parseIP(opts.nasIP))
	_ = rfc2865.NASPort_Set(pkt, rfc2865.NASPort(uint32(opts.nasPort)))
	_ = rfc2865.NASPortType_Set(pkt, rfc2865.NASPortType(uint32(opts.nasPortType)))
}

func sendPacket(pkt *radius.Packet, addr string, timeout time.Duration) (*radius.Packet, error) {
	client := &radius.Client{Retry: 0, InsecureSkipVerify: true}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return client.Exchange(ctx, pkt, addr)
}

func parseAcctStatus(kind string) (rfc2866.AcctStatusType, error) {
	switch strings.ToLower(kind) {
	case "start":
		return rfc2866.AcctStatusType_Value_Start, nil
	case "stop":
		return rfc2866.AcctStatusType_Value_Stop, nil
	case "interim":
		return rfc2866.AcctStatusType_Value_InterimUpdate, nil
	default:
		return 0, fmt.Errorf("invalid acct-type %q", kind)
	}
}

func ensureSessionID(opts *options) string {
	if opts.sessionID == "" {
		opts.sessionID = uuid.NewString()
	}
	return opts.sessionID
}

func parseIP(value string) net.IP {
	ip := net.ParseIP(value)
	if ip == nil {
		return net.IPv4zero
	}
	return ip
}

func logResponse(resp *radius.Packet) {
	logPacket("<- Response", resp)
}

func logPacket(label string, pkt *radius.Packet) {
	if pkt == nil {
		return
	}
	formatted := strings.TrimRight(radiusd.FmtPacket(pkt), "\n")
	log.Printf("%s\n%s", label, formatted)
}

func printUsage() {
	fmt.Println("radtest - simple RADIUS client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  radtest auth [flags]      # Send Access-Request")
	fmt.Println("  radtest acct [flags]      # Send Accounting-Request (start/stop/interim)")
	fmt.Println("  radtest flow [flags]      # Run auth + acct-start + acct-stop")
	fmt.Println()
	fmt.Println("Common flags:")
	fmt.Println("  -server 127.0.0.1      RADIUS server host")
	fmt.Println("  -secret testing123     Shared secret")
	fmt.Println("  -username test1        User-Name attribute")
	fmt.Println("  -password 111111       User-Password (auth/flow)")
	fmt.Println("  -calling-station MAC   Calling-Station-ID")
	fmt.Println("  -framed-ip 172.16.0.10 Framed-IP-Address (accounting)")
	fmt.Println("  -session-id <id>       Custom session id (optional)")
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampUint32(v uint64) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}

func toUint32FromInt(v int) uint32 {
	if v < 0 {
		return 0
	}
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}
