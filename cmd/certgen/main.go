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
		// 通用参数
		certType    = flag.String("type", "all", "证书类型: ca, server, client, all")
		outputDir   = flag.String("output", "./certs", "输出目录")
		validDays   = flag.Int("days", 3650, "证书有效期(天)")
		keySize     = flag.Int("keysize", 2048, "RSA密钥大小")
		showVersion = flag.Bool("version", false, "显示版本信息")

		// CA 参数
		caCommonName = flag.String("ca-cn", "ToughRADIUS CA", "CA证书的CommonName")

		// 服务器证书参数
		serverCommonName = flag.String("server-cn", "radius.example.com", "服务器证书的CommonName")
		serverDNS        = flag.String("server-dns", "radius.example.com,*.radius.example.com,localhost", "服务器证书的DNS名称(逗号分隔)")
		serverIPs        = flag.String("server-ips", "127.0.0.1", "服务器证书的IP地址(逗号分隔)")

		// 客户端证书参数
		clientCommonName = flag.String("client-cn", "radius-client", "客户端证书的CommonName")
		clientDNS        = flag.String("client-dns", "", "客户端证书的DNS名称(逗号分隔)")
		clientIPs        = flag.String("client-ips", "", "客户端证书的IP地址(逗号分隔)")

		// 组织信息
		organization = flag.String("org", "ToughRADIUS", "组织名称")
		orgUnit      = flag.String("ou", "IT", "组织单元")
		country      = flag.String("country", "CN", "国家代码")
		province     = flag.String("province", "Shanghai", "省份")
		locality     = flag.String("locality", "Shanghai", "城市")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ToughRADIUS 证书生成工具 v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  # 生成所有证书(CA + 服务器 + 客户端)\n")
		fmt.Fprintf(os.Stderr, "  %s -type all\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 仅生成CA证书\n")
		fmt.Fprintf(os.Stderr, "  %s -type ca -ca-cn \"My CA\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 生成服务器证书(需要先有CA证书)\n")
		fmt.Fprintf(os.Stderr, "  %s -type server -server-cn radius.mycompany.com -server-dns \"radius.mycompany.com,*.radius.mycompany.com\" -server-ips \"192.168.1.100,10.0.0.1\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 生成客户端证书(需要先有CA证书)\n")
		fmt.Fprintf(os.Stderr, "  %s -type client -client-cn my-radius-client\n\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("ToughRADIUS 证书生成工具 v%s\n", version)
		os.Exit(0)
	}

	// 创建基础配置
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
			log.Fatalf("生成证书失败: %v", err)
		}
	case "ca":
		if err := generateCA(baseConfig, *outputDir, *caCommonName); err != nil {
			log.Fatalf("生成CA证书失败: %v", err)
		}
	case "server":
		if err := generateServer(baseConfig, *outputDir, *serverCommonName, *serverDNS, *serverIPs); err != nil {
			log.Fatalf("生成服务器证书失败: %v", err)
		}
	case "client":
		if err := generateClient(baseConfig, *outputDir, *clientCommonName, *clientDNS, *clientIPs); err != nil {
			log.Fatalf("生成客户端证书失败: %v", err)
		}
	default:
		log.Fatalf("未知的证书类型: %s (支持: ca, server, client, all)", *certType)
	}

	fmt.Printf("\n✓ 证书生成完成! 输出目录: %s\n", *outputDir)
}

// generateAll 生成所有证书
func generateAll(baseConfig certgen.CertConfig, outputDir, caCN, serverCN, serverDNS, serverIPs, clientCN, clientDNS, clientIPs string) error {
	fmt.Println("=== 开始生成所有证书 ===\n")

	// 1. 生成 CA
	if err := generateCA(baseConfig, outputDir, caCN); err != nil {
		return fmt.Errorf("生成CA失败: %w", err)
	}
	fmt.Println()

	// 2. 生成服务器证书
	if err := generateServer(baseConfig, outputDir, serverCN, serverDNS, serverIPs); err != nil {
		return fmt.Errorf("生成服务器证书失败: %w", err)
	}
	fmt.Println()

	// 3. 生成客户端证书
	if err := generateClient(baseConfig, outputDir, clientCN, clientDNS, clientIPs); err != nil {
		return fmt.Errorf("生成客户端证书失败: %w", err)
	}

	return nil
}

// generateCA 生成CA证书
func generateCA(baseConfig certgen.CertConfig, outputDir, commonName string) error {
	fmt.Println(">>> 生成CA证书")

	config := certgen.CAConfig{
		CertConfig: baseConfig,
		OutputDir:  outputDir,
	}
	config.CommonName = commonName

	return certgen.GenerateCA(config)
}

// generateServer 生成服务器证书
func generateServer(baseConfig certgen.CertConfig, outputDir, commonName, dnsNames, ipAddrs string) error {
	fmt.Println(">>> 生成服务器证书")

	config := certgen.ServerConfig{
		CertConfig: baseConfig,
		CAKeyPath:  outputDir + "/ca.key",
		CACertPath: outputDir + "/ca.crt",
		OutputDir:  outputDir,
	}
	config.CommonName = commonName

	// 解析 DNS 名称
	if dnsNames != "" {
		config.DNSNames = strings.Split(dnsNames, ",")
		for i := range config.DNSNames {
			config.DNSNames[i] = strings.TrimSpace(config.DNSNames[i])
		}
	}

	// 解析 IP 地址
	if ipAddrs != "" {
		ipList := strings.Split(ipAddrs, ",")
		for _, ip := range ipList {
			ip = strings.TrimSpace(ip)
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				config.IPAddresses = append(config.IPAddresses, parsedIP)
			} else {
				return fmt.Errorf("无效的IP地址: %s", ip)
			}
		}
	}

	return certgen.GenerateServerCert(config)
}

// generateClient 生成客户端证书
func generateClient(baseConfig certgen.CertConfig, outputDir, commonName, dnsNames, ipAddrs string) error {
	fmt.Println(">>> 生成客户端证书")

	config := certgen.ClientConfig{
		CertConfig: baseConfig,
		CAKeyPath:  outputDir + "/ca.key",
		CACertPath: outputDir + "/ca.crt",
		OutputDir:  outputDir,
	}
	config.CommonName = commonName

	// 解析 DNS 名称
	if dnsNames != "" {
		config.DNSNames = strings.Split(dnsNames, ",")
		for i := range config.DNSNames {
			config.DNSNames[i] = strings.TrimSpace(config.DNSNames[i])
		}
	}

	// 解析 IP 地址
	if ipAddrs != "" {
		ipList := strings.Split(ipAddrs, ",")
		for _, ip := range ipList {
			ip = strings.TrimSpace(ip)
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				config.IPAddresses = append(config.IPAddresses, parsedIP)
			} else {
				return fmt.Errorf("无效的IP地址: %s", ip)
			}
		}
	}

	return certgen.GenerateClientCert(config)
}
