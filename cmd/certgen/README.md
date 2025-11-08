# ToughRADIUS 证书生成工具

这是一个用于生成 ToughRADIUS 所需 TLS/SSL 证书的命令行工具，支持生成 CA 证书、服务器证书和客户端证书，并且完全支持 SAN (Subject Alternative Name) 扩展。

## 功能特性

- ✅ 生成 CA 根证书
- ✅ 生成服务器证书（支持 SAN）
- ✅ 生成客户端证书（支持 SAN）
- ✅ 支持一次性生成所有证书
- ✅ 支持自定义证书参数
- ✅ 支持多个 DNS 名称和 IP 地址（SAN）

## 编译

```bash
# 从项目根目录编译
go build -o certgen ./cmd/certgen

# 或使用交叉编译
GOOS=linux GOARCH=amd64 go build -o certgen-linux ./cmd/certgen
```

## 快速开始

### 生成所有证书（推荐）

```bash
./certgen -type all -output ./mycerts
```

这将在 `./mycerts` 目录下生成：
- `ca.crt` 和 `ca.key` - CA 根证书
- `server.crt` 和 `server.key` - 服务器证书
- `client.crt` 和 `client.key` - 客户端证书

### 仅生成 CA 证书

```bash
./certgen -type ca -ca-cn "My Company CA" -output ./mycerts
```

### 生成服务器证书（需要先有 CA）

```bash
./certgen -type server \
  -server-cn radius.example.com \
  -server-dns "radius.example.com,*.radius.example.com,localhost" \
  -server-ips "192.168.1.100,10.0.0.1,127.0.0.1" \
  -output ./mycerts
```

### 生成客户端证书（需要先有 CA）

```bash
./certgen -type client \
  -client-cn my-radius-client \
  -output ./mycerts
```

## 命令行参数

### 通用参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-type` | `all` | 证书类型: `ca`, `server`, `client`, `all` |
| `-output` | `./certs` | 输出目录 |
| `-days` | `3650` | 证书有效期(天)，默认10年 |
| `-keysize` | `2048` | RSA密钥大小 |
| `-version` | - | 显示版本信息 |

### 组织信息参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-org` | `ToughRADIUS` | 组织名称 |
| `-ou` | `IT` | 组织单元 |
| `-country` | `CN` | 国家代码 |
| `-province` | `Shanghai` | 省份 |
| `-locality` | `Shanghai` | 城市 |

### CA 证书参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-ca-cn` | `ToughRADIUS CA` | CA证书的CommonName |

### 服务器证书参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-server-cn` | `radius.example.com` | 服务器证书的CommonName |
| `-server-dns` | `radius.example.com,*.radius.example.com,localhost` | DNS名称(逗号分隔) |
| `-server-ips` | `127.0.0.1` | IP地址(逗号分隔) |

### 客户端证书参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-client-cn` | `radius-client` | 客户端证书的CommonName |
| `-client-dns` | - | DNS名称(逗号分隔) |
| `-client-ips` | - | IP地址(逗号分隔) |

## 使用示例

### 示例1：为生产环境生成证书

```bash
./certgen -type all \
  -ca-cn "MyCompany RADIUS CA" \
  -server-cn radius.mycompany.com \
  -server-dns "radius.mycompany.com,radius-backup.mycompany.com" \
  -server-ips "192.168.1.100,192.168.1.101" \
  -client-cn radius-nas-client \
  -org "MyCompany Inc." \
  -ou "Network Infrastructure" \
  -country US \
  -province "California" \
  -locality "San Francisco" \
  -days 1825 \
  -output /etc/toughradius/certs
```

### 示例2：为开发环境生成证书

```bash
./certgen -type all \
  -server-cn localhost \
  -server-dns "localhost,*.localhost,127.0.0.1" \
  -server-ips "127.0.0.1,::1" \
  -output ./dev-certs
```

### 示例3：为多个 NAS 生成客户端证书

```bash
# 首先确保已有 CA 证书
./certgen -type ca -output ./certs

# 为第一个 NAS 生成客户端证书
./certgen -type client \
  -client-cn nas-01.example.com \
  -output ./certs

# 重命名证书
mv ./certs/client.crt ./certs/nas-01.crt
mv ./certs/client.key ./certs/nas-01.key

# 为第二个 NAS 生成客户端证书
./certgen -type client \
  -client-cn nas-02.example.com \
  -output ./certs

mv ./certs/client.crt ./certs/nas-02.crt
mv ./certs/client.key ./certs/nas-02.key
```

## SAN (Subject Alternative Name) 支持

SAN 允许一个证书用于多个域名和 IP 地址，这对于以下场景非常有用：

1. **多域名支持**：同一个服务器使用多个域名
2. **通配符域名**：支持 `*.example.com` 格式
3. **IP 地址**：直接使用 IP 地址访问
4. **负载均衡**：多个服务器共享证书

### SAN 示例

```bash
./certgen -type server \
  -server-cn radius.example.com \
  -server-dns "radius.example.com,radius1.example.com,radius2.example.com,*.radius.example.com" \
  -server-ips "192.168.1.100,192.168.1.101,10.0.0.50" \
  -output ./certs
```

生成的证书将包含：
- **CN**: radius.example.com
- **DNS SANs**: radius.example.com, radius1.example.com, radius2.example.com, *.radius.example.com
- **IP SANs**: 192.168.1.100, 192.168.1.101, 10.0.0.50

## 证书验证

### 查看证书信息

```bash
# 查看服务器证书
openssl x509 -in certs/server.crt -text -noout

# 查看 CA 证书
openssl x509 -in certs/ca.crt -text -noout

# 查看客户端证书
openssl x509 -in certs/client.crt -text -noout
```

### 验证证书链

```bash
# 验证服务器证书
openssl verify -CAfile certs/ca.crt certs/server.crt

# 验证客户端证书
openssl verify -CAfile certs/ca.crt certs/client.crt
```

### 测试 TLS 连接

```bash
# 启动测试服务器
openssl s_server -accept 8443 \
  -cert certs/server.crt \
  -key certs/server.key \
  -CAfile certs/ca.crt

# 使用客户端证书连接
openssl s_client -connect localhost:8443 \
  -cert certs/client.crt \
  -key certs/client.key \
  -CAfile certs/ca.crt
```

## 在 ToughRADIUS 中使用

### RadSec 配置

在 `toughradius.yml` 中配置 RadSec：

```yaml
radsec:
  enabled: true
  addr: :2083
  cert_file: /path/to/certs/server.crt
  key_file: /path/to/certs/server.key
  ca_file: /path/to/certs/ca.crt
```

### FreeRADIUS 客户端配置

如果使用 FreeRADIUS 作为 RadSec 客户端：

```conf
# /etc/freeradius/sites-enabled/tls
home_server toughradius {
    type = auth+acct
    ipaddr = radius.example.com
    port = 2083
    proto = tcp
    
    secret = radsec
    
    tls {
        private_key_file = /path/to/certs/client.key
        certificate_file = /path/to/certs/client.crt
        ca_file = /path/to/certs/ca.crt
    }
}
```

## 安全建议

1. **私钥保护**：生成的 `.key` 文件权限应设为 `600`
   ```bash
   chmod 600 certs/*.key
   ```

2. **CA 私钥**：CA 私钥 (`ca.key`) 应妥善保管，不要部署到服务器
   ```bash
   # 生成后立即备份并删除
   cp certs/ca.key ~/secure-backup/
   rm certs/ca.key
   ```

3. **证书有效期**：定期更新证书，建议有效期不超过2年

4. **证书吊销**：如果证书泄露，应重新生成整个证书链

## 故障排除

### 错误：CA 证书不存在

```
Error: read CA cert failed: open ./certs/ca.crt: no such file or directory
```

**解决方案**：先生成 CA 证书
```bash
./certgen -type ca -output ./certs
```

### 错误：无效的 IP 地址

```
Error: 无效的IP地址: 192.168.1.256
```

**解决方案**：检查 IP 地址格式是否正确

### 证书验证失败

```
error 20 at 0 depth lookup: unable to get local issuer certificate
```

**解决方案**：确保使用正确的 CA 证书进行验证

## 核心库 API 使用

如果需要在代码中使用证书生成功能：

```go
import "github.com/talkincode/toughradius/pkg/certgen"

// 生成 CA
caConfig := certgen.CAConfig{
    CertConfig: certgen.DefaultCertConfig(),
    OutputDir:  "./certs",
}
caConfig.CommonName = "My CA"
err := certgen.GenerateCA(caConfig)

// 生成服务器证书
serverConfig := certgen.ServerConfig{
    CertConfig: certgen.DefaultCertConfig(),
    CAKeyPath:  "./certs/ca.key",
    CACertPath: "./certs/ca.crt",
    OutputDir:  "./certs",
}
serverConfig.CommonName = "radius.example.com"
serverConfig.DNSNames = []string{"radius.example.com", "*.radius.example.com"}
serverConfig.IPAddresses = []net.IP{net.ParseIP("192.168.1.100")}
err = certgen.GenerateServerCert(serverConfig)
```

## 许可证

与 ToughRADIUS 项目保持一致。
