# Performance Benchmark Report - 2026-06-15

## English

**Verdict:** RECORDED

This report is informational. GitHub hosted runners can vary, so this workflow records benchmark visibility and trend signals without failing on timing changes.

### Run Context

| Field | Value |
| --- | --- |
| Started | 2026-06-15T16:09:19Z |
| Finished | 2026-06-15T16:09:53Z |
| Commit | 48cae220d4b1 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/27559638220) |
| Runner OS | Linux |
| Go | go version go1.25.11 linux/amd64 |
| GOOS/GOARCH | linux/amd64 |
| CPU | INTEL(R) XEON(R) PLATINUM 8573C |

### Summary

| Metric | Value |
| --- | ---: |
| Benchmarks | 15 |
| Packages | 6 |
| Slowest | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |
| Highest B/op | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |

### Slowest Benchmarks

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 398 | 2994473 | 1543207 | 15438 | -9.5% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 239148 | 5124 | 1465 | 7.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 246180 | 4655 | 1360 | 24.00 | -24.7% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 268670 | 4549 | 3501 | 52.00 | -16.0% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 292104 | 4045 | 4065 | 60.00 | -19.8% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 886989 | 1178 | 424.00 | 4.00 | -17.7% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1657244 | 731.30 | 736.00 | 25.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 3382060 | 356.30 | 128.00 | 1.00 | -18.8% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3521974 | 346.00 | 23.00 | 1.00 | -3.6% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 9676465 | 125.80 | 23.00 | 2.00 | +5.4% |

### Highest Allocation Benchmarks

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 398 | 2994473 | 1543207 | 15438 | -9.5% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 292104 | 4045 | 4065 | 60.00 | -19.8% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 268670 | 4549 | 3501 | 52.00 | -16.0% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 239148 | 5124 | 1465 | 7.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 246180 | 4655 | 1360 | 24.00 | -24.7% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1657244 | 731.30 | 736.00 | 25.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 886989 | 1178 | 424.00 | 4.00 | -17.7% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 3382060 | 356.30 | 128.00 | 1.00 | -18.8% |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 14829061 | 81.41 | 96.00 | 1.00 | +18.3% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10958541 | 107.30 | 32.00 | 1.00 | -8.5% |

### All Benchmark Results

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 268670 | 4549 | 3501 | 52.00 | -16.0% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Get | 21686259 | 57.08 | 0.00 | 0.00 | -20.0% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10958541 | 107.30 | 32.00 | 1.00 | -8.5% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_ConcurrentGet | 18651390 | 63.60 | 0.00 | 0.00 | +40.3% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 239148 | 5124 | 1465 | 7.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 246180 | 4655 | 1360 | 24.00 | -24.7% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 886989 | 1178 | 424.00 | 4.00 | -17.7% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 9676465 | 125.80 | 23.00 | 2.00 | +5.4% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3521974 | 346.00 | 23.00 | 1.00 | -3.6% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 3382060 | 356.30 | 128.00 | 1.00 | -18.8% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1657244 | 731.30 | 736.00 | 25.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkPacketLength | 240783642 | 4.93 | 0.00 | 0.00 | +18.0% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 292104 | 4045 | 4065 | 60.00 | -19.8% |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 14829061 | 81.41 | 96.00 | 1.00 | +18.3% |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 398 | 2994473 | 1543207 | 15438 | -9.5% |

## 中文

**结论：** 已记录

本报告只提供性能可见性。GitHub 托管 runner 存在波动，因此当前工作流记录趋势信号，但不会因为耗时变化直接失败。

### 运行上下文

| Field | Value |
| --- | --- |
| Started | 2026-06-15T16:09:19Z |
| Finished | 2026-06-15T16:09:53Z |
| Commit | 48cae220d4b1 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/27559638220) |
| Runner OS | Linux |
| Go | go version go1.25.11 linux/amd64 |
| GOOS/GOARCH | linux/amd64 |
| CPU | INTEL(R) XEON(R) PLATINUM 8573C |

### 摘要

| 指标 | 值 |
| --- | ---: |
| Benchmark 数量 | 15 |
| 包数量 | 6 |
| 最慢项 | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |
| 最高 B/op | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |

### 最慢 Benchmark

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 398 | 2994473 | 1543207 | 15438 | -9.5% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 239148 | 5124 | 1465 | 7.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 246180 | 4655 | 1360 | 24.00 | -24.7% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 268670 | 4549 | 3501 | 52.00 | -16.0% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 292104 | 4045 | 4065 | 60.00 | -19.8% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 886989 | 1178 | 424.00 | 4.00 | -17.7% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1657244 | 731.30 | 736.00 | 25.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 3382060 | 356.30 | 128.00 | 1.00 | -18.8% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3521974 | 346.00 | 23.00 | 1.00 | -3.6% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 9676465 | 125.80 | 23.00 | 2.00 | +5.4% |

### 最高内存分配 Benchmark

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 398 | 2994473 | 1543207 | 15438 | -9.5% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 292104 | 4045 | 4065 | 60.00 | -19.8% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 268670 | 4549 | 3501 | 52.00 | -16.0% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 239148 | 5124 | 1465 | 7.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 246180 | 4655 | 1360 | 24.00 | -24.7% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1657244 | 731.30 | 736.00 | 25.00 | -19.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 886989 | 1178 | 424.00 | 4.00 | -17.7% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 3382060 | 356.30 | 128.00 | 1.00 | -18.8% |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 14829061 | 81.41 | 96.00 | 1.00 | +18.3% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10958541 | 107.30 | 32.00 | 1.00 | -8.5% |
