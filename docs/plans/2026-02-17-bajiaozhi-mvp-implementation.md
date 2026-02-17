# 八角志 MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 交付「八角志」首版可上线 MVP：微信小程序（资讯/赛程/战卡/选手搜索）、后台管理（抓取源/审核/赛事/选手）、Go 后端（模块化单体 + Redis 队列）并支持 live 赛事 30 秒赛果更新。

**Architecture:** 采用模块化单体后端（API + Worker 分进程）与共享 MySQL/Redis。抓取任务入 Redis 队列，Worker 消费后写入待审核池，审核通过后发布到小程序。live 赛事由高频任务每 30 秒更新战卡结果并幂等入库。

**Tech Stack:** Go 1.24+, Gin, GORM/sqlc, MySQL 8, Redis 7, Vue 3 + Vite + Element Plus, 微信原生小程序, Vitest, Go test, Playwright（可选）

---

### Task 1: Monorepo Scaffolding

**Files:**
- Create: `backend/cmd/api/main.go`
- Create: `backend/cmd/worker/main.go`
- Create: `backend/go.mod`
- Create: `admin/package.json`
- Create: `miniapp/project.config.json`
- Create: `Makefile`
- Test: `backend/internal/bootstrap/bootstrap_test.go`

**Step 1: Write the failing test**

```go
package bootstrap

import "testing"

func TestLoadConfig_MissingEnvFails(t *testing.T) {
    _, err := LoadConfigFromEnv()
    if err == nil {
        t.Fatalf("expected error when required env is missing")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/bootstrap -v`
Expected: FAIL with `undefined: LoadConfigFromEnv`

**Step 3: Write minimal implementation**

```go
func LoadConfigFromEnv() (Config, error) {
    dsn := os.Getenv("MYSQL_DSN")
    if dsn == "" {
        return Config{}, errors.New("MYSQL_DSN is required")
    }
    return Config{MySQLDSN: dsn}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/bootstrap -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/cmd backend/go.mod backend/internal/bootstrap admin/package.json miniapp/project.config.json Makefile
git commit -m "chore: scaffold monorepo for bajiaozhi mvp"
```

### Task 2: Base API + Health Check

**Files:**
- Create: `backend/internal/http/server.go`
- Create: `backend/internal/http/health_handler.go`
- Modify: `backend/cmd/api/main.go`
- Test: `backend/internal/http/health_handler_test.go`

**Step 1: Write the failing test**

```go
func TestHealthHandler(t *testing.T) {
    r := gin.New()
    RegisterRoutes(r)
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/http -v`
Expected: FAIL with `undefined: RegisterRoutes`

**Step 3: Write minimal implementation**

```go
func RegisterRoutes(r *gin.Engine) {
    r.GET("/healthz", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/http -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/cmd/api/main.go backend/internal/http
git commit -m "feat: add api health endpoint"
```

### Task 3: Schema Migrations (资讯/赛事/战卡/选手/审核)

**Files:**
- Create: `backend/migrations/0001_init_schema.up.sql`
- Create: `backend/migrations/0001_init_schema.down.sql`
- Create: `backend/internal/storage/migrate_test.go`
- Create: `backend/internal/storage/testcontainers.go`

**Step 1: Write the failing test**

```go
func TestMigrations_CreateCoreTables(t *testing.T) {
    db := setupMySQLForTest(t)
    applyMigrations(t, db)
    mustHaveTable(t, db, "articles")
    mustHaveTable(t, db, "events")
    mustHaveTable(t, db, "bouts")
    mustHaveTable(t, db, "fighters")
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/storage -run TestMigrations_CreateCoreTables -v`
Expected: FAIL table not found

**Step 3: Write minimal implementation**

```sql
CREATE TABLE events (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  org VARCHAR(32) NOT NULL,
  name VARCHAR(255) NOT NULL,
  status ENUM('scheduled','live','completed') NOT NULL DEFAULT 'scheduled',
  starts_at DATETIME NOT NULL,
  venue VARCHAR(255) NOT NULL,
  updated_at DATETIME NOT NULL
);
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/storage -run TestMigrations_CreateCoreTables -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/migrations backend/internal/storage
git commit -m "feat: add initial mysql schema"
```

### Task 4: Queue-Based Ingest Pipeline

**Files:**
- Create: `backend/internal/ingest/enqueue.go`
- Create: `backend/internal/ingest/worker.go`
- Create: `backend/internal/ingest/repository.go`
- Test: `backend/internal/ingest/worker_test.go`

**Step 1: Write the failing test**

```go
func TestWorker_StoresPendingReviewOnSuccess(t *testing.T) {
    queue := newFakeQueue()
    repo := newFakeRepo()
    queue.Push(FetchJob{SourceID: 1, URL: "https://example.com/a"})

    w := NewWorker(queue, repo, fakeParserSuccess())
    w.RunOnce(context.Background())

    if repo.pendingCount != 1 {
        t.Fatalf("expected 1 pending record, got %d", repo.pendingCount)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/ingest -v`
Expected: FAIL with `undefined: NewWorker`

**Step 3: Write minimal implementation**

```go
func (w *Worker) RunOnce(ctx context.Context) {
    job, ok := w.queue.Pop(ctx)
    if !ok { return }
    rec, err := w.parser.Parse(ctx, job.URL)
    if err != nil { return }
    _ = w.repo.SavePending(ctx, rec)
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/ingest -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/ingest
git commit -m "feat: implement queue-based ingest worker"
```

### Task 5: Review & Publish Workflow

**Files:**
- Create: `backend/internal/review/service.go`
- Create: `backend/internal/review/http_handler.go`
- Test: `backend/internal/review/service_test.go`
- Test: `backend/internal/review/http_handler_test.go`

**Step 1: Write the failing test**

```go
func TestApprove_PublishesArticle(t *testing.T) {
    repo := newFakeReviewRepo()
    svc := NewService(repo)
    err := svc.Approve(context.Background(), 101, 9001)
    if err != nil { t.Fatal(err) }
    if !repo.articlePublished {
        t.Fatalf("expected article to be published")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/review -v`
Expected: FAIL with `undefined: NewService`

**Step 3: Write minimal implementation**

```go
func (s *Service) Approve(ctx context.Context, pendingID int64, reviewerID int64) error {
    rec, err := s.repo.GetPending(ctx, pendingID)
    if err != nil { return err }
    if err := s.repo.PublishArticle(ctx, rec); err != nil { return err }
    return s.repo.MarkApproved(ctx, pendingID, reviewerID)
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/review -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/review
git commit -m "feat: add review and publish workflow"
```

### Task 6: Event Card + Fighter Search APIs

**Files:**
- Create: `backend/internal/event/service.go`
- Create: `backend/internal/event/http_handler.go`
- Create: `backend/internal/fighter/service.go`
- Create: `backend/internal/fighter/http_handler.go`
- Test: `backend/internal/event/http_handler_test.go`
- Test: `backend/internal/fighter/http_handler_test.go`

**Step 1: Write the failing test**

```go
func TestGetEventCard_ReturnsBouts(t *testing.T) {
    r := gin.New()
    RegisterEventRoutes(r, fakeEventService())
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/events/10", nil)
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/event -v`
Expected: FAIL with route/handler missing

**Step 3: Write minimal implementation**

```go
r.GET("/api/events/:id", func(c *gin.Context) {
    card := svc.GetEventCard(c.Request.Context(), c.Param("id"))
    c.JSON(http.StatusOK, card)
})
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/event ./internal/fighter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/event backend/internal/fighter
git commit -m "feat: add event card and fighter search apis"
```

### Task 7: Live Result Updater (30s)

**Files:**
- Create: `backend/internal/live/scheduler.go`
- Create: `backend/internal/live/updater.go`
- Test: `backend/internal/live/updater_test.go`

**Step 1: Write the failing test**

```go
func TestUpdater_UpdatesBoutResultIdempotently(t *testing.T) {
    repo := newFakeLiveRepo()
    client := fakeLiveClientWinner("fighter_a")
    u := NewUpdater(repo, client)

    _ = u.UpdateEvent(context.Background(), 10)
    _ = u.UpdateEvent(context.Background(), 10)

    if repo.updateCount != 1 {
        t.Fatalf("expected idempotent update count 1, got %d", repo.updateCount)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/live -v`
Expected: FAIL with `undefined: NewUpdater`

**Step 3: Write minimal implementation**

```go
func (u *Updater) UpdateEvent(ctx context.Context, eventID int64) error {
    results, err := u.client.FetchEventResults(ctx, eventID)
    if err != nil { return err }
    for _, r := range results {
        _ = u.repo.UpsertBoutResult(ctx, r)
    }
    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/live -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/live
git commit -m "feat: add 30s live result updater"
```

### Task 8: Admin Web - Review Queue & Event Management

**Files:**
- Create: `admin/src/pages/review/ReviewQueue.vue`
- Create: `admin/src/pages/events/EventEditor.vue`
- Create: `admin/src/api/review.ts`
- Create: `admin/src/api/events.ts`
- Test: `admin/src/pages/review/ReviewQueue.spec.ts`

**Step 1: Write the failing test**

```ts
it('shows pending items and approves one item', async () => {
  mockListPending([{ id: 1, title: 'news-a' }])
  const wrapper = mount(ReviewQueue)
  await flushPromises()
  expect(wrapper.text()).toContain('news-a')
  await wrapper.get('[data-test="approve-1"]').trigger('click')
  expect(mockApprove).toHaveBeenCalledWith(1)
})
```

**Step 2: Run test to verify it fails**

Run: `cd admin && pnpm vitest src/pages/review/ReviewQueue.spec.ts`
Expected: FAIL component/api missing

**Step 3: Write minimal implementation**

```ts
export async function approvePending(id: number) {
  return request.post(`/admin/review/${id}/approve`)
}
```

**Step 4: Run test to verify it passes**

Run: `cd admin && pnpm vitest src/pages/review/ReviewQueue.spec.ts`
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src
git commit -m "feat: add admin review queue and event editor"
```

### Task 9: Mini Program - Schedule, Card, Fighter Search

**Files:**
- Create: `miniapp/pages/schedule/index.js`
- Create: `miniapp/pages/event-detail/index.js`
- Create: `miniapp/pages/fighter/index.js`
- Create: `miniapp/pages/search-fighter/index.js`
- Create: `miniapp/services/api.js`
- Test: `miniapp/tests/navigation.spec.js`

**Step 1: Write the failing test**

```js
test('from schedule to event card to fighter detail', async () => {
  const app = launchMiniApp()
  await app.open('/pages/schedule/index')
  await app.tap('[data-test="event-10"]')
  await app.tap('[data-test="fighter-20"]')
  expect(app.currentPage()).toBe('/pages/fighter/index?id=20')
})
```

**Step 2: Run test to verify it fails**

Run: `cd miniapp && npm test -- navigation.spec.js`
Expected: FAIL pages missing

**Step 3: Write minimal implementation**

```js
wx.navigateTo({
  url: `/pages/fighter/index?id=${fighterId}`,
})
```

**Step 4: Run test to verify it passes**

Run: `cd miniapp && npm test -- navigation.spec.js`
Expected: PASS

**Step 5: Commit**

```bash
git add miniapp
git commit -m "feat: add miniapp schedule card fighter-search flow"
```

### Task 10: End-to-End Verification & Release Checklist

**Files:**
- Create: `docs/release/mvp-checklist.md`
- Modify: `README.md`
- Test: `backend/tests/e2e/live_update_e2e_test.go`
- Test: `admin/e2e/review_publish.spec.ts`

**Step 1: Write the failing test**

```go
func TestE2E_LiveEventUpdatesEvery30Seconds(t *testing.T) {
    // seed live event + mock upstream result change
    // assert API returns updated bout winner within 35 seconds
}
```

**Step 2: Run test to verify it fails**

Run: `cd backend && go test ./tests/e2e -run TestE2E_LiveEventUpdatesEvery30Seconds -v`
Expected: FAIL missing harness/assertion

**Step 3: Write minimal implementation**

```go
// Test harness starts api + worker, injects fake provider,
// waits polling window, and verifies winner field changed.
```

**Step 4: Run test to verify it passes**

Run: `make test-e2e`
Expected: PASS with live update and review publish scenarios green

**Step 5: Commit**

```bash
git add docs/release/mvp-checklist.md README.md backend/tests/e2e admin/e2e
git commit -m "test: add e2e verification and release checklist"
```

## Execution Notes
- 全程遵循 `@test-driven-development`：每个功能先写失败测试，再写最小实现。
- 发生异常时使用 `@systematic-debugging`，禁止跳过复现与定位。
- 完成前必须执行 `@verification-before-completion`，以命令输出作为验收证据。
- 若需要并行推进前后端任务，使用 `@subagent-driven-development` 做任务级并行。

## Done Criteria
- 小程序可完成：资讯浏览、赛程浏览、战卡查看、点击选手查看详情、选手名称搜索
- 后台可完成：抓取源管理、待审核处理、赛事与选手维护
- 后端可完成：抓取入队、审核发布、live 赛果 30 秒更新、幂等写入
- 关键测试通过：单元/集成/E2E 全绿
