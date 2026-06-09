# 协议与 RFC 索引

> English version: [Protocol & RFC Reference](../en/rfc-index.md)

ToughRADIUS 实现了标准 RADIUS、EAP、动态授权与安全传输等协议。本章是项目所依赖标准的
**精选、面向实现**的索引，将每个 RFC 映射到它在代码中的使用位置及
[路线图](https://github.com/talkincode/toughradius/blob/main/docs/roadmap.md)里程碑。

完整 RFC 文本归档于
[`docs/rfcs/`](https://github.com/talkincode/toughradius/tree/main/docs/rfcs)；
逐文件的原始目录见
[`docs/rfcs/README.md`](https://github.com/talkincode/toughradius/blob/main/docs/rfcs/README.md)。
当两者出现差异时，以本章的引用为准。

## 已实现的标准

### RADIUS 核心

- **RFC 2865** —— RADIUS 认证；基础的请求/响应协议。
- **RFC 2866** —— RADIUS 计费；会话开始 / 中间更新 / 结束记录。

### RADIUS + EAP 集成

- **RFC 3579** —— RADIUS 对 EAP 的支持（EAP-Message 与 Message-Authenticator）。
- **RFC 3580** —— IEEE 802.1X 的 RADIUS 使用指南。

### EAP 框架与方法

- **RFC 3748** —— 可扩展认证协议（EAP）。框架本身，同时定义了 EAP-MD5，即
  MD5-Challenge 方法（§5.4）。
- **RFC 5216** —— EAP-TLS；基于证书的双向认证（里程碑 M1）。
- **RFC 5281** —— EAP-TTLSv0；用 TLS 隧道承载内层 PAP / MS-CHAPv2（M9）。
- **RFC 2759** —— MS-CHAP-V2，用作 EAP-MSCHAPv2 与 PEAPv0 的内层方法（M8）。
  MS-MPPE 会话密钥按 RFC 2548 / RFC 5705 派生。

> PEAPv0 没有独立 RFC，遵循 Microsoft/Cisco 的 PEAP 定义，内层承载 EAP-MSCHAPv2。
> 它以兼容为先——MS-CHAPv2 交换存在类似 NTLMv1 的攻击面——在可控证书环境下优先选用
> EAP-TLS。

### 动态授权

- **RFC 5176** —— CoA 与 Disconnect-Request，取代 RFC 3576（里程碑 M2）。

### 安全传输

- **RFC 6614** —— 基于 TLS 的 RADIUS（RadSec）。
- **RFC 6613** —— 基于 TCP 的 RADIUS。

### 厂商私有属性

- **RFC 2548** —— Microsoft 厂商私有属性，包含承载 EAP 密钥材料的
  MS-MPPE-Send/Recv-Key 属性。

### IPv6

- **RFC 3162**、**RFC 4818**、**RFC 6911** —— RADIUS 的 IPv6 地址与委派前缀属性
  （里程碑 M3）。

## 路线图标准

| RFC | 标准 | 里程碑 |
| --- | --- | --- |
| RFC 9190（+ RFC 9427） | EAP-TLS 1.3 与 TLS 1.3 密钥派生 | M10 |
| RFC 7170 / RFC 9930 | TEAP v1 —— 隧道 EAP，machine + user chaining | M11 |
| RFC 5931 | EAP-PWD —— 基于口令，无需为每客户端签发证书 | M12 |

> **目录准确性说明。** 在 `docs/rfcs/` 中，文件 `rfc7542-eap-pwd.txt` 标注有误：
> **RFC 7542 是网络访问标识符（NAI）**，而 **EAP-PWD 是 RFC 5931**。同样，EAP-MD5
> 由 **RFC 3748 §5.4** 定义，而非 RFC 3851（一份 S/MIME 规范）。本章采用正确的引用。

## 另见

- [概述](./overview.md) —— 含 EAP 套件的能力概览。
- [文档地图](./documentation-map.md) —— 各源文档的存放位置。
- [RFC Editor](https://www.rfc-editor.org/) —— 权威的在线 RFC 文本。
