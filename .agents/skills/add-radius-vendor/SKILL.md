---
name: add-radius-vendor
description: 新增或扩展厂商 VSA 解析与响应增强 (TR-F005)。当需要让 ToughRADIUS 识别某厂商请求 VSA，或在 Access-Accept 中下发该厂商专有属性时使用。
---

# 技能：新增厂商 VSA 解析 / 响应增强

> 关联功能编号：`TR-F005`　适用里程碑：M5

## 何时使用
需要让 ToughRADIUS 识别某厂商的请求 VSA，或在 Access-Accept 中下发该厂商特有属性时。

## 前置检索（先读后写）
```text
grep_search "VendorCode" --include internal/radiusd/**
file_search "internal/radiusd/plugins/vendorparsers/parsers/*_parser.go"
file_search "internal/radiusd/plugins/auth/enhancers/*_enhancer.go"
view internal/radiusd/plugins/vendorparsers/parsers/init.go      # 注册位置
view internal/radiusd/vendors/<已有厂商>/                          # 字典与常量
```
重点理解：字典 ≠ 解析。字典只描述属性，必须有 parser 才会被提取。

## 实现步骤
1. **常量 / 字典**：在 `internal/radiusd/vendors/<vendor>/` 定义 VendorCode 与 VSA 常量（参考 huawei）。若 `vendors` 包缺少 `Code<Vendor>` 常量，先补充。
2. **Parser**：在 `internal/radiusd/plugins/vendorparsers/parsers/<vendor>_parser.go` 实现解析器，模仿 `huawei_parser.go` 的接口与字段提取。
3. **注册 Parser**：在 `parsers/init.go` 的 `init()` 中 `vendors.Register(&vendors.VendorInfo{Code: vendors.Code<Vendor>, ...Parser: &<Vendor>Parser{}})`。
4. **Enhancer（如需下发响应属性）**：在 `internal/radiusd/plugins/auth/enhancers/<vendor>_enhancer.go` 实现，模仿 `huawei_enhancer.go`（速率、VLAN 等下发）。
5. **样例包测试**：新增 `<vendor>_parser_test.go` / `<vendor>_enhancer_test.go`，用真实属性样例覆盖解析与下发。

## 约定
- Huawei 等厂商带宽单位差异、二进制 vs 十进制换算必须加内联注释说明 "why"。
- 速率/VLAN/MAC 绑定行为必须与已有厂商语义一致。

## 验收
- [ ] `go test ./internal/radiusd/...` 通过
- [ ] `golangci-lint run` 无新增问题
- [ ] 新厂商解析与响应均有样例测试覆盖
- [ ] PR 引用 `TR-F005` 与里程碑编号
