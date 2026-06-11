# Cookbook: MikroTik RouterOS

> 中文版本：[实战手册：MikroTik RouterOS](../zh/cookbook-mikrotik.md)
>
> This chapter is part of the [Scenario Cookbook](./cookbook.md) and follows its
> [five-part shape and reading conventions](./cookbook.md#the-five-part-shape-of-every-scenario).

MikroTik RouterOS (vendor code **14988**) is the most common integration
target. ToughRADIUS registers a dedicated vendor enhancer for it; on a
successful auth it emits:

- `Mikrotik-Rate-Limit = "{up}k/{down}k"` — a string rate limit (produced by
  `mikrotik_enhancer.go`; `rx/tx` is from the **router's perspective**, so the
  first field is the subscriber's upload).
- Plus the standard attributes common to all devices: `Session-Timeout`,
  `Acct-Interim-Interval`, `Framed-Pool`, `Framed-IP-Address`, etc. (produced by
  `default_enhancer.go`).

> **Prerequisite**: register this router under **NAS devices** with *vendor =
> MikroTik*, the correct source IP and shared secret; ToughRADIUS must be
> reachable by the device (auth 1812, accounting 1813). If you register it as
> `Standard`, auth still succeeds but `Mikrotik-Rate-Limit` is **not** emitted.

---

## Scenario A: PPPoE broadband ISP — speed tiers + address pool + expiry disconnect + concurrency

### Need / scenario

A neighbourhood / small ISP serves Internet over PPPoE and needs: several speed
tiers (e.g. Home = 30M down, Business = 100M down), accounts that disconnect on
expiry and cannot redial, a concurrency cap per account (to stop credential
sharing), and IPs assigned from a shared address pool.

### On the ToughRADIUS side

1. **Create one rate profile per tier** (**Rate profiles → New**):
   - **Up / down rate**: the unit is **Kbps**. 30M down means `30720`, **not**
     `30`; 10M up means `10240`.
   - **Concurrency `active_num`**: e.g. `1` = at most one online session per
     account (`0` = unlimited).
   - **Address pool**: a pool name (e.g. `pppoe-pool`) that must match the
     `/ip pool` name on the router.
2. **Create a user** (**Users → New**): username / password, the matching **rate
   profile** (rate, concurrency, pool are inherited from it), an **expiry time**,
   status = enabled. For a fixed IP, set a static IPv4 on the user (it overrides
   the pool).
3. **Accounting interval**: the config item `radius.AcctInterimInterval`
   (default `120` seconds) sets the emitted `Acct-Interim-Interval`.

After a successful auth, the `Access-Accept` actually carries (anchored to
code):

| Attribute | Value | Source |
| --------- | ----- | ------ |
| `Mikrotik-Rate-Limit` | `"{up_kbps}k/{down_kbps}k"`, e.g. `10240k/30720k` | `mikrotik_enhancer.go` |
| `Session-Timeout` | seconds remaining until expiry (drops the current session at the deadline) | `default_enhancer.go` |
| `Acct-Interim-Interval` | `radius.AcctInterimInterval` (default 120) | `default_enhancer.go` |
| `Framed-Pool` | the pool name from the profile / user (emitted only if set) | `default_enhancer.go` |
| `Framed-IP-Address` | the user's static IPv4 (emitted only if set) | `default_enhancer.go` |

> **The concurrency cap is not implemented by emitting an attribute**: it is
> enforced by `online_count_checker` at auth time, comparing `active_num` to the
> current online count — an over-limit new session is rejected
> (`Access-Reject`).

### On the device side (RouterOS, reference example, verify on your firmware)

```routeros
# Point at ToughRADIUS (same shared secret for auth/accounting)
/radius add service=ppp address=<TOUGHRADIUS_IP> secret=<SECRET> timeout=3s
/radius incoming set accept=yes port=3799

# Enable RADIUS auth + accounting + periodic interim updates
/ppp aaa set use-radius=yes accounting=yes interim-update=5m

# The pool name must match the emitted Framed-Pool
/ip pool add name=pppoe-pool ranges=10.10.0.2-10.10.255.254

# PPPoE service (let the RADIUS Framed-Pool/Framed-IP decide remote-address)
/ppp profile add name=radius-pppoe local-address=10.10.0.1
/interface pppoe-server server add service-name=isp interface=<bridge> \
    default-profile=radius-pppoe disabled=no
```

> No manual queue is needed for rate limiting: on receiving `Mikrotik-Rate-Limit`
> the router creates a dynamic simple queue automatically.

### Verification

- **radtest (server side)**:
  ```bash
  go run ./cmd/radtest auth -server <TOUGHRADIUS_IP> -nas-ip <NAS_IP> \
    -username <username> -password <password> -secret <SECRET>
  ```
  On success it prints `Access-Accept`; you should see `Mikrotik-Rate-Limit`,
  `Session-Timeout`, and (if a pool is set) `Framed-Pool`.
- **Router side**: `/ppp active print` (online sessions), `/queue simple print`
  (the dynamic queue and its rate), `/log print where topics~"radius"`.
- **Admin UI**: the **Online sessions** page should list the session (Framed IP,
  duration, up/down traffic).

### Troubleshooting (symptom → locate → fix)

- **Rate limit has no effect at all** → ① confirm the NAS is registered as
  *MikroTik* (registered as Standard emits no VSA); ② wrong rate unit (`30`
  instead of `30720`); ③ direction reversed (`rx/tx` is the router's view, the
  first field is the subscriber's **upload**).
- **Dials up but gets no / wrong IP** → the `Framed-Pool` name does not match the
  `/ip pool` name, or no pool is set on the profile; align the names or set a
  static IP on the user.
- **Account still online after expiry** → `Session-Timeout` only drops the
  **current** session at the deadline; an already-online session waits for the
  timeout or a manual **forced disconnect** on the sessions page. A **redial**
  after expiry is rejected by `expire_checker` (counted under `user expire`).
- **The second connection is rejected** → with `active_num=1`,
  `online_count_checker` rejects the concurrent session — this is expected; raise
  the profile's concurrency to allow multiple dials.
- **Server slows down / stops replying after repeated auth failures** →
  `reject_delay_guard` introduces a delay once a username is rejected more than a
  threshold (default 7) in a row, to throttle brute force; it recovers
  automatically.

---

## Scenario B: Hotspot + MAC authentication

### Need / scenario

In a public WiFi / Hotspot environment, you want certain enrolled devices
(printers, IoT, long-term guest devices) to **skip the portal** and be admitted
and rate-limited by MAC address.

### On the ToughRADIUS side

ToughRADIUS decides a request is **MAC authentication** by this rule (anchored to
`auth_stages.go` / `eap_helper.go`):

> When the MAC address parsed from the request **equals the username**, the
> request is treated as MAC auth; the password check then compares against the
> user record's **MAC address** field rather than an ordinary password.

So configure it as follows:

1. **Create a user** with **username = the device MAC** (the string must match
   exactly what the router sends), and put the same MAC in the user record's
   **MAC address** field.
2. Assign the user a rate profile (rate and concurrency apply as usual).

> **The biggest trap is MAC format**: case and separators must exactly match the
> `User-Name` the RouterOS hotspot sends (ToughRADIUS compares strings exactly).
> RouterOS's format is influenced by settings such as `mac-auth-mode` — **what
> you store is what must be sent**.

### On the device side (RouterOS, reference example, verify on your firmware)

```routeros
/radius add service=hotspot address=<TOUGHRADIUS_IP> secret=<SECRET> timeout=3s

# Enable RADIUS and MAC login on the hotspot server profile
/ip hotspot profile set <profile> use-radius=yes login-by=mac,http-chap
# mac-auth-mode / mac-auth-password depend on firmware
```

### Verification

- **radtest**: use `-user <MAC> -pwd <MAC>` to simulate a MAC auth and check for
  `Access-Accept`.
- **Router side**: `/ip hotspot active print` should list the device;
  `/log print` for the auth log.
- **Admin UI**: the session appears on the Online sessions page.

### Troubleshooting (symptom → locate → fix)

- **MAC auth always fails** → the username / MAC field does not match the MAC
  string the router actually sends (case, separators). Capture one request, read
  the literal `User-Name`, and recreate the user from that exact string.
- **It is being treated as an ordinary account** → only "request MAC ==
  username" takes the MAC-auth path; if the hotspot does not log in by MAC, the
  request goes through the portal username / password logic.

---

## Scenario C: Live control — CoA, forced disconnect and FUP

### Need / scenario

Control online users in real time: rate-limit after a quota is exceeded (FUP),
force a re-authentication, or simply kick a session offline.

### On the ToughRADIUS side

Select a session on the **Online sessions** page and run one of two actions
(anchored to `session_actions.go` and
[Admin UI Manual · Online sessions](./admin-manual.md#online-sessions)):

- **Change of Authorization (CoA-Request)**: currently carries only
  **`Session-Timeout` (#27)** and / or **`Filter-Id` (#11)**.
  - Use `Session-Timeout` to shorten the session's life and force the client to
    re-authenticate sooner.
  - Use `Filter-Id` to have RouterOS apply a **pre-defined** filter /
    address-list (e.g. a rate or site-restriction rule).
- **Forced disconnect (Disconnect-Request)**: terminate the session directly
  (with a confirmation step).

> **The correct path to live FUP "speed change"**: ToughRADIUS's CoA does **not**
> rewrite `Mikrotik-Rate-Limit`. To change a user's speed live, the standard
> approach is — change the rate on the profile / user first, **then force a
> disconnect**; the client redials automatically and is re-authorized at the new
> speed. This path matches the system's actual capability and works for any
> vendor.

Operator-initiated CoA / Disconnect uses a short timeout with one automatic
retry, targeting the **CoA port (default 3799)** on the NAS record.

### On the device side (RouterOS, reference example, verify on your firmware)

```routeros
# Required, otherwise CoA/Disconnect is not received
/radius incoming set accept=yes port=3799
```

- The firewall must allow **inbound UDP 3799** (from ToughRADIUS to the router).
- For the `Filter-Id` approach, pre-define a filter / queue / address-list of the
  same name on RouterOS (verify on your firmware).

### Verification

- Click **Change authorization** / **Forced disconnect** on the sessions page;
  the result is reported as a notification.
- On the router, `/log print` shows the incoming request; after a **forced
  disconnect**, the session disappears from `/ppp active print` and the client
  then redials automatically.

### Troubleshooting (symptom → locate → fix)

- **CoA / Disconnect does not respond or times out** → ① RouterOS did not set
  `radius incoming accept=yes`; ② the firewall blocks UDP 3799; ③ the NAS
  record's CoA port differs from the device's `incoming port`; ④ the shared
  secret differs.
- **Changed the rate but the online user's speed did not change** → expected: CoA
  does not change the rate; **force a disconnect** so the client redials.
- **Port confusion** → the CoA `1700` you often see online is a client-side local
  port; this system and RFC 5176 use **3799**.

---

## Related chapters

- [Vendor Integration Guide · MikroTik](./vendor-guide.md) — attribute reference
  card.
- [Admin UI Manual](./admin-manual.md) — users / rate profiles / online sessions
  / CoA forms.
- [FAQ](./faq.md) — more cross-scenario troubleshooting Q&A.
