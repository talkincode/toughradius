# EAP Acceptance Reports

Weekly EAP acceptance runs validate ToughRADIUS with an external `eapol_test` supplicant and publish the latest retained reports here.

**Latest verdict:** PARTIAL

Coverage note: PEAP/MSCHAPv2 external `eapol_test` scenarios are still skipped and tracked by [#495](https://github.com/talkincode/toughradius/issues/495), so this report is partial external coverage rather than complete PEAP acceptance.

## Latest Scenario Summary

| Scenario | Method | Expected | Status | Duration | Detail |
| --- | --- | --- | --- | ---: | --- |
| EAP-TLS valid client certificate | EAP-TLS | Access-Accept | passed | 183 ms | external supplicant received the expected Access-Accept |
| EAP-TLS untrusted client certificate | EAP-TLS | Access-Reject | passed | 31 ms | external supplicant was rejected as expected |
| PEAP/MSCHAPv2 valid credentials | PEAP/MSCHAPv2 | Access-Accept | skipped | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. Tracking issue: https://github.com/talkincode/toughradius/issues/495. |
| PEAP/MSCHAPv2 wrong password | PEAP/MSCHAPv2 | Access-Reject | skipped | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. Tracking issue: https://github.com/talkincode/toughradius/issues/495. |
| EAP-TTLS/PAP valid credentials | EAP-TTLS/PAP | Access-Accept | passed | 133 ms | external supplicant received the expected Access-Accept |
| EAP-TTLS/MSCHAPv2 valid credentials | EAP-TTLS/MSCHAPv2 | Access-Accept | passed | 133 ms | external supplicant received the expected Access-Accept |
| Malformed external EAP client config | tooling | documented skip | skipped | 0 ms | Skipped intentionally: eapol_test parser failures do not exercise ToughRADIUS over RADIUS/EAP. Negative server behavior is covered by untrusted certificate and wrong password scenarios. |

## Retained Reports

- [2026-06-29](https://github.com/talkincode/toughradius/blob/main/docs/reports/eap/2026-06-29.md)
- [2026-06-22](https://github.com/talkincode/toughradius/blob/main/docs/reports/eap/2026-06-22.md)
- [2026-06-15](https://github.com/talkincode/toughradius/blob/main/docs/reports/eap/2026-06-15.md)
