# CLIENTS KNOWLEDGE BASE

## OVERVIEW
`internal/clients` 放第三方服务适配器。当前目录下有 `paymentgateway`、`sms`、`email`，用于隔离外部协议细节。

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| 支付渠道适配 | `paymentgateway/client.go` | 外部支付能力边界 |
| 短信适配 | `sms/client.go` | 通知渠道边界 |
| 邮件适配 | `email/client.go` | 通知渠道边界 |

## CONVENTIONS
- 业务层依赖接口，不直接依赖第三方 SDK 或协议细节。
- 外部请求要有 timeout、错误映射、可观测字段和必要的脱敏。
- client 只封装第三方交互，不承载订单/支付/库存/通知业务决策。

## ANTI-PATTERNS
- 不要让 handler/service 直接拼第三方请求协议。
- 不要把渠道错误原样透传给上层；映射为仓库稳定错误码体系。
- 不要在 client 中写业务状态流转或跨模块编排。
