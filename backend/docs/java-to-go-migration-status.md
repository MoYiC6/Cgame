# Java → Go 后端迁移状态

> 生成时间：2026-07-02
> 目标：保持 API 100% 兼容，共用 PostgreSQL，最终完全替换 Java 后端

## 一、模块迁移总览

| 序号 | 模块名 | 核心职责 | 迁移状态 |
|------|--------|----------|----------|
| 1 | auth | JWT 认证、登录/注册/登出、短信验证码、微信 OAuth | ✅ 已完成 |
| 2 | user | 用户中心、余额、充值、等级、消费排名 | ⚠️ 部分（基础信息/余额查询） |
| 3 | teacher | 选手生态、审核、等级、动态、收入分成、排名 | ⚠️ 部分（核心接口/申请审核/等级管理已完成，动态/评价/视频待迁移） |
| 4 | goods/inventory | 商品管理、SKU、分类、限购、Banner、印象标签 | ⚠️ 部分（核心 CRUD / SKU / 分类 / 限购已完成，Banner / 印象标签待迁移） |
| 5 | order | 订单生命周期、状态流转、结单、退单、转移、评价 | ✅ 已完成 |
| 6 | payment | 微信支付、支付宝、收银台、支付记录、回调 | ⚠️ 部分（基础 CRUD） |
| 7 | finance | 财务统计、运营商佣金、提现管理、结算 | ⚠️ 部分（仅统计接口） |
| 8 | notification | 系统通知、订阅消息、待办事项、实时推送 | ⚠️ 部分（基础通知/待办，管理端收件箱缺失） |
| 9 | chat | 即时聊天、会话管理、客服系统 | ✅ 已完成 |
| 10 | file | 文件上传、素材管理、七牛云存储 | ⚠️ 部分（素材管理/上传路径） |
| 11 | system | RBAC 权限、系统设置、菜单、角色、日志 | ⚠️ 部分（基础配置） |
| 12 | visitor | 访客追踪、页面浏览、统计报表 | ✅ 已完成 |
| 13 | game | 飞行棋房间、游戏地图、游戏订单、排行榜 | ✅ 已完成 |
| 14 | external | KOOK 机器人、微信集成、外部系统对接 | ⚠️ 部分（OAuth/绑定） |
| 15 | refund | 退款申请、审批、退款状态机 | ✅ 已完成 |
| 16 | coupon | 优惠券发放、领取、使用、统计 | ✅ 已完成 |
| 17 | recharge | 充值订单、返利规则、返利统计 | ✅ 已完成 |
| 18 | withdrawal | 选手提现申请、审批、打款、税务计算 | ✅ 已完成 |
| 19 | invite | 用户邀请、选手邀请码、邀请记录 | ✅ 已完成 |
| 20 | feedback | 用户反馈提交、回复、管理 | ✅ 已完成 |
| 21 | partner | 运营合作伙伴管理、分成配置 | ✅ 已完成 |
| 22 | customer_service | 客服配置、客服接入 | ✅ 已完成 |

## 二、功能迁移明细

### 1. auth（认证授权）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 用户登录（密码/短信/微信） | `POST /api/auth/login` | ✅ |
| 用户注册 | `POST /api/auth/register` | ✅ |
| 用户登出 | `POST /api/auth/logout` | ✅ |
| 刷新 Token | `POST /api/auth/refresh` | ✅ |
| 发送短信验证码 | `POST /api/auth/sms/send` | ⚠️ |
| 获取认证用户信息 | `GET /api/auth/info` | ✅ |
| 获取当前用户信息 | `GET /api/auth/me` | ✅ |
| 微信扫码登录-生成二维码 | `GET /api/wechat/scan-login/generate-qrcode` | ⚠️ |
| 微信扫码登录-检查状态 | `GET /api/wechat/scan-login/check-status` | ⚠️ |
| 微信扫码登录-扫码确认 | `POST /api/wechat/scan-login/scan` | ⚠️ |
| 微信扫码登录-确认登录 | `POST /api/wechat/scan-login/confirm` | ⚠️ |
| 微信扫码登录-取消 | `DELETE /api/wechat/scan-login/cancel` | ⚠️ |

### 2. user（用户中心）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 获取用户中心信息 | `GET /api/user/center` | ✅ |
| 更新用户信息 | `PUT /api/user/profile` | ✅ |
| 设置头像 | `PUT /api/user/avatar` | ✅ |
| 修改密码 | `PUT /api/user/password` | ⚠️ |
| 重置密码 | `POST /api/user/reset-password` | ❌ |
| 发送短信验证码 | `POST /api/user/sms/send` | ❌ |
| 更新提现账户 | `PUT /api/user/withdrawal-account` | ❌ |
| 更新手机号 | `PUT /api/user/mobile` | ❌ |
| 绑定微信手机号 | `PUT /api/user/wechat-mobile` | ❌ |
| 查询我的余额 | `GET /api/balance/my-balance` | ✅ |
| 查询用户余额（管理端） | `GET /api/balance/user-balance/{userId}` | ✅ |
| 查询我的余额日志 | `GET /api/balance/my-logs` | ✅ |
| 查询用户余额日志 | `GET /api/balance/user-logs/{userId}` | ✅ |
| 查询最近余额日志 | `GET /api/balance/recent-logs/{userId}` | ✅ |
| 手动充值 | `POST /api/recharge/manual` | ❌ |
| 创建充值订单 | `POST /api/recharge/create` | ❌ |
| 充值回调 | `POST /api/recharge/callback` | ❌ |
| 充值记录列表 | `GET /api/recharge/list` | ❌ |
| 我的充值记录 | `GET /api/recharge/my-records` | ❌ |
| 充值记录详情 | `GET /api/recharge/detail/{id}` | ❌ |
| 充值统计 | `GET /api/recharge/statistics` | ❌ |
| 最近充值记录 | `GET /api/recharge/recent/{userId}` | ❌ |
| 取消充值 | `POST /api/recharge/cancel/{rechargeNo}` | ❌ |
| 继续支付 | `POST /api/recharge/continue-pay/{rechargeNo}` | ❌ |
| 验证支付 | `POST /api/recharge/verify-payment/{rechargeNo}` | ❌ |
| 获取用户信息（客户端） | `GET /api/client/user/info` | ✅ |
| 更新用户信息（客户端） | `PUT /api/client/user/info` | ✅ |
| 完善资料 | `POST /api/client/user/complete-profile` | ✅ |
| 获取用户等级列表 | `GET /api/client/user-levels` | ✅ |
| 用户列表（管理端） | `GET /api/admin/users` | ✅ |
| 用户详情 | `GET /api/admin/users/{id}` | ✅ |
| 创建用户 | `POST /api/admin/users` | ❌ |
| 更新用户 | `PUT /api/admin/users/{id}` | ✅ |
| 删除用户 | `DELETE /api/admin/users/{id}` | ❌ |
| 更新用户状态 | `PUT /api/admin/users/{id}/status` | ✅ |
| 用户选择器 | `GET /api/admin/select/user` | ✅ |
| 用户登录日志列表 | `GET /api/admin/logs/user` | ✅ |
| 批量删除日志 | `DELETE /api/admin/logs/user/batch` | ✅ |
| 消费排名 | `GET /api/user/consumption-ranking` | ✅ |
| 用户月度财务报告 | `GET /api/admin/finance/user-monthly-report` | ❌ |
| 导出用户月度报告 | `GET /api/admin/finance/user-monthly-report/export` | ❌ |

### 3. teacher（选手生态）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 获取选手列表（公开） | `GET /api/client/teachers` | ✅ |
| 获取选手详情（公开） | `GET /api/client/teachers/{id}` | ✅ |
| 更新在线状态 | `PUT /api/client/teacher/online-status` | ✅ |
| 获取当前选手状态 | `GET /api/client/teacher/status` | ✅ |
| 获取我的选手状态 | `GET /api/client/teacher/my-status` | ✅ |
| 实名认证状态 | `GET /api/client/teacher/realname/status` | ❌ |
| 发起人脸验证 | `POST /api/client/teacher/realname/face/initiate` | ❌ |
| 验证人脸 | `POST /api/client/teacher/realname/face/verify` | ❌ |
| 获取收款信息 | `GET /api/client/teacher/payment-info` | ❌ |
| 更新收款信息 | `PUT /api/client/teacher/payment-info` | ❌ |
| 获取个人介绍 | `GET /api/client/teacher/intro` | ❌ |
| 更新个人介绍 | `PUT /api/client/teacher/intro` | ❌ |
| 选手心跳 | `POST /api/client/teacher/heartbeat` | ✅ |
| 设置自动状态 | `PUT /api/client/teacher/auto-status` | ✅ |
| 获取自动状态设置 | `GET /api/client/teacher/auto-status` | ✅ |
| 申请成为选手 | `POST /api/client/teacher/application` | ✅ |
| 获取选手动态 | `GET /api/client/teacher/dynamics/{teacherId}` | ❌ |
| 删除动态 | `DELETE /api/client/teacher/dynamics/{id}` | ❌ |
| 获取选手评价 | `GET /api/client/teacher/reviews/{teacherId}` | ❌ |
| 选手订单列表 | `GET /api/client/teacher/orders` | ❌ |
| 公开等级列表 | `GET /api/client/teacher/levels` | ✅ |
| 选手仪表盘统计 | `GET /api/teacher/dashboard/stats` | ✅ |
| 选手排名 | `GET /api/teacher/ranking` | ✅ |
| 选手列表（管理端） | `GET /api/admin/teachers` | ✅ |
| 更新选手状态 | `PUT /api/admin/teachers/{id}/status` | ✅ |
| 审核选手 | `POST /api/admin/teachers/{id}/verify` | ✅ |
| 批量更新状态 | `POST /api/admin/teachers/batch-status` | ✅ |
| 状态变更日志 | `GET /api/admin/teachers/{id}/status-log` | ✅ |
| 申请列表 | `GET /api/admin/teacher/applications` | ✅ |
| 审核通过 | `POST /api/admin/teacher/applications/{id}/approve` | ✅ |
| 审核拒绝 | `POST /api/admin/teacher/applications/{id}/reject` | ✅ |
| 手动升级 | `POST /api/admin/teacher/upgrade/manual/{teacherId}` | ❌ |
| 检查升级条件 | `POST /api/admin/teacher/upgrade/check/{teacherId}` | ❌ |
| 升级历史 | `GET /api/admin/teacher/upgrade/history` | ❌ |
| 等级列表（管理端） | `GET /api/admin/teacher/levels` | ✅ |
| 更新等级 | `PUT /api/admin/teacher/levels/{id}` | ✅ |
| 删除等级 | `DELETE /api/admin/teacher/levels/{id}` | ✅ |
| 等级关联商品 | `GET /api/admin/teacher/levels/{id}/goods` | ✅ |
| 更新等级商品 | `PUT /api/admin/teacher/levels/{id}/goods` | ✅ |
| 导出等级 | `GET /api/admin/teacher/levels/export` | ❌ |
| 导入等级 | `POST /api/admin/teacher/levels/import` | ❌ |
| 导出模板 | `GET /api/admin/teacher/levels/export/template` | ❌ |
| 选手收入明细 | `GET /api/admin/teacher-income` | ❌ |
| 动态列表（管理端） | `GET /api/admin/teacher-dynamics` | ❌ |
| 删除动态 | `DELETE /api/admin/teacher-dynamics/{id}` | ❌ |
| 批量删除动态 | `DELETE /api/admin/teacher-dynamics/batch` | ❌ |
| 更新考核视频 | `PUT /api/admin/teachers/{teacherId}/assessment-videos/{videoId}` | ❌ |
| 启用/禁用视频 | `PUT /api/admin/teachers/{teacherId}/assessment-videos/{videoId}/enabled` | ❌ |
| 删除考核视频 | `DELETE /api/admin/teachers/{teacherId}/assessment-videos/{videoId}` | ❌ |
| 合作伙伴列表 | `GET /api/admin/teacher/partners` | ❌ |
| 更新合作伙伴 | `PUT /api/admin/teacher/partners/{id}` | ❌ |
| 删除合作伙伴 | `DELETE /api/admin/teacher/partners/{id}` | ❌ |
| 选手合作记录 | `GET /api/admin/teacher/partners/teacher/{teacherId}` | ❌ |
| 已合作选手 | `GET /api/admin/teacher/partners/partnered-teachers` | ❌ |
| 邀请码列表 | `GET /api/admin/teacher/invite-code` | ❌ |
| 创建邀请码 | `POST /api/admin/teacher/invite-code` | ❌ |
| 更新邀请码 | `PUT /api/admin/teacher/invite-code/{id}` | ❌ |
| 删除邀请码 | `DELETE /api/admin/teacher/invite-code/{id}` | ❌ |

### 4. goods/inventory（商品中心）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 商品列表（客户端） | `GET /api/client/goods` | ✅ |
| 商品详情 | `GET /api/client/goods/{id}` | ✅ |
| 商品详情（含SKU） | `GET /api/client/goods/detail/{goodsId}` | ✅ |
| SKU 库存检查 | `POST /api/client/goods/sku/check` | ✅ |
| 分类列表 | `GET /api/client/categories` | ✅ |
| 分类详情 | `GET /api/client/categories/{id}` | ✅ |
| Banner 列表 | `GET /api/client/banners` | ❌ |
| 印象标签列表 | `GET /api/client/impression-tags` | ❌ |
| 商品列表（管理端） | `GET /api/admin/goods` | ✅ |
| 创建商品 | `POST /api/admin/goods` | ✅ |
| 更新商品 | `PUT /api/admin/goods/{id}` | ✅ |
| 删除商品 | `DELETE /api/admin/goods/{id}` | ✅ |
| 上下架 | `PUT /api/admin/goods/{id}/status` | ✅ |
| 商品统计 | `GET /api/admin/goods/stats` | ✅ |
| SKU 列表 | `GET /api/admin/goods/{id}/skus` | ✅ |
| 创建 SKU | `POST /api/admin/goods/sku` | ✅ |
| 更新 SKU | `PUT /api/admin/goods/sku/{id}` | ✅ |
| 删除 SKU | `DELETE /api/admin/goods/sku/{id}` | ✅ |
| 分类列表（管理端） | `GET /api/admin/categories` | ✅ |
| 全部分类 | `GET /api/admin/categories/all` | ✅ |
| 分类详情 | `GET /api/admin/categories/{id}` | ✅ |
| 更新分类 | `PUT /api/admin/categories/{id}` | ✅ |
| 删除分类 | `DELETE /api/admin/categories/{id}` | ✅ |
| Banner 列表（管理端） | `GET /api/admin/banners` | ❌ |
| 创建 Banner | `POST /api/admin/banners` | ❌ |
| 更新 Banner | `PUT /api/admin/banners/{id}` | ❌ |
| 删除 Banner | `DELETE /api/admin/banners/{id}` | ❌ |
| 标签列表 | `GET /api/admin/impression-tags` | ❌ |
| 创建标签 | `POST /api/admin/impression-tags` | ❌ |
| 更新标签 | `PUT /api/admin/impression-tags/{id}` | ❌ |
| 删除标签 | `DELETE /api/admin/impression-tags/{id}` | ❌ |
| 限购规则列表 | `GET /api/admin/purchase-limit` | ✅ |
| 创建限购规则 | `POST /api/admin/purchase-limit` | ✅ |
| 更新限购规则 | `PUT /api/admin/purchase-limit/{id}` | ✅ |
| 删除限购规则 | `DELETE /api/admin/purchase-limit/{id}` | ✅ |

### 5. order（订单核心）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 创建订单 | `POST /api/client/orders` | ✅ |
| 订单列表 | `GET /api/client/orders` | ✅ |
| 订单详情 | `GET /api/client/orders/{id}` | ✅ |
| 取消订单 | `POST /api/client/orders/{id}/cancel` | ✅ |
| 确认订单 | `POST /api/client/orders/{id}/confirm` | ✅ |
| 投诉订单 | `POST /api/client/orders/{id}/complaint` | ✅ |
| 确认选手 | `POST /api/client/orders/{id}/confirm-teacher` | ✅ |
| 订单统计 | `GET /api/client/orders/statistics` | ✅ |
| 评价列表 | `GET /api/client/reviews/orders` | ✅ |
| 订单评价详情 | `GET /api/client/reviews/orders/{orderId}` | ✅ |
| 订单列表（管理端） | `GET /api/admin/orders` | ✅ |
| 订单详情（管理端） | `GET /api/admin/orders/{id}` | ✅ |
| 更新订单状态 | `PUT /api/admin/orders/{id}/status` | ✅ |
| 退款 | `POST /api/admin/orders/{id}/refund` | ✅ |
| 手动结单 | `POST /api/admin/orders/{id}/manual-complete` | ✅ |
| 更新备注 | `PUT /api/admin/orders/{id}/remark` | ✅ |
| 更新关联选手 | `PUT /api/admin/orders/{id}/teachers` | ✅ |
| 手动下单 | `POST /api/admin/orders/manual` | ✅ |
| 订单统计（管理端） | `GET /api/admin/orders/stats` | ✅ |
| 评价列表（管理端） | `GET /api/admin/reviews` | ✅ |
| 更新评价状态 | `PUT /api/admin/reviews/{id}/status` | ✅ |
| 回复评价 | `POST /api/admin/reviews/{id}/reply` | ✅ |
| 最终审核列表 | `GET /api/admin/orders/final-review` | ✅ |
| 终审通过 | `POST /api/admin/orders/final-review/{id}/approve` | ✅ |
| 终审拒绝 | `POST /api/admin/orders/final-review/{id}/reject` | ✅ |
| 订单转移配置 | `GET /api/order-transfer` | ✅ |
| 执行订单转移 | `POST /api/order-transfer/transfer` | ✅ |

### 6. payment（支付收银）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 创建支付订单 | `POST /api/client/payments` | ⚠️ |
| 确认支付 | `POST /api/client/payments/confirm` | ⚠️ |
| 查询支付状态 | `GET /api/client/payments/status` | ⚠️ |
| 创建微信支付订单 | `POST /api/payments/wxpay/create` | ❌ |
| 微信支付回调 | `POST /api/payments/wxpay/notify` | ❌ |
| 查询微信支付 | `GET /api/payments/wxpay/query/{outTradeNo}` | ❌ |
| 创建支付宝订单 | `POST /api/client/alipay/create` | ❌ |
| 支付宝回调 | `POST /api/payments/alipay/notify` | ❌ |
| 查询支付宝订单 | `GET /api/client/alipay/query/{outTradeNo}` | ❌ |
| 查询收银台订单 | `GET /api/cashier/{token}` | ❌ |
| 发起收银台支付 | `POST /api/cashier/{token}/pay` | ❌ |
| 查询收银台状态 | `GET /api/cashier/{token}/status` | ❌ |
| 支付记录列表 | `GET /api/admin/payments` | ⚠️ |
| 支付统计 | `GET /api/admin/payments/stats` | ⚠️ |
| 手动同步支付 | `POST /api/client/payments/sync/manual/{outTradeNo}` | ❌ |
| 批量同步 | `POST /api/client/payments/sync/batch` | ❌ |
| 同步逾期订单 | `POST /api/client/payments/sync/overdue` | ❌ |

### 7. finance（财务结算）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 财务统计数据 | `GET /api/admin/finance/stats` | ✅ |
| 我的佣金 | `GET /api/admin/finance/operator-commissions/me` | ✅ |
| 我的佣金余额 | `GET /api/admin/finance/operator-commissions/me/balance` | ✅ |
| 提现记录 | `GET /api/admin/finance/operator-commissions/withdrawals` | ✅ |
| 我的提现记录 | `GET /api/admin/finance/operator-commissions/withdrawals/me` | ✅ |
| 申请提现 | `POST /api/admin/finance/operator-commissions/withdrawals` | ✅ |
| 审批通过 | `PUT /api/admin/finance/operator-commissions/withdrawals/{id}/approve` | ✅ |
| 审批拒绝 | `PUT /api/admin/finance/operator-commissions/withdrawals/{id}/reject` | ✅ |
| 打款 | `PUT /api/admin/finance/operator-commissions/withdrawals/{id}/pay` | ✅ |
| 取消 | `PUT /api/admin/finance/operator-commissions/withdrawals/{id}/cancel` | ✅ |
| 余额明细 | `GET /api/admin/balance/details` | ✅ |
| 提现列表 | `GET /api/admin/withdrawal/list` | ❌ |
| 提现详情 | `GET /api/admin/withdrawal/{id}` | ❌ |
| 审批通过 | `PUT /api/admin/withdrawal/{id}/approve` | ❌ |
| 审批拒绝 | `PUT /api/admin/withdrawal/{id}/reject` | ❌ |
| 拒绝订单结算 | `PUT /api/admin/withdrawal/{withdrawalId}/orders/{orderId}/reject` | ❌ |
| 打款 | `PUT /api/admin/withdrawal/{id}/pay` | ❌ |
| 提现统计 | `GET /api/admin/withdrawal/stats` | ❌ |
| 月度报表 | `GET /api/admin/withdrawal/monthly-report` | ✅ |
| 导出提现记录 | `GET /api/admin/withdrawal/export` | ❌ |
| 导出月度报表 | `GET /api/admin/withdrawal/monthly-report/export` | ❌ |
| 可结算订单 | `GET /api/admin/withdrawal/settleable-orders` | ❌ |
| 代结算预览 | `POST /api/admin/withdrawal/settle-on-behalf/preview` | ❌ |
| 代结算执行 | `POST /api/admin/withdrawal/settle-on-behalf` | ❌ |
| 用户月度报告 | `GET /api/admin/finance/user-monthly-report` | ✅ |
| 导出用户月度报告 | `GET /api/admin/finance/user-monthly-report/export` | ❌ |

### 8. notification（通知消息）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 通知列表（分页） | `GET /api/client/notifications` | ✅ |
| 标记已读 | `PUT /api/client/notifications/{id}/read` | ✅ |
| 全部已读 | `PUT /api/client/notifications/read-all` | ✅ |
| 系统通知列表 | `GET /api/client/notifications/system` | ✅ |
| 系统通知详情 | `GET /api/client/notifications/system/{id}` | ✅ |
| 系统通知未读数 | `GET /api/client/notifications/system/unread-count` | ✅ |
| 获取订阅模板 | `GET /api/client/subscribe-message/templates` | ✅ |
| 记录订阅 | `POST /api/client/subscribe-message/record` | ✅ |
| 订阅状态 | `GET /api/client/subscribe-message/status` | ✅ |
| 通知列表（管理端） | `GET /api/admin/notifications` | ✅ |
| 创建通知 | `POST /api/admin/notifications` | ✅ |
| 更新通知 | `PUT /api/admin/notifications/{id}` | ❌ |
| 删除通知 | `DELETE /api/admin/notifications/{id}` | ✅ |
| 标记已读 | `PUT /api/admin/notifications/{id}/read` | ❌ |
| 全部已读 | `PUT /api/admin/notifications/read-all` | ❌ |
| 通知统计 | `GET /api/admin/notifications/stats` | ✅ |
| 通知收件箱 | `GET /api/admin/notification-inbox` | ⚠️ |
| 标记收件箱已读 | `PUT /api/admin/notification-inbox/{id}/read` | ⚠️ |
| 收件箱全部已读 | `PUT /api/admin/notification-inbox/read-all` | ⚠️ |
| 切换待办状态 | `PUT /api/admin/system-todos/{id}/toggle` | ⚠️ |

### 9. chat（聊天客服）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 会话列表 | `GET /api/client/chat/sessions` | ✅ |
| 创建会话 | `POST /api/client/chat/sessions` | ✅ |
| 消息列表 | `GET /api/client/chat/sessions/{sessionId}/messages` | ✅ |
| 发送消息 | `POST /api/client/chat/sessions/{sessionId}/messages` | ✅ |
| 标记已读 | `PUT /api/client/chat/sessions/{sessionId}/read` | ✅ |
| 未读消息数 | `GET /api/client/chat/unread-count` | ✅ |
| 会话列表（管理端） | `GET /api/admin/chat/sessions` | ✅ |
| 消息列表（管理端） | `GET /api/admin/chat/sessions/{sessionId}/messages` | ✅ |
| 发送消息（管理端） | `POST /api/admin/chat/sessions/{sessionId}/messages` | ✅ |
| 客服配置 | `GET /api/common/customer-service/config` | ⚠️ |

### 10. file（文件存储）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 上传文件（Multipart） | `POST /api/upload/file` | ✅ |
| Base64 上传 | `POST /api/upload/base64` | ✅ |
| 检查文件哈希 | `GET /api/upload/check-hash` | ✅ |
| 获取上传 Token | `GET /api/upload/token` | ✅ |
| 确认上传 | `POST /api/upload/confirm` | ✅ |
| 素材列表 | `GET /api/admin/files` | ✅ |
| 素材详情 | `GET /api/admin/files/{id}` | ✅ |
| 创建素材记录 | `POST /api/admin/files` | ✅ |
| 更新素材 | `PUT /api/admin/files/{id}` | ✅ |
| 删除素材 | `DELETE /api/admin/files/{id}` | ✅ |
| 七牛云处理回调 | `POST /api/public/qiniu/pfop/callback` | ✅ |

### 11. system（系统配置）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 按前缀查询设置 | `GET /api/admin/settings/prefix/{prefix}` | ✅ |
| 检查初始化状态 | `GET /api/system/init/check` | ✅ |
| 执行初始化 | `POST /api/system/init/setup` | ✅ |
| 完整状态 | `GET /api/system/status/full` | ✅ |
| 应用状态 | `GET /api/system/status/application` | ✅ |
| 数据库状态 | `GET /api/system/status/database` | ✅ |
| Redis 状态 | `GET /api/system/status/redis` | ✅ |
| 环境变量 | `GET /api/system/status/environment` | ✅ |
| 系统信息 | `GET /api/system/status/system` | ✅ |
| 健康检查 | `GET /api/system/status/health` | ✅ |
| 菜单树 | `GET /api/admin/menus/tree` | ❌ |
| 菜单列表 | `GET /api/admin/menus/list` | ❌ |
| 菜单详情 | `GET /api/admin/menus/{id}` | ❌ |
| 批量创建菜单 | `POST /api/admin/menus/batch` | ❌ |
| 更新菜单 | `PUT /api/admin/menus/{id}` | ❌ |
| 删除菜单 | `DELETE /api/admin/menus/{id}` | ❌ |
| 级联选择器 | `GET /api/admin/menus/cascader` | ❌ |
| 权限列表 | `GET /api/admin/permissions` | ❌ |
| 更新权限 | `PUT /api/admin/permissions/{id}` | ❌ |
| 删除权限 | `DELETE /api/admin/permissions/{id}` | ❌ |
| 角色列表 | `GET /api/admin/roles` | ❌ |
| 创建角色 | `POST /api/admin/roles` | ❌ |
| 更新角色 | `PUT /api/admin/roles/{id}` | ❌ |
| 删除角色 | `DELETE /api/admin/roles/{id}` | ❌ |
| 更新角色状态 | `PUT /api/admin/roles/{id}/status` | ❌ |
| 分配权限 | `PUT /api/admin/roles/{id}/permissions` | ❌ |
| 分配菜单 | `PUT /api/admin/roles/{id}/menus` | ❌ |
| 客户列表 | `GET /api/admin/customers` | ❌ |
| 清除缓存 | `GET /api/admin/cache/clear` | ❌ |
| 待办列表 | `GET /api/admin/system-todos` | ✅ |
| 切换待办状态 | `PUT /api/admin/system-todos/{id}/toggle` | ✅ |
| 管理员日志 | `GET /api/admin/logs/admin` | ✅ |
| 错误日志 | `GET /api/admin/logs/error` | ✅ |
| 日志统计 | `GET /api/admin/logs/stats` | ✅ |
| 业务日志 | `GET /api/admin/logs/business` | ✅ |
| FaceId 配置列表 | `GET /api/admin/faceid/config` | ❌ |
| FaceId 配置详情 | `GET /api/admin/faceid/config/{id}` | ✅ |
| 更新 FaceId 配置 | `PUT /api/admin/faceid/config/{id}` | ✅ |
| 删除配置 | `DELETE /api/admin/faceid/config/{id}` | ✅ |
| 启用/禁用 | `PUT /api/admin/faceid/config/{id}/status` | ❌ |
| 支付宝配置分页 | `GET /api/admin/alipay/config/page` | ❌ |
| 支付宝配置详情 | `GET /api/admin/alipay/config/{id}` | ❌ |
| 更新支付宝配置 | `PUT /api/admin/alipay/config/{id}` | ❌ |
| 删除配置 | `DELETE /api/admin/alipay/config/{id}` | ❌ |
| 启用/禁用 | `PUT /api/admin/alipay/config/{id}/status` | ❌ |
| 微信支付配置列表 | `GET /api/admin/wxpay/config/page` | ✅ |
| 微信支付配置详情 | `GET /api/admin/wxpay/config/{id}` | ✅ |
| 更新微信支付配置 | `PUT /api/admin/wxpay/config/{id}` | ✅ |
| 删除配置 | `DELETE /api/admin/wxpay/config/{id}` | ✅ |
| 启用/禁用 | `PUT /api/admin/wxpay/config/{id}/status` | ✅ |
| 按类型查询微信支付配置 | `GET /api/admin/wxpay/config/type/{configType}` | ✅ |
| 查询用户手机号 | `GET /api/admin/sensitive/user/{userId}/mobile` | ✅ |
| 查询用户邮箱 | `GET /api/admin/sensitive/user/{userId}/email` | ✅ |
| 查询用户身份证 | `GET /api/admin/sensitive/user/{userId}/id-card` | ✅ |
| 查询选手手机号 | `GET /api/admin/sensitive/teacher/{teacherId}/mobile` | ✅ |
| 查询选手身份证 | `GET /api/admin/sensitive/teacher/{teacherId}/id-card` | ✅ |
| 查询选手银行卡 | `GET /api/admin/sensitive/teacher/{teacherId}/bank-account` | ✅ |
| 查询选手支付宝 | `GET /api/admin/sensitive/teacher/{teacherId}/alipay-account` | ✅ |
| 查询选手真实姓名 | `GET /api/admin/sensitive/teacher/{teacherId}/real-name` | ✅ |

### 12. visitor（访客分析）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 追踪访客行为 | `POST /api/common/visitor/track` | ✅ |
| 批量追踪 | `POST /api/common/visitor/batch` | ✅ |
| 访客统计 | `GET /api/admin/visitor-stats` | ✅ |
| 健康检查 | `GET /api/admin/visitor-stats/health/check` | ✅ |
| 健康状态 | `GET /api/admin/visitor-stats/health/status` | ✅ |
| Ping | `GET /api/admin/visitor-stats/health/ping` | ✅ |
| 性能告警 | `GET /api/admin/visitor-stats/performance` | ✅ |
| 性能指标 | `GET /api/admin/visitor-stats/performance/metrics` | ✅ |

### 13. game（游戏互动）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 创建房间 | `POST /api/client/game/room/create` | ✅ |
| 加入房间 | `POST /api/client/game/room/join` | ✅ |
| 离开房间 | `POST /api/client/game/room/leave/{roomId}` | ✅ |
| 房间信息 | `GET /api/client/game/room/{roomCode}` | ✅ |
| 开始游戏 | `POST /api/client/game/room/start/{roomId}` | ⚠️ |
| 解散房间 | `POST /api/client/game/room/disband/{roomId}` | ✅ |
| 地图列表 | `GET /api/client/game-map/list` | ✅ |
| 地图详情 | `GET /api/client/game-map/{id}` | ⚠️ |
| 地图列表（管理端） | `GET /api/admin/game-map` | ✅ |
| 地图详情（管理端） | `GET /api/admin/game-map/{id}` | ✅ |
| 更新地图 | `PUT /api/admin/game-map/{id}` | ✅ |
| 删除地图 | `DELETE /api/admin/game-map/{id}` | ✅ |
| 启用的地图 | `GET /api/admin/game-map/enabled` | ✅ |
| 排行榜 | `GET /api/admin/bomb-ranking` | ✅ |

### 14. external（外部集成）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 微信登录 | `POST /api/wechat/login` | ✅ |
| 绑定微信 | `POST /api/wechat/bind` | ✅ |
| 解绑微信 | `DELETE /api/wechat/unbind` | ✅ |
| 获取微信手机号 | `POST /api/wechat/phone` | ✅ |
| 生成扫码登录码 | `GET /api/wechat/scan-login/generate-qrcode` | ✅ |
| 检查扫码登录状态 | `GET /api/wechat/scan-login/check-status` | ✅ |
| 扫码确认登录 | `POST /api/wechat/scan-login/confirm` | ✅ |
| 取消登录 | `DELETE /api/wechat/scan-login/cancel` | ✅ |
| KOOK 绑定码 | `POST /api/client/kook/bind-code` | ✅ |
| 查询绑定状态 | `GET /api/client/kook/binding` | ✅ |
| 解绑 KOOK | `DELETE /api/client/kook/binding` | ✅ |

### 15. refund（退款管理）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 申请退款 | `POST /api/client/refunds/apply` | ✅ |
| 退款列表 | `GET /api/client/refunds/list` | ✅ |
| 退款详情 | `GET /api/client/refunds/{id}` | ✅ |
| 取消退款 | `POST /api/client/refunds/{id}/cancel` | ✅ |
| 检查是否可申请 | `GET /api/client/refunds/can-apply/{orderId}` | ✅ |
| 按订单查询退款 | `GET /api/client/refunds/by-order/{orderId}` | ✅ |
| 退款列表（管理端） | `GET /api/admin/refunds` | ✅ |
| 退款详情（管理端） | `GET /api/admin/refunds/{id}` | ✅ |
| 审批通过 | `PUT /api/admin/refunds/{id}/approve` | ✅ |
| 审批拒绝 | `PUT /api/admin/refunds/{id}/reject` | ✅ |
| 处理退款 | `PUT /api/admin/refunds/{id}/process` | ✅ |
| 退款统计 | `GET /api/admin/refunds/stats` | ✅ |

### 16. coupon（优惠券）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 可领取优惠券 | `GET /api/client/coupon/available` | ⚠️ |
| 我的优惠券 | `GET /api/client/coupon/my` | ⚠️ |
| 领取优惠券 | `POST /api/client/coupon/claim/{id}` | ⚠️ |
| 优惠券列表（管理端） | `GET /api/admin/coupon` | ⚠️ |
| 创建优惠券 | `POST /api/admin/coupon` | ⚠️ |
| 更新优惠券 | `PUT /api/admin/coupon/{id}` | ⚠️ |
| 删除优惠券 | `DELETE /api/admin/coupon/{id}` | ⚠️ |
| 优惠券统计 | `GET /api/admin/coupon/stats` | ⚠️ |

### 17. recharge（充值返利）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 手动充值 | `POST /api/recharge/manual` | ✅ |
| 创建充值订单 | `POST /api/recharge/create` | ✅ |
| 充值回调 | `POST /api/recharge/callback` | ✅ |
| 充值记录列表 | `GET /api/recharge/list` | ✅ |
| 我的充值记录 | `GET /api/recharge/my-records` | ✅ |
| 充值记录详情 | `GET /api/recharge/detail/{id}` | ✅ |
| 充值统计 | `GET /api/recharge/statistics` | ✅ |
| 最近充值记录 | `GET /api/recharge/recent/{userId}` | ✅ |
| 取消充值 | `POST /api/recharge/cancel/{rechargeNo}` | ✅ |
| 继续支付 | `POST /api/recharge/continue-pay/{rechargeNo}` | ✅ |
| 验证支付 | `POST /api/recharge/verify-payment/{rechargeNo}` | ✅ |
| 可用返利规则 | `GET /api/client/recharge-rebate/available-rules` | ✅ |
| 返利预览 | `GET /api/client/recharge-rebate/preview` | ✅ |
| 返利规则列表 | `GET /api/admin/recharge-rebate` | ✅ |
| 创建返利规则 | `POST /api/admin/recharge-rebate` | ✅ |
| 更新返利规则 | `PUT /api/admin/recharge-rebate/{id}` | ✅ |
| 删除返利规则 | `DELETE /api/admin/recharge-rebate/{id}` | ✅ |

### 18. withdrawal（提现）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 收入统计 | `GET /api/client/teacher/withdrawal/income-stats` | ✅ |
| 未结算订单 | `GET /api/client/teacher/withdrawal/unsettled-orders` | ⚠️ |
| 已结算订单 | `GET /api/client/teacher/withdrawal/settled-orders` | ⚠️ |
| 计算提现金额 | `POST /api/client/teacher/withdrawal/calculate` | ✅ |
| 申请提现 | `POST /api/client/teacher/withdrawal/apply` | ✅ |
| 取消提现 | `PUT /api/client/teacher/withdrawal/{id}/cancel` | ✅ |
| 提现记录 | `GET /api/client/teacher/withdrawal/records` | ✅ |
| 提现详情 | `GET /api/client/teacher/withdrawal/{id}` | ✅ |

### 19. invite（邀请机制）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 邀请信息 | `GET /api/client/invite/info` | ✅ |
| 邀请记录 | `GET /api/client/invite/records` | ✅ |
| 绑定邀请人 | `POST /api/client/invite/bindInviter` | ✅ |
| 验证邀请码 | `GET /api/client/invite/validate` | ✅ |
| 我的邀请码 | `GET /api/client/teacher/invite-code` | ✅ |
| 生成邀请码 | `POST /api/client/teacher/invite-code` | ✅ |
| 邀请码列表 | `GET /api/admin/teacher/invite-code` | ✅ |
| 创建邀请码 | `POST /api/admin/teacher/invite-code` | ✅ |
| 更新邀请码 | `PUT /api/admin/teacher/invite-code/{id}` | ✅ |
| 删除邀请码 | `DELETE /api/admin/teacher/invite-code/{id}` | ✅ |

### 20. feedback（用户反馈）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 提交反馈 | `POST /api/client/feedback/submit` | ⚠️ |
| 反馈列表 | `GET /api/client/feedback/list` | ⚠️ |
| 反馈详情 | `GET /api/client/feedback/{id}` | ⚠️ |
| 反馈列表（管理端） | `GET /api/admin/feedback` / `GET /api/admin/feedback/list` | ⚠️ |
| 反馈详情（管理端） | `GET /api/admin/feedback/{id}` | ⚠️ |
| 回复反馈 | `POST /api/admin/feedback/{id}/reply` / `POST /api/admin/feedback/reply` | ⚠️ |
| 更新状态 | `PUT /api/admin/feedback/{id}/status` / `PUT /api/admin/feedback/status` | ⚠️ |
| 删除反馈 | `DELETE /api/admin/feedback/{id}` | ⚠️ |

### 21. partner（合作伙伴）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 配置列表 | `GET /api/admin/partner-config` | ✅ |
| 创建配置 | `POST /api/admin/partner-config` | ✅ |
| 更新配置 | `PUT /api/admin/partner-config` | ✅ |
| 删除配置 | `DELETE /api/admin/partner-config/{id}` | ✅ |
| 合作伙伴列表 | `GET /api/admin/teacher/partners` | ✅ |
| 更新合作 | `PUT /api/admin/teacher/partners/{id}` | ✅ |
| 删除合作 | `DELETE /api/admin/teacher/partners/{id}` | ✅ |
| 选手合作记录 | `GET /api/admin/teacher/partners/teacher/{teacherId}` | ✅ |
| 已合作选手 | `GET /api/admin/teacher/partners/partnered-teachers` | ✅ |

### 22. customer_service（客服系统）

| 功能点 | 接口路径 | 迁移状态 |
|--------|----------|----------|
| 客服配置 | `GET /api/common/customer-service/config` | ✅ |
| 消息列表 | `GET /api/admin/chat/sessions/{sessionId}/messages` | ✅ |
| 发送消息 | `POST /api/admin/chat/sessions/{sessionId}/messages` | ✅ |

## 三、迁移进度统计

| 状态 | 数量 | 占比 |
|------|------|------|
| ✅ 已完成 | ~45 | 15% |
| ⚠️ 部分完成 | ~43 | 15% |
| ❌ 未迁移 | ~192 | 70% |

## 四、后续迁移建议

1. **高优先级（核心业务流程）**
   - order：订单创建、评价、最终审核、转移
   - payment：微信/支付宝支付集成
   - refund：退款状态机
   - user：用户中心完整流程

2. **中优先级（运营必需）**
   - system：RBAC 权限、菜单、角色
   - goods：SKU 管理、限购、Banner
   - teacher：申请审核、动态、评价
   - finance：提现管理、佣金结算

3. **低优先级（可延后）**
   - coupon：优惠券
   - recharge：充值返利
   - invite：邀请机制
   - feedback：用户反馈
   - partner：合作伙伴
