# P1 Backend Foundation - Decisions

## 2026-05-14 Task 9
- 保留 tx.go 契约不变，在 tx_manager.go 新增 SQLTxManager/Beginner/sqlTx，避免影响 NoopTxManager 与既有测试。
- panic 路径先 Rollback 再 re-panic；普通错误路径 Rollback 后返回原 callback error。
## task-19 lint config decision
- 保留 `govet`、`errcheck`、`staticcheck`、`unused`，并将 `gofmt` 迁移到 `formatters.enable`。
- 不再尝试单独启用 `gosimple`，因为当前 golangci-lint 版本已将其规则纳入 `staticcheck`。
- Final QA F4 recorded APPROVE because scope checks found no runtime CRUD endpoints, no business tables, and OpenAPI remained limited to health/readyz/module ping plus unified APIResponse.
