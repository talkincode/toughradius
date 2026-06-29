# 性能基准测试报告

每周 benchmark 任务基于现有 Go `Benchmark*` 函数记录 ToughRADIUS 性能信号。报告仅用于观察趋势，不会因为 GitHub 托管 runner 的耗时波动直接失败。

**最近结论：** 已记录

## 最近摘要

| 指标 | 值 |
| --- | ---: |
| Benchmark 数量 | 15 |
| 包数量 | 6 |
| 最慢项 | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |
| 最高 B/op | github.com/talkincode/toughradius/v9/pkg/excel / BenchmarkWriteToFile |

## 保留报告

- [2026-06-29](https://github.com/talkincode/toughradius/blob/main/docs/reports/performance/2026-06-29.md)
- [2026-06-22](https://github.com/talkincode/toughradius/blob/main/docs/reports/performance/2026-06-22.md)
- [2026-06-15](https://github.com/talkincode/toughradius/blob/main/docs/reports/performance/2026-06-15.md)
