# 实战手册：H3C / 中兴 / 爱快 / Cisco

> English version: [Cookbook: H3C, ZTE, iKuai & Cisco](../en/cookbook-vendors.md)
>
> 本章是[场景实战手册](./cookbook.md)的一部分，沿用其[阅读约定](./cookbook.md#每个场景的五段式)。

这四家厂商跑的是与两本旗舰手册**完全相同的运维场景**——PPPoE / IPoE 分级套餐、
地址池、到期断网、单账号并发限制、MAC 绑定、CoA / 强制下线与 FUP。这些场景的*机制*
（一个套餐档、一次到期、一个并发上限、一次 CoA 如何生效）都**一致**，因为它们来自共享的
`default_enhancer` 与各 checker——而非来自厂商。

因此本章不重复四遍完整的五段式，而是各厂商的**差异速查**：对每台设备只说清
**（1）** ToughRADIUS 下发哪个限速属性、单位倍率，**（2）** 请求 MAC / VLAN 如何解析
（从而哪些绑定可用），以及 **（3）** 端到端步骤该照哪本旗舰手册做。

> **请配合一本旗舰手册阅读。** 任何场景的完整分步，请照
> [MikroTik](./cookbook-mikrotik.md) 或 [Huawei](./cookbook-huawei.md) 手册中对应小节
> 操作；只有下面列出的下发属性不同。

## 该照哪本手册

| 你想要… | 照做 | 各厂商差异（本章） |
| --- | --- | --- |
| 分级套餐 + 地址池 + 到期 + 并发 | [MikroTik 场景 A](./cookbook-mikrotik.md#场景-apppoe-宽带-isp--分级套餐--地址池--到期断网--并发限制) 或 [Huawei 场景 A](./cookbook-huawei.md#场景-apppoe--ipoe-宽带--分级套餐峰值速率与-aaa-域) | 下发的**限速属性**与单位 |
| 线路防盗用（MAC / VLAN 绑定） | [Huawei 场景 B](./cookbook-huawei.md#场景-b线路防盗用--mac--vlan-绑定与双栈-ipv6) | **MAC 解析格式** + 是否支持 **VLAN 绑定** |
| 在线管控（CoA / 下线 / FUP） | [MikroTik 场景 C](./cookbook-mikrotik.md#场景-c在线管控--coa强制下线与-fup) | 无——CoA 与厂商无关（Session-Timeout + Filter-Id） |

> **并发上限、到期断网、地址池与 CoA 对这里每家厂商都一样**——它们由共享 checker /
> `default_enhancer` 执行，而非靠某个厂商 VSA。唯一真正的逐厂商变量就是下面的限速属性
> 和 MAC / VLAN 解析。

---

## H3C —— 厂商代码 25506

**限速属性**（锚定 `h3c_enhancer.go`）——与华为**同样的四元组形态**：

| 属性 | 取值 |
| --- | --- |
| `H3C-Input-Average-Rate` | `上行kbps × 1024`（bit/s）—— 用户上行 |
| `H3C-Input-Peak-Rate` | `上行kbps × 1024 × 4` |
| `H3C-Output-Average-Rate` | `下行kbps × 1024`（bit/s）—— 下行 |
| `H3C-Output-Peak-Rate` | `下行kbps × 1024 × 4` |

即「下行 30M」档（`30720` Kbps）会下发 `H3C-Output-Average-Rate = 31457280`、
峰值 `125829120`；所有值钳制到 Int32 上限。与华为不同，H3C **不**下发域或 IPv6 VSA，
只有限速四元组。

**请求解析**（`h3c_parser.go`）：MAC 来自 **`H3C-IP-Host-Addr`**（取末 17 位，即尾部的
`aa:bb:cc:dd:ee:ff`），该 VSA 缺失时回退到 `Calling-Station-Id`（`-`→`:`）；内 / 外层
VLAN 来自 `NAS-Port-Id`。**MAC 与 VLAN 绑定均支持。**

**设备侧（H3C Comware，参考，以实际固件为准）：**

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

> **坑**：把 NAS 登记为 *H3C*（不是 *Huawei*）。虽然限速算法一致，但 VSA 的**厂商 ID
> 不同**（25506 vs 2011）；把 H3C 设备登记成 Huawei，会收到它忽略的华为 VSA，限速将
> 不生效。

---

## ZTE 中兴 —— 厂商代码 3902

**限速属性**（锚定 `zte_enhancer.go`）——**两个属性，无峰值**：

| 属性 | 取值 |
| --- | --- |
| `ZTE-Rate-Ctrl-SCR-Up` | `上行kbps × 1024`（bit/s） |
| `ZTE-Rate-Ctrl-SCR-Down` | `下行kbps × 1024`（bit/s） |

「下行 30M」档下发 `ZTE-Rate-Ctrl-SCR-Down = 31457280`。没有峰值 / 突发属性——平均
速率即为上限。

**请求解析**（`zte_parser.go`）：MAC 来自 `Calling-Station-Id`，但中兴发的是**12 位
纯十六进制字符串**；ToughRADIUS 会重排成 `aa:bb:cc:dd:ee:ff`。VLAN 从 `NAS-Port-Id`
解析。**MAC 与 VLAN 绑定均支持**——但做 MAC 绑定时，所存 MAC 要用 `aa:bb:cc:dd:ee:ff`
形式（这是解析器产出的格式）。

**设备侧**：中兴 BRAS 与华为同样使用 radius-template + domain 模式——把认证 / 计费
模板绑定到 ToughRADIUS 地址、共享密钥与端口 1812 / 1813（以实际固件为准）。

---

## iKuai 爱快 —— 厂商代码 10055

国内常见的 SMB / SOHO 网关。

**限速属性**（锚定 `ikuai_enhancer.go`）——**两个属性，倍率不同**：

| 属性 | 取值 |
| --- | --- |
| `RP-Upstream-Speed-Limit` | `上行kbps × 8192`（= `× 1024 × 8`） |
| `RP-Downstream-Speed-Limit` | `下行kbps × 8192` |

即「下行 30M」档（`30720` Kbps）下发 `RP-Downstream-Speed-Limit = 251658240`。

> **坑 —— 高档位会被钳制。** 因为倍率是 `× 8192` 且取值钳制到 **Int32 上限
> （2147483647）**，任何高于约 **256 Mbps**（`262144` Kbps）的档位都会溢出并被**钳制**
> ——300M 档会按钳制上限下发，而非 300M。超高档位请改在 iKuai 设备侧做策略。

**请求解析**：iKuai 使用**默认解析器**——MAC 来自 `Calling-Station-Id`（`-`→`:`），
且**不解析 VLAN**（恒为 `0`）。故 **MAC 绑定可用，但 VLAN 绑定不可用**（VLAN 检查总被
跳过）。

**设备侧**：在 iKuai web 控制台进入 **认证计费 → RADIUS 计费**，设置服务器地址、端口
1812 / 1813 与共享密钥，再在 PPPoE 服务器设置中启用 RADIUS。

---

## Cisco —— 厂商代码 9

**限速属性**：**无**——Cisco 无可移植的数值限速 VSA，带宽仍在设备侧。`accept-cisco`
enhancer 会在套餐配置了地址池时下发一个非限速属性 `Cisco-AVPair="ip:addr-pool=<pool>"`，
与标准 `Framed-Pool` 互补；否则 Cisco NAS 只会收到 `default_enhancer.go` 下发的标准属性
（`Session-Timeout`、`Acct-Interim-Interval`、`Framed-Pool`、`Framed-IP-Address` 等）。认证
（PAP / CHAP / MS-CHAPv2 / EAP）、会话控制、计费与 CoA 都正常工作——**限速请在设备侧做**，
或用随仓库附带的 `Cisco-AVPair` 字典做自定义集成。

**请求解析**：默认解析器——MAC 来自 `Calling-Station-Id`，不解析 VLAN。故 **MAC 绑定
可用，VLAN 绑定不可用**。

**设备侧（Cisco IOS，参考，以实际平台为准）：**

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

> `aaa server radius dynamic-author` 启用 CoA / Disconnect（默认端口 3799）。不要指望
> Cisco 从 RADIUS 拿限速——用设备侧 service-policy / QoS 设置。

---

## 厂商能力矩阵

| 厂商 | 限速属性 | 单位倍率 | 峰值 | MAC 绑定 | VLAN 绑定 |
| --- | --- | --- | --- | --- | --- |
| H3C（25506） | `H3C-Input/Output-Average-Rate` + 峰值 | `× 1024`（bit/s） | ✅ 平均 × 4 | ✅ | ✅ |
| 中兴（3902） | `ZTE-Rate-Ctrl-SCR-Up/Down` | `× 1024`（bit/s） | ❌ | ✅ | ✅ |
| 爱快（10055） | `RP-Upstream/Downstream-Speed-Limit` | `× 8192` | ❌ | ✅ | ❌ |
| Cisco（9） | 无（设备侧 QoS） | — | — | ✅ | ❌ |

（对比：MikroTik 下发 Kbps 单位的 `Mikrotik-Rate-Limit` 字符串；华为下发 `× 1024`、
峰值 `× 4` 的限速四元组外加域 / IPv6。）

## 排障（症状 → 定位 → 解决）

- **限速不生效** → ① NAS 没按该厂商精确登记（VSA 厂商 ID 必须匹配——即便算法相同，
  H3C ≠ Huawei）；② Cisco **本就没有** RADIUS 限速 VSA——请在设备侧做 QoS；③ 设备侧
  单位搞错（这里除 iKuai 是 `× 8192` 外，其余都是 bit/s）。
- **iKuai 高档速率不对 / 被压低** → `× 8192` 的值溢出了 Int32 而被钳制；高档请在设备侧
  做。
- **iKuai / Cisco 的 VLAN 绑定从不触发** → 这是预期：它们的解析器不提取 VLAN（恒为
  `0`），故 VLAN 检查总被跳过。改用 MAC 绑定，或换支持 VLAN 解析的厂商（华为 / H3C /
  中兴）。
- **中兴 MAC 绑定从不匹配** → 把 MAC 按 `aa:bb:cc:dd:ee:ff` 存储（中兴发 12 位纯十六
  进制，ToughRADIUS 比对前会重排成冒号形式）。

---

## 相关章节

- [实战手册：MikroTik RouterOS](./cookbook-mikrotik.md) —— 完整 PPPoE / Hotspot / CoA
  剧本。
- [实战手册：华为 BRAS / NetEngine](./cookbook-huawei.md) —— 完整分级套餐 / 线路绑定 /
  CoA 剧本。
- [厂商对接指南](./vendor-guide.md) —— 各厂商属性参考卡。
- [常见问题解答](./faq.md) —— 更多跨场景排障问答。
