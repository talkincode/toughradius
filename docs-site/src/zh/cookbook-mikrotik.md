# 实战手册：MikroTik RouterOS

> English version: [Cookbook: MikroTik RouterOS](../en/cookbook-mikrotik.md)
>
> 本章是[场景实战手册](./cookbook.md)的一部分，沿用其[五段式与阅读约定](./cookbook.md#每个场景的五段式)。

MikroTik RouterOS（厂商代码 **14988**）是最常见的对接对象。ToughRADIUS 为其
注册了专门的厂商增强器，认证通过后下发：

- `Mikrotik-Rate-Limit = "{上行}k/{下行}k"` —— 字符串限速（由
  `mikrotik_enhancer.go` 产生；`rx/tx` 按**路由器视角**，即第一段是用户上行）。
- 以及所有设备通用的标准属性：`Session-Timeout`、`Acct-Interim-Interval`、
  `Framed-Pool`、`Framed-IP-Address` 等（由 `default_enhancer.go` 产生）。

> **前置条件**：已在 **NAS 设备** 中把这台路由器登记为 *厂商 = MikroTik*、填写
> 正确的源 IP 与共享密钥；ToughRADIUS 可被设备访问（认证 1812、计费 1813）。
> 若登记成 `Standard`，认证仍可成功，但**不会**下发 `Mikrotik-Rate-Limit`。

---

## 场景 A：PPPoE 宽带 ISP —— 分级套餐 + 地址池 + 到期断网 + 并发限制

### 需求 / 场景

一个小区 / 小型 ISP 用 PPPoE 拨号上网，需要：多档套餐（如家庭版下行 30M、
企业版下行 100M），账号到期自动断网且无法再拨，每账号限定并发会话数（防一号
多拨），IP 由统一地址池分配。

### ToughRADIUS 侧

1. **为每档套餐建一个计费策略**（**计费策略 → 新建**）：
   - **上行 / 下行速率**：单位为 **Kbps**。下行 30M 应填 `30720`，**不是** `30`；
     上行 10M 填 `10240`。
   - **并发数 `active_num`**：如 `1` 表示单账号最多 1 个在线会话（`0` = 不限）。
   - **地址池**：填地址池名（如 `pppoe-pool`），需与路由器上的 `/ip pool` 同名。
2. **创建用户**（**用户 → 新建**）：用户名 / 密码、选择对应**计费策略**（速率、
   并发、地址池由策略继承）、设置**过期时间**、状态 = 启用。如需固定 IP，可在用户
   上设置静态 IPv4（将覆盖地址池）。
3. **计费上报间隔**：系统配置项 `radius.AcctInterimInterval`（默认 `120` 秒）决定
   下发的 `Acct-Interim-Interval`。

认证通过后，`Access-Accept` 实际携带（锚定代码）：

| 属性 | 取值 | 来源 |
| ---- | ---- | ---- |
| `Mikrotik-Rate-Limit` | `"{上行kbps}k/{下行kbps}k"`，如 `10240k/30720k` | `mikrotik_enhancer.go` |
| `Session-Timeout` | 距过期时间的剩余秒数（到点断开当前会话） | `default_enhancer.go` |
| `Acct-Interim-Interval` | `radius.AcctInterimInterval`（默认 120） | `default_enhancer.go` |
| `Framed-Pool` | 策略 / 用户设置的地址池名（设置了才下发） | `default_enhancer.go` |
| `Framed-IP-Address` | 用户的静态 IPv4（设置了才下发） | `default_enhancer.go` |

> **并发限制不是靠下发属性实现的**：它由 `online_count_checker` 在认证阶段依据
> `active_num` 与当前在线数判定——超限的新会话直接被拒（`Access-Reject`）。

### 设备侧（RouterOS，参考示例，以实际固件为准）

```routeros
# 指向 ToughRADIUS（认证/计费同一共享密钥）
/radius add service=ppp address=<TOUGHRADIUS_IP> secret=<SECRET> timeout=3s
/radius incoming set accept=yes port=3799

# 开启 RADIUS 认证 + 计费 + 周期上报
/ppp aaa set use-radius=yes accounting=yes interim-update=5m

# 地址池名必须与下发的 Framed-Pool 一致
/ip pool add name=pppoe-pool ranges=10.10.0.2-10.10.255.254

# PPPoE 服务（remote-address 留给 RADIUS 的 Framed-Pool/Framed-IP 决定）
/ppp profile add name=radius-pppoe local-address=10.10.0.1
/interface pppoe-server server add service-name=isp interface=<bridge> \
    default-profile=radius-pppoe disabled=no
```

> 限速无需手工建队列：路由器收到 `Mikrotik-Rate-Limit` 后会自动生成动态
> simple queue。

### 验证

- **radtest（服务端）**：
  ```bash
  go run ./cmd/radtest auth -server <TOUGHRADIUS_IP> -nas-ip <NAS_IP> \
    -username <用户名> -password <密码> -secret <SECRET>
  ```
  成功时打印 `Access-Accept`，应能看到 `Mikrotik-Rate-Limit`、`Session-Timeout`，
  以及（若设置了池）`Framed-Pool`。
- **路由器侧**：`/ppp active print`（在线连接）、`/queue simple print`（动态队列
  及其速率）、`/log print where topics~"radius"`。
- **管理后台**：**在线会话** 页应出现该会话（Framed IP、时长、上下行流量）。

### 排障（症状 → 定位 → 解决）

- **限速完全不生效** → ① 确认 NAS 登记为 *MikroTik*（登记成 Standard 不下发 VSA）；
  ② 速率单位填错（填了 `30` 而非 `30720`）；③ 方向记反（`rx/tx` 是路由器视角，
  第一段是用户**上行**）。
- **拨号成功但拿不到 IP / 地址不对** → `Framed-Pool` 名与 `/ip pool` 名不一致，
  或策略未设地址池；改为同名，或在用户上设静态 IP。
- **账号到期后仍能上网** → `Session-Timeout` 只在到点时断开**当前**会话；已在线
  的旧会话需等超时或在会话页手动**强制下线**。到期后**再次拨号**会被
  `expire_checker` 拒（拒绝指标计入 `user expire`）。
- **第二条连接被拒** → `active_num=1` 时 `online_count_checker` 拒绝并发，这是预期
  行为；要允许多拨，调大计费策略的并发数。
- **连续认证失败后服务端变慢 / 不回包** → `reject_delay_guard` 在某用户名连续被拒
  超过阈值（默认 7 次）后引入延迟以抑制暴力破解，稍后自动恢复。

---

## 场景 B：Hotspot + MAC 认证

### 需求 / 场景

公共 WiFi / Hotspot 环境下，希望部分已登记设备（打印机、IoT、长期访客设备）
**免门户**、直接按 MAC 地址放行并限速。

### ToughRADIUS 侧

ToughRADIUS 判定一次请求是否为 **MAC 认证**的条件是（锚定代码
`auth_stages.go` / `eap_helper.go`）：

> 当请求中解析出的 MAC 地址 **等于用户名** 时，即视为 MAC 认证；此时密码比对使用
> 该用户记录的 **MAC 地址** 字段，而非普通密码。

因此配置方式为：

1. **创建用户**，**用户名 = 设备 MAC**（字符串需与路由器实际发送的格式完全一致），
   并在用户记录的 **MAC 地址** 字段填入同一 MAC。
2. 给该用户分配计费策略（速率、并发照常生效）。

> **最大的坑是 MAC 格式**：大小写与分隔符必须与 RouterOS hotspot 发送的
> `User-Name` 完全一致（ToughRADIUS 按字符串精确匹配）。RouterOS 的发送格式受
> `mac-auth-mode` 等设置影响——**存什么就必须发什么**。

### 设备侧（RouterOS，参考示例，以实际固件为准）

```routeros
/radius add service=hotspot address=<TOUGHRADIUS_IP> secret=<SECRET> timeout=3s

# 在 hotspot server profile 上启用 RADIUS 与 MAC 登录
/ip hotspot profile set <profile> use-radius=yes login-by=mac,http-chap
# mac-auth-mode / mac-auth-password 视固件而定
```

### 验证

- **radtest**：用 `-user <MAC> -pwd <MAC>` 模拟一次 MAC 认证，观察是否
  `Access-Accept`。
- **路由器侧**：`/ip hotspot active print` 应出现该设备；`/log print` 查看认证日志。
- **管理后台**：在线会话页出现对应会话。

### 排障（症状 → 定位 → 解决）

- **MAC 认证总是失败** → 用户名 / MAC 字段与路由器实际发送的 MAC 字符串
  格式不一致（大小写、分隔符）。抓一次请求看 `User-Name` 原文，按原文重建用户。
- **被当成普通账号在认证** → 只有「请求 MAC == 用户名」才触发 MAC 认证路径；若
  hotspot 未按 MAC 登录，请求会走门户用户名 / 密码逻辑。

---

## 场景 C：在线管控 —— CoA、强制下线与 FUP

### 需求 / 场景

对在线用户做实时管控：超出流量配额后限速（FUP）、临时强制重新认证，或直接把
某会话踢下线。

### ToughRADIUS 侧

在 **在线会话** 页选中某会话，可执行两类动作（锚定 `session_actions.go` 与
[管理系统用户手册 · 在线会话](./admin-manual.md#在线会话)）：

- **修改授权（CoA-Request）**：当前仅携带 **`Session-Timeout`（#27）** 和 / 或
  **`Filter-Id`（#11）**。
  - 用 `Session-Timeout` 缩短会话寿命，促使客户端尽快重新认证。
  - 用 `Filter-Id` 让 RouterOS 套用一个**预先定义好的** filter / address-list
    （例如一条限速或限站规则）。
- **强制下线（Disconnect-Request）**：直接终止该会话（操作需确认）。

> **实现 FUP「在线变速」的正确路径**：ToughRADIUS 的 CoA **不直接改写**
> `Mikrotik-Rate-Limit`。要让某用户实时变速，标准做法是——先在计费策略 / 用户上
> 改速率，**再对其发强制下线**；客户端自动重拨后即按新速率重新授权。这条路径与
> 系统代码能力一致，适用于任意厂商。

操作员发起的 CoA / Disconnect 使用较短超时并自动重传一次，目标为 NAS 记录上的
**CoA 端口（默认 3799）**。

### 设备侧（RouterOS，参考示例，以实际固件为准）

```routeros
# 必须开启，否则收不到 CoA/Disconnect
/radius incoming set accept=yes port=3799
```

- 防火墙需放行**入向 UDP 3799**（从 ToughRADIUS 到路由器）。
- 若使用 `Filter-Id` 方案，需在 RouterOS 预先定义同名的 filter / queue /
  address-list（以实际固件为准）。

### 验证

- 在会话页点击 **修改授权** / **强制下线**，结果以通知形式反馈。
- 路由器侧 `/log print` 可见 incoming 请求；**强制下线**后 `/ppp active print`
  中该会话消失，随后客户端自动重拨。

### 排障（症状 → 定位 → 解决）

- **CoA / Disconnect 无响应或超时** → ① RouterOS 未 `radius incoming accept=yes`；
  ② 防火墙挡了 UDP 3799；③ NAS 记录的 CoA 端口与设备 `incoming port` 不一致；
  ④ 共享密钥不一致。
- **改了速率但在线用户没变速** → 这是预期：CoA 不改限速，需**强制下线**让客户端
  重拨后生效。
- **端口困惑** → 网上常说的 CoA `1700` 是某些客户端的本地端口；本系统与 RFC 5176
  使用 **3799**。

---

## 场景 D：WPA2/WPA3 企业级 Wi-Fi —— 802.1X EAP 透传到 ToughRADIUS

### 需求 / 场景

你在 MikroTik AP 上跑企业级 Wi-Fi（`WPA2-EAP` / `WPA3-EAP`，即 802.1X），希望把
账号、证书与策略**集中到一处**：ToughRADIUS。员工既可以用**客户端证书**登录
（EAP-TLS，无需密码），也可以用**用户名 + 密码**登录（面向 Windows/AD 风格客户端
的 PEAP-MSCHAPv2，或面向老账号库 / LDAP 后端的 EAP-TTLS）。

> **本场景与
> [`multiduplikator/mikrotik_EAP`](https://github.com/multiduplikator/mikrotik_EAP)
> 的区别。**
> 那篇广为流传的指南让 RouterOS **自己**充当 EAP 服务器（ROS6 直连证书库，或 ROS7
> 用 **User Manager v5**）。而这里 MikroTik 只是**认证者（authenticator）**：用
> `eap-methods=passthrough` 把 802.1X/EAP 会话**透传**给 ToughRADIUS，由
> **ToughRADIUS 终结 EAP**。由此带来两个必须提前规划的差异：
>
> 1. **信任锚转移到 ToughRADIUS。** 客户端不再信任 `RouterCA`，而要信任
>    **签发 ToughRADIUS 服务器证书（`EapTlsCertFile`）的 CA**。你要把*那个* CA 分发
>    到客户端设备。
> 2. **路由器不再保存任何用于认证的用户/证书材料。** 所有身份都在 ToughRADIUS 上，
>    因此账号生命周期、限速策略、在线会话视图与计费都直接复用 ToughRADIUS。

### ToughRADIUS 侧

#### 0. 准备证书（磁盘上的 PEM 文件）

ToughRADIUS 从磁盘加载 PEM 文件（路径由下方配置项指定）。用 `openssl` 建一个最小自
建 CA —— 用 EC 密钥可让 TLS 记录更小，对 EAP 分片更友好：

```bash
# 根 CA（把 ca.pem 分发到每台客户端设备）
openssl ecparam -name prime256v1 -genkey -noout -out ca.key
openssl req -x509 -new -key ca.key -sha256 -days 3650 -out ca.pem \
  -subj "/CN=ToughRADIUS EAP Root CA"

# RADIUS 服务器证书（CN/SAN 即客户端钉住的服务器身份）
openssl ecparam -name prime256v1 -genkey -noout -out server.key
openssl req -new -key server.key -out server.csr -subj "/CN=radius.example.com"
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
  -days 825 -sha256 -out server.pem \
  -extfile <(printf "subjectAltName=DNS:radius.example.com\nextendedKeyUsage=serverAuth\nkeyUsage=digitalSignature,keyEncipherment")

# EAP-TLS 客户端证书（仅 EAP-TLS 用户需要）。SAN email 将成为 RADIUS 身份
# —— 见下方身份绑定规则。
openssl ecparam -name prime256v1 -genkey -noout -out alice.key
openssl req -new -key alice.key -out alice.csr -subj "/CN=alice"
openssl x509 -req -in alice.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
  -days 825 -sha256 -out alice.pem \
  -extfile <(printf "subjectAltName=email:alice@example.com\nextendedKeyUsage=clientAuth\nkeyUsage=digitalSignature")

# 把客户端 私钥+证书+CA 打包成 PKCS#12，便于 Windows/Android/iOS 导入
openssl pkcs12 -export -inkey alice.key -in alice.pem -certfile ca.pem \
  -out alice.p12 -passout pass:'<长口令>'
```

> **EAP-TLS 身份绑定（锚定 `tlsengine/identity.go`，RFC 5216 §5.2）。**
> 客户端证书通过链校验后，ToughRADIUS 按此顺序推导 **Peer-Id**：**SAN `rfc822Name`
>（email）→ SAN `dnsName` → 主题 `CN`**，并要求其与 RADIUS `User-Name`
> 大小写不敏感地相等。当存在 SAN 时，**不再**接受 `CN` 作为备选。因此用上面这张证书，
> ToughRADIUS 里要匹配的用户名是 **`alice@example.com`**，不是 `alice`。

#### 1. 上传材料并设置配置项

把 `server.pem`、`server.key`、`ca.pem` 拷到 ToughRADIUS 主机（如
`/var/toughradius/eap/`），然后在 **系统配置 → RADIUS** 中设置以下项（锚定
`internal/app/config_schemas.json` / `eap/handlers/tls_config.go`）：

| 配置项（`radius.*`） | EAP-TLS | PEAP / TTLS | 说明 |
| --- | --- | --- | --- |
| **EAP 认证方式**（`EapMethod`） | `eap-tls` | `eap-peap` / `eap-ttls` | EAP-Identity 阶段**首先提供**的方式。 |
| **启用的 EAP 处理器**（`EapEnabledHandlers`） | `eap-tls,eap-peap,eap-ttls` | 同上 | 允许清单；`*` 表示全部。客户端可 **NAK** 改用其它方式，但仅当此处启用了该方式才生效。 |
| **EAP-TLS 服务器证书**（`EapTlsCertFile`） | `/var/toughradius/eap/server.pem` | 同上 | 三种隧道方式**都**出示它。 |
| **EAP-TLS 服务器私钥**（`EapTlsKeyFile`） | `/var/toughradius/eap/server.key` | 同上 | —— |
| **EAP-TLS 客户端 CA 证书**（`EapTlsCaFile`） | `/var/toughradius/eap/ca.pem` | *(不使用)* | **EAP-TLS 必需**（校验客户端证书）；PEAP/TTLS 仅服务器认证，忽略它。 |
| **EAP-TLS 最低 TLS 版本**（`EapTlsMinVersion`） | `1.2` | `1.2` | PEAP/TTLS 无论如何都钉死在 **TLS 1.2**。 |

> 证书/私钥/CA 会在**两次握手之间**重新读取，因此可不重启 RADIUS 服务直接轮换。
> 在所需文件未全部配置前，EAP 会**安全拒绝**（`ErrTLSNotConfigured`），而不会在缺少
> 信任锚的情况下放行 —— 半配置状态绝不会放任何人进来。

#### 2. 创建用户

| 方式 | 要创建的用户名 | 密码字段 | 出处 |
| --- | --- | --- | --- |
| **EAP-TLS** | 证书 **Peer-Id**，如 `alice@example.com` | **不使用**（任意占位） | `tlsengine/identity.go` 中匹配证书身份 |
| **PEAP-MSCHAPv2** | 登录名，如 `bob` | 用户真实密码 | `peap_inner.go` 内层 MSCHAPv2 比对 `GetPassword` |
| **EAP-TTLS**（PAP 或 MS-CHAP-V2） | 登录名，如 `carol` | 用户真实密码 | `ttls_inner.go` 内层 PAP/MSCHAPv2 比对 `GetPassword` |

每个用户仍需 **状态 = 启用**、**未过期** 且绑定 **限速策略**（速率 / 地址池 / 并发与
场景 A 完全一致）。

> **透传场景最大的坑 —— 外层（匿名）身份。**
> ToughRADIUS 用**外层 `User-Name`** 加载用户记录，并从*该*记录取密码；把单独的
> *匿名*外层身份映射到真实账号的能力尚未实现（推迟到 M8.4）。因此
> **PEAP / TTLS 的外层身份必须等于真实用户名** —— 在客户端上把“匿名身份”**留空**
>（此时它会在明文外层身份里发送真实用户名），**或**把它设成与用户名相同。外层身份写成
> `anonymous` 会被判为*用户不存在*而拒绝。EAP-TLS 不受影响（其身份来自证书）。

### 设备侧（RouterOS，参考示例，请按你的固件核对）

先在 ToughRADIUS 的 **NAS 设备** 中登记这台路由器（源 IP + 共享密钥；若你还想下发
`Mikrotik-Rate-Limit` 就选厂商 *MikroTik*，否则 *Standard* —— EAP 本身不需要 VSA）。
然后把 AP 指向 ToughRADIUS，并让安全配置文件**透传 EAP**：

```routeros
# 1) wireless 服务的 RADIUS 服务器（密钥与 NAS 记录一致）
/radius add service=wireless address=<TOUGHRADIUS_IP> secret=<SECRET> timeout=3s

# 2a) 经典 /interface wireless —— passthrough 是关键词
/interface wireless security-profiles add name=eap-passthrough \
    authentication-types=wpa2-eap eap-methods=passthrough \
    radius-eap-accounting=yes \
    unicast-ciphers=aes-ccm group-ciphers=aes-ccm
/interface wireless set wlan1 security-profile=eap-passthrough \
    ssid="ToughRADIUS-EAP" mode=ap-bridge disabled=no

# 2b) …或 CAPsMAN（与参考仓库的拓扑一致）
/caps-man security add name=eap-passthrough authentication-types=wpa2-eap \
    encryption=aes-ccm group-encryption=aes-ccm eap-methods=passthrough \
    eap-radius-accounting=yes
# 把 security=eap-passthrough 挂到你的 /caps-man configuration，再 provision

# 2c) …或新版 /interface wifi（wifiwave2，ROS 7.13+）：透传是隐式的
/interface wifi security add name=eap-sec authentication-types=wpa2-eap,wpa3-eap
/interface wifi set wifi1 security=eap-sec \
    configuration.ssid="ToughRADIUS-EAP" disabled=no
```

> `eap-methods=passthrough`（经典 / CAPsMAN）正是让路由器从 EAP 终结者变成中继的开关。
> 新版 `/interface wifi` 没有 `eap-methods` 选项 —— 选用 `wpa2-eap`/`wpa3-eap` 安全配置
> 后，它会自动把请求中继给配置了 `service=wireless` 的 RADIUS。

### 验证

`radtest` **无法**驱动 EAP。请用 `eapol_test`（来自 `wpa_supplicant` / hostap）——
即本项目 [EAP 验收测试报告](./eap-acceptance-reports.md) 所用的工具（v2.10）。它直接对
ToughRADIUS 讲 RADIUS，因此你能在**接触真实射频之前**先验证服务器。存成下列任一文件后运行
`eapol_test -c <文件>.conf -a <TOUGHRADIUS_IP> -p 1812 -s <SECRET>`，通过会打印
`SUCCESS`：

```ini
# eap-tls.conf  —— 证书，无需密码
network={
    key_mgmt=WPA-EAP
    eap=TLS
    identity="alice@example.com"      # == 证书 SAN email == TR 用户名
    ca_cert="/etc/eap/ca.pem"         # 信任 ToughRADIUS 的服务器证书
    client_cert="/etc/eap/alice.pem"
    private_key="/etc/eap/alice.key"
}
```

```ini
# peap-mschapv2.conf
network={
    key_mgmt=WPA-EAP
    eap=PEAP
    identity="bob"                    # 外层 == 真实用户名（不要匿名身份）
    password="<bob-密码>"
    ca_cert="/etc/eap/ca.pem"
    phase2="auth=MSCHAPV2"
}
```

```ini
# ttls-pap.conf
network={
    key_mgmt=WPA-EAP
    eap=TTLS
    identity="carol"
    anonymous_identity="carol"        # 必须等于 identity（见上面的坑）
    password="<carol-密码>"
    ca_cert="/etc/eap/ca.pem"
    phase2="auth=PAP"                 # 或 auth=MSCHAPV2；不支持 CHAP/MS-CHAP
}
```

- **ToughRADIUS 日志**：成功时记录
  `radius auth success … is_eap=true result=success`；会话出现在**在线会话**页
  （射频上线后即带计费）。
- **路由器侧**：`/log print where topics~"radius,wireless"`；查看已关联客户端用
  `/interface wireless registration-table print`（经典）或
  `/interface wifi registration-table print`（wifi）。
- **EAP-TLS 与 EAP-TTLS（PAP 与 MS-CHAP-V2）** 用 `eapol_test` 可端到端干净通过。
  而 **PEAP-MSCHAPv2** 目前在 `eapol_test` 上存在已记录的内层封帧互通缺口，因此验收流水
  线改用进程内集成测试验证 PEAP（见
  [EAP 验收测试报告](./eap-acceptance-reports.md)）；真实的 Windows / Android / iOS
  PEAP 请求方可正常互通，因此 PEAP 请用真实客户端测试。

### 排障（症状 → 定位 → 解决）

- **EAP-TLS 被拒，回包 `… handshake failed`** → 客户端证书未链到 **`EapTlsCaFile`**
  的 CA（或配错了 CA 包）。用同一张 `ca.pem` 重签客户端证书，或修正 CA 路径。
- **EAP-TLS 被拒，回包 `… identity … does not match`** → 用户**用户名 ≠ 证书
  Peer-Id**。记住顺序 **SAN email → SAN DNS → CN**，且带 SAN 的证书会忽略 `CN`：对示例
  证书，用户名须为 `alice@example.com`。要么改用户名，要么签一张仅含 CN 的证书并把用户名
  设为该 CN。
- **客户端拒绝连接 / “无法验证服务器”** → 设备不信任 **ToughRADIUS** 的 CA。在客户端
  安装 `ca.pem`；在 **Android 11+** 上还要把 **域（Domain）** 字段填成服务器证书的
  CN/SAN（`radius.example.com`），这是 Android 现在强制要求的。
- **PEAP / TTLS 密码正确却被拒为“用户不存在”** → 用了**匿名外层身份**。清空它（或设为真实
  用户名），让外层 `User-Name` 等于账号 —— 见上面的坑。
- **完全没有 EAP 挑战 / 立即拒绝** → `EapTlsCertFile` + `EapTlsKeyFile`（EAP-TLS 还需
  `EapTlsCaFile`）未**全部**配置，于是 EAP 安全拒绝。补全路径即可，无需重启。
- **想要的方式始终没被提供** → 默认 `EapMethod` 是 `eap-md5`，对 WPA-Enterprise 无效。
  把 `EapMethod` 设成你的隧道方式，并把它（以及你允许客户端 NAK 切换到的方式）列入
  **启用的 EAP 处理器**。
- **EAP-TTLS 内层认证被拒** → 仅实现了**内层 PAP 与 MS-CHAP-V2**；不支持内层 CHAP /
  MS-CHAP / 隧道内 EAP。此外 TTLS 隧道钉死在 **TLS 1.2** —— 仅支持 TLS 1.3 的请求方无法
  完成第二阶段。
- **认证成功但无计费 / 无在线会话** → 启用 `radius-eap-accounting=yes`（经典）/
  `eap-radius-accounting=yes`（CAPsMAN），并确保 UDP **1813** 能到达 ToughRADIUS。

---

## 相关章节

- [厂商对接指南 · MikroTik](./vendor-guide.md) —— 属性参考卡。
- [管理系统用户手册](./admin-manual.md) —— 用户 / 计费策略 / 在线会话 / CoA 表单。
- [常见问题解答](./faq.md) —— 更多跨场景排障问答。
