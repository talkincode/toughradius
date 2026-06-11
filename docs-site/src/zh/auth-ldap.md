# LDAP / AD 认证后端

> English version: [LDAP / AD Authentication Backend](../en/auth-ldap.md)

ToughRADIUS 可以通过执行 LDAP **bind（绑定）** 操作，针对外部 LDAP 目录或
Microsoft Active Directory 校验用户口令，而不必（或不仅仅）使用自身数据库中
存储的口令。这样即可复用既有的企业目录，无需复制或迁移口令。

> **仅支持 PAP 族。** LDAP 后端通过「以明文口令执行 bind」来认证，因此只服务
> **裸 PAP** 与 **EAP-TTLS 内层 PAP**（RFC 5281）。挑战/响应类方法——CHAP、
> MS-CHAP、MS-CHAPv2、EAP-MD5、PEAP-MSCHAPv2——**无法**被服务，因为服务端
> 从不持有重算其响应所需的明文口令。启用 LDAP 后，这些方法会被有意拒绝并记录
> 可诊断原因；详见下文「认证流程」一节。

## 何时使用（以及何时不要用）

当你必须服务那些口令已存在于目录（OpenLDAP、FreeIPA 或 Active Directory）中
的用户，又不想给每个客户端签发证书时，就该用 LDAP 后端。它是混合环境与遗留
系统的务实桥梁：先用服务器证书保护隧道，再把用户名和口令塞进隧道里传输。

但要正视这笔权衡。**EAP-TTLS/PAP 在 TLS 隧道内传输明文口令**——它*只*受该隧道
保护。因此部署的安全性完全系于一张强壮且被正确校验的服务器证书与 TLS 1.2+。
如果每个客户端都能出示证书，请优先选 EAP-TLS；如果你需要为遗留客户端提供
目录后端的口令校验，那么 LDAP + EAP-TTLS/PAP 正是合适的工具——只是别对它的
安全性抱有不切实际的期待。

## 部署模型

LDAP 后端**只替换口令校验**这一步。它**不会**创建账号，也**不携带**授权数据。
每个用户仍需要一条本地 `RadiusUser` 记录，因为授权信息（套餐/资费、限速、
到期时间、并发会话上限、地址池、VLAN、MAC 绑定）是在认证*之前*由本地数据库
加载的。

换句话说：

- **认证**（口令对不对？）→ 由 LDAP 目录通过 bind 完成。
- **授权**（这个用户能做什么？）→ 仍由本地 `RadiusUser` 记录决定，与不启用
  LDAP 时完全一致。

全局开关 `ldap.Enabled` 对整台服务器启用或关闭该后端；没有「按用户指定认证源」
的字段。MAC 地址认证始终绕过 LDAP。

## 配置

在管理后台的**系统配置**页面、**LDAP** 分组下配置该后端；所有配置项也可通过
设置 API 修改。后端在*每次*认证尝试时都会重新读取配置，因此改动**即时生效**、
无需重启。

| 配置键 | 类型 | 默认值 | 适用模式 | 说明 |
| --- | --- | --- | --- | --- |
| `ldap.Enabled` | bool | `false` | 两者 | 启用 LDAP 后端。默认关闭。 |
| `ldap.ServerURL` | string | _(空)_ | 两者 | 目录 URL，如 `ldap://dc.example.com:389` 或 `ldaps://dc.example.com:636`。 |
| `ldap.BindMode` | 枚举 | `template` | 两者 | `template` 或 `search`（见下文）。 |
| `ldap.BindDNTemplate` | string | _(空)_ | template | 含单个 `%s`（代入用户名）的 DN 模板，如 `uid=%s,ou=people,dc=example,dc=com`，或 AD 的 UPN 形式 `%s@example.com`。 |
| `ldap.BaseDN` | string | _(空)_ | search | 用户查找的子树基点，如 `dc=example,dc=com`。 |
| `ldap.UserFilter` | string | `(uid=%s)` | search | 含单个 `%s`（代入用户名，代入前已转义）的过滤器，如 `(uid=%s)` 或 AD 的 `(sAMAccountName=%s)`。 |
| `ldap.SearchBindDN` | string | _(空)_ | search | 用于查找用户的只读服务账号 DN，如 `cn=svc-radius,ou=svc,dc=example,dc=com`。 |
| `ldap.SearchBindPassword` | string | _(空)_ | search | 服务账号 DN 的口令。 |
| `ldap.StartTLS` | bool | `false` | 两者 | 在 bind 前用 StartTLS 将 `ldap://` 连接升级为 TLS（RFC 4513 §3）。用 `ldaps://` 时请关闭。 |
| `ldap.TLSSkipVerify` | bool | `false` | 两者 | 跳过 TLS 证书校验。**不安全——仅限实验/自签名环境。** |
| `ldap.Timeout` | int（秒） | `5` | 两者 | 拨号与每次操作的超时秒数（1–60）。 |

### 两种 bind 模式

**模板模式**（`ldap.BindMode = template`）最简单：把用户名代入 `BindDNTemplate`
组成 DN，服务端直接以该 DN 加上所给口令执行 bind。适用于每个用户的 DN 都遵循
固定模式的场景。

```text
BindMode       = template
BindDNTemplate = uid=%s,ou=people,dc=example,dc=com
# Active Directory 写法：
# BindDNTemplate = %s@example.com
```

**搜索模式**（`ldap.BindMode = search`）适配 DN 不可预测的目录。服务端先以只读
**服务账号**（`SearchBindDN` / `SearchBindPassword`）bind，在 `BaseDN` 下用
`UserFilter` 搜索定位用户的 DN，然后**以用户身份重新 bind**（用所给口令）来
完成校验。

```text
BindMode           = search
ServerURL          = ldaps://dc.example.com:636
BaseDN             = dc=example,dc=com
UserFilter         = (sAMAccountName=%s)      # Active Directory
SearchBindDN       = cn=svc-radius,ou=svc,dc=example,dc=com
SearchBindPassword = ********
```

### 传输安全

- 使用 `ldaps://` URL 走隐式 TLS（636 端口），**或者**使用 `ldap://` URL 并设
  `StartTLS = true` 将明文连接升级（RFC 4513 §3）。不要在 `ldaps://` URL 上再
  开 StartTLS。
- 生产环境请保持 `TLSSkipVerify = false`。仅在实验或自签名目录上启用它——它会
  关闭证书校验，使 bind 暴露于被截获的风险。

## 认证流程

| 方法 | `ldap.Enabled = true` 时的行为 |
| --- | --- |
| 裸 PAP | 通过 LDAP bind 认证。 |
| EAP-TTLS / 内层 PAP | 通过 LDAP bind 认证；仍会为隧道派生 MS-MPPE 密钥。 |
| CHAP / MS-CHAP / MS-CHAPv2 | **拒绝**并记录可诊断原因（无明文口令可用于 bind）。 |
| EAP-MD5 / PEAP-MSCHAPv2 | 同理**拒绝**。 |
| MAC 认证 | 完全绕过 LDAP。 |

对挑战/响应类方法的拒绝是有意为之的，且集中在「取口令」边界统一处理，而不是
在协议入口分叉。启用 LDAP 后，本地 `RadiusUser.Password` 通常为空，挑战/响应
方法若仍以空口令推算期望值，可能**误判通过任意用户**——因此这些方法被「失败
即关闭」。

## 安全模型

该后端在设计上保守：

- **空用户名或空口令在任何网络操作前即被拒绝。** 带 DN 而口令为空的 bind 会被
  许多服务器当作*匿名*绑定而通过（RFC 4513 §5.1.2），因此在最前端直接拒绝。
- **注入被中和。** 搜索模式下用户名经 LDAP 过滤器转义（RFC 4515 §3）；模板模式
  下代入前经 DN 转义。
- **歧义搜索被拒绝。** 若 `UserFilter` 命中超过一条记录，则拒绝该次尝试而非
  猜测。
- **目录故障绝不被报告为口令错误。** 凭证无效（LDAP code 49）映射为口令拒绝；
  其余任何失败——目录不可达、配置错误、服务账号 bind 失败——均映射为
  *后端不可用*。这一区分同样驱动下文的指标。

## 可观测性

拒绝按原因计数（指标端点见[运维指南](./ops-guide.md)）：

- `radus_reject_passwd_error`——口令错误（bind 返回凭证无效）。
- `radus_reject_ldap_error`——后端无法给出答案（目录不可达、StartTLS 失败、
  服务账号 bind 失败、配置错误）。

将二者分开，意味着目录故障会体现为 `radus_reject_ldap_error`，而不是一片
「口令错误」的尖峰——于是对 `radus_reject_ldap_error` 的告警能清晰地指向目录
问题而非用户失误。认证成功会自增 `radus_accept`。

## 故障排查

| 现象 | 可能原因 |
| --- | --- |
| 所有登录都失败，`radus_reject_ldap_error` 持续上升 | `ServerURL` 错误/不可达、TLS 握手失败，或（搜索模式）服务账号无法 bind。 |
| Windows/AD 客户端被拒，PAP 用户正常 | 这些客户端在用 PEAP-MSCHAPv2 或 MS-CHAPv2，LDAP 无法服务——改用 EAP-TTLS/PAP，或改用本地口令后端。 |
| 搜索模式找不到用户 | 检查 `BaseDN` 与 `UserFilter`；确认服务账号能读取用户条目。 |
| 实验环境 bind 正常、生产失败 | `TLSSkipVerify` 此前掩盖了无效证书——安装受信任证书并关闭该开关。 |
| 口令通过了但会话没有套餐/限速 | 缺少本地 `RadiusUser` 记录——LDAP 只校验口令，请创建本地账号以提供授权。 |

## 参见

- [运维指南](./ops-guide.md)——EAP/TLS 证书、指标、进程模型。
- [核心术语与概念](./concepts.md)——PAP、EAP、EAP-TTLS。
- [协议与 RFC 索引](./rfc-index.md)——RFC 5281（EAP-TTLS）、RFC 4511/4513（LDAP）。
