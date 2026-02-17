# 八角志全量持久化与缓存加速 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将八角志后端从内存实现迁移为 MySQL + Redis 持久化架构，并为小程序读取增加 Redis 缓存层。

**Architecture:** API 与 Worker 共用 MySQL/Redis 基础设施。业务写入统一落 MySQL，读取采用 Cache-Aside。抓取与 live 任务通过 Redis Stream 持久化，Worker 消费并 ACK。写成功后执行缓存失效与延迟双删，确保一致性。

**Tech Stack:** Go 1.24+, Gin, GORM, go-redis/v9, MySQL 8, Redis 7, Docker Compose

---

### Task 1: 基础设施引导（MySQL + Redis + Compose）

**Files:**
- Create: `backend/internal/bootstrap/config.go`
- Create: `backend/internal/bootstrap/mysql.go`
- Create: `backend/internal/bootstrap/redis.go`
- Create: `docker-compose.yml`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/cmd/worker/main.go`
- Test: `backend/internal/bootstrap/config_test.go`

**Step 1: Write the failing test**

```go
func TestLoadConfig_RequiresMySQLAndRedis(t *testing.T) {
    os.Clearenv()
    _, err := LoadConfigFromEnv()
    if err == nil {
        t.Fatalf("expected error when required env missing")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/bootstrap -v`
Expected: FAIL with missing Redis config fields or parser

**Step 3: Write minimal implementation**

```go
type Config struct {
    MySQLDSN  string
    RedisAddr string
    RedisPass string
    RedisDB   int
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/bootstrap -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/bootstrap backend/cmd docker-compose.yml
git commit -m "feat: bootstrap mysql redis and compose runtime"
```

### Task 2: 持久化 schema 与模型（GORM）

**Files:**
- Create: `backend/migrations/0002_persistence_schema.up.sql`
- Create: `backend/migrations/0002_persistence_schema.down.sql`
- Create: `backend/internal/model/article.go`
- Create: `backend/internal/model/event.go`
- Create: `backend/internal/model/fighter.go`
- Test: `backend/internal/storage/persistence_schema_test.go`

**Step 1: Write the failing test**

```go
func TestSchema_HasPendingArticlesAndFighterUpdates(t *testing.T) {
    db := setupMySQLForTest(t)
    applyMigration(t, db, "../../migrations/0002_persistence_schema.up.sql")
    mustHaveTable(t, db, "pending_articles")
    mustHaveTable(t, db, "fighter_updates")
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/storage -run TestSchema_HasPendingArticlesAndFighterUpdates -v`
Expected: FAIL table not found

**Step 3: Write minimal implementation**

```sql
CREATE TABLE pending_articles (...);
CREATE TABLE fighter_updates (...);
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/storage -run TestSchema_HasPendingArticlesAndFighterUpdates -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/migrations backend/internal/model backend/internal/storage
git commit -m "feat: add persistence schema and gorm models"
```

### Task 3: Review/Article 仓储持久化 + 缓存

**Files:**
- Create: `backend/internal/repository/mysql/article_repository.go`
- Create: `backend/internal/repository/cache/article_cache.go`
- Modify: `backend/internal/review/service.go`
- Modify: `backend/internal/review/http_handler.go`
- Test: `backend/internal/review/service_persistence_test.go`
- Test: `backend/internal/repository/cache/article_cache_test.go`

**Step 1: Write the failing test**

```go
func TestApprove_PersistsArticleAndInvalidatesCache(t *testing.T) {
    repo := newFakeArticleRepo()
    cache := newFakeCache()
    svc := NewService(repo, cache)
    _ = svc.Approve(context.Background(), 1, 9001)
    if !repo.published { t.Fatal("not persisted") }
    if !cache.deleted["cache:articles:list:v1"] { t.Fatal("cache not invalidated") }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/review -run TestApprove_PersistsArticleAndInvalidatesCache -v`
Expected: FAIL constructor/signature mismatch

**Step 3: Write minimal implementation**

```go
if err := tx.Commit().Error; err != nil { return err }
cache.Del(ctx, "cache:articles:list:v1")
go delayedDelete(...)
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/review ./internal/repository/cache -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/review backend/internal/repository/mysql backend/internal/repository/cache
git commit -m "feat: persist review workflow and cache invalidation"
```

### Task 4: Event/Bout 持久化 + 缓存读写

**Files:**
- Create: `backend/internal/repository/mysql/event_repository.go`
- Create: `backend/internal/repository/cache/event_cache.go`
- Modify: `backend/internal/event/service.go`
- Modify: `backend/internal/event/http_handler.go`
- Test: `backend/internal/event/service_persistence_test.go`

**Step 1: Write the failing test**

```go
func TestGetEventCard_ReadsFromCacheThenDB(t *testing.T) {
    cache := newFakeEventCacheMiss()
    repo := newFakeEventRepo()
    svc := NewService(repo, cache)
    _, _ = svc.GetEventCard(context.Background(), 10)
    if repo.getCardCalls != 1 { t.Fatal("expected db fallback") }
    if !cache.setCalled { t.Fatal("expected cache fill") }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/event -run TestGetEventCard_ReadsFromCacheThenDB -v`
Expected: FAIL cache dependency missing

**Step 3: Write minimal implementation**

```go
if card, ok := cache.GetEventCard(...); ok { return card, nil }
card, err := repo.GetEventCard(...)
cache.SetEventCard(...)
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/event -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/event backend/internal/repository/mysql backend/internal/repository/cache
git commit -m "feat: persist event card and add redis cache-aside"
```

### Task 5: Fighter 持久化 + 缓存读写

**Files:**
- Create: `backend/internal/repository/mysql/fighter_repository.go`
- Create: `backend/internal/repository/cache/fighter_cache.go`
- Modify: `backend/internal/fighter/service.go`
- Modify: `backend/internal/fighter/http_handler.go`
- Test: `backend/internal/fighter/service_persistence_test.go`

**Step 1: Write the failing test**

```go
func TestSearch_UsesCacheAside(t *testing.T) {
    cache := newFakeFighterCacheMiss()
    repo := newFakeFighterRepo()
    svc := NewService(repo, cache)
    _, _ = svc.Search(context.Background(), "Alex")
    if repo.searchCalls != 1 { t.Fatal("expected db search") }
    if !cache.searchSetCalled { t.Fatal("expected cache fill") }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/fighter -run TestSearch_UsesCacheAside -v`
Expected: FAIL cache wiring missing

**Step 3: Write minimal implementation**

```go
if items, ok := cache.GetSearch(...); ok { return items, nil }
items, err := repo.SearchByName(...)
cache.SetSearch(...)
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/fighter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/fighter backend/internal/repository/mysql backend/internal/repository/cache
git commit -m "feat: persist fighter module and cache search/detail"
```

### Task 6: Redis Stream 队列化 ingest 与 live

**Files:**
- Create: `backend/internal/queue/stream_queue.go`
- Modify: `backend/internal/ingest/enqueue.go`
- Modify: `backend/internal/ingest/worker.go`
- Modify: `backend/internal/live/scheduler.go`
- Modify: `backend/internal/live/updater.go`
- Test: `backend/internal/queue/stream_queue_test.go`
- Test: `backend/internal/ingest/worker_queue_test.go`

**Step 1: Write the failing test**

```go
func TestConsume_AcksMessageOnSuccess(t *testing.T) {
    q := newFakeStreamQueue()
    q.Publish(...)
    handled := false
    _ = q.Consume(context.Background(), func(_ Job) error { handled = true; return nil })
    if !handled || !q.acked { t.Fatal("consume/ack failed") }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/queue -v`
Expected: FAIL missing stream queue implementation

**Step 3: Write minimal implementation**

```go
msg := XReadGroup(...)
if err := handler(job); err == nil { XAck(...) }
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/queue ./internal/ingest ./internal/live -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/queue backend/internal/ingest backend/internal/live
git commit -m "feat: migrate ingest and live tasks to redis streams"
```

### Task 7: API/Worker 依赖注入重构

**Files:**
- Modify: `backend/internal/http/server.go`
- Modify: `backend/internal/http/health_handler.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/cmd/worker/main.go`
- Test: `backend/internal/http/server_integration_test.go`

**Step 1: Write the failing test**

```go
func TestServer_UsesPersistentRepositories(t *testing.T) {
    srv := NewServer(newFakeDeps())
    if srv == nil { t.Fatal("server not built") }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/http -run TestServer_UsesPersistentRepositories -v`
Expected: FAIL constructor mismatch

**Step 3: Write minimal implementation**

```go
func NewServer(deps Dependencies) *gin.Engine { ... }
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/http -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/http backend/cmd
git commit -m "refactor: wire api and worker with mysql redis dependencies"
```

### Task 8: 本地部署与验收文档

**Files:**
- Modify: `README.md`
- Modify: `docs/release/mvp-checklist.md`
- Create: `ops/.env.example`
- Test: `backend/tests/e2e/persistence_e2e_test.go`

**Step 1: Write the failing test**

```go
func TestE2E_PersistedArticleSurvivesProcessRestart(t *testing.T) {
    // publish article -> restart api -> query /api/articles -> still exists
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./tests/e2e -run TestE2E_PersistedArticleSurvivesProcessRestart -v`
Expected: FAIL current in-memory implementation

**Step 3: Write minimal implementation**

```go
// e2e harness with compose services, API restart, persisted read assertion
```

**Step 4: Run verification**

Run: `cd /Users/weizhenguo/ai_coding_projects/w-mma && make test-e2e`
Expected: PASS

**Step 5: Commit**

```bash
git add README.md docs/release/mvp-checklist.md ops/.env.example backend/tests/e2e
git commit -m "test: add persistence e2e and local deploy guide"
```

## Execution Notes
- 全程遵循 `@test-driven-development`。
- 失败时遵循 `@systematic-debugging`，禁止跳过复现。
- 完成声明前遵循 `@verification-before-completion`。
- 每个任务完成后执行对应测试并提交小步 commit。
