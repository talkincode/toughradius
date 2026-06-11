# Scenario Cookbook

> 中文版本：[场景实战手册](../zh/cookbook.md)

The [Vendor Integration Guide](./vendor-guide.md) is a **reference card** — it
tells you which attributes ToughRADIUS sends to / parses for a given vendor.
This cookbook goes one step further: it is organized around **real operational
scenarios** and translates a business need, end to end, into "server config +
device config + verification + troubleshooting".

## The five-part shape of every scenario

So you can follow along and debug effectively, every scenario uses the same
structure:

1. **Need / scenario** — the problem in business language, no protocol detail.
2. **On the ToughRADIUS side** — exactly what to configure in the admin UI, and
   which attributes are **actually emitted** after a successful auth, produced
   by which piece of code.
3. **On the device side** — reference configuration for the NAS/router.
4. **Verification** — how to confirm it really works (radtest, device commands,
   admin UI).
5. **Troubleshooting** — the most common traps for that scenario, as
   "symptom → locate → fix".

## Reading conventions

- **Every ToughRADIUS-side claim is anchored to code**: emitted attributes come
  from the enhancers in `internal/radiusd/plugins/auth/enhancers/`; the
  accept/reject decisions come from the checkers in
  `internal/radiusd/plugins/auth/checkers/`. This describes the system's **real
  behaviour**, not an aspiration.
- **Device-side config is always a reference example**: command syntax varies by
  model and OS version — defer to the vendor docs and your actual firmware.
- **CoA / Disconnect port is 3799** (RFC 5176). The `1700` you often see online
  is a client-side local port, not the destination port this system uses.
- Rates are stored in **Kbps** in the rate profile (the UI labels the unit).
  See [Vendor Integration Guide · Rate units](./vendor-guide.md#rate-units) for
  the per-vendor conversion.

## Available cookbooks

- [MikroTik RouterOS](./cookbook-mikrotik.md) — PPPoE broadband ISP speed tiers,
  Hotspot + MAC authentication, CoA / forced disconnect and FUP.

> **Planned (roadmap M13.8 later batches)**: Huawei, H3C, ZTE, iKuai (each has a
> dedicated vendor enhancer) and Cisco / standard-attribute scenarios. Until
> those chapters land, use the attribute reference in the
> [Vendor Integration Guide](./vendor-guide.md).

## Related chapters

- [Quick Start](./quickstart.md) — install, first login, `radtest` verification.
- [Vendor Integration Guide](./vendor-guide.md) — per-vendor attribute reference.
- [Admin UI Manual](./admin-manual.md) — users, rate profiles, online sessions,
  CoA.
- [FAQ](./faq.md) — cross-scenario troubleshooting Q&A.
