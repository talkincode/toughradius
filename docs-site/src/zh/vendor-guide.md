# 厂商对接指南

> English version: [Vendor Integration Guide](../en/vendor-guide.md)

ToughRADIUS 对所有设备都讲标准 RADIUS，并为其认识的厂商追加**厂商私有属性
（VSA）**。本章先介绍所有设备通用的对接步骤，再按厂商给出对接案例：
ToughRADIUS 下发什么、解析什么，以及设备侧的参考配置。

> **NAS 记录上的厂商代码决定一切。** 属性增强按管理界面中 NAS 设备记录的
> *厂商* 字段选择，而不是靠嗅探报文。如果把一台 MikroTik 登记为 `Standard`，
> 认证不受影响，但**不会**下发 `Mikrotik-Rate-Limit`（即不限速）。请先选对厂商。

> 📖 想要端到端的运维范例（PPPoE 分级套餐、Hotspot + MAC 认证、CoA / 强制下线）？
> 见[场景实战手册](./cookbook.md)。本章是属性参考卡，实战手册是照着做的剧本。

## 任意设备的通用对接步骤

1. **登记 NAS**：在 **NAS 设备 → 新建** 填写源 IP 地址（或 identifier）、
   共享密钥，并选择正确的*厂商*。
2. **设备指向服务器**：认证 UDP `1812`、计费 UDP `1813`、相同的共享密钥。
3. **可选 CoA**：ToughRADIUS 默认向 NAS 的 UDP `3799` 发送 CoA/Disconnect
   （RFC 5176）；若设备监听其他端口，请在 NAS 记录上设置 *CoA 端口*。
   每次交互最多等待 5 秒并重传 2 次。
4. **创建计费策略与用户**，用 `go run ./cmd/radtest auth …` 验证
   （见[快速开始](./quickstart.md)）。

### 所有厂商都会收到的标准属性

无论何种厂商，`Access-Accept` 中都可能携带：`Session-Timeout`（距账号过期的
秒数）、`Acct-Interim-Interval`、`Framed-Pool`、`Framed-IP-Address`（静态
IPv4）、`Framed-IPv6-Prefix` / `Framed-IPv6-Address`（RFC 6911）、
`Framed-IPv6-Pool`、`Delegated-IPv6-Prefix`（RFC 4818）与
`Delegated-IPv6-Prefix-Pool` —— 取决于用户/策略中设置了哪些字段。

### 限速单位

策略中的速率以 **Kbps** 存储，各厂商增强器按如下规则换算：

| 厂商 | 属性 | 下发值 |
| ---- | ---- | ------ |
| 华为（2011） | `Huawei-Input/Output-Average-Rate`、`Huawei-Input/Output-Peak-Rate` | 平均 = `速率_kbps × 1024`（bit/s）；峰值再 `× 4`；上限 Int32 |
| MikroTik（14988） | `Mikrotik-Rate-Limit` | 字符串 `"{上行}k/{下行}k"`，如 `51200k/102400k`（按路由器视角 rx/tx） |
| H3C（25506） | `H3C-Input/Output-Average-Rate` 及峰值 | 与华为相同的 ×1024 / ×4 规则 |
| 中兴（3902） | `ZTE-Rate-Ctrl-SCR-Up/Down` | `速率_kbps × 1024` |
| 爱快（10055） | `RP-Upstream/Downstream-Speed-Limit` | `速率_kbps × 1024 × 8`，上限 Int32 |
| Cisco（9）、标准（0） | —— | 仅标准属性；限速依赖设备侧策略，或基于内置 `Cisco-AVPair` 字典自行扩展 |

### 请求解析（MAC 与 VLAN）

默认解析器从 `Calling-Station-Id` 读取 MAC 地址。厂商解析器会在设备族存在
稳定编码时，额外处理请求侧 VSA 或 `NAS-Port-Id`：

- `slot/subslot/port:vlan[.vlan2]` —— 如 `3/0/1:2814.727`
- `vlanid=<n>;vlanid2=<n>;` —— 如 `slot=2;...;vlanid=503;vlanid2=100;`

| 厂商 | MAC 来源 | VLAN 来源 | 说明 |
| ---- | -------- | --------- | ---- |
| 华为（2011） | `Calling-Station-Id` | `NAS-Port-Id` | 支持上面两种常见编码。 |
| H3C（25506） | `Calling-Station-Id` | `NAS-Port-Id` | 支持上面两种常见编码。 |
| 中兴（3902） | `Calling-Station-Id` | `NAS-Port-Id` | 支持上面两种常见编码。 |
| Radback（2352） | `Mac-Addr` VSA，缺省回退到 `Calling-Station-Id` | `Bind-Dot1q-Vlan-Tag-Id` VSA，缺省回退到 `NAS-Port-Id` | 仅请求侧 parser；当前未内置 Radback 响应增强器。 |
| Alcatel（3041） | `AAT-User-MAC-Address` VSA，缺省回退到 `Calling-Station-Id` | 有值时解析 `NAS-Port-Id` | 仅请求侧 parser；响应属性仍按部署需求定制。 |
| Aruba/HPE（14823） | `Calling-Station-Id` | `Aruba-User-Vlan` VSA，缺省回退到 `NAS-Port-Id` | 同时具备 Access-Accept 增强器，见下文。 |
| Juniper（2636） | `Calling-Station-Id` | `Juniper-VoIP-Vlan` VSA，缺省回退到 `NAS-Port-Id` | 仅请求侧 parser；响应属性仍按部署需求定制。 |
| MikroTik、爱快、Cisco、标准 | `Calling-Station-Id` | —— | 无厂商专用 VLAN parser；请使用设备侧策略或二次开发。 |

只要设备发送可识别的 MAC，MAC 绑定即可工作；VLAN 绑定则要求 NAS 命中上表中
支持 VLAN 的解析器，并且请求报文使用对应编码。

> 下文设备侧片段均为**参考示例**——命令语法随型号与系统版本而异，请以厂商
> 文档为准。

---

## MikroTik（RouterOS）—— 厂商代码 14988

最经典的对接：PPPoE / Hotspot 配合 `Mikrotik-Rate-Limit`。

ToughRADIUS 下发 `Mikrotik-Rate-Limit = "{up}k/{down}k"`，RouterOS 将其应用为
动态 simple queue（rx/tx 按路由器视角，即先用户上行）。

```routeros
/radius add service=ppp,hotspot address=<TOUGHRADIUS_IP> secret=<SECRET> \
    timeout=3s
/radius incoming set accept=yes port=3799
/ppp aaa set use-radius=yes accounting=yes interim-update=5m
```

- `radius incoming accept=yes` 开启 UDP 3799 上的 CoA/Disconnect。
- Hotspot 场景：在 hotspot server profile 中启用 RADIUS
  （`/ip hotspot profile set ... use-radius=yes`）。

## 华为 —— 厂商代码 2011

典型的 BRAS（ME60/NE 系列）/ 汇聚场景。ToughRADIUS 下发速率四元组
（`Huawei-Input/Output-Average-Rate`，峰值 ×4）、`Huawei-Domain-Name`
（用户/策略配置了域时）以及静态 IPv6 的 `Huawei-Framed-IPv6-Address`。
华为解析器可从 `NAS-Port-Id` 提取 VLAN，MAC 与 VLAN 绑定均可用。

```text
radius-server template tr_tpl
 radius-server shared-key cipher <SECRET>
 radius-server authentication <TOUGHRADIUS_IP> 1812
 radius-server accounting <TOUGHRADIUS_IP> 1813
#
aaa
 authentication-scheme auth_radius
  authentication-mode radius
 accounting-scheme acct_radius
  accounting-mode radius
  accounting interim interval 5
 domain default
  authentication-scheme auth_radius
  accounting-scheme acct_radius
  radius-server tr_tpl
```

如需 CoA/Disconnect，请在设备上启用 RADIUS 动态授权扩展
（`radius-server authorization` 指向服务器地址）。

## Cisco —— 厂商代码 9

Cisco 设备使用标准属性认证（PAP / CHAP / MS-CHAPv2 / EAP 均可；会话、计费、
CoA 同样可用）。默认不下发 Cisco 私有属性——带宽策略请在设备侧实施，或基于
内置的 `Cisco-AVPair` 字典自行扩展。

```text
aaa new-model
radius server TOUGHRADIUS
 address ipv4 <TOUGHRADIUS_IP> auth-port 1812 acct-port 1813
 key <SECRET>
aaa authentication ppp default group radius
aaa accounting network default start-stop group radius
aaa server radius dynamic-author
 client <TOUGHRADIUS_IP> server-key <SECRET>
```

`aaa server radius dynamic-author` 启用 CoA/Disconnect（默认端口 3799）。

## H3C —— 厂商代码 25506

速率语义与华为一致（`H3C-Input/Output-Average-Rate` ×1024，峰值 ×4）。
H3C 解析器可提取 VLAN，支持 VLAN 绑定。

```text
radius scheme tr_scheme
 primary authentication <TOUGHRADIUS_IP> 1812
 primary accounting <TOUGHRADIUS_IP> 1813
 key authentication simple <SECRET>
 key accounting simple <SECRET>
 user-name-format without-domain
#
domain default enable system
 authentication ppp radius-scheme tr_scheme
 accounting ppp radius-scheme tr_scheme
```

## 中兴 —— 厂商代码 3902

ToughRADIUS 下发 `ZTE-Rate-Ctrl-SCR-Up/Down`（速率 ×1024），并从
`NAS-Port-Id` 解析 VLAN。中兴 BRAS 的配置与华为类似，采用 radius 模板 + 域
的模式：将认证/计费模板绑定到服务器地址、密钥及 1812/1813 端口即可。

## 爱快（iKuai）—— 厂商代码 10055

国内常见的中小企业网关。ToughRADIUS 下发
`RP-Upstream-Speed-Limit` / `RP-Downstream-Speed-Limit`
（= `速率_kbps × 8192`，上限 Int32）。在爱快 Web 控制台：**认证计费 →
RADIUS 计费** —— 填写服务器地址、端口 1812/1813 与共享密钥，并在 PPPoE
服务端设置中启用 RADIUS。

## Aruba/HPE —— 厂商代码 14823

Aruba/HPE 同时具备请求解析与 `Access-Accept` 响应增强能力。请求解析器会优先
读取报文中的 `Aruba-User-Vlan`，并保留与其他 VLAN-aware 解析器一致的
`NAS-Port-Id` 回退。

`Access-Accept` 中 ToughRADIUS 会下发：

| 属性 | 来源 | 边界 / no-op 规则 |
| ---- | ---- | ----------------- |
| `Aruba-User-Vlan` | 用户或策略的 `Vlanid1` | 仅在 VLAN ID 为 `1..4094` 时下发；缺省、`0`、`4095` 或负值都会省略。`Vlanid2` 没有 Aruba 属性映射，不会下发。 |
| `Aruba-User-Role` | 用户或策略的 `Domain` | 继承后的 domain / role 字段非空且不为 `N/A` 时下发。 |

非 Aruba NAS 不会追加 Aruba VSA；字段未设置时也不会写入占位值。

## 标准 / 其他设备 —— 厂商代码 0

任何符合 RFC 的 NAS（pfSense、strongSwan、各类 Wi-Fi 控制器等）都能以
`Standard` 厂商代码对接 ToughRADIUS：凭据校验、会话控制、计费、IPv4/IPv6
属性一应俱全——只是没有私有限速属性。代码库还内置了更多厂商的属性**字典**
（Microsoft、F5、PfSense、Hillstone 等）供二次开发。对 Juniper、Alcatel、
Aruba、Radback 而言，当前能力不再只是“仅字典”：请以上文的请求解析器与
Aruba 增强器边界为准。单纯字典仍只定义属性；只有注册 parser 或 enhancer 后，
才会解析请求或增强响应。

## 对接排障

| 现象 | 可能原因 |
| ---- | -------- |
| 设备收到 `Access-Accept` 但没有限速 | NAS 记录厂商为 `Standard`，或设备忽略该 VSA——先检查厂商代码 |
| 所有请求被静默丢弃 | 源 IP 未登记为 NAS，或共享密钥不一致 |
| VLAN 绑定始终不匹配 | NAS 厂商未命中上表中的 VLAN-aware parser，或请求使用了非预期的 VSA / `NAS-Port-Id` 格式 |
| CoA/Disconnect 超时 | 设备未开启 CoA 监听，或端口非默认——在 NAS 记录上设置 *CoA 端口* |

更多见[常见问题](./faq.md)。
