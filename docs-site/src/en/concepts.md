# Concepts & Terminology

> 中文版本：[核心术语与概念](../zh/concepts.md)

This chapter explains the core AAA terminology used throughout ToughRADIUS and
maps each concept to where it lives in the product. For the authoritative list
of protocol specifications and how each one maps to the code, see the
[Protocol & RFC Reference](./rfc-index.md).

## AAA in one paragraph

RADIUS (Remote Authentication Dial In User Service) is the protocol that lets a
network device ask a central server three questions: *who is this user*
(**Authentication**), *what are they allowed to do* (**Authorization**), and
*what did they consume* (**Accounting**). ToughRADIUS answers all three: it
validates credentials, returns authorization attributes (IP address, bandwidth,
VLAN, session limits) in the `Access-Accept`, and records usage from accounting
packets.

## Core terms

| Term | Meaning in ToughRADIUS |
| ---- | ---------------------- |
| **NAS** (Network Access Server) | The network device — a router, switch, BRAS, or wireless controller — that sends RADIUS requests. Each NAS is registered under **NAS Devices** in the admin UI with its IP address, shared secret, and vendor code. Requests from unregistered addresses are dropped. |
| **Shared secret** | A per-NAS password that authenticates RADIUS packets themselves (RFC 2865 §3). It must match on both the NAS and the ToughRADIUS NAS record. |
| **Subscriber / RADIUS user** | An account under **RADIUS Users**: username, password, expiration, optional static IPv4/IPv6 addresses, MAC/VLAN bindings, and a billing profile. |
| **Billing profile** (`RadiusProfile`) | A reusable template under **Billing Profiles**: concurrent-session limit (`active_num`), upload/download rate in **Kbps**, address pool, IPv6 prefix, domain, and MAC/VLAN binding switches. A user inherits profile values for fields it leaves empty. |
| **VSA** (Vendor-Specific Attribute) | RADIUS attribute 26 (RFC 2865 §5.26) — a container that lets each vendor define private attributes. ToughRADIUS ships attribute dictionaries for 15+ vendors and emits vendor-specific authorization attributes based on the NAS vendor code. See the [Vendor Integration Guide](./vendor-guide.md). |
| **Vendor code** | The IANA Private Enterprise Number that selects vendor behavior, e.g. `9` Cisco, `2011` Huawei, `14988` MikroTik, `25506` H3C, `3902` ZTE, `10055` iKuai, `0` standard. Set per NAS record; it controls which request parser and which Access-Accept attributes are used. |
| **CoA / Disconnect** (RFC 5176) | Dynamic Authorization: server-initiated messages that change a live session (`CoA-Request`, e.g. new `Session-Timeout` or `Filter-Id`) or terminate it (`Disconnect-Request`). Sent from the **Online Sessions** page to the NAS on UDP port 3799 (overridable per NAS). |
| **RadSec** (RFC 6614) | RADIUS over TLS on TCP port 2083. Wraps RADIUS in a TLS tunnel so packets can safely cross untrusted networks; plain UDP RADIUS relies only on the shared secret. |
| **EAP** (RFC 3748) | The Extensible Authentication Protocol used by 802.1X networks. ToughRADIUS implements EAP-MD5, EAP-MSCHAPv2, EAP-TLS, PEAPv0/EAP-MSCHAPv2, and EAP-TTLS; the active method and certificates are chosen in **System Config → RADIUS**. |
| **Accounting session** | The lifecycle reported by the NAS via `Accounting-Request` packets: Start → Interim-Update(s) → Stop (RFC 2866). Live sessions appear under **Online Sessions** (`radius_online` table); history under **Accounting** (`radius_accounting`). |
| **Acct-Interim-Interval** | How often (seconds) the NAS should send interim accounting updates. Returned in every Access-Accept from the `radius.AcctInterimInterval` setting. |
| **Session-Timeout** | Maximum remaining session time in seconds. ToughRADIUS sets it to the time left until the user's expiration date, so a session never outlives the account. |
| **Address pool** (`Framed-Pool`) | A named IP pool configured on the NAS. ToughRADIUS only returns the pool *name*; the NAS allocates the actual address. Static addresses, by contrast, are returned directly as `Framed-IP-Address` / `Framed-IPv6-Address`. |
| **MAC binding** | When a profile enables `bind_mac`, the first calling MAC seen is stored on the user and later requests must match it. |
| **VLAN binding** | When a profile enables `bind_vlan`, the inner/outer VLAN IDs parsed from `NAS-Port-Id` are stored and enforced the same way. Requires a vendor parser that extracts VLANs (Huawei, H3C, ZTE). |

## How an authentication request flows

```text
NAS ──Access-Request──▶ UDP 1812
        │
        ▼
  goroutine pool (ants, TOUGHRADIUS_RADIUS_POOL, default 1024)
        │
        ▼
  1. NAS lookup ─ source IP / identifier must match a registered NAS record
  2. Vendor parser ─ extracts MAC (Calling-Station-Id) and VLANs (NAS-Port-Id)
       · Huawei / H3C / ZTE parsers extract VLAN IDs
       · all other vendors use the default parser (MAC only)
  3. Credential validation ─ PAP / CHAP / MS-CHAPv2, or the EAP state machine
  4. Checkers ─ account status, expiration, MAC bind, VLAN bind, concurrent
     session count (active_num)
  5. Accept enhancers ─ standard attributes (Session-Timeout, pools, static
     IPs, IPv6) plus vendor attributes selected by the NAS vendor code
        │
        ▼
NAS ◀─Access-Accept / Access-Reject──
```

Failed authentications are classified into Prometheus-style counters
(`radus_reject_passwd_error`, `radus_reject_expire`, …) that feed the
dashboard; see the [Operations Guide](./ops-guide.md#metrics).
A reject-delay guard slows brute-force attempts: after
`radius.RejectDelayMaxRejects` consecutive rejects (default 7) inside
`radius.RejectDelayWindowSeconds` (default 10 s), responses are delayed.

## Password protocols at a glance

| Protocol | Where the password travels | Notes |
| -------- | -------------------------- | ----- |
| **PAP** | In the request, XOR-protected by the shared secret (RFC 2865 §5.2) | Universally supported; pair with RadSec or a trusted network. |
| **CHAP** | Never — MD5 challenge/response (RFC 2865 §5.3) | Requires the server to know the cleartext password. |
| **MS-CHAPv2** | Never — NT-hash challenge/response (RFC 2548) | Used by Windows; carries a well-known NTLMv1-like attack surface. |
| **EAP (tunneled)** | Inside a TLS tunnel (PEAP / EAP-TTLS) or replaced by certificates (EAP-TLS) | Preferred for Wi-Fi / 802.1X deployments. |

## Related chapters

- [Protocol & RFC Reference](./rfc-index.md) — every RFC cited above, mapped to code.
- [Vendor Integration Guide](./vendor-guide.md) — per-vendor attributes and formulas.
- [Quick Start](./quickstart.md) — see these concepts in action in 10 minutes.
