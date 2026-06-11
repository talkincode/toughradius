# Cookbook: Huawei BRAS / NetEngine

> 中文版本：[实战手册：华为 BRAS / NetEngine](../zh/cookbook-huawei.md)
>
> This chapter is part of the [Scenario Cookbook](./cookbook.md) and follows its
> [five-part shape and reading conventions](./cookbook.md#the-five-part-shape-of-every-scenario).

Huawei (vendor code **2011**) is the dominant broadband BRAS / enterprise
gateway in Chinese carrier and enterprise networks (NetEngine / ME60 / older
MA5200 lines). ToughRADIUS registers a dedicated vendor enhancer for it; on a
successful auth it emits:

- A **four-attribute rate quartet** (produced by `huawei_enhancer.go`):
  `Huawei-Input-Average-Rate`, `Huawei-Input-Peak-Rate`,
  `Huawei-Output-Average-Rate`, `Huawei-Output-Peak-Rate`.
- `Huawei-Domain-Name` — only when the user / profile has a **domain** set.
- `Huawei-Framed-IPv6-Address` — only when the user has a **static IPv6**.
- Plus the standard attributes common to all devices (`Session-Timeout`,
  `Acct-Interim-Interval`, `Framed-Pool`, `Framed-IP-Address`, …) produced by
  `default_enhancer.go`.

On the request side the **Huawei parser** (`huawei_parser.go`) extracts the MAC
from `Calling-Station-Id` (normalising `-` to `:`) and the inner / outer VLAN
IDs from `NAS-Port-Id`. That is what makes MAC binding *and* VLAN binding
possible for Huawei devices.

> **Prerequisite**: register this BRAS under **NAS devices** with *vendor =
> Huawei*, the correct source IP and shared secret; ToughRADIUS must be
> reachable (auth 1812, accounting 1813). If you register it as `Standard`, auth
> still succeeds but **none** of the Huawei VSAs above are emitted, and the VLAN
> IDs are not parsed.

---

## Scenario A: PPPoE / IPoE broadband — speed tiers, peak rate and AAA domain

### Need / scenario

A carrier or enterprise runs broadband on a Huawei BRAS and needs several speed
tiers (e.g. Home = 30M down / 10M up, Business = 100M down). Huawei honours both
an **average** and a **peak** (burst) rate, and groups subscribers into AAA
**domains** so the BRAS applies the right domain policy.

### On the ToughRADIUS side

1. **Create one rate profile per tier** (**Rate profiles → New**):
   - **Up / down rate**: the unit is **Kbps**. 30M down means `30720`, **not**
     `30`; 10M up means `10240`.
   - **Domain** (optional): the Huawei AAA domain name (e.g. `isp`), pushed as
     `Huawei-Domain-Name` so the BRAS binds the session to that domain.
2. **Create a user** (**Users → New**): username / password, the matching **rate
   profile**, an **expiry time**, status = enabled. A per-user **domain** field
   overrides the profile domain.

How the stored Kbps rates become the four Huawei VSAs (anchored to
`huawei_enhancer.go`):

| Attribute | Value | Note |
| --------- | ----- | ---- |
| `Huawei-Input-Average-Rate` | `up_kbps × 1024` (bit/s) | "Input" = traffic **into** the BRAS = subscriber **upload** |
| `Huawei-Input-Peak-Rate` | `up_kbps × 1024 × 4` | average × 4 |
| `Huawei-Output-Average-Rate` | `down_kbps × 1024` (bit/s) | "Output" = traffic **out** to the subscriber = **download** |
| `Huawei-Output-Peak-Rate` | `down_kbps × 1024 × 4` | average × 4 |
| `Huawei-Domain-Name` | the user / profile domain | emitted **only** if a domain is set |
| `Session-Timeout`, `Acct-Interim-Interval`, `Framed-Pool`, `Framed-IP-Address` | standard | from `default_enhancer.go` |

> **Two traps unique to Huawei.** ① **Unit**: the rate VSAs are in **bit/s**,
> not Kbps — ToughRADIUS multiplies the stored Kbps by **1024** (binary), and
> the **peak** is the average **× 4**. So a "30M down" tier is sent as
> `Output-Average-Rate = 30720 × 1024 = 31457280` and
> `Output-Peak-Rate = 125829120`. ② **Direction naming**: Huawei "Input" is the
> subscriber's **upload** and "Output" is the **download** (BRAS perspective) —
> the opposite words from MikroTik's `rx/tx`, but the same physical meaning. All
> four values are clamped to the Int32 maximum.

### On the device side (Huawei VRP, reference example, verify on your firmware)

```text
# RADIUS server template
radius-server template tr-tmpl
 radius-server shared-key cipher <SECRET>
 radius-server authentication <TOUGHRADIUS_IP> 1812 weight 80
 radius-server accounting <TOUGHRADIUS_IP> 1813 weight 80
#
# AAA domain bound to the template (matches Huawei-Domain-Name)
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

> The `Huawei-Output-Average-Rate` / peak values map to the BRAS's per-user
> CAR / QoS; no manual per-user queue is required.

### Verification

- **radtest (server side)**:
  ```bash
  go run ./cmd/radtest auth -server <TOUGHRADIUS_IP> -nas-ip <NAS_IP> \
    -username <username> -password <password> -secret <SECRET>
  ```
  On success it prints `Access-Accept`; you should see the four Huawei rate
  attributes and (if set) `Huawei-Domain-Name`.
- **BRAS side**: `display access-user username <name>` (online session, rate,
  domain), `display aaa online-fail-record` for failures.
- **Admin UI**: the **Online sessions** page should list the session.

### Troubleshooting (symptom → locate → fix)

- **Rate limit has no effect** → ① confirm the NAS is registered as *Huawei*
  (registered as Standard emits no VSA); ② remember the value is bit/s
  (`× 1024`), so a tier that looks "1000× too small" usually means the unit was
  read as Kbps on the device side.
- **Speeds look swapped (upload capped at the download value)** → the
  Input/Output direction confusion: Huawei **Input = upload**, **Output =
  download**; check which profile field feeds which.
- **Domain policy not applied** → the user / profile has no **domain** set (then
  `Huawei-Domain-Name` is not emitted), or the domain name does not match a
  `domain` configured on the BRAS.
- **Peak / burst seems too high** → expected: peak is hard-coded as average
  **× 4** in the enhancer; tune the average tier, not the peak.

---

## Scenario B: Line anti-fraud — MAC + VLAN binding and dual-stack IPv6

### Need / scenario

A broadband operator wants to stop account sharing / theft by **binding each
account to its access line** — the subscriber's MAC and / or the access VLAN
(inner / outer, i.e. QinQ) — and to hand out a **static IPv6** for dual-stack
service.

### On the ToughRADIUS side

Huawei encodes the access line in attributes the parser already reads:

- the **MAC** comes from `Calling-Station-Id`;
- the **inner / outer VLAN** come from `NAS-Port-Id`.

Binding is then enforced by two checkers (anchored to `mac_bind_checker.go` /
`vlan_bind_checker.go`):

1. **MAC binding** — enable **bind MAC** on the user / profile and store the
   allowed **MAC** on the user. At auth time, if the request MAC differs from the
   stored MAC, the session is rejected (`MacBindError`).
2. **VLAN binding** — enable **bind VLAN** on the user / profile and store the
   allowed **VLAN ID 1 / VLAN ID 2** on the user. If a stored VLAN differs from
   the parsed one, the session is rejected (`VlanBindError`).
3. **Static IPv6** — set the user's **IPv6 address**; the enhancer then emits
   `Huawei-Framed-IPv6-Address` (the prefix length, if written as `addr/len`, is
   stripped before sending).

> **Binding only enforces what you have stored.** Each checker is **skipped**
> when its toggle is off, *and also* when either side (stored value or request
> value) is empty / zero — it never silently auto-learns. So to bind a line you
> must (a) turn the toggle on **and** (b) fill in the MAC / VLAN on the user
> record (typically captured from the first successful login).

### On the device side (Huawei VRP, reference example, verify on your firmware)

```text
# IPoE / PPPoE access that reports the access line to RADIUS.
# Huawei carries the inner/outer VLAN inside NAS-Port-Id and the
# subscriber MAC inside Calling-Station-Id by default on BRAS access.
radius-server template tr-tmpl
 radius-server shared-key cipher <SECRET>
 radius-server authentication <TOUGHRADIUS_IP> 1812 weight 80
 radius-server accounting <TOUGHRADIUS_IP> 1813 weight 80
#
# Enable IPv6 address/prefix delivery on the BRAS as required by your design
ipv6
```

### Verification

- Capture one real `Access-Request` (or read the auth log) and confirm the
  `Calling-Station-Id` and `NAS-Port-Id` values, then store **exactly** those on
  the user before turning binding on.
- **radtest** with the matching MAC succeeds; a different MAC / VLAN is rejected.
- **BRAS side**: `display access-user username <name>` shows the bound line and
  any delivered IPv6 address.

### Troubleshooting (symptom → locate → fix)

- **Binding rejects the legitimate user** → the stored MAC / VLAN does not match
  what the BRAS actually sends. Read the real `Calling-Station-Id` /
  `NAS-Port-Id`, store those exact values, then re-enable binding.
- **VLAN binding never triggers** → the parsed VLAN is `0` (the BRAS does not put
  the VLAN in `NAS-Port-Id` in your topology), or **bind VLAN** is off, or the
  user's stored VLAN is `0`; any of these skips the check.
- **No IPv6 delivered** → the user has no static **IPv6 address** set, or the BRAS
  is not configured for IPv6 / dual-stack on that domain.
- **Binding is silently ignored** → the toggle is off or the stored value is
  empty; the checker only enforces when the toggle is on **and** both sides are
  present.

---

## Scenario C: Live control — CoA, forced disconnect and FUP

### Need / scenario

Control online users on a Huawei BRAS in real time: shorten a session, force
re-authentication, rate-limit after a quota is exceeded (FUP), or kick a session
offline.

### On the ToughRADIUS side

Select a session on the **Online sessions** page and run one of two actions
(anchored to `session_actions.go` and
[Admin UI Manual · Online sessions](./admin-manual.md#online-sessions)):

- **Change of Authorization (CoA-Request)**: currently carries only
  **`Session-Timeout` (#27)** and / or **`Filter-Id` (#11)** — it does **not**
  rewrite the Huawei rate VSAs.
- **Forced disconnect (Disconnect-Request)**: terminate the session directly.

> **The correct path to live FUP "speed change" on Huawei.** Because CoA does
> not rewrite `Huawei-Output-Average-Rate`, changing a user's speed live follows
> the same vendor-agnostic path as everywhere else: change the rate on the
> profile / user first, **then force a disconnect**; the client redials and is
> re-authorized at the new rate quartet. (Some Huawei firmware also reacts to a
> vendor-specific CoA rate attribute, but ToughRADIUS does **not** emit one — so
> do not rely on it.)

Operator-initiated CoA / Disconnect uses a short timeout with one automatic
retry, targeting the **CoA port (default 3799)** on the NAS record.

### On the device side (Huawei VRP, reference example, verify on your firmware)

```text
# The RADIUS template must accept dynamic authorization (CoA/DM).
radius-server template tr-tmpl
 radius-server authorization <TOUGHRADIUS_IP> shared-key cipher <SECRET>
```

- The firewall must allow **inbound UDP 3799** from ToughRADIUS to the BRAS.
- For the `Filter-Id` approach, pre-define a matching ACL / user-group of the
  same name on the BRAS (verify on your firmware).

### Verification

- Click **Change authorization** / **Forced disconnect** on the sessions page;
  the result is reported as a notification.
- **BRAS side**: after a forced disconnect, `display access-user username <name>`
  no longer lists the session, and the client redials automatically.

### Troubleshooting (symptom → locate → fix)

- **CoA / Disconnect does not respond or times out** → ① the RADIUS template has
  no `radius-server authorization`; ② the firewall blocks UDP 3799; ③ the NAS
  record's CoA port differs from the BRAS's authorization port; ④ the shared
  secret differs.
- **Changed the rate but the online user's speed did not change** → expected: CoA
  does not change the rate; **force a disconnect** so the client redials.
- **Port confusion** → the CoA `1700` you often see online is a client-side local
  port; this system and RFC 5176 use **3799**.

---

## Related chapters

- [Vendor Integration Guide · Huawei](./vendor-guide.md) — attribute reference
  card.
- [Cookbook: MikroTik RouterOS](./cookbook-mikrotik.md) — the same scenarios on
  RouterOS.
- [Admin UI Manual](./admin-manual.md) — users / rate profiles / online sessions
  / CoA forms.
- [FAQ](./faq.md) — more cross-scenario troubleshooting Q&A.
