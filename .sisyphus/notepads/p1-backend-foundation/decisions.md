# P1 Backend Foundation - Decisions

## 2026-05-14 Task 9
- 保留 tx.go 契约不变，在 tx_manager.go 新增 SQLTxManager/Beginner/sqlTx，避免影响 NoopTxManager 与既有测试。
- panic 路径先 Rollback 再 re-panic；普通错误路径 Rollback 后返回原 callback error。
