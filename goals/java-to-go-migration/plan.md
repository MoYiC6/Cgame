# Java → Go 后端逐步迁移计划

## 1. 迁移方案概述

### 1.1 核心策略

- **逐步迁移，按域切分**：15 个业务域分批迁移，每批 1-2 周
- **数据库共享过渡**：Go 与 Java 共用 PostgreSQL，Go 端新增表，Java 端旧表逐步弃用
- **接口 100% 兼容**：保持相同路由前缀 `/api/*`、请求/响应格式、错误码，前端零改动
- **双写→切流→下线**：Java 主写 → 双写校验 → Go 主写 → Java 只读 → Java 下线
- **单体部署**：保持单体模式，不引入微服务基础设施

### 1.2 迁移优先级

| 阶段 | 业务域 | 表数 | 复杂度 | 依赖 |
|------|--------|------|--------|------|
| Phase 1 | 访客分析 (12) | 3 | 低 | 无 |
| Phase 1 | 文件存储 (13) | 2 | 低 | 无 |
| Phase 1 | 系统配置 (14) | 5 | 低 | 无 |
| Phase 2 | 通知消息 (9) | 4 | 中 | 系统权限 |
| Phase 2 | 聊天客服 (10) | 2 | 中 | 系统权限 |
| Phase 3 | 用户中心 (2) | 13 | 中 | 系统权限 |
| Phase 3 | 选手生态 (3) | 13 | 中 | 系统权限+用户中心 |
| Phase 4 | 商品中心 (4) | 9 | 中 | 系统权限 |
| Phase 4 | 营销引擎 (7) | 7 | 中 | 用户+商品 |
| Phase 5 | 订单核心 (5) | 7 | **高** | 用户+选手+商品+支付 |
| Phase 5 | 支付收银 (6) | 5 | **高** | 订单+用户 |
| Phase 6 | 财务结算 (8) | 5 | **高** | 订单+支付+选手 |
| Phase 7 | 游戏互动 (11) | 6 | 中 | 用户+选手 |
| Phase 7 | 外部集成 (15) | 2 | 中 | 系统权限+用户 |
| Phase 8 | 系统权限 (1) | 12 | **高** | 被所有域依赖（最后迁移） |

---

## 2. 单域迁移标准流程

每个域按以下 6 步执行，形成可复用的迁移 checklist：

### Step 1: DDL & Migration
- 分析 Java 端 Entity 类，提取表结构
- 编写 goose migration 文件（`backend/migrations/`）
- 关键：字段类型兼容、索引一致、软删除 `deleted SMALLINT DEFAULT 0`、乐观锁 `version INT DEFAULT 0`
- 验证：`make migrate-up` 执行成功，表结构与 Java 端一致

### Step 2: Model & Repository
- 定义 Go struct（`internal/modules/<domain>/model.go`）
- 编写 sqlc 查询定义（`sql/queries/<domain>.sql`）
- 生成 sqlc 代码：`make sqlc-generate`
- 实现 Repository 接口（`internal/modules/<domain>/repository.go`）
- 验证：`go test ./internal/modules/<domain>/` 通过

### Step 3: Service & DTO
- 实现 Service 层（`internal/modules/<domain>/service.go`）
- 定义请求/响应 DTO（`internal/modules/<domain>/dto.go`）
- 移植 Java Service 的业务逻辑，保持相同方法签名和返回值
- 验证：单元测试覆盖核心业务逻辑

### Step 4: Handler & Routing
- 实现 HTTP Handler（`internal/modules/<domain>/handler.go`）
- 注册路由到 `internal/bootstrap/server.go`，保持与 Java 端相同的路径和方法
- 验证：`go run ./cmd/api` 启动成功，路由列表匹配

### Step 5: 双写 & 数据同步
- 配置双写：Java 端继续写入，Go 端通过数据库触发器或应用层同步
- 编写数据校验脚本（Python/Go），对比 Java 和 Go 的表数据（行数、checksum、关键字段）
- 验证：连续 24 小时数据一致性校验通过

### Step 6: 切流 & 下线
- 灰度放量：10% → 50% → 100% 流量切到 Go（通过负载均衡或网关）
- 接口对比测试：对同一组输入，Java 和 Go 返回结果一致
- 切换主写：Go 主写，Java 降级只读
- Java 下线：删除 Java 端该域相关代码和表

---

## 3. 各 Phase 详细计划

### Phase 1：独立域（第 1-2 周）

**目标**：跑通全流程，建立迁移节奏

#### 3.1.1 访客分析（visitor）
- 表：`visitor_sessions`, `visitor_page_views`, `visitor_daily_stats`
- 迁移要点：
  - 纯追加写入，无更新/删除，最简单
  - Java 端 `VisitorTrackingService.trackVisitor()` → Go 端 `handler.go` POST `/api/admin/visitor-stats/track`
  - 数据校验：行数对比 + 每日 PV/UV 统计值对比
- 验证：`go test ./internal/modules/visitor/` + 数据脚本

#### 3.1.2 文件存储（file）
- 表：`files`, `file_categories`
- 迁移要点：
  - 七牛云上传逻辑移植到 Go（`internal/clients/storage/`）
  - Java 端 `FileUploadService.uploadFile()` → Go 端同名接口
  - 双写：Java 写 DB 后触发 Go 同步，或直接切 Go 写（文件元数据）
- 验证：上传文件后 DB 记录一致 + 文件可访问

#### 3.1.3 系统配置（system）
- 表：`system_settings`, `partner_config`, `kook_tickets`, `faceid_config`, `realname_verify_log`
- 迁移要点：
  - KV 模式，`SystemSettingsService.getValue/setValue` → Go 端实现
  - 缓存策略保持一致（Redis + 本地缓存）
- 验证：读写配置值一致 + 缓存命中率一致

### Phase 2：通知 & 聊天（第 3-4 周）

#### 3.2.1 通知消息（notification）
- 表：`notification`, `user_notification`, `system_todo`, `subscribe_message_log`
- 迁移要点：
  - 站内信 + 系统通知 + 订阅消息
  - Java `NotificationService.createNotification()` → Go 端
  - WebSocket 推送需同步迁移（`ChatWebSocketService`）
- 验证：通知创建/读取/推送逻辑一致

#### 3.2.2 聊天客服（chat）
- 表：`chat_session`, `chat_message`
- 迁移要点：
  - WebSocket 连接管理
  - 消息历史存储
- 验证：WebSocket 连接正常 + 消息收发一致

### Phase 3：用户 & 选手（第 5-8 周）

#### 3.3.1 用户中心（user）
- 表：`user_balance_log`, `user_login_log`, `user_notification`, `user_subscription`, `user_level`, `user_level_log`, `user_purchase_record`, `user_coupon`, `user_recharge_record`, `invite_record`, `feedback`, `feedback_reply`
- 迁移要点：
  - 余额变更（`UserBalanceService.increaseBalance/decreaseBalance`）需事务保证
  - 充值记录与支付回调关联
  - 用户等级自动计算（定时任务）
- 验证：余额变动一致性 + 充值流程端到端测试

#### 3.3.2 选手生态（teacher）
- 表：`teacher`, `teacher_applications`, `teacher_level`, `teacher_level_goods`, `teacher_level_upgrade_history`, `teacher_status_log`, `teacher_balance_log`, `teacher_dynamics`, `teacher_assessment_video`, `teacher_invite_code`, `teacher_map_permission`, `teacher_partner`, `impression_tag`
- 迁移要点：
  - 选手申请审核流程
  - 等级自动升级逻辑
  - 余额与订单结算关联
- 验证：选手接单→完成→结算全流程一致

### Phase 4：商品 & 营销（第 9-12 周）

#### 3.4.1 商品中心（inventory/goods）
- 表：`goods`, `goods_sku`, `goods_spec`, `goods_spec_value`, `goods_sku_spec_value`, `goods_category`, `goods_map`, `goods_sku_stock_log`, `purchase_limit_rule`, `user_purchase_record`
- 迁移要点：
  - SKU 体系（spec → spec_value → sku → sku_spec_value）
  - 库存扣减/回滚（`GoodsStockService.decreaseStock/increaseStock`）
  - 限购规则（`PurchaseLimitService`）
- 验证：库存扣减一致性 + SKU 查询结果一致

#### 3.4.2 营销引擎（marketing）
- 表：`coupon`, `coupon_grant_log`, `recharge_config`, `recharge_rebate_rule`, `recharge_rebate_log`, `purchase_limit_rule`
- 迁移要点：
  - 优惠券领取/使用/恢复
  - 充值返利规则匹配
- 验证：优惠计算一致性 + 返利到账正确

### Phase 5：订单 & 支付（第 13-18 周）

#### 3.5.1 订单核心（order）
- 表：`game_order`, `order_review`, `order_transfer_record`, `order_transfer_config`, `order_rejected_teacher`, `order_operation_log`, `manual_order_idempotency`
- 迁移要点：
  - 订单状态机（0-8 状态流转）
  - 转单/终审/老板确认/自动确认
  - 结算逻辑（`OrderSettlementService`）
  - 操作日志（`OrderLogService`）
- 验证：订单全生命周期状态流转一致 + 结算金额正确

#### 3.5.2 支付收银（payment）
- 表：`payment_record`, `refund_requests`, `refund_audits`, `wxpay_config`, `alipay_config`
- 迁移要点：
  - 微信支付 JSAPI/Native/H5 + 支付宝小程序/Wap/当面付
  - 支付回调通知处理
  - 退款状态机（`RefundStateMachineService`）
  - 支付同步（`PaymentSyncController`）
- 验证：支付/退款回调处理一致 + 状态机流转正确

### Phase 6：财务结算（第 19-22 周）

#### 3.6.1 财务结算（finance）
- 表：`operator_balance`, `operator_withdrawal`, `operator_commission_log`, `withdrawal_request`
- 迁移要点：
  - 运营分成结算（`OperatorCommissionSettlementService`）
  - 选手提现申请/审核/打款
  - 余额变动审计
- 验证：结算金额一致性 + 提现流程正确

### Phase 7：游戏 & 外部集成（第 23-26 周）

#### 3.7.1 游戏互动（game）
- 表：`game_room`, `game_room_player`, `game_map`, `game_move`, `game_record`, `bomb_ranking`
- 迁移要点：
  - WebSocket 游戏房间管理
  - 排行榜计算
- 验证：游戏房间创建/加入/开始流程一致

#### 3.7.2 外部集成（external）
- KOOK 机器人、微信 OAuth、扫码登录
- 迁移要点：
  - 微信 OAuth 流程移植
  - KOOK Bot API 适配
- 验证：OAuth 授权流程一致 + KOOK 消息接收正常

### Phase 8：系统权限 & 下线（第 27-28 周）

#### 3.8.1 系统权限（auth/system）
- 表：`sys_user`, `sys_role`, `sys_permission`, `sys_menu`, `sys_user_role`, `sys_role_permission`, `sys_role_menu`, `sys_user_token`, `sys_login_log`, `admin_log`, `error_log`, `sensitive_data_access_log`
- 迁移要点：
  - 最后迁移，因为被所有域依赖
  - 需确保所有其他域已迁移完成
  - 用户表需合并 Go 端 `users` 和 Java 端 `sys_user` 的字段
- 验证：登录/权限/菜单全流程一致

#### 3.8.2 Java 下线
- Java 端设置为只读（`spring.profiles.active=readonly`）
- 观察 1 周，无异常后停止 Java 服务
- 清理 Java 代码库

---

## 4. 验证体系

### 4.1 数据校验脚本

```bash
# 每个域迁移后运行
python scripts/validate_migration.py \
  --domain=visitor \
  --java-dsn="postgres://..." \
  --go-dsn="postgres://..." \
  --tables=visitor_sessions,visitor_page_views,visitor_daily_stats
```

脚本功能：
- 行数对比（Java vs Go）
- 关键字段 checksum 对比
- 缺失/多余数据检测
- 输出：通过/失败 + 差异明细

### 4.2 接口对比测试

```bash
# 对同一组输入，对比 Java 和 Go 的响应
python scripts/compare_responses.py \
  --java-base="http://localhost:8081" \
  --go-base="http://localhost:8082" \
  --endpoints=api/client/teachers,api/client/goods/1
```

### 4.3 灰度放量

- 阶段 1：10% 流量 → Go，90% → Java
- 阶段 2：50% 流量 → Go，50% → Java
- 阶段 3：100% 流量 → Go，Java 降级只读
- 监控指标：错误率、延迟、数据一致性

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 数据不一致 | 高 | 每 6 小时跑校验脚本，核心表加 triggers 审计 |
| 接口兼容性问题 | 中 | 迁移前录制 Java 端接口响应快照，逐条对比 |
| 性能回退 | 中 | 灰度放量，监控 P99 延迟 |
| 资金错误（支付/财务） | 极高 | 最后迁移，双写期间 Java 为主，Go 只读验证 |
| 迁移超期 | 中 | 每域 1-2 周固定周期，超期立即复盘 |

---

## 6. 下一步行动

1. **立即开始 Phase 1**：访客分析（visitor）
   - 创建 `internal/modules/visitor/` 目录结构
   - 编写 migration `00003_create_visitor_tables.sql`
   - 实现 model/repository/service/handler

2. **建立验证基础设施**：
   - 编写 `scripts/validate_migration.py` 数据校验脚本
   - 建立接口对比测试框架

3. **每周复盘**：每周末检查迁移进度，调整下周计划
