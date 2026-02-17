# 八角志全量持久化与缓存加速设计

## 1. 目标
将当前内存实现升级为生产可用的数据层：
- 全业务数据持久化到 MySQL（资讯、审核、赛事、战卡、选手、动态、live 结果）
- 抓取与 live 任务使用 Redis Stream 持久化队列
- 小程序读取链路增加 Redis 缓存层（Cache-Aside）
- 保证 Redis 与 MySQL 一致性（写库后删缓存 + 延迟双删）
- 提供 Docker Compose 一键本地部署与验收

## 2. 架构与数据流
### 2.1 组件
- `api`：Gin + GORM，提供后台与小程序接口
- `worker`：消费 Redis Stream，执行抓取与 live 更新
- `mysql`：业务主存储
- `redis`：队列与读取缓存

### 2.2 核心流程
1. 后台 `POST /admin/ingest/fetch` 入 `stream:ingest:fetch`
2. Worker 消费任务并抓真实 URL，入 `pending_articles`
3. 审核通过后写 `articles`，小程序经缓存读取
4. live 任务入 `stream:live:update`，worker 幂等更新 `bouts`
5. 小程序读取统一走 `Redis 命中 -> MySQL 回源 -> 回填缓存`

## 3. 数据模型（MySQL）
- `articles`
- `pending_articles`
- `events`
- `bouts`
- `fighters`
- `fighter_updates`

索引与约束：
- `articles.source_url` 唯一
- `bouts(event_id, sequence_no)` 唯一
- `pending_articles(status, created_at)` 索引
- `events(status, starts_at)` 索引
- `fighters(name)` 索引

## 4. Redis 设计
### 4.1 Stream
- `stream:ingest:fetch`
- `stream:live:update`
- `stream:dlq:*`

### 4.2 Cache Key
- `cache:articles:list:v1`
- `cache:events:list:v1:{org}:{date}`
- `cache:event:detail:v1:{event_id}`
- `cache:fighter:detail:v1:{fighter_id}`
- `cache:fighter:search:v1:{q}`

### 4.3 TTL
- 资讯列表：120s
- 赛事列表：60s
- 赛事详情（live）：20s，非 live：120s
- 选手详情：300s
- 搜索：120s

## 5. 一致性策略
采用 Cache-Aside：
- 读：先 Redis，未命中回 MySQL，再写缓存
- 写：MySQL 事务成功后立即删缓存
- 延迟双删：提交后 200-500ms 再删一次相关 key

失效规则：
- 审核发布资讯：删 `cache:articles:list:v1`
- 赛事/战卡更新：删 `cache:events:list:*` + `cache:event:detail:v1:{event_id}`
- 选手数据更新：删 `cache:fighter:detail:*` + `cache:fighter:search:*`
- live 更新：删赛事详情和赛事列表缓存

## 6. 迁移策略（分层迁移）
1. 抽象 repository/cache/queue 接口
2. 先落 MySQL + Redis 基础设施与 compose
3. 按模块替换实现：review/content -> event -> fighter -> live -> ingest
4. 每步保持测试可运行，逐步删除内存实现

## 7. 测试与验收
- 单元：缓存命中/回源、删缓存触发、幂等更新
- 集成：抓取-审核-发布链路，live 更新链路
- E2E：小程序关键接口可读到持久化数据
- 降级：Redis 故障回源 MySQL 仍可用

## 8. 本地部署
使用 Docker Compose 一键启动：
- `mysql`
- `redis`
- `api`
- `worker`

验收步骤：
1. 启动 compose
2. 提交抓取任务
3. 审核发布
4. 验证 `/api/articles` 数据持久化
5. 验证缓存命中与更新后失效
