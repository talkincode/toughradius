# 华为 ME60 新功能实现说明

## 📋 功能概述

本次更新为 ToughRADIUS v9 实现了三个华为 ME60 设备所需的关键功能：

1. ✅ **IPv4/IPv6 地址下发** - 在 RADIUS 认证响应中下发静态 IPv6 地址
2. ⚠️ **华为域属性下发** - 设置华为域名属性（需要架构调整）
3. ✅ **CoA 强制断线** - 通过 RADIUS CoA 协议实现真正的强制下线

---

## 🎯 功能详解

### 1. IPv6 地址下发

#### 支持的场景

- 用户配置了固定 IPv6 地址
- 华为 ME60 设备需要下发 IPv6 地址给用户
- 支持标准 RADIUS 协议和华为私有 VSA 属性

#### 配置方式

在用户管理界面的 `ipv6_addr` 字段中设置 IPv6 地址：

```
示例 1: 2001:db8::1/64   (带前缀长度)
示例 2: 2001:db8::1       (不带前缀，系统自动添加 /128)
示例 3: fe80::1/128       (链路本地地址)
```

#### 下发的 RADIUS 属性

**标准属性（所有设备）**:

- `Framed-IPv6-Prefix` (RFC 3162, Type 97) - 完整的 IPv6 前缀

**华为 VSA（仅华为设备）**:

- `Huawei-Framed-IPv6-Address` (Vendor-ID: 2011, Type: 158) - IPv6 地址（不含前缀长度）

#### 技术实现

- `default_enhancer.go`: 设置标准 `FramedIPv6Prefix`
- `huawei_enhancer.go`: 设置华为 `HuaweiFramedIPv6Address` VSA
- 自动检测华为设备（通过 NAS 的 `VendorCode = "2011"`）
- 智能处理有/无前缀长度的 IPv6 地址

---

### 2. 华为域属性下发

#### 当前状态

⚠️ **架构限制，暂未实现**

#### 问题描述

- `Domain` 字段存在于 `RadiusProfile` 表中
- `RadiusUser` 表中没有 `Domain` 字段
- 当前认证流程的 `AuthContext` 只包含 `User`，不包含 `Profile`

#### 解决方案

需要修改 `internal/radiusd/plugins/auth/auth.go`：

```go
type AuthContext struct {
    Request  *Request
    Response *radius.Packet
    User     *domain.RadiusUser
    Profile  *domain.RadiusProfile  // 新增这一行
    Nas      *domain.NetNas
    Metadata map[string]interface{}
}
```

然后在认证流程中根据 `user.ProfileId` 加载 `Profile` 数据。

#### 预留代码

已在 `huawei_enhancer.go` 中添加 TODO 注释：

```go
// TODO: Set Huawei Domain Name
// Note: Domain field exists in RadiusProfile, not RadiusUser
// Need to enhance AuthContext to include Profile information
// Then we can use: huawei.HuaweiDomainName_SetString(resp, profile.Domain)
```

---

### 3. CoA 强制断线功能

#### 功能描述

通过 RADIUS CoA（Change of Authorization）协议向 NAS 设备发送 Disconnect-Request，实现真正的用户强制下线。

#### 使用方式

**方法 1: Admin Web 界面**

1. 进入"在线用户"页面
2. 找到要下线的用户
3. 点击"删除"或"强制下线"按钮

**方法 2: REST API**

```bash
curl -X DELETE "http://localhost:1816/api/v1/sessions/{session_id}" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 执行流程

```
1. 客户端请求删除会话
   ↓
2. 查询会话信息（用户名、AcctSessionId、NAS 地址）
   ↓
3. 查询 NAS 配置（IP、Secret、CoA 端口）
   ↓
4. 删除数据库中的会话记录 ← 立即返回成功响应
   ↓
5. 异步发送 CoA Disconnect-Request 到 NAS:3799
   ↓
6. NAS 响应 Disconnect-ACK 或 Disconnect-NAK
   ↓
7. 记录日志（成功/失败/超时）
```

#### 技术特性

✅ **异步执行** - API 立即返回，不等待 NAS 响应  
✅ **超时保护** - 5 秒超时，避免长时间等待  
✅ **自动重试** - 2 秒内自动重试失败的请求  
✅ **优雅降级** - NAS 不可达时仍删除数据库记录  
✅ **完整日志** - 记录 ACK/NAK/错误所有情况

#### CoA 报文结构

```
Code: Disconnect-Request (40)
Secret: NAS 配置的 RADIUS Secret
Attributes:
  - Acct-Session-Id: 会话 ID（精确匹配用户）
  - User-Name: 用户名（辅助识别）
```

发送到: `NAS_IP:3799` (RFC 5176 标准 CoA 端口)

#### 日志查看

成功示例：

```
INFO  CoA Disconnect-Request ACK received
      nas_addr=192.168.1.1:3799
      username=test@example.com
```

失败示例：

```
WARN  CoA Disconnect-Request NAK received
      nas_addr=192.168.1.1:3799
      username=test@example.com
      response_code=41
```

错误示例：

```
ERROR Failed to send CoA Disconnect-Request
      nas_addr=192.168.1.1:3799
      error="context deadline exceeded"
```

---

## 📝 配置示例

### 用户配置（含 IPv6）

```json
{
  "username": "user001@example.com",
  "password": "SecurePass123",
  "profile_id": 1,
  "ip_addr": "10.0.1.100", // IPv4 地址
  "ipv6_addr": "2001:db8::100/64", // IPv6 地址
  "up_rate": 100, // 上行 100Mbps
  "down_rate": 100, // 下行 100Mbps
  "expire_time": "2025-12-31T23:59:59Z"
}
```

### NAS 配置（华为 ME60）

```json
{
  "name": "ME60-BJ-01",
  "identifier": "me60-bj-01",
  "ipaddr": "192.168.1.1",
  "secret": "YourRadiusSecret",
  "coa_port": 3799, // CoA 端口（可选，默认 3799）
  "model": "ME60",
  "vendor_code": "2011" // 华为厂商代码
}
```

### 套餐配置（含域属性）

```json
{
  "name": "企业套餐",
  "domain": "enterprise.example.com", // 华为域属性
  "ipv6_prefix": "2001:db8::/48", // IPv6 地址池
  "up_rate": 1000,
  "down_rate": 1000,
  "active_num": 2
}
```

---

## 🧪 测试验证

### 编译测试

```bash
CGO_ENABLED=0 go build -o ./release/toughradius main.go
```

✅ 编译通过，无错误

### 单元测试

```bash
go test ./internal/radiusd/plugins/auth/enhancers/... -v
```

✅ 59 个测试用例全部通过

### 功能测试

#### 测试 IPv6 下发

1. 创建用户，设置 `ipv6_addr = "2001:db8::1/64"`
2. 使用 radtest 工具认证
3. 抓包查看 Access-Accept 响应
4. 验证包含 `Framed-IPv6-Prefix` 属性
5. 华为设备验证 `Huawei-Framed-IPv6-Address` VSA

#### 测试 CoA 强制断线

1. 用户正常认证并上线
2. 查看"在线用户"列表
3. 执行强制下线操作
4. 检查数据库会话是否删除
5. 检查日志是否有 CoA 成功记录
6. 验证用户实际是否断线

---

## 🔍 故障排查

### IPv6 不下发

**问题**: 用户配置了 IPv6 但认证时未下发

**检查**:

1. 用户的 `ipv6_addr` 字段是否为空或 `N/A`
2. IPv6 地址格式是否正确（`2001:db8::1` 或 `2001:db8::1/64`）
3. 查看 RADIUS 日志是否有错误

### CoA 强制断线失败

**问题**: 删除会话后用户仍在线

**检查**:

1. NAS 设备是否开启 CoA 功能
2. NAS 的 CoA 端口是否为 3799
3. NAS 防火墙是否允许 UDP 3799
4. NAS 的 RADIUS Secret 是否正确
5. 查看 ToughRADIUS 日志中的 CoA 响应码

**日志示例**:

```bash
# 查看 CoA 相关日志
tail -f logs/toughradius.log | grep "CoA"

# 检查 NAS 配置
SELECT ipaddr, secret, coa_port FROM net_nas WHERE ipaddr = '192.168.1.1';
```

### 华为域属性问题

**问题**: 想要下发域属性但未实现

**临时方案**: 暂时不支持，需要等待架构调整

**长期方案**:

1. 修改 `AuthContext` 添加 `Profile` 字段
2. 在认证流程中加载用户的 `RadiusProfile`
3. 取消 `huawei_enhancer.go` 中的 TODO 注释

---

## 📊 性能影响

### IPv6 功能

- **CPU**: 几乎无影响（仅额外的属性设置）
- **内存**: +32 字节/会话（IPv6 地址存储）
- **网络**: +20-30 字节/认证响应（IPv6 属性）

### CoA 功能

- **响应时间**: API 立即返回（异步执行）
- **并发**: 每个 goroutine 占用约 2KB 内存
- **网络**: 每次断线发送约 100 字节 UDP 包

---

## 🔐 安全建议

1. **CoA Secret 安全**

   - 使用强 RADIUS Secret（至少 16 字符）
   - 定期更换 Secret
   - 不同 NAS 使用不同 Secret

2. **IPv6 地址规划**

   - 使用专用 IPv6 地址段
   - 避免使用公网可路由的 IPv6（除非必要）
   - 考虑使用 ULA（fd00::/8）用于内网

3. **CoA 访问控制**
   - NAS 设备应配置只接受来自 ToughRADIUS 服务器的 CoA
   - 启用防火墙规则限制 3799 端口访问

---

## 📚 相关 RFC 标准

- **RFC 2865**: RADIUS Authentication
- **RFC 2866**: RADIUS Accounting
- **RFC 3162**: RADIUS and IPv6
- **RFC 5176**: Dynamic Authorization Extensions to RADIUS (CoA)

---

## 🤝 技术支持

如有问题，请查看：

- GitHub Issues: https://github.com/talkincode/toughradius/issues
- 完整测试报告: `FEATURE_TEST_REPORT.md`
- 项目文档: `README.md`

---

**更新日期**: 2025 年 11 月 24 日  
**版本**: ToughRADIUS v9dev  
**作者**: GitHub Copilot
