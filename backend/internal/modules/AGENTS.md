# MODULES KNOWLEDGE BASE

## OVERVIEW
`internal/modules` 承载业务域代码；当前有 `order`、`payment`、`inventory`、`notification` 四个同构模块。

## STRUCTURE
```text
internal/modules/
├── order/
├── payment/
├── inventory/
└── notification/
```

每个模块当前都基本包含：
- `handler.go`
- `service.go`
- `repository.go`
- `dto.go`
- `model.go`
- `status.go`
- `events.go`

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| 看 HTTP 到业务的入口 | `*/handler.go` | `RegisterRoutes` + 具体 handler 方法 |
| 看业务规则 | `*/service.go` | 当前是业务骨架；新增真实流程时状态、幂等、事务边界应收敛到这里 |
| 看数据访问边界 | `*/repository.go` | 只做读写，不做业务决策 |
| 看输入输出模型 | `*/dto.go` | 请求/响应对象 |
| 看状态机 | `*/status.go` | 枚举和状态常量 |
| 看事件定义 | `*/events.go` | 域事件命名与 payload |

## CONVENTIONS
- 模块内部依赖顺序：`handler -> service -> repository`。
- Handler 接收 `context.Context` 的下游调用，通过 `c.Request.Context()` 传递。
- `RegisterRoutes` 只向父级 `/api/v1` group 挂本模块路由。
- `Ping` 这种健康样例可以存在，但新增业务时不要把示例式逻辑扩散成真实流程设计。
- DTO、Model、Status、Events 分文件放，避免把域内概念塞进一个大文件。
- 跨域协作必须走 Service 接口或事件；不要直接越过边界读对方 Repository / 表。

## ANTI-PATTERNS
- 不要在 Handler 拼 SQL、开事务、调用第三方 SDK。
- 不要在 Repository 做状态机、权限判断、幂等决策。
- 不要跨模块直接访问其他模块表或内部 struct。
- 不要为四个域复制粘贴不必要的“微差异”代码；若骨架一致，先考虑抽象共性规范。
- 不要把 trace / request metadata 丢掉；需要透传时从 `observability` 读取并写回响应。

## TESTING
- 单元测试优先与源码同目录放置：`*_test.go`。
- 集成测试放 `backend/test/integration/`，通过 bootstrap 装配完整路由链路。
- 新增业务域行为时，至少补 service 或 handler 层测试；不要只改样例 ping 而无验证。

## NOTES
- 当前四个模块高度同构，因此这里用一份共享规范覆盖全部模块；暂不为 `order/payment/inventory/notification` 单独建 AGENTS.md。
