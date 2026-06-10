# 核心术语与概念

> English version: [Concepts & Terminology](../en/concepts.md)

本章解释贯穿 ToughRADIUS 的 AAA 核心术语，并说明每个概念在产品中的落点。
协议规范的权威清单及其与代码的对应关系，请见[协议与 RFC 索引](./rfc-index.md)。

## 一段话理解 AAA

RADIUS（Remote Authentication Dial In User Service）协议让网络设备向中心服务器
提出三个问题：*这个用户是谁*（**认证 Authentication**）、*他能做什么*
（**授权 Authorization**）、*他用了多少*（**计费 Accounting**）。ToughRADIUS
对三者全部作答：校验凭据、在 `Access-Accept` 中下发授权属性（IP 地址、带宽、
VLAN、会话限制），并从计费报文中记录用量。

## 核心术语

| 术语 | 在 ToughRADIUS 中的含义 |
| ---- | ----------------------- |
| **NAS**（网络接入服务器） | 发出 RADIUS 请求的网络设备——路由器、交换机、BRAS 或无线控制器。每台 NAS 须在管理界面的 **NAS 设备** 中登记 IP 地址、共享密钥和厂商代码；未登记地址的请求会被丢弃。 |
| **共享密钥**（shared secret） | 每台 NAS 一个的口令，用于认证 RADIUS 报文本身（RFC 2865 §3），NAS 侧与 ToughRADIUS 的 NAS 记录必须一致。 |
| **拨入用户 / RADIUS 用户** | **RADIUS 用户** 中的账号：用户名、密码、过期时间、可选的静态 IPv4/IPv6 地址、MAC/VLAN 绑定以及计费策略。 |
| **计费策略**（`RadiusProfile`） | **计费策略** 中的可复用模板：并发会话数（`active_num`）、上行/下行速率（单位 **Kbps**）、地址池、IPv6 前缀、域、MAC/VLAN 绑定开关。用户留空的字段自动继承策略值。 |
| **VSA**（厂商私有属性） | RADIUS 26 号属性（RFC 2865 §5.26），允许各厂商定义私有属性的容器。ToughRADIUS 内置 15+ 厂商的属性字典，并按 NAS 厂商代码下发对应厂商的授权属性。详见[厂商对接指南](./vendor-guide.md)。 |
| **厂商代码**（vendor code） | 选择厂商行为的 IANA 私有企业号，例如 `9` Cisco、`2011` 华为、`14988` MikroTik、`25506` H3C、`3902` 中兴、`10055` 爱快、`0` 标准。按 NAS 记录设置，决定使用哪个请求解析器及下发哪些 Access-Accept 属性。 |
| **CoA / Disconnect**（RFC 5176） | 动态授权：由服务器主动发起、修改在线会话（`CoA-Request`，如新的 `Session-Timeout` 或 `Filter-Id`）或终止会话（`Disconnect-Request`）的报文。在 **在线会话** 页面发起，发往 NAS 的 UDP 3799 端口（可按 NAS 覆盖）。 |
| **RadSec**（RFC 6614） | TCP 2083 端口上的 RADIUS over TLS。用 TLS 隧道包裹 RADIUS，使报文可以安全穿越不可信网络；纯 UDP RADIUS 仅靠共享密钥保护。 |
| **EAP**（RFC 3748） | 802.1X 网络使用的可扩展认证协议。ToughRADIUS 实现 EAP-MD5、EAP-MSCHAPv2、EAP-TLS、PEAPv0/EAP-MSCHAPv2 和 EAP-TTLS；当前方法与证书在 **系统配置 → RADIUS** 中选择。 |
| **计费会话** | NAS 通过 `Accounting-Request` 报文汇报的生命周期：Start → Interim-Update（多次）→ Stop（RFC 2866）。在线会话见 **在线会话**（`radius_online` 表），历史记录见 **计费记录**（`radius_accounting` 表）。 |
| **Acct-Interim-Interval** | NAS 发送中间计费更新的间隔（秒）。每个 Access-Accept 都会携带，取自 `radius.AcctInterimInterval` 配置。 |
| **Session-Timeout** | 会话最长剩余时间（秒）。ToughRADIUS 将其设为距用户过期时间的剩余秒数，确保会话不会超过账号有效期。 |
| **地址池**（`Framed-Pool`） | 配置在 NAS 上的命名 IP 池。ToughRADIUS 只下发池*名称*，实际地址由 NAS 分配；静态地址则直接通过 `Framed-IP-Address` / `Framed-IPv6-Address` 下发。 |
| **MAC 绑定** | 策略开启 `bind_mac` 后，首次出现的主叫 MAC 被记录到用户上，后续请求必须匹配。 |
| **VLAN 绑定** | 策略开启 `bind_vlan` 后，从 `NAS-Port-Id` 解析出的内/外层 VLAN ID 会被记录并校验。需要支持 VLAN 提取的厂商解析器（华为、H3C、中兴）。 |

## 一次认证请求的流转

```text
NAS ──Access-Request──▶ UDP 1812
        │
        ▼
  goroutine 池（ants，TOUGHRADIUS_RADIUS_POOL，默认 1024）
        │
        ▼
  1. NAS 查找 ─ 源 IP / identifier 必须匹配已登记的 NAS 记录
  2. 厂商解析器 ─ 提取 MAC（Calling-Station-Id）与 VLAN（NAS-Port-Id）
       · 华为 / H3C / 中兴解析器可提取 VLAN ID
       · 其余厂商走默认解析器（仅 MAC）
  3. 凭据校验 ─ PAP / CHAP / MS-CHAPv2，或 EAP 状态机
  4. 检查器 ─ 账号状态、过期时间、MAC 绑定、VLAN 绑定、并发数（active_num）
  5. Accept 增强器 ─ 标准属性（Session-Timeout、地址池、静态 IP、IPv6）
     以及按 NAS 厂商代码选择的厂商属性
        │
        ▼
NAS ◀─Access-Accept / Access-Reject──
```

认证失败会被归类到 Prometheus 风格计数器（`radus_reject_passwd_error`、
`radus_reject_expire` 等）并呈现在仪表盘，见[运维指南](./ops-guide.md#指标)。
拒绝延迟防护可减缓暴力破解：在 `radius.RejectDelayWindowSeconds`（默认 10 秒）
窗口内连续拒绝达到 `radius.RejectDelayMaxRejects`（默认 7 次）后，响应将被延迟。

## 密码协议速览

| 协议 | 密码如何传输 | 说明 |
| ---- | ------------ | ---- |
| **PAP** | 随请求传输，由共享密钥 XOR 保护（RFC 2865 §5.2） | 兼容性最好；建议配合 RadSec 或可信网络使用。 |
| **CHAP** | 不传输——MD5 质询/应答（RFC 2865 §5.3） | 要求服务器保存明文口令。 |
| **MS-CHAPv2** | 不传输——NT 哈希质询/应答（RFC 2548） | Windows 常用；存在众所周知的类 NTLMv1 攻击面。 |
| **EAP（隧道类）** | 在 TLS 隧道内传输（PEAP / EAP-TTLS）或以证书代替（EAP-TLS） | Wi-Fi / 802.1X 部署的推荐选择。 |

## 相关章节

- [协议与 RFC 索引](./rfc-index.md) —— 上文引用的全部 RFC 与代码的对应关系。
- [厂商对接指南](./vendor-guide.md) —— 各厂商属性与计算公式。
- [快速开始](./quickstart.md) —— 用 10 分钟实际体验这些概念。
