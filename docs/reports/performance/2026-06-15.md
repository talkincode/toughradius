# Performance Benchmark Report - 2026-06-15

## English

**Verdict:** RECORDED

This report is informational. GitHub hosted runners can vary, so this workflow records benchmark visibility and trend signals without failing on timing changes.

### Run Context

| Field | Value |
| --- | --- |
| Started | 2026-06-15T09:36:23Z |
| Finished | 2026-06-15T09:37:04Z |
| Commit | 66bda4d0cf58 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/27537313130) |
| Runner OS | Linux |
| Go | go version go1.25.11 linux/amd64 |
| GOOS/GOARCH | linux/amd64 |
| CPU | AMD EPYC 9V74 80-Core Processor |

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
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 360 | 3307955 | 1484354 | 15437 | +26.9% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 186146 | 6337 | 1465 | 7.00 | +29.5% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 189478 | 6180 | 1360 | 24.00 | +28.6% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 232948 | 5414 | 3502 | 52.00 | +28.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 231930 | 5042 | 4065 | 60.00 | +27.3% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 728718 | 1432 | 424.00 | 4.00 | +27.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1535384 | 904.10 | 736.00 | 25.00 | +41.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2691984 | 439.00 | 128.00 | 1.00 | +28.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3369394 | 359.10 | 23.00 | 1.00 | +30.3% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 10317824 | 119.40 | 23.00 | 2.00 | +18.5% |

### Highest Allocation Benchmarks

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 360 | 3307955 | 1484354 | 15437 | +26.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 231930 | 5042 | 4065 | 60.00 | +27.3% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 232948 | 5414 | 3502 | 52.00 | +28.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 186146 | 6337 | 1465 | 7.00 | +29.5% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 189478 | 6180 | 1360 | 24.00 | +28.6% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1535384 | 904.10 | 736.00 | 25.00 | +41.9% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 728718 | 1432 | 424.00 | 4.00 | +27.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2691984 | 439.00 | 128.00 | 1.00 | +28.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 17156950 | 68.84 | 96.00 | 1.00 | +14.5% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10317580 | 117.30 | 32.00 | 1.00 | +26.7% |

### All Benchmark Results

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 232948 | 5414 | 3502 | 52.00 | +28.1% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Get | 16783470 | 71.35 | 0.00 | 0.00 | +29.0% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10317580 | 117.30 | 32.00 | 1.00 | +26.7% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_ConcurrentGet | 24856970 | 45.32 | 0.00 | 0.00 | +29.6% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 186146 | 6337 | 1465 | 7.00 | +29.5% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 189478 | 6180 | 1360 | 24.00 | +28.6% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 728718 | 1432 | 424.00 | 4.00 | +27.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 10317824 | 119.40 | 23.00 | 2.00 | +18.5% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3369394 | 359.10 | 23.00 | 1.00 | +30.3% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2691984 | 439.00 | 128.00 | 1.00 | +28.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1535384 | 904.10 | 736.00 | 25.00 | +41.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkPacketLength | 286870846 | 4.18 | 0.00 | 0.00 | +29.3% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 231930 | 5042 | 4065 | 60.00 | +27.3% |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 17156950 | 68.84 | 96.00 | 1.00 | +14.5% |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 360 | 3307955 | 1484354 | 15437 | +26.9% |

## 中文

**结论：** 已记录

本报告只提供性能可见性。GitHub 托管 runner 存在波动，因此当前工作流记录趋势信号，但不会因为耗时变化直接失败。

### 运行上下文

| Field | Value |
| --- | --- |
| Started | 2026-06-15T09:36:23Z |
| Finished | 2026-06-15T09:37:04Z |
| Commit | 66bda4d0cf58 |
| Ref | main |
| Workflow | [workflow run](https://github.com/talkincode/toughradius/actions/runs/27537313130) |
| Runner OS | Linux |
| Go | go version go1.25.11 linux/amd64 |
| GOOS/GOARCH | linux/amd64 |
| CPU | AMD EPYC 9V74 80-Core Processor |

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
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 360 | 3307955 | 1484354 | 15437 | +26.9% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 186146 | 6337 | 1465 | 7.00 | +29.5% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 189478 | 6180 | 1360 | 24.00 | +28.6% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 232948 | 5414 | 3502 | 52.00 | +28.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 231930 | 5042 | 4065 | 60.00 | +27.3% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 728718 | 1432 | 424.00 | 4.00 | +27.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1535384 | 904.10 | 736.00 | 25.00 | +41.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2691984 | 439.00 | 128.00 | 1.00 | +28.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Global | 3369394 | 359.10 | 23.00 | 1.00 | +30.3% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkAuthRateLimiter_Sharded | 10317824 | 119.40 | 23.00 | 2.00 | +18.5% |

### 最高内存分配 Benchmark

| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| github.com/talkincode/toughradius/v9/pkg/excel | BenchmarkWriteToFile | 360 | 3307955 | 1484354 | 15437 | +26.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkServeRADIUSAccessAccept | 231930 | 5042 | 4065 | 60.00 | +27.3% |
| github.com/talkincode/toughradius/v9/internal/adminapi | BenchmarkIssueToken | 232948 | 5414 | 3502 | 52.00 | +28.1% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_MarshalJSON | 186146 | 6337 | 1465 | 7.00 | +29.5% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkRadiusUser_UnmarshalJSON | 189478 | 6180 | 1360 | 24.00 | +28.6% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkParseTcpPacket | 1535384 | 904.10 | 736.00 | 25.00 | +41.9% |
| github.com/talkincode/toughradius/v9/internal/domain | BenchmarkSysOprLog_MarshalJSON | 728718 | 1432 | 424.00 | 4.00 | +27.1% |
| github.com/talkincode/toughradius/v9/internal/radiusd | BenchmarkCheckRequestSecret | 2691984 | 439.00 | 128.00 | 1.00 | +28.9% |
| github.com/talkincode/toughradius/v9/internal/radiusd/plugins/eap/statemanager | BenchmarkMemoryStateManager_GetStateParallel | 17156950 | 68.84 | 96.00 | 1.00 | +14.5% |
| github.com/talkincode/toughradius/v9/internal/app | BenchmarkProfileCache_Set | 10317580 | 117.30 | 32.00 | 1.00 | +26.7% |
