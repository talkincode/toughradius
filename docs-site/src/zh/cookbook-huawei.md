# 实战手册：华为 BRAS / NetEngine

> English version: [Cookbook: Huawei BRAS / NetEngine](../en/cookbook-huawei.md)
>
> 本章是[场景实战手册](./cookbook.md)的一部分，沿用其[五段式与阅读约定](./cookbook.md#每个场景的五段式)。

华为（厂商代码 **2011**）是国内运营商与企业网络中占主导地位的宽带 BRAS / 企业网关
（NetEngine / ME60 / 较老的 MA5200 系列）。ToughRADIUS 为其注册了专门的厂商增强器，
认证通过后下发：

- 一组**四属性限速四元组**（由 `huawei_enhancer.go` 产生）：
  `Huawei-Input-Average-Rate`、`Huawei-Input-Peak-Rate`、
  `Huawei-Output-Average-Rate`、`Huawei-Output-Peak-Rate`。
- `Huawei-Domain-Name` —— 仅当用户 / 策略设置了**域名（domain）**时下发。
- `Huawei-Framed-IPv6-Address` —— 仅当用户设置了**静态 IPv6** 时下发。
- 以及所有设备通用的标准属性（`Session-Timeout`、`Acct-Interim-Interval`、
  `Framed-Pool`、`Framed-IP-Address` 等），由 `default_enhancer.go` 产生。

请求侧，**华为解析器**（`huawei_parser.go`）从 `Calling-Station-Id` 提取 MAC
（把 `-` 规整为 `:`），从 `NAS-Port-Id` 提取内 / 外层 VLAN ID。正是这两项让华为设备
具备了 MAC 绑定**与** VLAN 绑定能力。

> **前置条件**：已在 **NAS 设备** 中把这台 BRAS 登记为 *厂商 = Huawei*、填写正确的
> 源 IP 与共享密钥；ToughRADIUS 可被访问（认证 1812、计费 1813）。若登记成
> `Standard`，认证仍可成功，但上述华为 VSA **全都不会**下发，VLAN 也不会被解析。

---

## 场景 A：PPPoE / IPoE 宽带 —— 分级套餐、峰值速率与 AAA 域

### 需求 / 场景

运营商或企业用华为 BRAS 跑宽带，需要多档套餐（如家庭版下行 30M / 上行 10M、
企业版下行 100M）。华为同时支持**平均**速率与**峰值**（突发）速率，并把用户归入
AAA **域（domain）**，以便 BRAS 套用对应的域策略。

### ToughRADIUS 侧

1. **为每档套餐建一个计费策略**（**计费策略 → 新建**）：
   - **上行 / 下行速率**：单位为 **Kbps**。下行 30M 应填 `30720`，**不是** `30`；
     上行 10M 填 `10240`。
   - **域名 domain**（可选）：华为 AAA 域名（如 `isp`），将作为 `Huawei-Domain-Name`
     下发，让 BRAS 把会话绑定到该域。
2. **创建用户**（**用户 → 新建**）：用户名 / 密码、对应**计费策略**、**过期时间**、
   状态 = 启用。用户级的**域名**字段会覆盖策略上的域名。

存储的 Kbps 速率如何变成四个华为 VSA（锚定 `huawei_enhancer.go`）：

| 属性 | 取值 | 说明 |
| ---- | ---- | ---- |
| `Huawei-Input-Average-Rate` | `上行kbps × 1024`（bit/s） | 「Input」= 进入 BRAS 的流量 = 用户**上行** |
| `Huawei-Input-Peak-Rate` | `上行kbps × 1024 × 4` | 平均值 × 4 |
| `Huawei-Output-Average-Rate` | `下行kbps × 1024`（bit/s） | 「Output」= 流向用户的流量 = **下行** |
| `Huawei-Output-Peak-Rate` | `下行kbps × 1024 × 4` | 平均值 × 4 |
| `Huawei-Domain-Name` | 用户 / 策略的域名 | **设置了才**下发 |
| `Session-Timeout`、`Acct-Interim-Interval`、`Framed-Pool`、`Framed-IP-Address` | 标准 | 由 `default_enhancer.go` 产生 |

> **华为独有的两个坑。** ① **单位**：限速 VSA 是 **bit/s** 而非 Kbps——ToughRADIUS
> 把存储的 Kbps 乘以 **1024**（二进制），且**峰值**是平均值的 **× 4**。所以「下行
> 30M」会被下发为 `Output-Average-Rate = 30720 × 1024 = 31457280`、
> `Output-Peak-Rate = 125829120`。② **方向命名**：华为「Input」是用户**上行**，
> 「Output」是**下行**（BRAS 视角）——与 MikroTik 的 `rx/tx` 用词相反，但物理含义
> 相同。四个值都会被钳制到 Int32 上限。

### 设备侧（华为 VRP，参考示例，以实际固件为准）

```text
# RADIUS 服务器模板
radius-server template tr-tmpl
 radius-server shared-key cipher <SECRET>
 radius-server authentication <TOUGHRADIUS_IP> 1812 weight 80
 radius-server accounting <TOUGHRADIUS_IP> 1813 weight 80
#
# 绑定到模板的 AAA 域（与 Huawei-Domain-Name 对应）
aaa
 authentication-scheme radius-auth
  authentication-mode radius
 accounting-scheme radius-acct
  accounting-mode radius
 domain isp
  authentication-scheme radius-auth
  accounting-scheme radius-acct
  radius-server tr-tmpl
```

> `Huawei-Output-Average-Rate` / 峰值会映射到 BRAS 的每用户 CAR / QoS，无需手工
> 为每个用户建队列。

### 验证

- **radtest（服务端）**：
  ```bash
  go run ./cmd/radtest auth -user <用户名> -pwd <密码> -nasip <NAS_IP> -secret <SECRET>
  ```
  成功时打印 `Access-Accept`，应能看到四个华为限速属性以及（若设置了）
  `Huawei-Domain-Name`。
- **BRAS 侧**：`display access-user username <name>`（在线会话、速率、域），
  `display aaa online-fail-record` 查看失败记录。
- **管理后台**：**在线会话** 页应出现该会话。

### 排障（症状 → 定位 → 解决）

- **限速不生效** → ① 确认 NAS 登记为 *Huawei*（登记成 Standard 不下发 VSA）；
  ② 记住取值是 bit/s（`× 1024`），若某档看起来「小了 1000 倍」，多半是设备侧把单位
  当成了 Kbps。
- **速率方向反了（上行被限到下行的值）** → Input/Output 方向搞混：华为
  **Input = 上行**、**Output = 下行**；检查哪个策略字段喂给了哪个属性。
- **域策略未生效** → 用户 / 策略没设**域名**（则不下发 `Huawei-Domain-Name`），
  或域名与 BRAS 上配置的 `domain` 不一致。
- **峰值 / 突发看起来过高** → 这是预期：增强器把峰值硬编码为平均值的 **× 4**；
  要调就调平均档，而非峰值。

---

## 场景 B：线路防盗用 —— MAC + VLAN 绑定与双栈 IPv6

### 需求 / 场景

宽带运营商希望通过**把账号绑定到接入线路**来防止共享 / 盗用——绑定用户 MAC 和 / 或
接入 VLAN（内 / 外层，即 QinQ），并为双栈业务下发**静态 IPv6**。

### ToughRADIUS 侧

华为把接入线路编码在解析器已读取的属性里：

- **MAC** 来自 `Calling-Station-Id`；
- **内 / 外层 VLAN** 来自 `NAS-Port-Id`。

随后由两个检查器执行绑定（锚定 `mac_bind_checker.go` / `vlan_bind_checker.go`）：

1. **MAC 绑定** —— 在用户 / 策略上开启 **绑定 MAC**，并在用户上存好允许的 **MAC**。
   认证时若请求 MAC 与所存 MAC 不同，会话被拒（`MacBindError`）。
2. **VLAN 绑定** —— 在用户 / 策略上开启 **绑定 VLAN**，并存好允许的 **VLAN ID 1 /
   VLAN ID 2**。若所存 VLAN 与解析出的不一致，会话被拒（`VlanBindError`）。
3. **静态 IPv6** —— 设置用户的 **IPv6 地址**；增强器随即下发
   `Huawei-Framed-IPv6-Address`（若写成 `地址/前缀长度`，下发前会去掉前缀长度）。

> **绑定只校验你存了的东西。** 每个检查器在其开关关闭时会**跳过**，*并且*在任一侧
> （所存值或请求值）为空 / 为 0 时也会跳过——它绝不会悄悄自动学习。所以要绑定一条
> 线路，必须（a）打开开关**并且**（b）在用户记录里填上 MAC / VLAN（通常取自首次
> 成功登录的抓取值）。

### 设备侧（华为 VRP，参考示例，以实际固件为准）

```text
# IPoE / PPPoE 接入，向 RADIUS 上报接入线路。
# BRAS 接入默认把内/外层 VLAN 放进 NAS-Port-Id、把用户 MAC 放进 Calling-Station-Id。
radius-server template tr-tmpl
 radius-server shared-key cipher <SECRET>
 radius-server authentication <TOUGHRADIUS_IP> 1812 weight 80
 radius-server accounting <TOUGHRADIUS_IP> 1813 weight 80
#
# 按设计需要在 BRAS 上启用 IPv6 地址/前缀下发
ipv6
```

### 验证

- 抓一次真实 `Access-Request`（或看认证日志），确认 `Calling-Station-Id` 与
  `NAS-Port-Id` 的取值，把它们**原样**存到用户上，再开启绑定。
- 用匹配的 MAC 做 **radtest** 成功；换一个 MAC / VLAN 则被拒。
- **BRAS 侧**：`display access-user username <name>` 可见绑定的线路与下发的 IPv6。

### 排障（症状 → 定位 → 解决）

- **绑定把合法用户也拒了** → 所存 MAC / VLAN 与 BRAS 实际发送的不一致。读取真实的
  `Calling-Station-Id` / `NAS-Port-Id`，按原值存好，再重新开启绑定。
- **VLAN 绑定从不触发** → 解析出的 VLAN 为 `0`（你的组网里 BRAS 没把 VLAN 放进
  `NAS-Port-Id`），或 **绑定 VLAN** 没开，或用户存的 VLAN 为 `0`；任一情况都会跳过。
- **没有下发 IPv6** → 用户没设静态 **IPv6 地址**，或该域在 BRAS 上没配 IPv6 / 双栈。
- **绑定被静默忽略** → 开关没开或所存值为空；检查器只有在开关开启**且**两侧都有值时
  才生效。

---

## 场景 C：在线管控 —— CoA、强制下线与 FUP

### 需求 / 场景

对华为 BRAS 上的在线用户做实时管控：缩短会话、强制重新认证、超出配额后限速
（FUP），或把某会话踢下线。

### ToughRADIUS 侧

在 **在线会话** 页选中某会话，可执行两类动作（锚定 `session_actions.go` 与
[管理系统用户手册 · 在线会话](./admin-manual.md#在线会话)）：

- **修改授权（CoA-Request）**：当前仅携带 **`Session-Timeout`（#27）** 和 / 或
  **`Filter-Id`（#11）**——**不会**改写华为的限速 VSA。
- **强制下线（Disconnect-Request）**：直接终止该会话。

> **华为上实现 FUP「在线变速」的正确路径。** 由于 CoA 不改写
> `Huawei-Output-Average-Rate`，在线变速沿用与其他厂商一致的与厂商无关路径：先在
> 计费策略 / 用户上改速率，**再对其发强制下线**；客户端重拨后即按新的限速四元组重新
> 授权。（部分华为固件也会响应某个厂商私有的 CoA 限速属性，但 ToughRADIUS **不**
> 下发这种属性——不要依赖它。）

操作员发起的 CoA / Disconnect 使用较短超时并自动重传一次，目标为 NAS 记录上的
**CoA 端口（默认 3799）**。

### 设备侧（华为 VRP，参考示例，以实际固件为准）

```text
# RADIUS 模板必须接受动态授权（CoA/DM）。
radius-server template tr-tmpl
 radius-server authorization <TOUGHRADIUS_IP> shared-key cipher <SECRET>
```

- 防火墙需放行从 ToughRADIUS 到 BRAS 的**入向 UDP 3799**。
- 若使用 `Filter-Id` 方案，需在 BRAS 预先定义同名的 ACL / 用户组（以实际固件为准）。

### 验证

- 在会话页点击 **修改授权** / **强制下线**，结果以通知形式反馈。
- **BRAS 侧**：强制下线后 `display access-user username <name>` 不再列出该会话，
  客户端随后自动重拨。

### 排障（症状 → 定位 → 解决）

- **CoA / Disconnect 无响应或超时** → ① RADIUS 模板未配 `radius-server
  authorization`；② 防火墙挡了 UDP 3799；③ NAS 记录的 CoA 端口与 BRAS 的
  authorization 端口不一致；④ 共享密钥不一致。
- **改了速率但在线用户没变速** → 这是预期：CoA 不改限速，需**强制下线**让客户端
  重拨后生效。
- **端口困惑** → 网上常说的 CoA `1700` 是某些客户端的本地端口；本系统与 RFC 5176
  使用 **3799**。

---

## 相关章节

- [厂商对接指南 · Huawei](./vendor-guide.md) —— 属性参考卡。
- [实战手册：MikroTik RouterOS](./cookbook-mikrotik.md) —— RouterOS 上的同类场景。
- [管理系统用户手册](./admin-manual.md) —— 用户 / 计费策略 / 在线会话 / CoA 表单。
- [常见问题解答](./faq.md) —— 更多跨场景排障问答。
