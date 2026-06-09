# Protocol & RFC Reference

> 中文版本：[协议与 RFC 索引](../zh/rfc-index.md)

ToughRADIUS implements standard RADIUS, EAP, dynamic-authorization, and
secure-transport protocols. This chapter is the curated, implementation-oriented
index of the standards the project relies on, with each RFC mapped to where it is
used in the code and on the
[roadmap](https://github.com/talkincode/toughradius/blob/main/docs/roadmap.md).

The full RFC texts are archived under
[`docs/rfcs/`](https://github.com/talkincode/toughradius/tree/main/docs/rfcs);
the raw catalog of every file lives in
[`docs/rfcs/README.md`](https://github.com/talkincode/toughradius/blob/main/docs/rfcs/README.md).
Where the two differ, the citations in this chapter take precedence.

## Implemented standards

### RADIUS core

- **RFC 2865** — RADIUS authentication; the base request/response protocol.
- **RFC 2866** — RADIUS accounting; session start / interim-update / stop records.

### RADIUS + EAP integration

- **RFC 3579** — RADIUS support for EAP (EAP-Message and Message-Authenticator).
- **RFC 3580** — IEEE 802.1X RADIUS usage guidelines.

### EAP framework and methods

- **RFC 3748** — Extensible Authentication Protocol (EAP). The framework itself;
  also defines EAP-MD5, the MD5-Challenge method (§5.4).
- **RFC 5216** — EAP-TLS; certificate-based mutual authentication (milestone M1).
- **RFC 5281** — EAP-TTLSv0; a TLS tunnel carrying inner PAP / MS-CHAPv2 (M9).
- **RFC 2759** — MS-CHAP-V2, used as the inner method of EAP-MSCHAPv2 and PEAPv0
  (M8). MS-MPPE session keys are derived per RFC 2548 / RFC 5705.

> PEAPv0 has no standalone RFC; it follows the Microsoft/Cisco PEAP definition and
> carries inner EAP-MSCHAPv2. It is compatibility-oriented — MS-CHAPv2 exchanges
> carry an NTLMv1-like attack surface — so prefer EAP-TLS where you control client
> certificates.

### Dynamic authorization

- **RFC 5176** — CoA and Disconnect-Request, superseding RFC 3576 (milestone M2).

### Secure transport

- **RFC 6614** — RADIUS over TLS (RadSec).
- **RFC 6613** — RADIUS over TCP.

### Vendor-specific attributes

- **RFC 2548** — Microsoft VSAs, including the MS-MPPE-Send/Recv-Key attributes
  used to carry EAP key material.

### IPv6

- **RFC 3162**, **RFC 4818**, **RFC 6911** — RADIUS IPv6 addressing and
  delegated-prefix attributes (milestone M3).

## Roadmap standards

| RFC | Standard | Milestone |
| --- | --- | --- |
| RFC 9190 (+ RFC 9427) | EAP-TLS 1.3 and TLS 1.3 key derivation | M10 |
| RFC 7170 / RFC 9930 | TEAP v1 — tunnel EAP, machine + user chaining | M11 |
| RFC 5931 | EAP-PWD — password-based, no per-client certificate | M12 |

> **Catalog accuracy note.** In `docs/rfcs/`, the file `rfc7542-eap-pwd.txt` is
> mislabeled: **RFC 7542 is the Network Access Identifier (NAI)**, whereas
> **EAP-PWD is RFC 5931**. Likewise, EAP-MD5 is defined by **RFC 3748 §5.4**, not
> RFC 3851 (an S/MIME specification). This chapter uses the correct citations.

## See also

- [Overview](./overview.md) — capability summary, including the EAP suite.
- [Documentation Map](./documentation-map.md) — where every source document lives.
- [RFC Editor](https://www.rfc-editor.org/) — authoritative online RFC texts.
