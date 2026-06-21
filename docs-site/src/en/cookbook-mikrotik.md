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

## Scenario D: WPA2/WPA3-Enterprise Wi-Fi — 802.1X EAP passthrough to ToughRADIUS

### Need / scenario

You run enterprise Wi-Fi (`WPA2-EAP` / `WPA3-EAP`, i.e. 802.1X) on MikroTik APs
and want **one central place** for accounts, certificates and policy: ToughRADIUS.
Staff authenticate either with a **client certificate** (EAP-TLS, password-less),
or with **username + password** (PEAP-MSCHAPv2 for Windows/AD-style clients, or
EAP-TTLS for legacy / LDAP back ends).

> **How this differs from
> [`multiduplikator/mikrotik_EAP`](https://github.com/multiduplikator/mikrotik_EAP).**
> That well-known guide makes RouterOS itself the EAP server (ROS6 against the
> certificate store, or ROS7 via **User Manager v5**). Here MikroTik is **only the
> authenticator**: its `eap-methods=passthrough` relays the 802.1X/EAP conversation
> to ToughRADIUS, and **ToughRADIUS terminates EAP**. Two consequences follow that
> you must plan for:
>
> 1. **The trust anchor moves to ToughRADIUS.** Clients no longer trust `RouterCA`;
>    they must trust the **CA that signed ToughRADIUS's server certificate**
>    (`EapTlsCertFile`). You distribute *that* CA to client devices.
> 2. **The router holds no user/cert material for auth.** All identities live on
>    ToughRADIUS, so you get its account lifecycle, rate profiles, online-session
>    view and accounting for free.

### On the ToughRADIUS side

#### 0. Prepare the certificates (PEM on disk)

ToughRADIUS loads PEM files from disk (paths set in the config items below). A
minimal in-house CA with `openssl` — EC keys keep the TLS records small, which
matters for EAP fragmentation:

```bash
# Root CA (distribute ca.pem to every client device)
openssl ecparam -name prime256v1 -genkey -noout -out ca.key
openssl req -x509 -new -key ca.key -sha256 -days 3650 -out ca.pem \
  -subj "/CN=ToughRADIUS EAP Root CA"

# RADIUS server certificate (CN/SAN is what clients pin as the server identity)
openssl ecparam -name prime256v1 -genkey -noout -out server.key
openssl req -new -key server.key -out server.csr -subj "/CN=radius.example.com"
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
  -days 825 -sha256 -out server.pem \
  -extfile <(printf "subjectAltName=DNS:radius.example.com\nextendedKeyUsage=serverAuth\nkeyUsage=digitalSignature,keyEncipherment")

# EAP-TLS client certificate (only for EAP-TLS users). The SAN email becomes the
# RADIUS identity — see the identity-binding rule below.
openssl ecparam -name prime256v1 -genkey -noout -out alice.key
openssl req -new -key alice.key -out alice.csr -subj "/CN=alice"
openssl x509 -req -in alice.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
  -days 825 -sha256 -out alice.pem \
  -extfile <(printf "subjectAltName=email:alice@example.com\nextendedKeyUsage=clientAuth\nkeyUsage=digitalSignature")

# Bundle the client key+cert+CA as PKCS#12 for Windows/Android/iOS enrolment
openssl pkcs12 -export -inkey alice.key -in alice.pem -certfile ca.pem \
  -out alice.p12 -passout pass:'<long-passphrase>'
```

> **EAP-TLS identity binding (anchored to `tlsengine/identity.go`, RFC 5216 §5.2).**
> After the client certificate passes chain validation, ToughRADIUS derives the
> **Peer-Id** in this order: **SAN `rfc822Name` (email) → SAN `dnsName` → subject
> `CN`**, and it must equal the RADIUS `User-Name` (case-insensitive). When a SAN
> is present the `CN` is **not** accepted as an alternate. So with the cert above
> the matching ToughRADIUS username is **`alice@example.com`**, not `alice`.

#### 1. Upload the material and set the config items

Copy `server.pem`, `server.key`, and `ca.pem` onto the ToughRADIUS host (e.g.
`/var/toughradius/eap/`), then set these on **System config → RADIUS**
(anchored to `internal/app/config_schemas.json` /
`eap/handlers/tls_config.go`):

| Config item (`radius.*`) | EAP-TLS | PEAP / TTLS | Notes |
| --- | --- | --- | --- |
| **EAP Method** (`EapMethod`) | `eap-tls` | `eap-peap` / `eap-ttls` | The method **offered first** on EAP-Identity. |
| **Enabled EAP Handlers** (`EapEnabledHandlers`) | `eap-tls,eap-peap,eap-ttls` | same | Allow-list; `*` = all. A client may **NAK** to another method, honoured only if it is enabled here. |
| **EAP-TLS Server Certificate** (`EapTlsCertFile`) | `/var/toughradius/eap/server.pem` | same | Presented for **all three** tunnelled methods. |
| **EAP-TLS Server Private Key** (`EapTlsKeyFile`) | `/var/toughradius/eap/server.key` | same | — |
| **EAP-TLS Client CA Bundle** (`EapTlsCaFile`) | `/var/toughradius/eap/ca.pem` | *(unused)* | **Required for EAP-TLS** (verifies client certs); PEAP/TTLS are server-only and ignore it. |
| **EAP-TLS Minimum TLS Version** (`EapTlsMinVersion`) | `1.2` | `1.2` | PEAP/TTLS are pinned to **TLS 1.2** regardless. |

> Certificate/key/CA are re-read **between handshakes**, so you can rotate them
> without restarting the RADIUS service. Until the required files are all set,
> EAP **safe-rejects** (`ErrTLSNotConfigured`) rather than authenticating without
> trust anchors — so a half-configured server never lets anyone in.

#### 2. Create the users

| Method | Username to create | Password field | Comes from |
| --- | --- | --- | --- |
| **EAP-TLS** | the cert **Peer-Id**, e.g. `alice@example.com` | **unused** (any placeholder) | cert identity matched in `tlsengine/identity.go` |
| **PEAP-MSCHAPv2** | the login name, e.g. `bob` | the user's real password | inner MSCHAPv2 vs `GetPassword` in `peap_inner.go` |
| **EAP-TTLS** (PAP or MS-CHAP-V2) | the login name, e.g. `carol` | the user's real password | inner PAP/MSCHAPv2 vs `GetPassword` in `ttls_inner.go` |

Each user still needs **status = enabled**, a **future expiry**, and a **rate
profile** (rate / pool / concurrency apply exactly as in Scenario A).

> **The single biggest passthrough trap — the outer (anonymous) identity.**
> ToughRADIUS loads the user record from the **outer `User-Name`** and looks up
> the password from *that* record; mapping a separate *anonymous* outer identity
> to the real account is not yet implemented (deferred to M8.4). So for
> **PEAP / TTLS the outer identity must equal the real username** — on the
> supplicant, **leave “anonymous identity” blank** (it then sends the real
> username in the clear outer identity) **or set it equal to the username**. An
> outer identity of `anonymous` is rejected as *user not found*. EAP-TLS is
> unaffected (its identity comes from the certificate).

### On the device side (RouterOS, reference example, verify on your firmware)

First register this router under **NAS devices** in ToughRADIUS (its source IP +
shared secret; vendor *MikroTik* if you also want `Mikrotik-Rate-Limit`, otherwise
*Standard* — EAP itself needs no VSA). Then point the AP at ToughRADIUS and make
the security profile **pass EAP through**:

```routeros
# 1) RADIUS server for the wireless service (same secret as the NAS record)
/radius add service=wireless address=<TOUGHRADIUS_IP> secret=<SECRET> timeout=3s

# 2a) CLASSIC /interface wireless — passthrough is the key word
/interface wireless security-profiles add name=eap-passthrough \
    authentication-types=wpa2-eap eap-methods=passthrough \
    radius-eap-accounting=yes \
    unicast-ciphers=aes-ccm group-ciphers=aes-ccm
/interface wireless set wlan1 security-profile=eap-passthrough \
    ssid="ToughRADIUS-EAP" mode=ap-bridge disabled=no

# 2b) …or CAPsMAN (matches the reference repo's topology)
/caps-man security add name=eap-passthrough authentication-types=wpa2-eap \
    encryption=aes-ccm group-encryption=aes-ccm eap-methods=passthrough \
    eap-radius-accounting=yes
# attach security=eap-passthrough to your /caps-man configuration, then provision

# 2c) …or the new /interface wifi (wifiwave2, ROS 7.13+): passthrough is IMPLICIT
/interface wifi security add name=eap-sec authentication-types=wpa2-eap,wpa3-eap
/interface wifi set wifi1 security=eap-sec \
    configuration.ssid="ToughRADIUS-EAP" disabled=no
```

> `eap-methods=passthrough` (classic / CAPsMAN) is exactly what makes the router a
> relay instead of an EAP terminator. The new `/interface wifi` stack has no
> `eap-methods` knob — selecting a `wpa2-eap`/`wpa3-eap` security profile makes it
> relay to the configured `service=wireless` RADIUS automatically.

### Verification

`radtest` **cannot** drive EAP. Use `eapol_test` (from `wpa_supplicant` /
hostap) — the same tool the project's
[EAP acceptance reports](./eap-acceptance-reports.md) run (v2.10). It talks
RADIUS straight to ToughRADIUS, so you can validate the server **before**
touching a real radio. Save one of these and run
`eapol_test -c <file>.conf -a <TOUGHRADIUS_IP> -p 1812 -s <SECRET>` — a pass
prints `SUCCESS`:

```ini
# eap-tls.conf  — certificate, password-less
network={
    key_mgmt=WPA-EAP
    eap=TLS
    identity="alice@example.com"      # == the cert SAN email == the TR username
    ca_cert="/etc/eap/ca.pem"         # trust ToughRADIUS's server cert
    client_cert="/etc/eap/alice.pem"
    private_key="/etc/eap/alice.key"
}
```

```ini
# peap-mschapv2.conf
network={
    key_mgmt=WPA-EAP
    eap=PEAP
    identity="bob"                    # outer == real username (no anonymous id)
    password="<bob-password>"
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
    anonymous_identity="carol"        # must equal identity (see the trap above)
    password="<carol-password>"
    ca_cert="/etc/eap/ca.pem"
    phase2="auth=PAP"                 # or auth=MSCHAPV2; CHAP/MS-CHAP unsupported
}
```

- **ToughRADIUS log**: a successful run logs
  `radius auth success … is_eap=true result=success`; the session appears on the
  **Online sessions** page (with accounting once the radio is live).
- **Router side**: `/log print where topics~"radius,wireless"`; for an associated
  client `/interface wireless registration-table print` (classic) or
  `/interface wifi registration-table print` (wifi).
- **EAP-TLS and EAP-TTLS (PAP & MS-CHAP-V2)** verify cleanly end-to-end with
  `eapol_test`. For **PEAP-MSCHAPv2**, `eapol_test` currently hits a documented
  inner-framing interop gap, so the acceptance harness validates PEAP with the
  in-process integration test instead (see
  [EAP Acceptance Reports](./eap-acceptance-reports.md)); real Windows / Android /
  iOS PEAP supplicants interoperate, so test PEAP with an actual client.

### Troubleshooting (symptom → locate → fix)

- **EAP-TLS reject, reply `… handshake failed`** → the client certificate does not
  chain to the **`EapTlsCaFile`** CA (or the wrong CA bundle is configured).
  Re-issue the client cert from the same `ca.pem`, or fix the CA path.
- **EAP-TLS reject, reply `… identity … does not match`** → the user's
  **username ≠ the cert Peer-Id**. Remember the order **SAN email → SAN DNS → CN**
  and that a SAN cert's `CN` is ignored: for the sample cert the username must be
  `alice@example.com`. Either rename the user or issue a CN-only cert and set the
  username to the CN.
- **Client refuses to connect / “can’t verify server”** → the device does not
  trust **ToughRADIUS's** CA. Install `ca.pem` on the client; on **Android 11+**
  also set the **Domain** field to the server cert CN/SAN (`radius.example.com`),
  which Android now enforces.
- **PEAP / TTLS reject as “user not found” although the password is right** → an
  **anonymous outer identity** was used. Clear it (or set it to the real username)
  so the outer `User-Name` equals the account — see the trap above.
- **No EAP challenge at all / immediate reject** → `EapTlsCertFile` + `EapTlsKeyFile`
  (and `EapTlsCaFile` for EAP-TLS) are not **all** set, so EAP safe-rejects. Set the
  paths; no restart needed.
- **The method you want is never offered** → the default `EapMethod` is `eap-md5`,
  which is invalid for WPA-Enterprise. Set `EapMethod` to your tunnelled method and
  include it (and any you accept via client NAK) in **Enabled EAP Handlers**.
- **EAP-TTLS inner auth rejected** → only **inner PAP and MS-CHAP-V2** are
  implemented; inner CHAP / MS-CHAP / tunnelled-EAP are not. Also the TTLS tunnel is
  pinned to **TLS 1.2** — a TLS-1.3-only supplicant won't complete phase 2.
- **Auth works but no accounting / no online session** → enable
  `radius-eap-accounting=yes` (classic) / `eap-radius-accounting=yes` (CAPsMAN) and
  make sure UDP **1813** reaches ToughRADIUS.

---

## Related chapters

- [Vendor Integration Guide · MikroTik](./vendor-guide.md) — attribute reference
  card.
- [Admin UI Manual](./admin-manual.md) — users / rate profiles / online sessions
  / CoA forms.
- [FAQ](./faq.md) — more cross-scenario troubleshooting Q&A.
