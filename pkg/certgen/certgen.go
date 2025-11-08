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

// CertConfig 证书配置
type CertConfig struct {
	CommonName         string
	Organization       []string
	OrganizationalUnit []string
	Country            []string
	Province           []string
	Locality           []string
	DNSNames           []string // SAN DNS 名称
	IPAddresses        []net.IP // SAN IP 地址
	ValidDays          int      // 有效天数
	KeySize            int      // RSA 密钥大小
}

// CAConfig CA 证书配置
type CAConfig struct {
	CertConfig
	OutputDir string // 输出目录
}

// ServerConfig 服务器证书配置
type ServerConfig struct {
	CertConfig
	CAKeyPath  string // CA 私钥路径
	CACertPath string // CA 证书路径
	OutputDir  string // 输出目录
}

// ClientConfig 客户端证书配置
type ClientConfig struct {
	CertConfig
	CAKeyPath  string // CA 私钥路径
	CACertPath string // CA 证书路径
	OutputDir  string // 输出目录
}

// DefaultCertConfig 返回默认证书配置
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

// GenerateCA 生成 CA 证书
func GenerateCA(config CAConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output directory failed: %w", err)
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("generate private key failed: %w", err)
	}

	// 创建证书模板
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

	// 自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("create certificate failed: %w", err)
	}

	// 保存证书
	certPath := filepath.Join(config.OutputDir, "ca.crt")
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// 保存私钥
	keyPath := filepath.Join(config.OutputDir, "ca.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create key file failed: %w", err)
	}
	defer keyOut.Close()

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

// GenerateServerCert 生成服务器证书（支持 SAN）
func GenerateServerCert(config ServerConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output directory failed: %w", err)
	}

	// 加载 CA 证书和私钥
	caCert, caKey, err := loadCAFiles(config.CACertPath, config.CAKeyPath)
	if err != nil {
		return err
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("generate private key failed: %w", err)
	}

	// 创建证书模板
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

	// 使用 CA 签名
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create certificate failed: %w", err)
	}

	// 保存证书
	certPath := filepath.Join(config.OutputDir, "server.crt")
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// 保存私钥
	keyPath := filepath.Join(config.OutputDir, "server.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create key file failed: %w", err)
	}
	defer keyOut.Close()

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

// GenerateClientCert 生成客户端证书（支持 SAN）
func GenerateClientCert(config ClientConfig) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output directory failed: %w", err)
	}

	// 加载 CA 证书和私钥
	caCert, caKey, err := loadCAFiles(config.CACertPath, config.CAKeyPath)
	if err != nil {
		return err
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("generate private key failed: %w", err)
	}

	// 创建证书模板
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

	// 使用 CA 签名
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create certificate failed: %w", err)
	}

	// 保存证书
	certPath := filepath.Join(config.OutputDir, "client.crt")
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("create cert file failed: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("encode certificate failed: %w", err)
	}

	// 保存私钥
	keyPath := filepath.Join(config.OutputDir, "client.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create key file failed: %w", err)
	}
	defer keyOut.Close()

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

// loadCAFiles 加载 CA 证书和私钥文件
func loadCAFiles(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// 读取 CA 证书
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

	// 读取 CA 私钥
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
