# TR-F005 / M5.1 Vendor VSA Gap Baseline

This document is the M5.1 inventory for "vendor backlog and dictionary gaps"
under TR-F005. It is intentionally snapshot-style so the next M5 batch can
execute without re-discovering the same scope.

> Refreshed after the M5.2 request-parser batch (#449–#454) and the M5.3 Aruba
> response enhancer (#456) landed. The first baseline (#433) predated that work
> and is superseded by the matrix below; see "Delta since the first baseline".

## Scope and Method

Inspection sources:

- `internal/radiusd/vendors/*`
- `internal/radiusd/vendors/codes.go`
- `internal/radiusd/plugins/vendorparsers/parsers/init.go` (parser registration —
  the source of truth for what is actually active, not just the `*_parser.go`
  files present)
- `internal/radiusd/plugins/init.go` (`RegisterResponseEnhancer` registration)
- `share/dictionary` and `share/dictionary.*`

Snapshot commands (reproducible):

```bash
git rev-parse --short HEAD                                  # 9882f79e
ls internal/radiusd/vendors
grep -E 'Code[A-Z]' internal/radiusd/vendors/codes.go
grep -E 'Code(Huawei|H3C|ZTE|Radback|Alcatel|Aruba|Juniper)' \
  internal/radiusd/plugins/vendorparsers/parsers/init.go    # registered parsers
grep RegisterResponseEnhancer internal/radiusd/plugins/init.go
grep -c '^\$INCLUDE' share/dictionary                       # 213
```

## Coverage Snapshot (HEAD 9882f79e)

- Generated vendor dictionary packages: `15`
- Registered vendor parsers (`parsers/init.go`): `default + huawei + h3c + zte +
  radback + alcatel + aruba + juniper` (7 vendors + default)
- Registered response enhancers (`plugins/init.go`): `default + huawei + h3c +
  zte + mikrotik + ikuai + aruba` (6 vendors + default)
- `share/dictionary` includes `213` `$INCLUDE dictionary.*` entries; the repo
  ships generated packages for `15` of those vendors.

## Why a parser is not always worth adding

`vendorparsers.VendorRequest` carries only `MacAddr`, `Vlanid1`, `Vlanid2`. A
vendor **parser** therefore adds value only when the vendor encodes MAC/VLAN with
a vendor-specific **request** VSA; otherwise it merely duplicates `DefaultParser`,
which already reads the standard `Calling-Station-Id`. The response **enhancer**
(rate / VLAN / `avpair` reply attributes) is a separate, demand-driven track and
does not depend on a parser.

## M5.1 Gap Matrix (current)

| Vendor Package | Vendor ID | `Code*` Constant | Parser | Enhancer | Dictionary Source |
| --- | ---: | --- | --- | --- | --- |
| `alcatel` | 3041 | present | present | missing | `share/dictionary.alcatel` |
| `aruba` | 14823 | present | present | present | `share/dictionary.aruba` |
| `cisco` | 9 | present | missing | missing | `share/dictionary.cisco` |
| `f5` | 3375 | present | missing | missing | `share/dictionary.f5` |
| `h3c` | 25506 | present | present | present | `share/dictionary.h3c` |
| `hillstone` | 28557 | present | missing | missing | `share/dictionary.hillstone` |
| `huawei` | 2011 | present | present | present | `share/dictionary.huawei` |
| `ikuai` | 10055 | present | missing | present | no `share/dictionary.ikuai` in tree |
| `juniper` | 2636 | present | present | missing | `share/dictionary.juniper` |
| `microsoft` | 311 | present | missing | missing | `share/dictionary.microsoft` |
| `mikrotik` | 14988 | present | missing | present | `share/dictionary.mikrotik` |
| `pfSense` | 13644 | present | missing | missing | `share/dictionary.pfsense` (case mapping) |
| `radback` | 2352 | present | present | missing | no `share/dictionary.radback`/`.redback` in tree |
| `unix` | 4 | missing (`CodeUnix`) | missing | missing | `share/dictionary.unix` |
| `zte` | 3902 | present | present | present | `share/dictionary.zte` |

## Delta since the first baseline (#433)

The M5.2 / M5.3 batches closed the high-value request-side parser gaps and the
first response enhancer that the first baseline listed as pending:

- `radback` request parser — MAC + VLAN (#449)
- `alcatel` request parser — MAC, plus the `vendors.CodeAlcatel` constant (#450)
- `aruba` request parser — VLAN, plus the `vendors.CodeAruba` constant (#451)
- `juniper` request parser — VoIP VLAN (#453)
- `aruba` Access-Accept response enhancer (#456, M5.3)

The first baseline also reported `alcatel` / `aruba` as missing their `Code*`
constant; both now exist in `codes.go`. Only `unix` still lacks a `Code*`
constant.

## Remaining gaps / backlog for the next M5 batch

The high-value request-side parsers are now delivered (see Delta above), so the
remaining TR-F005 work is response-enhancer and hygiene work:

1. **Response enhancers (demand-driven), no parser dependency.** Highest value
   first:
   - `cisco` `cisco-avpair` — the most common vendor reply attribute; `cisco`
     already has its `Code*` constant and dictionary package, only the enhancer
     is missing. Tracked as **M5.4**.
   - `alcatel`, `juniper`, `radback` — already have a request parser but no
     response enhancer; add when a deployment needs their reply attributes.
   - `microsoft`, `f5`, `hillstone`, `pfSense` — neither parser nor enhancer;
     close as enhancer / registry work when demand surfaces.
2. **Parsers — high-value request-side MAC/VLAN vendors are done.** Add a new
   parser only when a vendor dictionary proves a vendor-specific MAC/VLAN
   **request** encoding (otherwise it duplicates `DefaultParser`).
3. **Symmetry-only parsers stay deferred.** `mikrotik`, `ikuai` already ship
   response enhancers; their VLAN VSAs (e.g. `Mikrotik-Wireless-VLANID`) are
   Access-Accept reply attributes and their request side uses the standard
   `Calling-Station-Id`, so a dedicated parser would behave identically to
   `DefaultParser`.
4. **Hygiene.** `unix` (VendorID 4) has a generated package and a
   `share/dictionary.unix` source but no `vendors.CodeUnix` constant; add the
   constant only if/when a `unix` parser or enhancer is actually wired.
5. **Per-vendor acceptance set (unchanged).**
   - parser + registration in `parsers/init.go`, and/or enhancer + registration
     in `plugins/init.go`
   - sample-based parser / enhancer tests with real attribute samples
   - `go test ./internal/radiusd/...` and `golangci-lint` (v2.12.2) green

## Non-Goals of M5.1

- No runtime behavior changes
- No new parser/enhancer implementation in this round
- No expansion beyond TR-F005 scope
