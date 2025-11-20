// Package main implements radtest, a simple RADIUS protocol testing client.
//
// radtest provides three subcommands for testing RADIUS servers:
//   - auth: Send Access-Request (authentication only)
//   - acct: Send Accounting-Request (start/stop/interim)
//   - flow: Complete authentication + accounting flow (auth → acct-start → acct-stop)
//
// The tool outputs formatted RADIUS packet details using the same formatter
// as the server, making it ideal for debugging RADIUS protocol interactions.
//
// Example usage:
//
//	radtest auth -username test1 -password 111111 -server 127.0.0.1
//	radtest acct -acct-type start -username test1 -session-id abc123
//	radtest flow -username test1 -password 111111 -flow-delay 2s
//
// Protocol conformance:
//   - RFC 2865: RADIUS authentication (Access-Request/Accept/Reject)
//   - RFC 2866: RADIUS accounting (Accounting-Request/Response)
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

// Default configuration values for RADIUS test packets.
// These match common test environment setups and can be overridden via flags.
const (
	defaultServer          = "127.0.0.1"             // Default RADIUS server IP
	defaultSecret          = "testing123"            // Standard test shared secret
	defaultUsername        = "test1"                 // Test username
	defaultPassword        = "111111"                // Test password
	defaultCallingStation  = "AA-BB-CC-DD-EE-FF"     // Simulated client MAC address
	defaultFramedIP        = "172.16.0.10"           // Allocated IP for accounting
	defaultNASIdentifier   = "radtest"               // NAS identification string
	defaultNASIP           = "127.0.0.1"             // NAS IP address
	defaultFlowDelay       = 1500 * time.Millisecond // Delay between flow stages
	defaultTimeout         = 3 * time.Second         // Request timeout
	defaultAcctSessionTime = 60                      // Default session duration (seconds)
)

// command represents the radtest subcommand type.
type command string

// Supported subcommands for RADIUS testing.
const (
	cmdAuth command = "auth" // Authentication-only test (Access-Request)
	cmdAcct command = "acct" // Accounting test (Accounting-Request)
	cmdFlow command = "flow" // Full flow test (auth + acct-start + acct-stop)
)

// options holds all configuration parameters for RADIUS test operations.
// Fields are populated from command-line flags or default values.
type options struct {
	server          string        // RADIUS server IP or hostname
	authPort        int           // Authentication port (default 1812)
	acctPort        int           // Accounting port (default 1813)
	secret          string        // Shared secret for packet authentication
	username        string        // User-Name attribute (RFC 2865)
	password        string        // User-Password for authentication
	callingStation  string        // Calling-Station-Id (typically MAC address)
	framedIP        string        // Framed-IP-Address for accounting
	nasIdentifier   string        // NAS-Identifier attribute
	nasIP           string        // NAS-IP-Address attribute
	nasPort         uint          // NAS-Port attribute
	nasPortType     uint          // NAS-Port-Type attribute (RFC 2865)
	sessionID       string        // Acct-Session-Id (auto-generated if empty)
	timeout         time.Duration // Request timeout
	acctType        string        // Accounting type: "start", "stop", or "interim"
	acctSessionTime int           // Acct-Session-Time in seconds
	inOctets        uint64        // Acct-Input-Octets (bytes received)
	outOctets       uint64        // Acct-Output-Octets (bytes sent)
	inPackets       uint64        // Acct-Input-Packets
	outPackets      uint64        // Acct-Output-Packets
	flowDelay       time.Duration // Delay between flow test stages
}

// defaultOptions creates an options instance with sensible defaults.
// All default values can be overridden via command-line flags.
//
// Returns:
//   - *options: Initialized options with default configuration
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

// parseAuthFlags parses command-line arguments for the "auth" subcommand.
//
// Parameters:
//   - args: Command-line arguments (excluding subcommand)
//
// Returns:
//   - *options: Parsed configuration
//   - error: Parse error or nil on success
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

// parseAcctFlags parses command-line arguments for the "acct" subcommand.
//
// Parameters:
//   - args: Command-line arguments (excluding subcommand)
//
// Returns:
//   - *options: Parsed configuration
//   - error: Parse error or nil on success
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

// parseFlowFlags parses command-line arguments for the "flow" subcommand.
// Flow tests require both authentication and accounting parameters.
//
// Parameters:
//   - args: Command-line arguments (excluding subcommand)
//
// Returns:
//   - *options: Parsed configuration
//   - error: Parse error or nil on success
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

// bindCommonFlags registers common RADIUS attributes as command-line flags.
// These flags are shared across all subcommands (auth, acct, flow).
//
// Parameters:
//   - fs: FlagSet to bind flags to
//   - opts: Options struct to populate from flags
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

// bindAccountingFlags registers accounting-specific flags.
// Used by "acct" and "flow" subcommands for session statistics.
//
// Parameters:
//   - fs: FlagSet to bind flags to
//   - opts: Options struct to populate from flags
func bindAccountingFlags(fs *flag.FlagSet, opts *options) {
	fs.StringVar(&opts.acctType, "acct-type", opts.acctType, "Accounting request type: start|stop|interim")
	fs.IntVar(&opts.acctSessionTime, "session-time", opts.acctSessionTime, "Acct-Session-Time (seconds)")
	fs.Uint64Var(&opts.inOctets, "in-octets", opts.inOctets, "Acct-Input-Octets")
	fs.Uint64Var(&opts.outOctets, "out-octets", opts.outOctets, "Acct-Output-Octets")
	fs.Uint64Var(&opts.inPackets, "in-packets", opts.inPackets, "Acct-Input-Packets")
	fs.Uint64Var(&opts.outPackets, "out-packets", opts.outPackets, "Acct-Output-Packets")
}

// runAuth sends an Access-Request packet and validates the response.
// Implements RADIUS authentication protocol per RFC 2865.
//
// Parameters:
//   - opts: Configuration including username, password, and server address
//
// Returns:
//   - error: nil if Access-Accept received, error otherwise
//
// Output:
//   - Logs formatted request packet before sending
//   - Logs formatted response packet after receiving
//   - Reports authentication success or failure
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

// runAcct sends an Accounting-Request packet and validates the response.
// Implements RADIUS accounting protocol per RFC 2866.
//
// Parameters:
//   - opts: Configuration including acct-type (start/stop/interim) and session data
//
// Returns:
//   - error: nil if Accounting-Response received, error otherwise
//
// Side effects:
//   - Auto-generates session ID if not provided
//   - Logs formatted request and response packets
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

// runFlow executes a complete RADIUS session flow simulation.
// Sequence: Access-Request → Accounting-Start → [delay] → Accounting-Stop
//
// This simulates a real client session lifecycle for integration testing.
//
// Parameters:
//   - opts: Configuration for authentication and accounting
//
// Returns:
//   - error: nil if all stages succeed, error from first failure otherwise
//
// Side effects:
//   - Generates session ID shared across all stages
//   - Sleeps for opts.flowDelay between acct-start and acct-stop
//   - Sets minimum realistic traffic values for acct-stop if not specified
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

// buildAuthPacket constructs an Access-Request packet per RFC 2865.
//
// Includes standard attributes:
//   - User-Name, User-Password (PAP authentication)
//   - NAS attributes (Identifier, IP, Port, Port-Type)
//   - Calling-Station-Id, Called-Station-Id
//   - NAS-Port-Id (fixed to "slot=0/port=1")
//
// Parameters:
//   - opts: Options containing username, password, and NAS configuration
//
// Returns:
//   - *radius.Packet: Constructed packet with random authenticator
//   - error: Attribute encoding error or nil on success
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

// buildAcctPacket constructs an Accounting-Request packet per RFC 2866.
//
// Includes standard accounting attributes:
//   - Acct-Status-Type (Start/Stop/Interim-Update)
//   - Acct-Session-Id, Acct-Session-Time
//   - Acct-Input-Octets, Acct-Output-Octets
//   - Acct-Input-Packets, Acct-Output-Packets
//   - Framed-IP-Address, User-Name
//   - NAS attributes
//
// Parameters:
//   - opts: Configuration including session ID and traffic counters
//   - status: Accounting status type (Start/Stop/Interim)
//
// Returns:
//   - *radius.Packet: Constructed packet with MD5-hashed authenticator
//   - error: Attribute encoding error or nil on success
//
// Side effects:
//   - Auto-generates session ID if opts.sessionID is empty
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

// setCommonNasAttributes adds NAS-related attributes to a RADIUS packet.
// These attributes identify the Network Access Server to the RADIUS server.
//
// Attributes set:
//   - NAS-Identifier: Text name of the NAS
//   - NAS-IP-Address: IPv4 address of the NAS
//   - NAS-Port: Physical port number
//   - NAS-Port-Type: Port type enumeration (RFC 2865)
//
// Parameters:
//   - pkt: RADIUS packet to modify
//   - opts: Options containing NAS configuration
//
// Note: Errors are intentionally ignored as these attributes are optional.
func setCommonNasAttributes(pkt *radius.Packet, opts *options) {
	_ = rfc2865.NASIdentifier_Set(pkt, []byte(opts.nasIdentifier))
	_ = rfc2865.NASIPAddress_Set(pkt, parseIP(opts.nasIP))
	_ = rfc2865.NASPort_Set(pkt, rfc2865.NASPort(uint32(opts.nasPort)))
	_ = rfc2865.NASPortType_Set(pkt, rfc2865.NASPortType(uint32(opts.nasPortType)))
}

// sendPacket transmits a RADIUS packet via UDP and waits for response.
//
// Parameters:
//   - pkt: RADIUS packet to send
//   - addr: Server address in "host:port" format
//   - timeout: Maximum wait time for response
//
// Returns:
//   - *radius.Packet: Response packet from server
//   - error: Network error, timeout, or nil on success
//
// Note: InsecureSkipVerify is enabled for testing purposes.
func sendPacket(pkt *radius.Packet, addr string, timeout time.Duration) (*radius.Packet, error) {
	client := &radius.Client{Retry: 0, InsecureSkipVerify: true}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return client.Exchange(ctx, pkt, addr)
}

// parseAcctStatus converts string accounting type to RFC 2866 status value.
//
// Parameters:
//   - kind: Accounting type string ("start", "stop", or "interim")
//
// Returns:
//   - rfc2866.AcctStatusType: Numeric status type constant
//   - error: Invalid type error or nil on success
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

// ensureSessionID returns an existing session ID or generates a new UUID.
//
// Parameters:
//   - opts: Options that may contain a pre-set session ID
//
// Returns:
//   - string: Session ID (existing or newly generated)
//
// Side effects:
//   - Modifies opts.sessionID if it was empty
func ensureSessionID(opts *options) string {
	if opts.sessionID == "" {
		opts.sessionID = uuid.NewString()
	}
	return opts.sessionID
}

// parseIP parses an IP address string, returning IPv4 zero on failure.
//
// Parameters:
//   - value: IP address string (e.g., "192.168.1.1")
//
// Returns:
//   - net.IP: Parsed IP address or 0.0.0.0 if invalid
func parseIP(value string) net.IP {
	ip := net.ParseIP(value)
	if ip == nil {
		return net.IPv4zero
	}
	return ip
}

// logResponse logs a RADIUS response packet using structured formatting.
// Delegates to logPacket with "<- Response" prefix.
//
// Parameters:
//   - resp: RADIUS response packet to log
func logResponse(resp *radius.Packet) {
	logPacket("<- Response", resp)
}

// logPacket outputs formatted RADIUS packet details to stdout.
// Uses the same formatter as the server for consistency in debugging.
//
// Parameters:
//   - label: Prefix label (e.g., "-> Access-Request", "<- Response")
//   - pkt: RADIUS packet to format and log
//
// Output format includes:
//   - Packet identifier, code, authenticator
//   - All attributes with type names and formatted values
//   - Vendor-Specific attributes with vendor code and type
func logPacket(label string, pkt *radius.Packet) {
	if pkt == nil {
		return
	}
	formatted := strings.TrimRight(radiusd.FmtPacket(pkt), "\n")
	log.Printf("%s\n%s", label, formatted)
}

// printUsage displays command-line help text to stdout.
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

// max returns the larger of two uint64 values.
func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// maxInt returns the larger of two int values.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// clampUint32 safely converts uint64 to uint32, clamping to MaxUint32.
// Prevents overflow when encoding large traffic counters to RADIUS attributes.
//
// Parameters:
//   - v: 64-bit unsigned value
//
// Returns:
//   - uint32: Clamped value fitting in 32 bits
func clampUint32(v uint64) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}

// toUint32FromInt safely converts signed int to uint32 with bounds checking.
//
// Parameters:
//   - v: Signed integer value
//
// Returns:
//   - uint32: Converted value (0 if negative, clamped if > MaxUint32)
func toUint32FromInt(v int) uint32 {
	if v < 0 {
		return 0
	}
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}
