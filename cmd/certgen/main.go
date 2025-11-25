// Package main provides a command-line tool for generating TLS/SSL certificates
// required by ToughRADIUS for RadSec (RADIUS over TLS) and other TLS scenarios.
//
// This tool supports generating CA root certificates, server certificates, and
// client certificates with full SAN (Subject Alternative Name) extension support.
//
// Key Features:
//   - Generate CA, server, and client certificates
//   - Support for multiple DNS names and IP addresses (SAN)
//   - Customizable organization information and validity period
//   - One-command generation of complete certificate chain
//
// Usage Examples:
//
//	# Generate all certificates (CA + server + client)
//	certgen -type all -output ./certs
//
//	# Generate only CA certificate
//	certgen -type ca -ca-cn "My Company CA"
//
//	# Generate server certificate with custom DNS/IP
//	certgen -type server \
//	  -server-cn radius.example.com \
//	  -server-dns "radius.example.com,*.example.com" \
//	  -server-ips "192.168.1.100,10.0.0.1"
//
// For detailed usage and examples, run:
//
//	certgen -h
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/talkincode/toughradius/v9/pkg/certgen"
)

// version is the current version of the certificate generator tool.
// This is displayed when using the -version flag.
const version = "1.0.0"

// main is the entry point of the certificate generator.
// It parses command-line flags, constructs certificate configurations,
// and routes to the appropriate generation function based on the -type parameter.
//
// The function supports four certificate generation modes:
//   - "all": Generate complete certificate chain (CA + server + client)
//   - "ca": Generate only the CA root certificate
//   - "server": Generate server certificate (requires existing CA)
//   - "client": Generate client certificate (requires existing CA)
//
// Exit codes:
//   - 0: Success or version information displayed
//   - 1: Error occurred (invalid parameters or generation failure)
//
// The function uses log.Fatalf for error handling, which immediately
// terminates the program with exit code 1 and prints the error message.
func main() {
	var (
		// Common parameters
		certType    = flag.String("type", "all", "Certificate type: ca, server, client, all")
		outputDir   = flag.String("output", "./certs", "Output directory")
		validDays   = flag.Int("days", 3650, "Certificate validity (days)")
		keySize     = flag.Int("keysize", 2048, "RSA key size")
		showVersion = flag.Bool("version", false, "Show version info")

		// CA options
		caCommonName = flag.String("ca-cn", "ToughRADIUS CA", "CA certificate CommonName")

		// Server certificate parameters
		serverCommonName = flag.String("server-cn", "radius.example.com", "Server certificate CommonName")
		serverDNS        = flag.String("server-dns", "radius.example.com,*.radius.example.com,localhost", "Server certificate DNS names (comma-separated)")
		serverIPs        = flag.String("server-ips", "127.0.0.1", "Server certificate IP addresses (comma-separated)")

		// Client certificate parameters
		clientCommonName = flag.String("client-cn", "radius-client", "Client certificate CommonName")
		clientDNS        = flag.String("client-dns", "", "Client certificate DNS names (comma-separated)")
		clientIPs        = flag.String("client-ips", "", "Client certificate IP addresses (comma-separated)")

		// Organization information
		organization = flag.String("org", "ToughRADIUS", "Organization name")
		orgUnit      = flag.String("ou", "IT", "Organizational unit")
		country      = flag.String("country", "CN", "Country code")
		province     = flag.String("province", "Shanghai", "Province/State")
		locality     = flag.String("locality", "Shanghai", "City")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ToughRADIUS Certificate Generator v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate all certificates (CA + server + client)\n")
		fmt.Fprintf(os.Stderr, "  %s -type all\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate only the CA certificate\n")
		fmt.Fprintf(os.Stderr, "  %s -type ca -ca-cn \"My CA\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate the server certificate (requires CA cert)\n")
		fmt.Fprintf(os.Stderr, "  %s -type server -server-cn radius.mycompany.com -server-dns \"radius.mycompany.com,*.radius.mycompany.com\" -server-ips \"192.168.1.100,10.0.0.1\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate the client certificate (requires CA cert)\n")
		fmt.Fprintf(os.Stderr, "  %s -type client -client-cn my-radius-client\n\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("ToughRADIUS Certificate Generator v%s\n", version)
		os.Exit(0)
	}

	// Create base configuration
	baseConfig := certgen.CertConfig{
		Organization:       []string{*organization},
		OrganizationalUnit: []string{*orgUnit},
		Country:            []string{*country},
		Province:           []string{*province},
		Locality:           []string{*locality},
		ValidDays:          *validDays,
		KeySize:            *keySize,
	}

	switch *certType {
	case "all":
		if err := generateAll(baseConfig, *outputDir, *caCommonName, *serverCommonName, *serverDNS, *serverIPs, *clientCommonName, *clientDNS, *clientIPs); err != nil {
			log.Fatalf("failed to generate certificates: %v", err)
		}
	case "ca":
		if err := generateCA(baseConfig, *outputDir, *caCommonName); err != nil {
			log.Fatalf("failed to generate CA certificate: %v", err)
		}
	case "server":
		if err := generateServer(baseConfig, *outputDir, *serverCommonName, *serverDNS, *serverIPs); err != nil {
			log.Fatalf("failed to generate server certificate: %v", err)
		}
	case "client":
		if err := generateClient(baseConfig, *outputDir, *clientCommonName, *clientDNS, *clientIPs); err != nil {
			log.Fatalf("failed to generate client certificate: %v", err)
		}
	default:
		log.Fatalf("unknown certificate type: %s (supported: ca, server, client, all)", *certType)
	}

	fmt.Printf("\n✓ Certificate generation complete! Output directory: %s\n", *outputDir)
}

// generateAll generates a complete certificate chain including CA, server, and client certificates.
// This is the recommended approach for initial deployment as it creates all necessary
// certificates in a single operation.
//
// The function executes the following sequence:
//  1. Generate CA root certificate (self-signed)
//  2. Generate server certificate (signed by CA)
//  3. Generate client certificate (signed by CA)
//
// If any step fails, the function returns immediately with a wrapped error,
// preventing incomplete certificate chains.
//
// Parameters:
//   - baseConfig: Base certificate configuration containing organization info, validity period, and key size
//   - outputDir: Directory where certificates will be saved (created if doesn't exist)
//   - caCN: CommonName for the CA certificate (e.g., "ToughRADIUS CA")
//   - serverCN: CommonName for the server certificate (e.g., "radius.example.com")
//   - serverDNS: Comma-separated DNS names for server SAN (e.g., "radius.example.com,*.radius.example.com")
//   - serverIPs: Comma-separated IP addresses for server SAN (e.g., "192.168.1.100,127.0.0.1")
//   - clientCN: CommonName for the client certificate (e.g., "radius-client")
//   - clientDNS: Comma-separated DNS names for client SAN (usually empty for RADIUS clients)
//   - clientIPs: Comma-separated IP addresses for client SAN (usually empty for RADIUS clients)
//
// Returns:
//   - error: Returns wrapped error if any generation step fails, nil on success
//
// Generated files in outputDir:
//   - ca.crt, ca.key: CA certificate and private key
//   - server.crt, server.key: Server certificate and private key
//   - client.crt, client.key: Client certificate and private key
//
// Concurrency: This function is not thread-safe. Certificate generation must
// be sequential as server and client certificates depend on the CA certificate.
//
// Example:
//
//	config := certgen.CertConfig{
//	    Organization: []string{"MyCompany"},
//	    ValidDays: 365,
//	    KeySize: 2048,
//	}
//	err := generateAll(config, "./certs", "My CA", "radius.example.com",
//	    "radius.example.com,localhost", "127.0.0.1",
//	    "nas-client", "", "")
func generateAll(baseConfig certgen.CertConfig, outputDir, caCN, serverCN, serverDNS, serverIPs, clientCN, clientDNS, clientIPs string) error {
	fmt.Println("=== Starting to generate all certificates ===")

	// 1. Generate the CA certificate
	if err := generateCA(baseConfig, outputDir, caCN); err != nil {
		return fmt.Errorf("failed to generate CA: %w", err)
	}
	fmt.Println()

	// 2. Generate the server certificate
	if err := generateServer(baseConfig, outputDir, serverCN, serverDNS, serverIPs); err != nil {
		return fmt.Errorf("failed to generate server certificate: %w", err)
	}
	fmt.Println()

	// 3. Generate the client certificate
	if err := generateClient(baseConfig, outputDir, clientCN, clientDNS, clientIPs); err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	return nil
}

// generateCA generates a self-signed CA (Certificate Authority) root certificate.
// This certificate is used to sign server and client certificates, forming a
// complete PKI (Public Key Infrastructure) chain.
//
// The generated CA certificate has both ServerAuth and ClientAuth extended key usages,
// allowing it to sign both server and client certificates. The CA private key is
// saved with 0600 permissions (owner read/write only) for security.
//
// Parameters:
//   - baseConfig: Base certificate configuration (organization info, validity, key size)
//   - outputDir: Directory where CA certificate and key will be saved
//   - commonName: The CommonName field for the CA certificate (e.g., "ToughRADIUS CA")
//
// Returns:
//   - error: Returns error if directory creation, key generation, or certificate
//     creation fails, nil on success
//
// Generated files:
//   - <outputDir>/ca.crt: CA certificate in PEM format
//   - <outputDir>/ca.key: CA private key in PKCS#8 PEM format (mode 0600)
//
// Security considerations:
//   - The CA private key should be backed up and stored securely offline
//   - If the CA key is compromised, the entire certificate chain becomes untrusted
//   - Consider removing ca.key from production servers after generating certificates
//
// Side effects:
//   - Creates outputDir if it doesn't exist
//   - Prints certificate and key paths to stdout
//
// Example:
//
//	config := certgen.CertConfig{
//	    Organization: []string{"MyCompany"},
//	    ValidDays: 3650,  // 10 years
//	    KeySize: 4096,    // High security
//	}
//	err := generateCA(config, "/etc/toughradius/certs", "MyCompany Root CA")
func generateCA(baseConfig certgen.CertConfig, outputDir, commonName string) error {
	fmt.Println(">>> Generating CA certificate")

	config := certgen.CAConfig{
		CertConfig: baseConfig,
		OutputDir:  outputDir,
	}
	config.CommonName = commonName

	return certgen.GenerateCA(config)
}

// generateServer generates a server certificate signed by the CA certificate.
// The certificate supports SAN (Subject Alternative Name) extension for multiple
// DNS names and IP addresses, which is required by modern TLS clients.
//
// This function parses comma-separated DNS names and IP addresses, validates IP
// formats, and generates a certificate suitable for TLS server authentication
// (e.g., RadSec servers, HTTPS servers).
//
// Parameters:
//   - baseConfig: Base certificate configuration (organization info, validity, key size)
//   - outputDir: Directory containing CA files and where server cert will be saved
//   - commonName: Server's CommonName, typically the primary domain (e.g., "radius.example.com")
//   - dnsNames: Comma-separated DNS names for SAN extension (e.g., "radius.example.com,*.radius.example.com,localhost")
//     Whitespace around commas is automatically trimmed. Empty string is allowed.
//   - ipAddrs: Comma-separated IP addresses for SAN extension (e.g., "192.168.1.100,127.0.0.1,2001:db8::1")
//     Supports both IPv4 and IPv6. Whitespace is trimmed. Empty string is allowed.
//
// Returns:
//   - error: Returns error if CA files are missing, IP address format is invalid,
//     or certificate generation fails
//
// Required files in outputDir:
//   - ca.crt: CA certificate (must exist before calling this function)
//   - ca.key: CA private key (must exist before calling this function)
//
// Generated files:
//   - <outputDir>/server.crt: Server certificate in PEM format
//   - <outputDir>/server.key: Server private key in PKCS#8 PEM format (mode 0600)
//
// Common errors:
//   - "read CA cert failed": CA certificate not found (run generateCA first)
//   - "invalid IP address": IP format is incorrect (check for typos)
//
// SAN (Subject Alternative Name) usage:
//
//	Modern TLS clients (Chrome, Firefox, OpenSSL 1.1+) require SAN extension.
//	The CommonName field alone is deprecated and may be rejected.
//	Always include at least the CN value in the DNS SAN list.
//
// Example:
//
//	config := certgen.CertConfig{
//	    Organization: []string{"MyCompany"},
//	    ValidDays: 730,  // 2 years
//	    KeySize: 2048,
//	}
//	err := generateServer(config, "./certs", "radius.example.com",
//	    "radius.example.com,radius-backup.example.com,*.radius.example.com",
//	    "192.168.1.100,192.168.1.101,10.0.0.50")
func generateServer(baseConfig certgen.CertConfig, outputDir, commonName, dnsNames, ipAddrs string) error {
	fmt.Println(">>> Generating server certificate")

	config := certgen.ServerConfig{
		CertConfig: baseConfig,
		CAKeyPath:  outputDir + "/ca.key",
		CACertPath: outputDir + "/ca.crt",
		OutputDir:  outputDir,
	}
	config.CommonName = commonName

	// Parse DNS names
	if dnsNames != "" {
		config.DNSNames = strings.Split(dnsNames, ",")
		for i := range config.DNSNames {
			config.DNSNames[i] = strings.TrimSpace(config.DNSNames[i])
		}
	}

	// Parse IP addresses
	if ipAddrs != "" {
		ipList := strings.Split(ipAddrs, ",")
		for _, ip := range ipList {
			ip = strings.TrimSpace(ip)
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				config.IPAddresses = append(config.IPAddresses, parsedIP)
			} else {
				return fmt.Errorf("invalid IP address: %s", ip)
			}
		}
	}

	return certgen.GenerateServerCert(config)
}

// generateClient generates a client certificate signed by the CA certificate.
// Client certificates are used for mutual TLS authentication, commonly required
// by RadSec NAS (Network Access Server) devices connecting to RADIUS servers.
//
// Unlike server certificates, client certificates typically don't require SAN
// extensions (DNS/IP), but this function supports them for flexibility.
//
// Parameters:
//   - baseConfig: Base certificate configuration (organization info, validity, key size)
//   - outputDir: Directory containing CA files and where client cert will be saved
//   - commonName: Client's CommonName, typically the device identifier (e.g., "nas-mikrotik-001")
//   - dnsNames: Comma-separated DNS names for SAN extension (usually empty for RADIUS clients)
//   - ipAddrs: Comma-separated IP addresses for SAN extension (usually empty for RADIUS clients)
//
// Returns:
//   - error: Returns error if CA files are missing, IP address format is invalid,
//     or certificate generation fails
//
// Required files in outputDir:
//   - ca.crt: CA certificate (must exist before calling this function)
//   - ca.key: CA private key (must exist before calling this function)
//
// Generated files:
//   - <outputDir>/client.crt: Client certificate in PEM format
//   - <outputDir>/client.key: Client private key in PKCS#8 PEM format (mode 0600)
//
// Multiple client certificate workflow:
//
//	  Each NAS device should have its own unique client certificate. Since this
//	  function always outputs to "client.crt" and "client.key", you must rename
//	  the files immediately after generation when creating multiple certificates:
//
//		generateClient(config, "./certs", "nas-device-01", "", "")
//		// Rename: client.crt -> nas-device-01.crt
//		generateClient(config, "./certs", "nas-device-02", "", "")
//		// Rename: client.crt -> nas-device-02.crt
//
// Common errors:
//   - "read CA cert failed": CA certificate not found (run generateCA first)
//   - "invalid IP address": IP format is incorrect (check for typos)
//
// Security best practices:
//   - Generate separate certificates for each NAS device
//   - Use device-specific CommonNames for easier identification in logs
//   - Distribute only the client certificate and key to each NAS (not the CA key)
//   - Keep a record of issued certificates for audit and revocation purposes
//
// Example:
//
//	config := certgen.CertConfig{
//	    Organization: []string{"MyCompany"},
//	    ValidDays: 365,
//	    KeySize: 2048,
//	}
//	// Generate certificate for Mikrotik NAS device
//	err := generateClient(config, "./certs", "nas-mikrotik-001", "", "")
//	// Generate certificate for Cisco NAS device
//	err = generateClient(config, "./certs", "nas-cisco-002", "", "")
func generateClient(baseConfig certgen.CertConfig, outputDir, commonName, dnsNames, ipAddrs string) error {
	fmt.Println(">>> Generating client certificate")

	config := certgen.ClientConfig{
		CertConfig: baseConfig,
		CAKeyPath:  outputDir + "/ca.key",
		CACertPath: outputDir + "/ca.crt",
		OutputDir:  outputDir,
	}
	config.CommonName = commonName

	// Parse DNS names
	if dnsNames != "" {
		config.DNSNames = strings.Split(dnsNames, ",")
		for i := range config.DNSNames {
			config.DNSNames[i] = strings.TrimSpace(config.DNSNames[i])
		}
	}

	// Parse IP addresses
	if ipAddrs != "" {
		ipList := strings.Split(ipAddrs, ",")
		for _, ip := range ipList {
			ip = strings.TrimSpace(ip)
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				config.IPAddresses = append(config.IPAddresses, parsedIP)
			} else {
				return fmt.Errorf("invalid IP address: %s", ip)
			}
		}
	}

	return certgen.GenerateClientCert(config)
}
