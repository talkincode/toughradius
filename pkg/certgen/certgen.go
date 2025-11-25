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

// CertConfig Certificate configuration
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

// CAConfig CA Certificate configuration
type CAConfig struct {
	CertConfig
	OutputDir string // Output directory
}

// ServerConfig holds server certificate configuration
type ServerConfig struct {
	CertConfig
	CAKeyPath  string // CA private key path
	CACertPath string // CA certificate path
	OutputDir  string // Output directory
}

// ClientConfig holds client certificate configuration
type ClientConfig struct {
	CertConfig
	CAKeyPath  string // CA private key path
	CACertPath string // CA certificate path
	OutputDir  string // Output directory
}

// DefaultCertConfig Returns default certificate configuration
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

// GenerateCA Generate CA certificate
func GenerateCA(config CAConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
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
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer func() { _ = certOut.Close() }() //nolint:errcheck

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(config.OutputDir, "ca.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
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

// GenerateServerCert generates the server certificate (supports SAN)
func GenerateServerCert(config ServerConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
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
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer func() { _ = certOut.Close() }() //nolint:errcheck

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(config.OutputDir, "server.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
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

// GenerateClientCert generates the client certificate (supports SAN)
func GenerateClientCert(config ClientConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
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
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer func() { _ = certOut.Close() }() //nolint:errcheck

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// Save private key
	keyPath := filepath.Join(config.OutputDir, "client.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
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
	certPEM, err := os.ReadFile(certPath)
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
	keyPEM, err := os.ReadFile(keyPath)
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
