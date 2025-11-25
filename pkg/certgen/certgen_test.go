package certgen

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCA(t *testing.T) {
	tmpDir := t.TempDir()

	config := CAConfig{
		CertConfig: DefaultCertConfig(),
		OutputDir:  tmpDir,
	}
	config.CommonName = "Test CA"

	err := GenerateCA(config)
	if err != nil {
		t.Fatalf("GenerateCA failed: %v", err)
	}

	// Validate that the file exists
	certPath := filepath.Join(tmpDir, "ca.crt")
	keyPath := filepath.Join(tmpDir, "ca.key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("CA certificate file not created")
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("CA key file not created")
	}

	// Validate the certificate contents
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read CA cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("Failed to decode CA cert PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse CA cert: %v", err)
	}

	if cert.Subject.CommonName != "Test CA" {
		t.Errorf("Expected CommonName 'Test CA', got '%s'", cert.Subject.CommonName)
	}

	if !cert.IsCA {
		t.Error("Certificate is not a CA")
	}
}

func TestGenerateServerCert(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate the CA first
	caConfig := CAConfig{
		CertConfig: DefaultCertConfig(),
		OutputDir:  tmpDir,
	}
	caConfig.CommonName = "Test CA"

	err := GenerateCA(caConfig)
	if err != nil {
		t.Fatalf("GenerateCA failed: %v", err)
	}

	// Generate the server certificate
	serverConfig := ServerConfig{
		CertConfig: DefaultCertConfig(),
		CAKeyPath:  filepath.Join(tmpDir, "ca.key"),
		CACertPath: filepath.Join(tmpDir, "ca.crt"),
		OutputDir:  tmpDir,
	}
	serverConfig.CommonName = "radius.example.com"
	serverConfig.DNSNames = []string{"radius.example.com", "*.radius.example.com", "localhost"}
	serverConfig.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("192.168.1.100")}

	err = GenerateServerCert(serverConfig)
	if err != nil {
		t.Fatalf("GenerateServerCert failed: %v", err)
	}

	// Validate that the file exists
	certPath := filepath.Join(tmpDir, "server.crt")
	keyPath := filepath.Join(tmpDir, "server.key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Server certificate file not created")
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Server key file not created")
	}

	// Validate the certificate contents and SAN
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read server cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("Failed to decode server cert PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse server cert: %v", err)
	}

	if cert.Subject.CommonName != "radius.example.com" {
		t.Errorf("Expected CommonName 'radius.example.com', got '%s'", cert.Subject.CommonName)
	}

	// Validate SAN
	if len(cert.DNSNames) != 3 {
		t.Errorf("Expected 3 DNS names, got %d", len(cert.DNSNames))
	}

	if len(cert.IPAddresses) != 2 {
		t.Errorf("Expected 2 IP addresses, got %d", len(cert.IPAddresses))
	}

	// Validate ExtKeyUsage
	hasServerAuth := false
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			hasServerAuth = true
			break
		}
	}
	if !hasServerAuth {
		t.Error("Certificate does not have ServerAuth ExtKeyUsage")
	}
}

func TestGenerateClientCert(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate the CA first
	caConfig := CAConfig{
		CertConfig: DefaultCertConfig(),
		OutputDir:  tmpDir,
	}
	caConfig.CommonName = "Test CA"

	err := GenerateCA(caConfig)
	if err != nil {
		t.Fatalf("GenerateCA failed: %v", err)
	}

	// Generate the client certificate
	clientConfig := ClientConfig{
		CertConfig: DefaultCertConfig(),
		CAKeyPath:  filepath.Join(tmpDir, "ca.key"),
		CACertPath: filepath.Join(tmpDir, "ca.crt"),
		OutputDir:  tmpDir,
	}
	clientConfig.CommonName = "client.example.com"
	clientConfig.DNSNames = []string{"client.example.com"}

	err = GenerateClientCert(clientConfig)
	if err != nil {
		t.Fatalf("GenerateClientCert failed: %v", err)
	}

	// Validate that the file exists
	certPath := filepath.Join(tmpDir, "client.crt")
	keyPath := filepath.Join(tmpDir, "client.key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Client certificate file not created")
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Client key file not created")
	}

	// Validate the certificate contents
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read client cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("Failed to decode client cert PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse client cert: %v", err)
	}

	if cert.Subject.CommonName != "client.example.com" {
		t.Errorf("Expected CommonName 'client.example.com', got '%s'", cert.Subject.CommonName)
	}

	// Validate ExtKeyUsage
	hasClientAuth := false
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
			break
		}
	}
	if !hasClientAuth {
		t.Error("Certificate does not have ClientAuth ExtKeyUsage")
	}
}
