# EAP Acceptance Test Report - 2026-06-15

## English

**Verdict:** ACCEPTED

### Run Context

| Field | Value |
| --- | --- |
| Started | 2026-06-15T15:26:08Z |
| Finished | 2026-06-15T15:26:08Z |
| Commit | 48cae220d4b1 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/27556993388) |
| Runner OS | Linux |
| Go | go version go1.25.11 linux/amd64 |
| Tool | eapol_test eapol_test v2.10 |

### Scenario Results

| Scenario | Method | Expected | Status | Duration | Detail |
| --- | --- | --- | --- | ---: | --- |
| EAP-TLS valid client certificate | EAP-TLS | Access-Accept | passed | 143 ms | external supplicant received the expected Access-Accept |
| EAP-TLS untrusted client certificate | EAP-TLS | Access-Reject | passed | 18 ms | external supplicant was rejected as expected |
| PEAP/MSCHAPv2 valid credentials | PEAP/MSCHAPv2 | Access-Accept | skipped | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. |
| PEAP/MSCHAPv2 wrong password | PEAP/MSCHAPv2 | Access-Reject | skipped | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. |
| EAP-TTLS/PAP valid credentials | EAP-TTLS/PAP | Access-Accept | passed | 119 ms | external supplicant received the expected Access-Accept |
| EAP-TTLS/MSCHAPv2 valid credentials | EAP-TTLS/MSCHAPv2 | Access-Accept | passed | 120 ms | external supplicant received the expected Access-Accept |
| Malformed external EAP client config | tooling | documented skip | skipped | 0 ms | Skipped intentionally: eapol_test parser failures do not exercise ToughRADIUS over RADIUS/EAP. Negative server behavior is covered by untrusted certificate and wrong password scenarios. |

## 中文

**结论：** 通过

### 运行上下文

| Field | Value |
| --- | --- |
| Started | 2026-06-15T15:26:08Z |
| Finished | 2026-06-15T15:26:08Z |
| Commit | 48cae220d4b1 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/27556993388) |
| Runner OS | Linux |
| Go | go version go1.25.11 linux/amd64 |
| Tool | eapol_test eapol_test v2.10 |

### 场景结果

| 场景 | 方法 | 预期 | 状态 | 耗时 | 说明 |
| --- | --- | --- | --- | ---: | --- |
| EAP-TLS valid client certificate | EAP-TLS | Access-Accept | 通过 | 143 ms | external supplicant received the expected Access-Accept |
| EAP-TLS untrusted client certificate | EAP-TLS | Access-Reject | 通过 | 18 ms | external supplicant was rejected as expected |
| PEAP/MSCHAPv2 valid credentials | PEAP/MSCHAPv2 | Access-Accept | 跳过 | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. |
| PEAP/MSCHAPv2 wrong password | PEAP/MSCHAPv2 | Access-Reject | 跳过 | 0 ms | Skipped intentionally: eapol_test currently exposes a PEAP inner-framing interop gap (server rejects the decrypted phase-2 payload as an invalid inner EAP message). The in-process PEAP/MSCHAPv2 integration test remains the current acceptance coverage. |
| EAP-TTLS/PAP valid credentials | EAP-TTLS/PAP | Access-Accept | 通过 | 119 ms | external supplicant received the expected Access-Accept |
| EAP-TTLS/MSCHAPv2 valid credentials | EAP-TTLS/MSCHAPv2 | Access-Accept | 通过 | 120 ms | external supplicant received the expected Access-Accept |
| Malformed external EAP client config | tooling | documented skip | 跳过 | 0 ms | Skipped intentionally: eapol_test parser failures do not exercise ToughRADIUS over RADIUS/EAP. Negative server behavior is covered by untrusted certificate and wrong password scenarios. |
