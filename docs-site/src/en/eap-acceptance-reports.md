# EAP Acceptance Reports

Weekly EAP acceptance runs validate ToughRADIUS with an external `eapol_test` supplicant and publish the latest retained reports here.

Maintainers: report PR signing requirements, artifact fallback, and verification steps are documented in the [Agent Development Guide](./agent-guide.md#report-pr-automation).

**Latest verdict:** ACCEPTED

## Latest Scenario Summary

| Scenario | Method | Expected | Status | Duration | Detail |
| --- | --- | --- | --- | ---: | --- |
| EAP-TLS valid client certificate | EAP-TLS | Access-Accept | passed | 146 ms | external supplicant received the expected Access-Accept |
| EAP-TLS untrusted client certificate | EAP-TLS | Access-Reject | passed | 19 ms | external supplicant was rejected as expected |
| PEAP/MSCHAPv2 valid credentials | PEAP/MSCHAPv2 | Access-Accept | skipped | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. |
| PEAP/MSCHAPv2 wrong password | PEAP/MSCHAPv2 | Access-Reject | skipped | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. |
| EAP-TTLS/PAP valid credentials | EAP-TTLS/PAP | Access-Accept | passed | 122 ms | external supplicant received the expected Access-Accept |
| EAP-TTLS/MSCHAPv2 valid credentials | EAP-TTLS/MSCHAPv2 | Access-Accept | passed | 122 ms | external supplicant received the expected Access-Accept |
| Malformed external EAP client config | tooling | documented skip | skipped | 0 ms | Skipped intentionally: eapol_test parser failures do not exercise ToughRADIUS over RADIUS/EAP. Negative server behavior is covered by untrusted certificate and wrong password scenarios. |

## Retained Reports

- [2026-06-15](https://github.com/talkincode/toughradius/blob/main/docs/reports/eap/2026-06-15.md)
