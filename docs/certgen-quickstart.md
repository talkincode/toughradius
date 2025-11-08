# 证书生成工具快速指南

ToughRADIUS 提供了一个内置的证书生成工具，用于快速生成 TLS/SSL 证书，支持 SAN (Subject Alternative Name) 扩展。

## 快速开始

### 编译工具

```bash
go build -o certgen ./cmd/certgen
```

### 生成所有证书（一键生成）

```bash
./certgen -type all -output ./certs
```

这将生成：

- CA 根证书 (`ca.crt`, `ca.key`)
- 服务器证书 (`server.crt`, `server.key`) - 支持 SAN
- 客户端证书 (`client.crt`, `client.key`) - 支持 SAN

### 自定义参数生成

```bash
./certgen -type all \
  -server-cn radius.example.com \
  -server-dns "radius.example.com,*.radius.example.com,localhost" \
  -server-ips "192.168.1.100,10.0.0.1,127.0.0.1" \
  -org "MyCompany" \
  -output ./mycerts
```

## 主要功能

✅ **完整的 SAN 支持** - 支持多个 DNS 名称和 IP 地址  
✅ **一键生成** - 可一次性生成所有证书  
✅ **灵活配置** - 支持自定义组织信息、有效期等  
✅ **独立工具** - 核心功能在 `pkg/certgen` 包中，可在代码中直接使用

## 详细文档

完整的使用文档和 API 说明请参考：[cmd/certgen/README.md](../cmd/certgen/README.md)

## 在 ToughRADIUS 中使用

生成证书后，在 `toughradius.yml` 中配置 RadSec：

```yaml
radsec:
  enabled: true
  addr: :2083
  cert_file: ./certs/server.crt
  key_file: ./certs/server.key
  ca_file: ./certs/ca.crt
```

## 证书验证

```bash
# 验证证书有效性
openssl verify -CAfile certs/ca.crt certs/server.crt

# 查看证书详情（包括 SAN）
openssl x509 -in certs/server.crt -text -noout | grep -A 3 "Subject Alternative Name"
```
