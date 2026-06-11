# TR-F005 / M5.1 Vendor VSA Gap Baseline

This document records the M5.1 baseline for "vendor backlog and dictionary gaps"
under TR-F005. It is intentionally snapshot-style so M5.2 can execute by batch
without re-discovering the same scope.

## Scope and Method

Inspection sources:

- `internal/radiusd/vendors/*`
- `internal/radiusd/vendors/codes.go`
- `internal/radiusd/plugins/vendorparsers/parsers/*_parser.go`
- `internal/radiusd/plugins/auth/enhancers/*_enhancer.go`
- `share/dictionary` and `share/dictionary.*`

Snapshot commands (reproducible):

```bash
git rev-parse --short HEAD
ls internal/radiusd/vendors
ls internal/radiusd/plugins/vendorparsers/parsers/*_parser.go
ls internal/radiusd/plugins/auth/enhancers/*_enhancer.go
rg '^\$INCLUDE\s+dictionary\.' share/dictionary
```

## Coverage Snapshot

- Generated vendor dictionary packages: `15`
- Registered vendor parsers: `default + huawei + h3c + zte`
- Registered vendor enhancers: `default + huawei + h3c + zte + mikrotik + ikuai`
- `share/dictionary` includes `213` dictionary entries; the repo currently ships
  generated packages for `15` vendors.

## M5.1 Gap Matrix

| Vendor Package | Vendor ID | `Code*` Constant | Parser | Enhancer | Dictionary Source Gap |
| --- | ---: | --- | --- | --- | --- |
| `alcatel` | 3041 | missing | missing | missing | `share/dictionary.alcatel` exists |
| `aruba` | 14823 | missing | missing | missing | `share/dictionary.aruba` exists |
| `cisco` | 9 | present | missing | missing | `share/dictionary.cisco` exists |
| `f5` | 3375 | present | missing | missing | `share/dictionary.f5` exists |
| `h3c` | 25506 | present | present | present | none |
| `hillstone` | 28557 | present | missing | missing | `share/dictionary.hillstone` exists |
| `huawei` | 2011 | present | present | present | none |
| `ikuai` | 10055 | present | missing | present | no `share/dictionary.ikuai` file in tree |
| `juniper` | 2636 | present | missing | missing | `share/dictionary.juniper` exists |
| `microsoft` | 311 | present | missing | missing | `share/dictionary.microsoft` exists |
| `mikrotik` | 14988 | present | missing | present | `share/dictionary.mikrotik` exists |
| `pfSense` | 13644 | present | missing | missing | source file name is `dictionary.pfsense` (case mapping) |
| `radback` | 2352 | present | missing | missing | no `share/dictionary.radback` or `dictionary.redback` file |
| `unix` | 4 | missing | missing | missing | `share/dictionary.unix` exists |
| `zte` | 3902 | present | present | present | none |

## Prioritized Backlog for M5.2

> Prioritization corrected during M5.1 → M5.2 grooming, anchored to the gap
> matrix above and to the `vendorparsers.VendorRequest` model (it carries only
> `MacAddr` + two `Vlanid` fields). A vendor **parser** therefore adds value only
> when the vendor encodes MAC/VLAN with a vendor-specific VSA **on the request
> side**; otherwise it merely duplicates `DefaultParser` (which already reads the
> standard `Calling-Station-Id`). The response **enhancer** (rate/VLAN reply
> attributes) is a separate, demand-driven track and does not depend on a parser.

1. Parsers with genuine request-side value — vendors whose dictionary exposes a
   vendor-specific MAC/VLAN request attribute (verify request/response semantics
   per vendor dictionary + spec before implementing):
   - `radback` (`Mac-Addr` / `Bind-Dot1q-Vlan-Tag-Id`)
   - `alcatel` (`AAT-User-MAC-Address`)
   - `aruba` (`Aruba-User-Vlan`)
   - `juniper` (`Juniper-VoIP-Vlan`)
2. Enhancers (response attributes) — drive by deployment demand, independently of
   parser work; `cisco` (`cisco-avpair`), `juniper`, `microsoft` are common asks.
3. De-prioritized "symmetry-only" parsers — `mikrotik`, `ikuai` use the standard
   `Calling-Station-Id` on the request side (their VLAN VSAs such as
   `Mikrotik-Wireless-VLANID` are Access-Accept reply attributes), so a dedicated
   parser would be behaviorally identical to `DefaultParser`. Add one only if a
   concrete deployment proves a vendor-specific request encoding. (The original
   baseline listed these as batch 1 — that was a misjudgment, now corrected.)
4. Remaining generated vendors with no current request-side parser value:
   `cisco`, `f5`, `hillstone`, `pfSense`, `unix`, `microsoft` — close as
   enhancer/registry work or when a deployment surfaces a real attribute need.
   Note: `alcatel`, `aruba`, `unix` still lack a `vendors.Code*` constant; add it
   before registering a parser/enhancer.
5. For each vendor in M5.2, follow the same acceptance set:
   - parser + registration + enhancer (if response VSA required)
   - sample-based parser/enhancer tests
   - `go test ./internal/radiusd/...` and `golangci-lint` green

## Non-Goals of M5.1

- No runtime behavior changes
- No new parser/enhancer implementation in this round
- No expansion beyond TR-F005 scope
