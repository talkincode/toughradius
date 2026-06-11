# Cookbook: H3C, ZTE, iKuai & Cisco

> 中文版本：[实战手册：H3C / 中兴 / 爱快 / Cisco](../zh/cookbook-vendors.md)
>
> This chapter is part of the [Scenario Cookbook](./cookbook.md) and follows its
> [reading conventions](./cookbook.md#the-five-part-shape-of-every-scenario).

These four vendors run the **same operational scenarios** as the two flagship
cookbooks — PPPoE / IPoE speed tiers, address pool, expiry disconnect,
per-account concurrency, MAC binding, CoA / forced disconnect and FUP. The
*mechanics* (how a tier, an expiry, a concurrency cap, or a CoA behaves) are
**identical**, because they come from the shared `default_enhancer` and the
checkers — not from the vendor.

So instead of repeating the full five-part playbook four times, this chapter is
the **per-vendor diff**: for each device it states exactly **(1)** which rate
attributes ToughRADIUS emits and the unit multiplier, **(2)** how the request
MAC / VLAN is parsed (hence which bindings work), and **(3)** which flagship
scenario to follow for the end-to-end steps.

> **Read this with a flagship chapter open.** For the full step-by-step of any
> scenario, follow the matching section in the
> [MikroTik](./cookbook-mikrotik.md) or [Huawei](./cookbook-huawei.md) cookbook;
> only the emitted attributes below differ.

## Which playbook applies

| You want to… | Follow | What changes per vendor (this chapter) |
| --- | --- | --- |
| Speed tiers + address pool + expiry + concurrency | [MikroTik Scenario A](./cookbook-mikrotik.md#scenario-a-pppoe-broadband-isp--speed-tiers--address-pool--expiry-disconnect--concurrency) or [Huawei Scenario A](./cookbook-huawei.md#scenario-a-pppoe--ipoe-broadband--speed-tiers-peak-rate-and-aaa-domain) | the **rate attribute(s)** emitted + unit |
| Line anti-fraud (MAC / VLAN binding) | [Huawei Scenario B](./cookbook-huawei.md#scenario-b-line-anti-fraud--mac--vlan-binding-and-dual-stack-ipv6) | **MAC parse format** + whether **VLAN binding** is available |
| Live control (CoA / disconnect / FUP) | [MikroTik Scenario C](./cookbook-mikrotik.md#scenario-c-live-control--coa-forced-disconnect-and-fup) | nothing — CoA is vendor-agnostic (Session-Timeout + Filter-Id) |

> **The concurrency cap, expiry disconnect, address pool and CoA work the same
> for every vendor here** — they are enforced by the shared checkers /
> `default_enhancer`, not by a vendor VSA. The only real per-vendor variable is
> the rate-limit attribute and the MAC / VLAN parsing below.

---

## H3C — vendor code 25506

**Rate attributes** (anchored to `h3c_enhancer.go`) — the **same quartet shape as
Huawei**:

| Attribute | Value |
| --- | --- |
| `H3C-Input-Average-Rate` | `up_kbps × 1024` (bit/s) — subscriber upload |
| `H3C-Input-Peak-Rate` | `up_kbps × 1024 × 4` |
| `H3C-Output-Average-Rate` | `down_kbps × 1024` (bit/s) — download |
| `H3C-Output-Peak-Rate` | `down_kbps × 1024 × 4` |

So a "30M down" tier (`30720` Kbps) is sent as `H3C-Output-Average-Rate =
31457280`, peak `125829120`; all values clamp to the Int32 max. Unlike Huawei,
H3C emits **no** domain or IPv6 VSA — only the rate quartet.

**Request parsing** (`h3c_parser.go`): MAC comes from **`H3C-IP-Host-Addr`** (the
last 17 characters, i.e. the trailing `aa:bb:cc:dd:ee:ff`) and falls back to
`Calling-Station-Id` (`-`→`:`) when that VSA is absent; the inner / outer VLAN
come from `NAS-Port-Id`. **Both MAC and VLAN binding are supported.**

**Device side (H3C Comware, reference, verify on your firmware):**

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

> **Trap**: register the NAS as *H3C* (not *Huawei*). Although the rate maths is
> identical, the VSA **vendor IDs differ** (25506 vs 2011); a Huawei-registered
> H3C device would receive Huawei VSAs it ignores, and the rate limit would not
> apply.

---

## ZTE — vendor code 3902

**Rate attributes** (anchored to `zte_enhancer.go`) — **two attributes, no peak**:

| Attribute | Value |
| --- | --- |
| `ZTE-Rate-Ctrl-SCR-Up` | `up_kbps × 1024` (bit/s) |
| `ZTE-Rate-Ctrl-SCR-Down` | `down_kbps × 1024` (bit/s) |

A "30M down" tier is sent as `ZTE-Rate-Ctrl-SCR-Down = 31457280`. There is no
peak / burst attribute — the average rate is the cap.

**Request parsing** (`zte_parser.go`): MAC comes from `Calling-Station-Id`, but
ZTE sends it as a **bare 12-hex-digit string**; ToughRADIUS reformats it to
`aa:bb:cc:dd:ee:ff`. VLAN is parsed from `NAS-Port-Id`. **Both MAC and VLAN
binding are supported** — but for MAC binding, store the MAC in the
`aa:bb:cc:dd:ee:ff` form (that is what the parser produces).

**Device side**: ZTE BRAS uses the same radius-template + domain pattern as
Huawei — bind the authentication / accounting template to the ToughRADIUS
address, shared secret and ports 1812 / 1813 (verify on your firmware).

---

## iKuai — vendor code 10055

A popular SMB / SOHO gateway in China.

**Rate attributes** (anchored to `ikuai_enhancer.go`) — **two attributes, a
different multiplier**:

| Attribute | Value |
| --- | --- |
| `RP-Upstream-Speed-Limit` | `up_kbps × 8192` (= `× 1024 × 8`) |
| `RP-Downstream-Speed-Limit` | `down_kbps × 8192` |

So a "30M down" tier (`30720` Kbps) is sent as `RP-Downstream-Speed-Limit =
251658240`.

> **Trap — high tiers clamp.** Because the multiplier is `× 8192` and the value
> is clamped to the **Int32 max (2147483647)**, any tier above roughly **256
> Mbps** (`262144` Kbps) overflows and is **clamped** — a 300M tier would be
> sent at the clamped ceiling, not 300M. For very high tiers, apply the policy on
> the iKuai device instead.

**Request parsing**: iKuai uses the **default parser** — MAC from
`Calling-Station-Id` (`-`→`:`), and **VLAN is not parsed** (always `0`). So **MAC
binding works, but VLAN binding does not** (the VLAN check is always skipped).

**Device side**: on the iKuai web console, go to **认证计费 → RADIUS 计费**, set
the server address, ports 1812 / 1813 and the shared secret, then enable RADIUS
in the PPPoE server settings.

---

## Cisco — vendor code 9

**Rate attributes**: **none**. ToughRADIUS has **no Cisco enhancer**, so a Cisco
NAS receives only the standard attributes (`Session-Timeout`,
`Acct-Interim-Interval`, `Framed-Pool`, `Framed-IP-Address`, …) from
`default_enhancer.go`. Authentication (PAP / CHAP / MS-CHAPv2 / EAP), session
control, accounting and CoA all work normally — **apply bandwidth policy on the
device**, or build a custom integration using the bundled `Cisco-AVPair`
dictionary.

**Request parsing**: default parser — MAC from `Calling-Station-Id`, no VLAN. So
**MAC binding works, VLAN binding does not**.

**Device side (Cisco IOS, reference, verify on your platform):**

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

> `aaa server radius dynamic-author` enables CoA / Disconnect (default port
> 3799). Do **not** expect a rate limit from RADIUS on Cisco — set it with a
> service-policy / QoS on the device.

---

## Vendor capability matrix

| Vendor | Rate attribute(s) | Unit multiplier | Peak | MAC bind | VLAN bind |
| --- | --- | --- | --- | --- | --- |
| H3C (25506) | `H3C-Input/Output-Average-Rate` + peaks | `× 1024` (bit/s) | ✅ avg × 4 | ✅ | ✅ |
| ZTE (3902) | `ZTE-Rate-Ctrl-SCR-Up/Down` | `× 1024` (bit/s) | ❌ | ✅ | ✅ |
| iKuai (10055) | `RP-Upstream/Downstream-Speed-Limit` | `× 8192` | ❌ | ✅ | ❌ |
| Cisco (9) | none (device-side QoS) | — | — | ✅ | ❌ |

(For comparison: MikroTik emits the `Mikrotik-Rate-Limit` string in Kbps;
Huawei emits the rate quartet `× 1024` with peak `× 4` plus domain / IPv6.)

## Troubleshooting (symptom → locate → fix)

- **Rate limit has no effect** → ① the NAS is not registered with this exact
  vendor (the VSA vendor ID must match — H3C ≠ Huawei even though the maths is
  the same); ② Cisco has **no** RADIUS rate VSA by design — set QoS on the
  device; ③ wrong unit on the device side (all values here are bit/s, except the
  iKuai `× 8192`).
- **iKuai high-tier speed is wrong / capped low** → the `× 8192` value overflowed
  Int32 and was clamped; apply high tiers on the device.
- **VLAN binding never triggers on iKuai / Cisco** → expected: their parser does
  not extract VLANs (always `0`), so the VLAN check is always skipped. Use MAC
  binding instead, or a vendor that parses VLAN (Huawei / H3C / ZTE).
- **ZTE MAC binding never matches** → store the MAC as `aa:bb:cc:dd:ee:ff`
  (ZTE sends 12 bare hex digits; ToughRADIUS reformats to colon form before
  comparing).

---

## Related chapters

- [Cookbook: MikroTik RouterOS](./cookbook-mikrotik.md) — full PPPoE / Hotspot /
  CoA playbook.
- [Cookbook: Huawei BRAS / NetEngine](./cookbook-huawei.md) — full speed-tier /
  line-binding / CoA playbook.
- [Vendor Integration Guide](./vendor-guide.md) — the attribute reference card
  for every vendor.
- [FAQ](./faq.md) — more cross-scenario troubleshooting Q&A.
