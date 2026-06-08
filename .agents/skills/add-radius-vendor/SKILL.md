---
name: add-radius-vendor
description: Add or extend vendor VSA parsing and response enhancement (TR-F005). Use when ToughRADIUS must recognize a vendor's request VSAs, or emit that vendor's proprietary attributes in Access-Accept.
---

# Skill: Add Vendor VSA Parsing / Response Enhancement

> Feature ID: `TR-F005` | Milestone: M5

## When to use
When ToughRADIUS must recognize a vendor's request VSAs, or emit that vendor's specific attributes in Access-Accept.

## Pre-research (read before writing)
```text
grep_search "VendorCode" --include internal/radiusd/**
file_search "internal/radiusd/plugins/vendorparsers/parsers/*_parser.go"
file_search "internal/radiusd/plugins/auth/enhancers/*_enhancer.go"
view internal/radiusd/plugins/vendorparsers/parsers/init.go      # registration point
view internal/radiusd/vendors/<existing-vendor>/                 # dictionary & constants
```
Key insight: a dictionary != a parser. A dictionary only describes attributes; without a parser they are never extracted.

## Implementation steps
1. **Constants / dictionary**: define the VendorCode and VSA constants under `internal/radiusd/vendors/<vendor>/` (reference huawei). If the `vendors` package lacks the `Code<Vendor>` constant, add it first.
2. **Parser**: implement the parser at `internal/radiusd/plugins/vendorparsers/parsers/<vendor>_parser.go`, mirroring `huawei_parser.go`'s interface and field extraction.
3. **Register the parser**: in `parsers/init.go`'s `init()`, call `vendors.Register(&vendors.VendorInfo{Code: vendors.Code<Vendor>, ...Parser: &<Vendor>Parser{}})`.
4. **Enhancer (if emitting response attributes)**: implement `internal/radiusd/plugins/auth/enhancers/<vendor>_enhancer.go`, mirroring `huawei_enhancer.go` (rate, VLAN, etc.).
5. **Sample-based tests**: add `<vendor>_parser_test.go` / `<vendor>_enhancer_test.go` covering parsing and emission with real attribute samples.

## Conventions
- Vendor bandwidth-unit differences and binary-vs-decimal conversions must carry inline comments explaining the "why".
- Rate / VLAN / MAC-binding behavior must be semantically consistent with existing vendors.

## Acceptance
- [ ] `go test ./internal/radiusd/...` passes
- [ ] `golangci-lint run` reports no new issues
- [ ] Both new vendor parsing and response have sample test coverage
- [ ] PR references `TR-F005` and the milestone ID
