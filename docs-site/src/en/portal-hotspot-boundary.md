# Portal / Hotspot Integration Boundary

> 中文版本：[Portal / Hotspot 对接边界](../zh/portal-hotspot-boundary.md)

This chapter defines the hard product boundary for captive portal and hotspot
deployments.

## Iron rule

**ToughRADIUS does not provide, host, or operate a captive portal login page.**

Portal server, guest onboarding, voucher issuance, SMS/WeChat login, payment
flows, advertising pages, and captive network enforcement belong to a separate
portal / gateway product. They are not part of the ToughRADIUS product scope.

ToughRADIUS stays a RADIUS AAA system:

- authentication on UDP `1812`;
- accounting on UDP `1813`;
- session policy and audit through RADIUS attributes, accounting records, and
  optional CoA / Disconnect;
- vendor-specific RADIUS attributes where they can be expressed safely inside
  the existing vendor parser / enhancer model.

## Supported shape

The supported integration model is:

```text
Client -> NAS / WLAN controller / gateway portal -> RADIUS -> ToughRADIUS
```

The NAS, WLAN controller, hotspot gateway, or external portal product owns:

- HTTP/HTTPS captive-portal redirection;
- the login page and user interaction;
- pre-auth / post-auth network enforcement;
- vendor portal callbacks or proprietary portal protocols;
- device-side session admission and release.

ToughRADIUS owns:

- user, profile, NAS, and policy data;
- Access-Accept / Access-Reject decisions;
- accounting records and online-session state;
- standard and vendor RADIUS attributes such as timeout, pool, rate, VLAN,
  role, or portal URL where a supported vendor enhancer emits them;
- CoA / Disconnect where the NAS supports it.

## What this means for Hotspot

MikroTik Hotspot, Huawei/H3C/iKuai/Cisco WLAN controllers, Aruba captive portal
flows, and similar devices can still integrate with ToughRADIUS when the device
uses RADIUS as its backend. The device remains the portal implementation; this
project remains the AAA backend.

Common supported cases:

- Hotspot / PPPoE / WLAN controller sends Access-Request to ToughRADIUS.
- MAC authentication admits known devices without showing a portal page.
- Accounting updates keep online-session and traffic data current.
- CoA / Disconnect removes or refreshes a session when supported by the NAS.
- Vendor attributes may steer device-side behavior, but only as RADIUS
  attributes. They do not make ToughRADIUS a portal server.

## Explicit non-goals

Do not add these to ToughRADIUS:

- hosted login pages for guests or subscribers;
- voucher, coupon, QR-code, SMS, WeChat, OAuth, or payment onboarding flows;
- captive-portal JavaScript applications or customer self-service portals;
- proprietary portal-server callback protocols as first-class subsystems;
- per-vendor portal state machines that duplicate the NAS / controller role;
- generic campaign, advertisement, CRM, or visitor-management features.

If a deployment needs those functions, use a dedicated portal product in front
of the NAS / controller and integrate it with ToughRADIUS through RADIUS.

## Allowed narrow extensions

Portal-related work is acceptable only when it stays inside the existing RADIUS
boundary:

1. Add or fix vendor dictionaries, parsers, or enhancers for RADIUS attributes
   such as captive-portal URL, user role, filter ID, VLAN, session timeout, or
   rate limit.
2. Document a specific NAS / controller configuration that uses ToughRADIUS as
   the RADIUS backend.
3. Add tests for request parsing, response attributes, accounting, or CoA
   behavior.

Any request to build a hosted portal must be rejected or moved to another
product before implementation starts.
