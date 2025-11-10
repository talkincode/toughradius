# RADIUS RFC 协议文档

本目录包含 RADIUS 相关的 RFC 协议文档，共 **50 个** RFC 文档，供开发和参考使用。

## 核心协议

- **RFC 2865** - Remote Authentication Dial In User Service (RADIUS)

  - RADIUS 核心协议，定义了基本的认证和授权机制
  - 文件：`rfc2865.txt`

- **RFC 2866** - RADIUS Accounting
  - RADIUS 计费协议，定义了会话计费和统计机制
  - 文件：`rfc2866.txt`

## 扩展协议

- **RFC 2867** - RADIUS Accounting Modifications for Tunnel Protocol Support

  - 隧道协议的计费扩展
  - 文件：`rfc2867.txt`

- **RFC 2868** - RADIUS Attributes for Tunnel Protocol Support

  - 隧道协议支持的 RADIUS 属性定义
  - 文件：`rfc2868.txt`

- **RFC 2869** - RADIUS Extensions

  - RADIUS 扩展协议，包括 EAP 支持等
  - 文件：`rfc2869.txt`

- **RFC 6929** - RADIUS Protocol Extensions

  - RADIUS 协议扩展（重要）
  - 文件：`rfc6929.txt`

- **RFC 7499** - Support of Fragmentation of RADIUS Packets
  - RADIUS 数据包分片支持
  - 文件：`rfc7499.txt`

## 网络接入服务器 (NAS)

- **RFC 2882** - Network Access Servers Requirements: Extended RADIUS Practices

  - 网络接入服务器需求和扩展 RADIUS 实践
  - 文件：`rfc2882.txt`

- **RFC 5607** - RADIUS Authorization for NAS Management
  - NAS 管理的 RADIUS 授权
  - 文件：`rfc5607.txt`

## IPv6 支持

- **RFC 3162** - RADIUS and IPv6

  - RADIUS 对 IPv6 的支持
  - 文件：`rfc3162.txt`

- **RFC 4818** - RADIUS Delegated-IPv6-Prefix Attribute

  - RADIUS IPv6 前缀委派属性
  - 文件：`rfc4818.txt`

- **RFC 5997** - Use of Status-Server Packets in RADIUS

  - RADIUS Status-Server 数据包的使用（支持 IPv6）
  - 文件：`rfc5997.txt`

- **RFC 6519** - RADIUS Extensions Support for Dual-Stack Lite

  - RADIUS 对双栈 Lite 的扩展支持
  - 文件：`rfc6519.txt`

- **RFC 6911** - RADIUS Attributes for IPv6 Access Networks

  - IPv6 接入网络的 RADIUS 属性
  - 文件：`rfc6911.txt`

- **RFC 6930** - RADIUS Attribute for IPv6 Rapid Deployment on IPv4 Infrastructures (6rd)

  - IPv6 快速部署的 RADIUS 属性
  - 文件：`rfc6930.txt`

- **RFC 5969** - IPv6 Rapid Deployment on IPv4 Infrastructures (6rd) -- Protocol Specification
  - 6rd 协议规范
  - 文件：`rfc5969.txt`

## EAP 相关

### EAP 核心协议

- **RFC 2284** - PPP Extensible Authentication Protocol (EAP) [已废弃]

  - EAP 原始规范（已被 RFC 3748 取代，保留用于历史参考）
  - 文件：`rfc2284.txt`

- **RFC 3748** - Extensible Authentication Protocol (EAP)

  - 可扩展认证协议，RADIUS EAP 的基础
  - 文件：`rfc3748.txt`

- **RFC 5247** - Extensible Authentication Protocol (EAP) Key Management Framework
  - EAP 密钥管理框架
  - 文件：`rfc5247.txt`

### RADIUS EAP 集成

- **RFC 3579** - RADIUS Support For Extensible Authentication Protocol (EAP)

  - RADIUS 对 EAP 的支持
  - 文件：`rfc3579.txt`

- **RFC 3580** - IEEE 802.1X Remote Authentication Dial In User Service (RADIUS) Usage Guidelines
  - IEEE 802.1X 与 RADIUS 集成的使用指南
  - 文件：`rfc3580.txt`

### EAP 认证方法

- **RFC 3851** - The Secure/Multipurpose Internet Mail Extensions (S/MIME) Version 3.1 Message Specification

  - 包含 EAP-MD5 Challenge 的相关定义
  - 文件：`rfc3851.txt`

- **RFC 5216** - The EAP-TLS Authentication Protocol

  - EAP-TLS 认证协议（基于 TLS 的 EAP 方法）
  - 文件：`rfc5216.txt`

- **RFC 5281** - Extensible Authentication Protocol Tunneled Transport Layer Security Authenticated Protocol Version 0 (EAP-TTLSv0)

  - EAP-TTLS 隧道 TLS 认证协议
  - 文件：`rfc5281.txt`

- **RFC 4186** - Extensible Authentication Protocol Method for Global System for Mobile Communications (GSM) Subscriber Identity Modules (EAP-SIM)

  - EAP-SIM 认证方法（GSM SIM 卡认证）
  - 文件：`rfc4186.txt`

- **RFC 4187** - Extensible Authentication Protocol Method for 3rd Generation Authentication and Key Agreement (EAP-AKA)

  - EAP-AKA 认证方法（3G UMTS 认证）
  - 文件：`rfc4187.txt`

- **RFC 4764** - The EAP-PSK Protocol: A Pre-Shared Key Extensible Authentication Protocol (EAP) Method

  - EAP-PSK 预共享密钥认证方法
  - 文件：`rfc4764.txt`

- **RFC 5448** - Improved Extensible Authentication Protocol Method for 3GPP Mobile Network Authentication and Key Agreement (EAP-AKA')

  - 改进的 EAP-IKEv2 认证方法
  - 文件：`rfc5448.txt`

- **RFC 7170** - Tunnel Extensible Authentication Protocol (TEAP) Version 1

  - TEAP 隧道 EAP 协议 v1（Cisco 提出的新一代隧道 EAP）
  - 文件：`rfc7170.txt`

- **RFC 7542** - The Network Access Identifier (NAI) for the Extensible Authentication Protocol (EAP)
  - EAP-PWD 基于密码的认证协议
  - 文件：`rfc7542.txt`

### EAP 扩展与应用

- **RFC 6124** - An EAP Authentication Method Based on the Encrypted Key Exchange (EKE) Protocol
  - 基于加密密钥交换的 EAP 认证方法（用于 Wi-Fi 集成）
  - 文件：`rfc6124.txt`

## 厂商特定属性

- **RFC 2548** - Microsoft Vendor-specific RADIUS Attributes

  - Microsoft 厂商特定 RADIUS 属性
  - 文件：`rfc2548.txt`

- **RFC 2759** - Microsoft PPP CHAP Extensions, Version 2 (MS-CHAP-V2)

  - Microsoft MS-CHAP-V2 认证协议
  - 文件：`rfc2759.txt`

- **RFC 4679** - DSL Forum Vendor-Specific RADIUS Attributes
  - DSL Forum（宽带论坛）厂商特定 RADIUS 属性
  - 文件：`rfc4679.txt`

## IEEE 802 网络支持

- **RFC 5904** - RADIUS Attributes for IEEE 802 Networks

  - IEEE 802 网络的 RADIUS 属性（包括 Wi-Fi）
  - 文件：`rfc5904.txt`

- **RFC 7268** - RADIUS Attributes for IEEE 802 Networks (更新版)
  - IEEE 802 网络的 RADIUS 属性（更新版本）
  - 文件：`rfc7268.txt`

## 隧道和 VPN 支持

- **RFC 2809** - Implementation of L2TP Compulsory Tunneling via RADIUS
  - 通过 RADIUS 实现 L2TP 强制隧道
  - 文件：`rfc2809.txt`

## 高级特性

- **RFC 4372** - Chargeable User Identity

  - 可计费用户标识
  - 文件：`rfc4372.txt`

- **RFC 3576** - Dynamic Authorization Extensions to Remote Authentication Dial In User Service (RADIUS)

  - RADIUS 动态授权扩展（早期版本）
  - 文件：`rfc3576.txt`

- **RFC 5176** - Dynamic Authorization Extensions to RADIUS

  - RADIUS 动态授权扩展（CoA/Disconnect，更新版本）
  - 文件：`rfc5176.txt`

- **RFC 4675** - RADIUS Attributes for Virtual LAN and Priority Support
  - RADIUS VLAN 和优先级支持属性
  - 文件：`rfc4675.txt`

## 认证扩展

- **RFC 4590** - RADIUS Extension for Digest Authentication

  - RADIUS 摘要认证扩展
  - 文件：`rfc4590.txt`

- **RFC 5090** - RADIUS Extension for Digest Authentication (更新版)
  - RADIUS 摘要认证扩展（更新版本）
  - 文件：`rfc5090.txt`

## 位置信息

- **RFC 5580** - Carrying Location Objects in RADIUS and Diameter
  - 在 RADIUS 和 Diameter 中携带位置对象
  - 文件：`rfc5580.txt`

## 移动网络支持

- **RFC 6572** - RADIUS Support for Proxy Mobile IPv6
  - RADIUS 对代理移动 IPv6 的支持
  - 文件：`rfc6572.txt`

## 安全传输

- **RFC 6613** - RADIUS over TCP

  - RADIUS over TCP（替代 UDP）
  - 文件：`rfc6613.txt`

- **RFC 6614** - Transport Layer Security (TLS) Encryption for RADIUS

  - RADIUS over TLS (RadSec)，提供加密传输
  - 文件：`rfc6614.txt`

- **RFC 7585** - Dynamic Peer Discovery for RADIUS/TLS and RADIUS/DTLS Based on the Network Access Identifier (NAI)
  - 基于 NAI 的 RADIUS/TLS 和 RADIUS/DTLS 动态对等发现
  - 文件：`rfc7585.txt`

## 实现指南

- **RFC 5080** - Common Remote Authentication Dial In User Service (RADIUS) Implementation Issues and Suggested Fixes
  - RADIUS 常见实现问题和建议修复方案
  - 文件：`rfc5080.txt`

## Diameter 协议（RADIUS 继任者）

- **RFC 7155** - Diameter Network Access Server Application
  - Diameter 网络接入服务器应用（RADIUS 的演进版本）
  - 文件：`rfc7155.txt`

## 使用说明

所有文档均为纯文本格式（.txt），可使用任何文本编辑器查看。

### ToughRADIUS 实现参考

ToughRADIUS 主要实现了以下 RFC 规范：

#### 已实现 - 核心协议

- ✅ RFC 2865 - RADIUS 核心认证协议
- ✅ RFC 2866 - RADIUS 计费协议
- ✅ RFC 2869 - RADIUS 扩展

#### 已实现 - EAP 支持

- ✅ RFC 3748 - EAP 核心协议
- ✅ RFC 3579 - RADIUS EAP 支持
- ✅ RFC 5216 - EAP-TLS
- ✅ RFC 2759 - MS-CHAP-V2 (用于 EAP-MSCHAPv2)

#### 已实现 - 高级功能

- ✅ RFC 6614 - RadSec (RADIUS over TLS)
- ✅ RFC 5176 - 动态授权（CoA/Disconnect，部分支持）
- ✅ RFC 2548 - Microsoft 厂商属性
- ✅ RFC 4675 - VLAN 支持

#### 计划支持

- 🔄 RFC 5281 - EAP-TTLS
- 🔄 RFC 7170 - TEAP
- 🔄 RFC 4186 - EAP-SIM
- 🔄 RFC 4187 - EAP-AKA
- 🔄 RFC 6613 - RADIUS over TCP
- 🔄 RFC 7499 - RADIUS 数据包分片

### 在线查看

您也可以在 IETF 官网在线查看这些 RFC 文档：
<https://www.rfc-editor.org/rfc/>

### RFC 文档分类汇总

| 分类            | 数量   | 说明                                                     |
| --------------- | ------ | -------------------------------------------------------- |
| 核心协议        | 2      | RFC 2865, 2866                                           |
| 扩展协议        | 5      | RFC 2867, 2868, 2869, 6929, 7499                         |
| NAS 支持        | 2      | RFC 2882, 5607                                           |
| IPv6 支持       | 7      | RFC 3162, 4818, 5997, 6519, 6911, 6930, 5969             |
| EAP 核心        | 3      | RFC 2284, 3748, 5247                                     |
| RADIUS-EAP 集成 | 2      | RFC 3579, 3580                                           |
| EAP 认证方法    | 9      | RFC 3851, 5216, 5281, 4186, 4187, 4764, 5448, 7170, 7542 |
| EAP 扩展        | 1      | RFC 6124                                                 |
| 厂商特定        | 3      | RFC 2548, 2759, 4679                                     |
| IEEE 802 网络   | 2      | RFC 5904, 7268                                           |
| 隧道/VPN        | 1      | RFC 2809                                                 |
| 高级特性        | 4      | RFC 4372, 3576, 5176, 4675                               |
| 认证扩展        | 2      | RFC 4590, 5090                                           |
| 位置信息        | 1      | RFC 5580                                                 |
| 移动网络        | 1      | RFC 6572                                                 |
| 安全传输        | 3      | RFC 6613, 6614, 7585                                     |
| 实现指南        | 1      | RFC 5080                                                 |
| Diameter        | 1      | RFC 7155                                                 |
| **总计**        | **50** |                                                          |

## 文档更新

最后更新时间：2025-11-10

如需下载其他 RADIUS 相关 RFC 文档，请访问：
<https://www.rfc-editor.org/search/rfc_search_detail.php?title=RADIUS>

## 重要 EAP 认证方法说明

### 常用 EAP 方法

1. **EAP-TLS (RFC 5216)** - 最安全，需要客户端证书
2. **EAP-TTLS (RFC 5281)** - 隧道方式，服务器证书 + 用户名密码
3. **EAP-MSCHAPv2 (基于 RFC 2759)** - 常与 PEAP 配合使用
4. **TEAP (RFC 7170)** - 新一代隧道 EAP，Cisco 推动

### 移动网络 EAP 方法

1. **EAP-SIM (RFC 4186)** - GSM SIM 卡认证
2. **EAP-AKA (RFC 4187)** - 3G/4G USIM 卡认证
3. **EAP-AKA' (RFC 5448)** - 改进的 EAP-AKA

### 其他 EAP 方法

1. **EAP-PSK (RFC 4764)** - 预共享密钥
2. **EAP-PWD (RFC 7542)** - 基于密码
3. **EAP-MD5 (RFC 3851)** - 不安全，不推荐生产环境使用

## 厂商特定扩展说明

### Microsoft 扩展

- **RFC 2548** - Microsoft 专有 RADIUS 属性

  - 包含 MS-CHAP 相关属性
  - MS-MPPE 加密密钥属性
  - Windows 域控集成相关属性

- **RFC 2759** - MS-CHAP-V2 协议
  - Windows 环境下广泛使用
  - 与 EAP-MSCHAPv2 相关

### DSL/宽带厂商

- **RFC 4679** - DSL Forum 属性
  - PPPoE 场景常用
  - 宽带接入服务提供商使用
  - 包含线路信息、速率限制等属性

### IEEE 802 网络（Wi-Fi）

- **RFC 5904/7268** - IEEE 802 网络属性
  - 无线网络 (Wi-Fi) 管理
  - VLAN 分配
  - QoS 策略
  - 802.11 特定参数

## 实用工具脚本

本目录包含自动下载脚本 `download-rfcs.sh`，可用于更新或重新下载所有 RFC 文档：

```bash
cd docs/rfcs
./download-rfcs.sh
```

脚本特性：

- ✅ 自动下载所有 50 个 RFC 文档
- ✅ 跳过已存在的文件（增量更新）
- ✅ 显示下载进度和统计信息
- ✅ 从官方 IETF RFC 编辑器网站下载
