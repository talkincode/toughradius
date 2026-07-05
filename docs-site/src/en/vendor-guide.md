# Vendor Integration Guide

> 中文版本：[厂商对接指南](../zh/vendor-guide.md)

ToughRADIUS speaks standard RADIUS to every device and adds **vendor-specific
attributes (VSAs)** for the vendors it knows. This chapter walks through the
integration steps shared by all devices, then gives a case study per vendor:
what ToughRADIUS sends, what it parses, and a reference configuration for the
device side.

> **The vendor code on the NAS record decides everything.** Attribute
> enhancement is selected by the *Vendor* field of the NAS device record in the
> admin UI — not by inspecting packets. If you leave a MikroTik registered as
> `Standard`, it will authenticate fine but receive **no** `Mikrotik-Rate-Limit`
> (no bandwidth enforcement). Pick the right vendor first.

> 📖 Want end-to-end operational examples (PPPoE speed tiers, Hotspot + MAC auth,
> CoA / forced disconnect)? See the [Scenario Cookbook](./cookbook.md). This
> chapter is the attribute reference card; the cookbook is the playbook.

> **Portal boundary:** ToughRADIUS is the RADIUS AAA backend. It does not host
> captive portal login pages or guest onboarding flows. See
> [Portal / Hotspot Integration Boundary](./portal-hotspot-boundary.md).

## Integration steps for any device

1. **Register the NAS** under **NAS Devices → Create**: source IP address (or
   identifier), shared secret, and the correct *Vendor*.
2. **Point the device** at the server: authentication UDP `1812`, accounting
   UDP `1813`, the same shared secret.
3. **Optional CoA**: ToughRADIUS sends CoA/Disconnect (RFC 5176) to the NAS on
   UDP `3799` by default; set the *CoA port* field on the NAS record if your
   device listens elsewhere. Each exchange waits up to 5 s and retransmits
   twice.
4. **Create a billing profile and users**, then test with
   `go run ./cmd/radtest auth …` ([Quick Start](./quickstart.md)).

### What every vendor receives (standard attributes)

Regardless of vendor, every `Access-Accept` may carry: `Session-Timeout`
(seconds until the account expires), `Acct-Interim-Interval`, `Framed-Pool`,
`Framed-IP-Address` (static IPv4), `Framed-IPv6-Prefix` / `Framed-IPv6-Address`
(RFC 6911), `Framed-IPv6-Pool`, `Delegated-IPv6-Prefix` (RFC 4818), and
`Delegated-IPv6-Prefix-Pool` — depending on which user/profile fields are set.

### Rate-limit units

Profile rates are stored in **Kbps**. Each vendor enhancer converts:

| Vendor | Attributes | Value sent |
| ------ | ---------- | ---------- |
| Huawei (2011) | `Huawei-Input/Output-Average-Rate`, `Huawei-Input/Output-Peak-Rate` | average = `rate_kbps × 1024` (bit/s); peak = `× 4` further; clamped to Int32 max |
| MikroTik (14988) | `Mikrotik-Rate-Limit` | string `"{up}k/{down}k"`, e.g. `51200k/102400k` (rx/tx from the router's view) |
| H3C (25506) | `H3C-Input/Output-Average-Rate`, peak variants | same ×1024 / ×4 scheme as Huawei |
| ZTE (3902) | `ZTE-Rate-Ctrl-SCR-Up/Down` | `rate_kbps × 1024` |
| iKuai (10055) | `RP-Upstream/Downstream-Speed-Limit` | `rate_kbps × 1024 × 8`, clamped to Int32 max |
| Cisco (9), Standard (0) | — | no rate VSA — bandwidth is device-side; the `accept-cisco` enhancer emits a non-rate `Cisco-AVPair="ip:addr-pool=<pool>"` |

### Request parsing (MAC and VLAN)

The default parser reads the MAC address from `Calling-Station-Id`. Vendor
parsers add request-side VSA or `NAS-Port-Id` handling where the device family
has a stable encoding:

- `slot/subslot/port:vlan[.vlan2]` — e.g. `3/0/1:2814.727`
- `vlanid=<n>;vlanid2=<n>;` — e.g. `slot=2;...;vlanid=503;vlanid2=100;`

| Vendor | MAC source | VLAN source | Notes |
| ------ | ---------- | ----------- | ----- |
| Huawei (2011) | `Calling-Station-Id` | `NAS-Port-Id` | Supports both encodings above. |
| H3C (25506) | `Calling-Station-Id` | `NAS-Port-Id` | Supports both encodings above. |
| ZTE (3902) | `Calling-Station-Id` | `NAS-Port-Id` | Supports both encodings above. |
| Radback (2352) | `Mac-Addr` VSA, falling back to `Calling-Station-Id` | `Bind-Dot1q-Vlan-Tag-Id` VSA, falling back to `NAS-Port-Id` | Request parser only; no Radback response enhancer is shipped. |
| Alcatel (3041) | `AAT-User-MAC-Address` VSA, falling back to `Calling-Station-Id` | `NAS-Port-Id` when present | Request parser only; response attributes remain deployment-specific. |
| Aruba/HPE (14823) | `Calling-Station-Id` | `Aruba-User-Vlan` VSA, falling back to `NAS-Port-Id` | Also has an Access-Accept enhancer; see below. |
| Juniper (2636) | `Calling-Station-Id` | `Juniper-VoIP-Vlan` VSA, falling back to `NAS-Port-Id` | Request parser only; response attributes remain deployment-specific. |
| MikroTik, iKuai, Cisco, Standard | `Calling-Station-Id` | — | No vendor-specific VLAN parser; use device-side policy or custom integration. |

MAC binding works with every vendor that sends a recognizable MAC address.
VLAN binding requires one of the VLAN-aware parsers above and a matching
request encoding from the NAS.

> Device-side snippets below are **reference examples** — command syntax varies
> by model and OS version; always consult the vendor documentation.

---

## MikroTik (RouterOS) — vendor code 14988

Best-known integration: PPPoE / Hotspot with `Mikrotik-Rate-Limit`.

ToughRADIUS sends `Mikrotik-Rate-Limit = "{up}k/{down}k"`; RouterOS applies it
as a dynamic simple queue (rx-rate/tx-rate from the router's perspective, i.e.
subscriber upload first).

```routeros
/radius add service=ppp,hotspot address=<TOUGHRADIUS_IP> secret=<SECRET> \
    timeout=3s
/radius incoming set accept=yes port=3799
/ppp aaa set use-radius=yes accounting=yes interim-update=5m
```

- `radius incoming accept=yes` enables CoA/Disconnect on UDP 3799.
- For Hotspot: enable RADIUS in the hotspot server profile
  (`/ip hotspot profile set ... use-radius=yes`).

## Huawei — vendor code 2011

Typical BRAS (ME60/NE) / aggregation deployments. ToughRADIUS sends the rate
quartet (`Huawei-Input/Output-Average-Rate`, peaks ×4), `Huawei-Domain-Name`
(when the user/profile has a domain), and `Huawei-Framed-IPv6-Address` for
static IPv6. The Huawei parser extracts VLANs from `NAS-Port-Id`, so MAC *and*
VLAN binding both work.

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

For CoA/Disconnect, enable the RADIUS dynamic authorization extension
(`radius-server authorization` on the device) toward the server address.

## Cisco — vendor code 9

ToughRADIUS authenticates Cisco devices with standard attributes (PAP / CHAP /
MS-CHAPv2 / EAP all work; sessions, accounting, CoA likewise). When the user's
plan has an address pool, the `accept-cisco` enhancer emits
`Cisco-AVPair="ip:addr-pool=<pool>"` alongside the standard `Framed-Pool`; no
other Cisco-specific attributes are sent. Apply bandwidth policy on the device
(Cisco has no portable numeric rate VSA), or extend via the bundled
`Cisco-AVPair` dictionary if you build a custom integration.

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

`aaa server radius dynamic-author` enables CoA/Disconnect (default port 3799).

## H3C — vendor code 25506

Same rate semantics as Huawei (`H3C-Input/Output-Average-Rate` ×1024, peak ×4).
The H3C parser extracts VLANs, so VLAN binding is supported.

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

## ZTE — vendor code 3902

ToughRADIUS sends `ZTE-Rate-Ctrl-SCR-Up/Down` (rate ×1024) and parses VLANs
from `NAS-Port-Id`. Configuration on ZTE BRAS follows the same
radius-template + domain pattern as Huawei; bind the authentication/accounting
template to the server address, secret, and ports 1812/1813.

## iKuai — vendor code 10055

Popular SMB gateway in China. ToughRADIUS sends
`RP-Upstream-Speed-Limit` / `RP-Downstream-Speed-Limit` (= `rate_kbps × 8192`,
clamped). On the iKuai web console: **认证计费 → RADIUS 计费** — set the server
address, ports 1812/1813, and the shared secret; enable RADIUS in the PPPoE
server settings.

## Aruba/HPE — vendor code 14823

Aruba/HPE devices have both request parsing and Access-Accept enhancement.
The parser extracts VLAN information from `Aruba-User-Vlan` when the request
carries it, with the same `NAS-Port-Id` fallback used by other VLAN-aware
parsers.

On `Access-Accept`, ToughRADIUS sends:

| Attribute | Source | Boundary / no-op rule |
| --------- | ------ | --------------------- |
| `Aruba-User-Vlan` | user or profile `Vlanid1` | Sent only for VLAN IDs `1..4094`; missing, `0`, `4095`, or negative values are omitted. `Vlanid2` has no Aruba attribute mapping and is not sent. |
| `Aruba-User-Role` | user or profile `Domain` | Sent when the inherited domain / role field is non-empty and not `N/A`. |

No Aruba VSAs are added for non-Aruba NAS records, and unset fields are left out
instead of emitting placeholder values.

## Standard / other devices — vendor code 0

Any RFC-compliant NAS (pfSense, strongSwan, FreeRADIUS clients, Wi-Fi
controllers, …) can authenticate against ToughRADIUS with vendor code
`Standard`: full credential validation, session control, accounting, IPv4/IPv6
attributes — but no proprietary rate attributes. Attribute **dictionaries** for
more vendors (Microsoft, F5, PfSense, Hillstone, …) ship in the codebase for
custom development. For Juniper, Alcatel, Aruba, and Radback, the shipped
capability is more specific than "dictionary only": see the request parser and
Aruba enhancer boundaries above. A dictionary alone still does not parse
requests or enhance accepts unless a parser or enhancer is registered.

## Troubleshooting an integration

| Symptom | Likely cause |
| ------- | ------------ |
| Device gets `Access-Accept` but no bandwidth limit | NAS record vendor is `Standard`, or the device ignores the VSA — check the vendor code first |
| All requests silently dropped | Source IP not registered as a NAS, or shared secret mismatch |
| VLAN binding never matches | NAS vendor is not one of the VLAN-aware parsers above, or the request uses an unexpected VSA / `NAS-Port-Id` format |
| CoA/Disconnect times out | Device CoA listener disabled, or non-default port — set *CoA port* on the NAS record |

More in the [FAQ](./faq.md).
