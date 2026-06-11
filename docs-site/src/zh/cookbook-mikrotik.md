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
  go run ./cmd/radtest auth -user <用户名> -pwd <密码> -nasip <NAS_IP> -secret <SECRET>
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

## 相关章节

- [厂商对接指南 · MikroTik](./vendor-guide.md) —— 属性参考卡。
- [管理系统用户手册](./admin-manual.md) —— 用户 / 计费策略 / 在线会话 / CoA 表单。
- [常见问题解答](./faq.md) —— 更多跨场景排障问答。
