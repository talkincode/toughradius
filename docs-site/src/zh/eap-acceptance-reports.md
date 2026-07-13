# EAP 验收测试报告

每周 EAP 验收任务使用外部 `eapol_test` supplicant 验证 ToughRADIUS，并在这里展示最近保留的报告。

**最近结论：** 部分通过

覆盖说明：PEAP/MSCHAPv2 外部 `eapol_test` 场景仍为 skipped，并由 [#495](https://github.com/talkincode/toughradius/issues/495) 跟踪，因此本报告代表部分外部覆盖，不宣称完整 PEAP 外部验收。

## 最近场景摘要

| 场景 | 方法 | 预期 | 状态 | 耗时 | 说明 |
| --- | --- | --- | --- | ---: | --- |
| EAP-TLS valid client certificate | EAP-TLS | Access-Accept | 通过 | 188 ms | external supplicant received the expected Access-Accept |
| EAP-TLS untrusted client certificate | EAP-TLS | Access-Reject | 通过 | 33 ms | external supplicant was rejected as expected |
| PEAP/MSCHAPv2 valid credentials | PEAP/MSCHAPv2 | Access-Accept | 跳过 | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. Tracking issue: https://github.com/talkincode/toughradius/issues/495. |
| PEAP/MSCHAPv2 wrong password | PEAP/MSCHAPv2 | Access-Reject | 跳过 | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. Tracking issue: https://github.com/talkincode/toughradius/issues/495. |
| EAP-TTLS/PAP valid credentials | EAP-TTLS/PAP | Access-Accept | 通过 | 135 ms | external supplicant received the expected Access-Accept |
| EAP-TTLS/MSCHAPv2 valid credentials | EAP-TTLS/MSCHAPv2 | Access-Accept | 通过 | 137 ms | external supplicant received the expected Access-Accept |
| Malformed external EAP client config | tooling | documented skip | 跳过 | 0 ms | Skipped intentionally: eapol_test parser failures do not exercise ToughRADIUS over RADIUS/EAP. Negative server behavior is covered by untrusted certificate and wrong password scenarios. |

## 保留报告

- [2026-07-13](https://github.com/talkincode/toughradius/blob/main/docs/reports/eap/2026-07-13.md)
- [2026-07-06](https://github.com/talkincode/toughradius/blob/main/docs/reports/eap/2026-07-06.md)
- [2026-06-29](https://github.com/talkincode/toughradius/blob/main/docs/reports/eap/2026-06-29.md)
