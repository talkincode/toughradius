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

const version = "1.0.0"

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

// generateAll generates all certificates
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

// generateCA generates the CA certificate
func generateCA(baseConfig certgen.CertConfig, outputDir, commonName string) error {
	fmt.Println(">>> Generating CA certificate")

	config := certgen.CAConfig{
		CertConfig: baseConfig,
		OutputDir:  outputDir,
	}
	config.CommonName = commonName

	return certgen.GenerateCA(config)
}

// generateServer generates the server certificate
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

// generateClient generates the client certificate
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
