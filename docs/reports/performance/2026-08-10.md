# Performance Benchmark Report - 2026-08-10

## English

**Verdict:** RECORDED

This report is informational. GitHub hosted runners can vary, so this workflow records benchmark visibility and trend signals without failing on timing changes.

The previous report used a different runner context, so `Delta ns/op` is marked `n/a (context changed)` instead of presenting cross-runner timing as a direct trend.

### Run Context

| Field | Value |
| --- | --- |
| Started | 2026-08-10T11:31:58Z |
| Finished | 2026-08-10T11:32:37Z |
| Commit | 2bf67750ac35 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/31383895276) |
| Runner OS | Linux |
| Go | go version go1.25.12 linux/amd64 |
| GOOS/GOARCH | linux/amd64 |
| CPU | AMD EPYC 7763 64-Core Processor |

### Summary

| Metric | Value |
| --- | ---: |
| Benchmarks | 15 |
| Packages | 6 |
| Slowest | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |
| Highest B/op | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |

### Slowest Benchmarks

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 338 | 3592567 | 1515132 | 15437 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 180847 | 6663 | 1465 | 7.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 202242 | 5795 | 1360 | 24.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 196759 | 5586 | 3502 | 52.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 224361 | 5233 | 4065 | 60.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 669920 | 1611 | 424.00 | 4.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1369281 | 884.50 | 736.00 | 25.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2867533 | 416.20 | 128.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3465751 | 325.50 | 23.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 9323724 | 151.50 | 23.00 | 2.00 | n/a (context changed) |

### Highest Allocation Benchmarks

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 338 | 3592567 | 1515132 | 15437 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 224361 | 5233 | 4065 | 60.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 196759 | 5586 | 3502 | 52.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 180847 | 6663 | 1465 | 7.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 202242 | 5795 | 1360 | 24.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1369281 | 884.50 | 736.00 | 25.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 669920 | 1611 | 424.00 | 4.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2867533 | 416.20 | 128.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 15861805 | 73.90 | 96.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10959326 | 109.80 | 32.00 | 1.00 | n/a (context changed) |

### All Benchmark Results

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 196759 | 5586 | 3502 | 52.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Get | 18286860 | 65.45 | 0.00 | 0.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10959326 | 109.80 | 32.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_ConcurrentGet | 27905439 | 41.57 | 0.00 | 0.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 180847 | 6663 | 1465 | 7.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 202242 | 5795 | 1360 | 24.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 669920 | 1611 | 424.00 | 4.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 9323724 | 151.50 | 23.00 | 2.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3465751 | 325.50 | 23.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2867533 | 416.20 | 128.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1369281 | 884.50 | 736.00 | 25.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkPacketLength | 300668373 | 4.01 | 0.00 | 0.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 224361 | 5233 | 4065 | 60.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 15861805 | 73.90 | 96.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 338 | 3592567 | 1515132 | 15437 | n/a (context changed) |

## 中文

**结论：** 已记录

本报告只提供性能可见性。GitHub 托管 runner 存在波动，因此当前工作流记录趋势信号，但不会因为耗时变化直接失败。

上一份报告的 runner 上下文不同，因此 `Delta ns/op` 标记为 `n/a (context changed)`，避免把跨 runner 耗时差异呈现为直接趋势。

### 运行上下文

| Field | Value |
| --- | --- |
| Started | 2026-08-10T11:31:58Z |
| Finished | 2026-08-10T11:32:37Z |
| Commit | 2bf67750ac35 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/31383895276) |
| Runner OS | Linux |
| Go | go version go1.25.12 linux/amd64 |
| GOOS/GOARCH | linux/amd64 |
| CPU | AMD EPYC 7763 64-Core Processor |

### 摘要

| 指标 | 值 |
| --- | ---: |
| Benchmark 数量 | 15 |
| 包数量 | 6 |
| 最慢项 | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |
| 最高 B/op | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |

### 最慢 Benchmark

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 338 | 3592567 | 1515132 | 15437 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 180847 | 6663 | 1465 | 7.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 202242 | 5795 | 1360 | 24.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 196759 | 5586 | 3502 | 52.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 224361 | 5233 | 4065 | 60.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 669920 | 1611 | 424.00 | 4.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1369281 | 884.50 | 736.00 | 25.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2867533 | 416.20 | 128.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3465751 | 325.50 | 23.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 9323724 | 151.50 | 23.00 | 2.00 | n/a (context changed) |

### 最高内存分配 Benchmark

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 338 | 3592567 | 1515132 | 15437 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 224361 | 5233 | 4065 | 60.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 196759 | 5586 | 3502 | 52.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 180847 | 6663 | 1465 | 7.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 202242 | 5795 | 1360 | 24.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1369281 | 884.50 | 736.00 | 25.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 669920 | 1611 | 424.00 | 4.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2867533 | 416.20 | 128.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 15861805 | 73.90 | 96.00 | 1.00 | n/a (context changed) |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10959326 | 109.80 | 32.00 | 1.00 | n/a (context changed) |
