# P1 Backend Foundation - Issues

## Task 12
- 初版失败点：测试错误假设 InMemoryWorker.Shutdown 应返回 context.Canceled，且入口误用了 logger.Warn；已统一为 no-op shutdown 契约与 Info 级 degrade 日志。
## task-19 lint issues
- 复现阶段证据文件已生成：`.sisyphus/evidence/task-19-lint-failure.txt`。
- 成功阶段证据文件已生成：`.sisyphus/evidence/task-19-lint-happy.txt`。
- 失败阶段包含 6 个问题：`errcheck` 4 个、`govet` 2 个；修复后 lint 清零。
