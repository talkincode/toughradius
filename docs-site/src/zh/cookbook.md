# 场景实战手册

> English version: [Scenario Cookbook](../en/cookbook.md)

[厂商对接指南](./vendor-guide.md)是一张**属性参考卡**——它告诉你某个厂商
ToughRADIUS 会下发/解析哪些属性。本手册更进一步：以**真实运维场景**为单位，
端到端地把「业务诉求」翻译成「服务端配置 + 设备侧配置 + 验证 + 排障」。

## 每个场景的五段式

为了便于照着做、也便于排错，本手册的每个场景都用同一结构：

1. **需求 / 场景** —— 用业务语言描述要解决的问题，不谈协议细节。
2. **ToughRADIUS 侧** —— 在管理后台具体配置什么；以及认证通过后**实际下发**
   哪些属性、由哪段代码产生。
3. **设备侧** —— NAS/路由器侧的参考配置。
4. **验证** —— 如何确认它真的生效（radtest、设备命令、管理后台）。
5. **排障** —— 以「症状 → 定位 → 解决」列出该场景最常见的坑。

## 阅读约定

- **ToughRADIUS 侧的每条声明都锚定代码**：下发的属性来自
  `internal/radiusd/plugins/auth/enhancers/` 的增强器，准入/拒绝判定来自
  `internal/radiusd/plugins/auth/checkers/` 的检查器。它描述的是系统**真实
  行为**，而非美好设想。
- **设备侧配置均为参考示例**：命令语法随型号与系统版本而异，请以厂商文档与
  实际固件为准。
- **CoA / Disconnect 端口为 3799**（RFC 5176）。网络上常见的 `1700` 是某些
  客户端的本地端口，不是本系统使用的目标端口。
- 速率在计费策略中以 **Kbps** 存储（界面标注单位）。换算规则见
  [厂商对接指南 · 限速单位](./vendor-guide.md#限速单位)。

## 现有实战手册

- [MikroTik RouterOS](./cookbook-mikrotik.md) —— PPPoE 宽带 ISP 分级套餐、
  Hotspot + MAC 认证、CoA / 强制下线与 FUP。
- [华为 BRAS / NetEngine](./cookbook-huawei.md) —— 带峰值速率与 AAA 域的宽带
  分级套餐、线路防盗用（MAC + VLAN 绑定）与双栈 IPv6、CoA / 强制下线与 FUP。

> **规划中（路线图 M13.8 后续批次）**：H3C、中兴、爱快（均有对应的
> 厂商增强器）以及 Cisco / 标准属性场景。在这些专章就绪前，请先参考
> [厂商对接指南](./vendor-guide.md)的属性参考。

## 相关章节

- [快速开始](./quickstart.md) —— 安装、首登、`radtest` 验证。
- [厂商对接指南](./vendor-guide.md) —— 各厂商属性与解析参考卡。
- [管理系统用户手册](./admin-manual.md) —— 用户、计费策略、在线会话、CoA。
- [常见问题解答](./faq.md) —— 跨场景的排障问答。
