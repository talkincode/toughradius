# Portal / Hotspot 对接边界

> English version: [Portal / Hotspot Integration Boundary](../en/portal-hotspot-boundary.md)

本章定义 Captive Portal / Hotspot 场景的产品边界。

## 铁律

**ToughRADIUS 不提供、不托管、不运营 Portal 登录页。**

Portal Server、访客开户注册、临时券码、短信/微信登录、支付流程、广告页、
准入放行控制等能力，属于独立的 Portal / 网关产品，不属于 ToughRADIUS 的产品
范围。

ToughRADIUS 只做 RADIUS AAA：

- UDP `1812` 上的认证；
- UDP `1813` 上的计费；
- 通过 RADIUS 属性、计费记录和可选 CoA / Disconnect 完成会话策略与审计；
- 在现有厂商 parser / enhancer 模型内，安全下发可表达为 RADIUS 属性的厂商能力。

## 支持的形态

支持的对接模型是：

```text
客户端 -> NAS / WLAN 控制器 / 网关 Portal -> RADIUS -> ToughRADIUS
```

NAS、WLAN 控制器、Hotspot 网关或外部 Portal 产品负责：

- HTTP/HTTPS Portal 跳转；
- 登录页与用户交互；
- 认证前 / 认证后的网络放行控制；
- 厂商 Portal 回调或私有 Portal 协议；
- 设备侧会话准入与释放。

ToughRADIUS 负责：

- 用户、Profile、NAS 与策略数据；
- Access-Accept / Access-Reject 决策；
- 计费记录和在线会话状态；
- 标准与厂商 RADIUS 属性，例如超时、地址池、速率、VLAN、角色，或已支持厂商
  enhancer 下发的 Portal URL；
- NAS 支持时的 CoA / Disconnect。

## 对 Hotspot 的含义

MikroTik Hotspot、华为 / H3C / 爱快 / Cisco WLAN 控制器、Aruba Captive
Portal 等设备，仍然可以在设备自身负责 Portal 的前提下对接 ToughRADIUS。
设备是 Portal 实现方；本项目是 AAA 后端。

常见支持场景：

- Hotspot / PPPoE / WLAN 控制器向 ToughRADIUS 发送 Access-Request。
- 已登记设备通过 MAC 认证跳过 Portal 页面。
- Accounting update 维护在线会话与流量数据。
- NAS 支持时，通过 CoA / Disconnect 刷新或踢下会话。
- 厂商属性可以引导设备侧行为，但只能作为 RADIUS 属性下发；这不会把
  ToughRADIUS 变成 Portal Server。

## 明确不做

不要把以下能力加入 ToughRADIUS：

- 面向访客或用户的托管登录页；
- 券码、二维码、短信、微信、OAuth 或支付开户注册流程；
- Captive Portal 前端应用或客户自助门户；
- 厂商私有 Portal Server 回调协议的一等子系统；
- 复制 NAS / 控制器职责的厂商 Portal 状态机；
- 通用营销、广告、CRM 或访客管理功能。

如果部署确实需要这些能力，应使用独立 Portal 产品，并通过 RADIUS 与
ToughRADIUS 对接。

## 允许的窄扩展

Portal 相关工作只有在留在 RADIUS 边界内时才允许：

1. 为 Captive Portal URL、用户角色、Filter-Id、VLAN、Session-Timeout、
   速率等 RADIUS 属性补充或修复厂商字典、parser、enhancer。
2. 编写特定 NAS / 控制器如何把 ToughRADIUS 作为 RADIUS 后端的配置文档。
3. 为请求解析、响应属性、计费或 CoA 行为补测试。

任何“让 ToughRADIUS 自带 Portal”的需求，在实现前都必须拒绝或转移到其他产品。
