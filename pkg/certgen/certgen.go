// Package certgen provides utilities for generating X.509 certificates for RadSec (RADIUS over TLS).
//
// This package simplifies TLS certificate generation for:
//   - CA (Certificate Authority) creation for internal PKI
//   - Server certificates for RadSec listeners (RFC 6614)
//   - Client certificates for mutual TLS authentication
//
// All certificates support Subject Alternative Names (SAN) for DNS and IP addresses,
// which is required by modern TLS clients.
//
// Example usage:
//
//	// 1. Generate CA certificate
//	caConfig := certgen.CAConfig{
//	    CertConfig: certgen.DefaultCertConfig(),
//	    OutputDir: "/etc/toughradius/certs",
//	}
//	caConfig.CommonName = "ToughRADIUS CA"
//	certgen.GenerateCA(caConfig)
//
//	// 2. Generate server certificate with SAN
//	serverConfig := certgen.ServerConfig{
//	    CertConfig: certgen.DefaultCertConfig(),
//	    CAKeyPath: "/etc/toughradius/certs/ca.key",
//	    CACertPath: "/etc/toughradius/certs/ca.crt",
//	    OutputDir: "/etc/toughradius/certs",
//	}
//	serverConfig.CommonName = "radius.example.com"
//	serverConfig.DNSNames = []string{"radius.example.com", "*.radius.example.com"}
//	serverConfig.IPAddresses = []net.IP{net.ParseIP("192.168.1.10")}
//	certgen.GenerateServerCert(serverConfig)
package certgen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CertConfig holds common X.509 certificate configuration shared across CA, server, and client certificates.
// It defines the certificate's subject (identity), validity period, and cryptographic parameters.
//
// Subject fields follow X.509 Distinguished Name (DN) structure:
//   - CommonName: Primary identifier (e.g., hostname for server certs, username for client certs)
//   - Organization: Organization name(s)
//   - OrganizationalUnit: Department or division
//   - Country: Two-letter country code (e.g., "CN", "US")
//   - Province: State or province
//   - Locality: City or locality
//
// SAN (Subject Alternative Names) support:
//   - DNSNames: Additional DNS names (wildcards supported: "*.example.com")
//   - IPAddresses: IP addresses for IP-based TLS connections
//
// Security parameters:
//   - ValidDays: Certificate validity period (default 3650 = 10 years)
//   - KeySize: RSA key size in bits (2048 or 4096 recommended)
type CertConfig struct {
	CommonName         string
	Organization       []string
	OrganizationalUnit []string
	Country            []string
	Province           []string
	Locality           []string
	DNSNames           []string // SAN DNS names
	IPAddresses        []net.IP // SAN IP addresses
	ValidDays          int      // Valid days
	KeySize            int      // RSA Key size
}

// CAConfig extends CertConfig for Certificate Authority generation.
// The CA is used to sign server and client certificates.
//
// Fields:
//   - CertConfig: Embedded base configuration
//   - OutputDir: Directory where ca.crt and ca.key will be saved
//
// Output files:
//   - ca.crt: PEM-encoded CA certificate (public, distribute to clients)
//   - ca.key: PEM-encoded CA private key (secret, protect carefully)
type CAConfig struct {
	CertConfig
	OutputDir string // Output directory
}

// ServerConfig extends CertConfig for server certificate generation.
// Server certificates are used by RadSec listeners for TLS encryption.
//
// Fields:
//   - CertConfig: Embedded base configuration (set DNSNames for SAN support)
//   - CAKeyPath: Path to CA private key (ca.key) for signing
//   - CACertPath: Path to CA certificate (ca.crt) for chain validation
//   - OutputDir: Directory where server.crt and server.key will be saved
//
// Output files:
//   - server.crt: PEM-encoded server certificate (signed by CA)
//   - server.key: PEM-encoded server private key (protect with 0600 permissions)
type ServerConfig struct {
	CertConfig
	CAKeyPath  string // CA private key path
	CACertPath string // CA certificate path
	OutputDir  string // Output directory
}

// ClientConfig extends CertConfig for client certificate generation.
// Client certificates enable mutual TLS authentication for RadSec clients.
//
// Fields:
//   - CertConfig: Embedded base configuration (CommonName = client identifier)
//   - CAKeyPath: Path to CA private key (ca.key) for signing
//   - CACertPath: Path to CA certificate (ca.crt) for chain validation
//   - OutputDir: Directory where client.crt and client.key will be saved
//
// Output files:
//   - client.crt: PEM-encoded client certificate (signed by CA)
//   - client.key: PEM-encoded client private key (protect with 0600 permissions)
type ClientConfig struct {
	CertConfig
	CAKeyPath  string // CA private key path
	CACertPath string // CA certificate path
	OutputDir  string // Output directory
}

// DefaultCertConfig returns a pre-configured CertConfig with sensible defaults.
// Use this as a starting point and override specific fields as needed.
//
// Default values:
//   - Organization: "ToughRADIUS"
//   - OrganizationalUnit: "IT"
//   - Country: "CN"
//   - Province: "Shanghai"
//   - Locality: "Shanghai"
//   - ValidDays: 3650 (10 years)
//   - KeySize: 2048 bits (RSA)
//
// Returns:
//   - CertConfig: Configuration with default values
//
// Example:
//
//	config := certgen.DefaultCertConfig()
//	config.CommonName = "radius.example.com"
//	config.ValidDays = 365  // Override to 1 year
func DefaultCertConfig() CertConfig {
	return CertConfig{
		Organization:       []string{"ToughRADIUS"},
		OrganizationalUnit: []string{"IT"},
		Country:            []string{"CN"},
		Province:           []string{"Shanghai"},
		Locality:           []string{"Shanghai"},
		ValidDays:          3650,
		KeySize:            2048,
	}
}

// GenerateCA creates a self-signed Certificate Authority (CA) for internal PKI.
// The CA is used to sign server and client certificates, establishing a trust chain.
//
// This generates:
//   - RSA private key (PKCS#8 format)
//   - Self-signed X.509 certificate with CA:TRUE constraint
//   - Certificate supports both server and client authentication
//
// Parameters:
//   - config: CA configuration (must set CommonName)
//
// Returns:
//   - error: File I/O, crypto generation, or encoding errors
//
// Output files:
//   - {OutputDir}/ca.crt: CA certificate (PEM, mode 0644)
//   - {OutputDir}/ca.key: CA private key (PEM, mode 0600 - protect carefully!)
//
// Side effects:
//   - Creates OutputDir if it doesn't exist (mode 0755)
//   - Prints file paths to stdout on success
//
// Example:
//
//	caConfig := certgen.CAConfig{
//	    CertConfig: certgen.DefaultCertConfig(),
//	    OutputDir: "/etc/toughradius/certs",
//	}
//	caConfig.CommonName = "ToughRADIUS Internal CA"
//	if err := certgen.GenerateCA(caConfig); err != nil {
//	    log.Fatalf("CA generation failed: %v", err)
//	}
func GenerateCA(config CAConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil { //nolint:gosec // G301: 0755 is standard for certificate directories
		return fmt.Errorf("create output directory failed: %w", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("generate private key failed: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate serial number failed: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         config.CommonName,
			Organization:       config.Organization,
			OrganizationalUnit: config.OrganizationalUnit,
			Country:            config.Country,
			Province:           config.Province,
			Locality:           config.Locality,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(config.ValidDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("create certificate failed: %w", err)
	}

	// Save certificate
	certPath := filepath.Join(config.OutputDir, "ca.crt")
	certOut, err := os.Create(certPath) //nolint:gosec // G304: path is constructed from validated config
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer func() { _ = certOut.Close() }() //nolint:errcheck

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(config.OutputDir, "ca.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) //nolint:gosec // G304: path is constructed from validated config
	if err != nil {
		return fmt.Errorf("create key file failed: %w", err)
	}
	defer func() { _ = keyOut.Close() }() //nolint:errcheck

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal private key failed: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("encode private key failed: %w", err)
	}

	fmt.Printf("CA certificate generated successfully:\n")
	fmt.Printf("  Certificate: %s\n", certPath)
	fmt.Printf("  Private Key: %s\n", keyPath)

	return nil
}

// GenerateServerCert creates a server certificate signed by the CA.
// This certificate is used by RadSec servers for TLS encryption (RFC 6614).
//
// Supports Subject Alternative Names (SAN):
//   - DNSNames: Multiple hostnames (wildcards supported)
//   - IPAddresses: Direct IP-based TLS connections
//
// Parameters:
//   - config: Server configuration (must set CommonName, CAKeyPath, CACertPath)
//
// Returns:
//   - error: File I/O, CA loading, crypto generation, or encoding errors
//
// Output files:
//   - {OutputDir}/server.crt: Server certificate (PEM, mode 0644)
//   - {OutputDir}/server.key: Server private key (PEM, mode 0600)
//
// Side effects:
//   - Creates OutputDir if it doesn't exist
//   - Prints file paths and SAN info to stdout on success
//
// Example:
//
//	serverConfig := certgen.ServerConfig{
//	    CertConfig: certgen.DefaultCertConfig(),
//	    CAKeyPath: "/etc/toughradius/certs/ca.key",
//	    CACertPath: "/etc/toughradius/certs/ca.crt",
//	    OutputDir: "/etc/toughradius/certs",
//	}
//	serverConfig.CommonName = "radius.example.com"
//	serverConfig.DNSNames = []string{"radius.example.com", "*.radius.example.com"}
//	serverConfig.IPAddresses = []net.IP{net.ParseIP("192.168.1.10")}
//	if err := certgen.GenerateServerCert(serverConfig); err != nil {
//	    log.Fatalf("Server cert generation failed: %v", err)
//	}
func GenerateServerCert(config ServerConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil { //nolint:gosec // G301: 0755 is standard for certificate directories
		return fmt.Errorf("create output directory failed: %w", err)
	}

	// Load CA certificateand private key
	caCert, caKey, err := loadCAFiles(config.CACertPath, config.CAKeyPath)
	if err != nil {
		return err
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("generate private key failed: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate serial number failed: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         config.CommonName,
			Organization:       config.Organization,
			OrganizationalUnit: config.OrganizationalUnit,
			Country:            config.Country,
			Province:           config.Province,
			Locality:           config.Locality,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(config.ValidDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              config.DNSNames,
		IPAddresses:           config.IPAddresses,
	}

	// Sign using the CA
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create certificate failed: %w", err)
	}

	// Save certificate
	certPath := filepath.Join(config.OutputDir, "server.crt")
	certOut, err := os.Create(certPath) //nolint:gosec // G304: path is constructed from validated config
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer func() { _ = certOut.Close() }() //nolint:errcheck

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(config.OutputDir, "server.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) //nolint:gosec // G304: path is constructed from validated config
	if err != nil {
		return fmt.Errorf("create key file failed: %w", err)
	}
	defer func() { _ = keyOut.Close() }() //nolint:errcheck

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal private key failed: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("encode private key failed: %w", err)
	}

	fmt.Printf("Server certificate generated successfully:\n")
	fmt.Printf("  Certificate: %s\n", certPath)
	fmt.Printf("  Private Key: %s\n", keyPath)
	if len(config.DNSNames) > 0 {
		fmt.Printf("  DNS Names: %v\n", config.DNSNames)
	}
	if len(config.IPAddresses) > 0 {
		fmt.Printf("  IP Addresses: %v\n", config.IPAddresses)
	}

	return nil
}

// GenerateClientCert creates a client certificate signed by the CA.
// This certificate enables mutual TLS authentication for RadSec clients.
//
// The CommonName typically identifies the client (e.g., NAS hostname or identifier).
// SAN support allows multiple identities in a single certificate.
//
// Parameters:
//   - config: Client configuration (must set CommonName, CAKeyPath, CACertPath)
//
// Returns:
//   - error: File I/O, CA loading, crypto generation, or encoding errors
//
// Output files:
//   - {OutputDir}/client.crt: Client certificate (PEM, mode 0644)
//   - {OutputDir}/client.key: Client private key (PEM, mode 0600)
//
// Side effects:
//   - Creates OutputDir if it doesn't exist
//   - Prints file paths and SAN info to stdout on success
//
// Example:
//
//	clientConfig := certgen.ClientConfig{
//	    CertConfig: certgen.DefaultCertConfig(),
//	    CAKeyPath: "/etc/toughradius/certs/ca.key",
//	    CACertPath: "/etc/toughradius/certs/ca.crt",
//	    OutputDir: "/etc/toughradius/clients",
//	}
//	clientConfig.CommonName = "nas-device-001"
//	if err := certgen.GenerateClientCert(clientConfig); err != nil {
//	    log.Fatalf("Client cert generation failed: %v", err)
//	}
func GenerateClientCert(config ClientConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil { //nolint:gosec // G301: 0755 is standard for certificate directories
		return fmt.Errorf("create output directory failed: %w", err)
	}

	// Load CA certificateand private key
	caCert, caKey, err := loadCAFiles(config.CACertPath, config.CAKeyPath)
	if err != nil {
		return err
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("generate private key failed: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate serial number failed: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         config.CommonName,
			Organization:       config.Organization,
			OrganizationalUnit: config.OrganizationalUnit,
			Country:            config.Country,
			Province:           config.Province,
			Locality:           config.Locality,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(config.ValidDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              config.DNSNames,
		IPAddresses:           config.IPAddresses,
	}

	// Sign using the CA
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create certificate failed: %w", err)
	}

	// Save certificate
	certPath := filepath.Join(config.OutputDir, "client.crt")
	certOut, err := os.Create(certPath) //nolint:gosec // G304: path is constructed from validated config
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer func() { _ = certOut.Close() }() //nolint:errcheck

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(config.OutputDir, "client.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) //nolint:gosec // G304: path is constructed from validated config
	if err != nil {
		return fmt.Errorf("create key file failed: %w", err)
	}
	defer func() { _ = keyOut.Close() }() //nolint:errcheck

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal private key failed: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("encode private key failed: %w", err)
	}

	fmt.Printf("Client certificate generated successfully:\n")
	fmt.Printf("  Certificate: %s\n", certPath)
	fmt.Printf("  Private Key: %s\n", keyPath)
	if len(config.DNSNames) > 0 {
		fmt.Printf("  DNS Names: %v\n", config.DNSNames)
	}
	if len(config.IPAddresses) > 0 {
		fmt.Printf("  IP Addresses: %v\n", config.IPAddresses)
	}

	return nil
}

// loadCAFiles loads the CA certificate and private key files
func loadCAFiles(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Read CA certificate
	certPEM, err := os.ReadFile(certPath) //nolint:gosec // G304: path is user-specified CA cert file
	if err != nil {
		return nil, nil, fmt.Errorf("read CA cert failed: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode CA cert PEM")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA cert failed: %w", err)
	}

	// Read CA private key
	keyPEM, err := os.ReadFile(keyPath) //nolint:gosec // G304: path is user-specified CA key file
	if err != nil {
		return nil, nil, fmt.Errorf("read CA key failed: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA key PEM")
	}

	caKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA key failed: %w", err)
	}

	rsaKey, ok := caKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("CA key is not RSA private key")
	}

	return caCert, rsaKey, nil
}
