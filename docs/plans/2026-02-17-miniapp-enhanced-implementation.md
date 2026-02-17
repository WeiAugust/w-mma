# 八角志小程序增强版 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成小程序增强版开发，直连后端真实数据并支持资讯、赛程、战卡、选手详情、选手搜索与页面状态管理。

**Architecture:** 基于微信小程序原生页面结构（`json/wxml/wxss/js`）构建五个页面，统一使用 `services/api.js` 访问后端。页面层采用一致的 `loading/error/empty` 状态模型，并通过事件跳转形成完整业务链路。测试层使用 Jest 对 API 与页面逻辑进行单元覆盖并保留导航链路回归测试。

**Tech Stack:** 微信小程序原生、Node.js、Jest

---

### Task 1: Miniapp 基础结构与页面注册

**Files:**
- Create: `miniapp/app.json`
- Create: `miniapp/app.js`
- Create: `miniapp/app.wxss`
- Create: `miniapp/pages/news/index.json`
- Create: `miniapp/pages/schedule/index.json`
- Create: `miniapp/pages/event-detail/index.json`
- Create: `miniapp/pages/fighter/index.json`
- Create: `miniapp/pages/search-fighter/index.json`
- Test: `miniapp/tests/app-config.spec.js`

**Step 1: Write the failing test**

```js
const appConfig = require('../app.json')

test('registers all mvp pages in app config', () => {
  expect(appConfig.pages).toEqual([
    'pages/news/index',
    'pages/schedule/index',
    'pages/event-detail/index',
    'pages/fighter/index',
    'pages/search-fighter/index',
  ])
})
```

**Step 2: Run test to verify it fails**

Run: `cd miniapp && npm test -- app-config.spec.js`
Expected: FAIL with `Cannot find module '../app.json'`

**Step 3: Write minimal implementation**

```json
{
  "pages": [
    "pages/news/index",
    "pages/schedule/index",
    "pages/event-detail/index",
    "pages/fighter/index",
    "pages/search-fighter/index"
  ]
}
```

**Step 4: Run test to verify it passes**

Run: `cd miniapp && npm test -- app-config.spec.js`
Expected: PASS

**Step 5: Commit**

```bash
git add miniapp/app.* miniapp/pages/*/index.json miniapp/tests/app-config.spec.js
git commit -m "feat: scaffold enhanced miniapp pages and app config"
```

### Task 2: 统一 API 请求层（直连后端）

**Files:**
- Modify: `miniapp/services/api.js`
- Test: `miniapp/tests/api.spec.js`

**Step 1: Write the failing test**

```js
const { listArticles } = require('../services/api')

test('listArticles requests backend articles endpoint', async () => {
  global.wx = {
    request: jest.fn(({ success }) => success({ data: { items: [] } })),
  }

  await listArticles()

  expect(global.wx.request).toHaveBeenCalledWith(
    expect.objectContaining({ url: 'http://localhost:8080/api/articles', method: 'GET' }),
  )
})
```

**Step 2: Run test to verify it fails**

Run: `cd miniapp && npm test -- api.spec.js`
Expected: FAIL with `listArticles is not a function`

**Step 3: Write minimal implementation**

```js
function listArticles() {
  return request('/api/articles')
}
```

**Step 4: Run test to verify it passes**

Run: `cd miniapp && npm test -- api.spec.js`
Expected: PASS

**Step 5: Commit**

```bash
git add miniapp/services/api.js miniapp/tests/api.spec.js
git commit -m "feat: add miniapp backend api client for real data"
```

### Task 3: 资讯页增强（加载/空态/错误态/刷新）

**Files:**
- Create: `miniapp/pages/news/index.js`
- Create: `miniapp/pages/news/index.wxml`
- Create: `miniapp/pages/news/index.wxss`
- Test: `miniapp/tests/news-page.spec.js`

**Step 1: Write the failing test**

```js
const newsPage = require('../pages/news/index')

test('loadArticles writes items on success', async () => {
  const ctx = createPageContext(newsPage)
  newsPage.__setApi({ listArticles: jest.fn().mockResolvedValue({ items: [{ id: 1, title: 'a' }] }) })

  await newsPage.loadArticles.call(ctx)

  expect(ctx.data.items).toHaveLength(1)
  expect(ctx.data.error).toBe('')
})
```

**Step 2: Run test to verify it fails**

Run: `cd miniapp && npm test -- news-page.spec.js`
Expected: FAIL with missing `__setApi`/`loadArticles`

**Step 3: Write minimal implementation**

```js
async function loadArticles() {
  this.setData({ loading: true, error: '' })
  try {
    const data = await api.listArticles()
    this.setData({ items: data.items || [], loading: false })
  } catch (err) {
    this.setData({ loading: false, error: err.message || '加载失败' })
  }
}
```

**Step 4: Run test to verify it passes**

Run: `cd miniapp && npm test -- news-page.spec.js`
Expected: PASS

**Step 5: Commit**

```bash
git add miniapp/pages/news miniapp/tests/news-page.spec.js
git commit -m "feat: implement miniapp news page states and refresh"
```

### Task 4: 赛程与战卡页面增强

**Files:**
- Modify: `miniapp/pages/schedule/index.js`
- Create: `miniapp/pages/schedule/index.wxml`
- Create: `miniapp/pages/schedule/index.wxss`
- Modify: `miniapp/pages/event-detail/index.js`
- Create: `miniapp/pages/event-detail/index.wxml`
- Create: `miniapp/pages/event-detail/index.wxss`
- Test: `miniapp/tests/schedule-event-page.spec.js`
- Modify: `miniapp/tests/navigation.spec.js`

**Step 1: Write the failing test**

```js
test('schedule page filters by org', async () => {
  const ctx = createPageContext(schedulePage)
  schedulePage.__setApi({ listEvents: jest.fn().mockResolvedValue({ items: [
    { id: 1, org: 'UFC', name: 'A' },
    { id: 2, org: 'ONE', name: 'B' },
  ]})})

  await schedulePage.loadEvents.call(ctx)
  schedulePage.onOrgChange.call(ctx, { detail: { value: 'UFC' } })

  expect(ctx.data.filteredItems).toHaveLength(1)
})
```

**Step 2: Run test to verify it fails**

Run: `cd miniapp && npm test -- schedule-event-page.spec.js`
Expected: FAIL with missing filter logic

**Step 3: Write minimal implementation**

```js
function applyOrgFilter(items, org) {
  if (!org || org === 'ALL') return items
  return items.filter((item) => item.org === org)
}
```

**Step 4: Run test to verify it passes**

Run: `cd miniapp && npm test -- schedule-event-page.spec.js`
Expected: PASS

**Step 5: Commit**

```bash
git add miniapp/pages/schedule miniapp/pages/event-detail miniapp/tests/schedule-event-page.spec.js miniapp/tests/navigation.spec.js
git commit -m "feat: enhance schedule and event card pages with states"
```

### Task 5: 选手详情与选手搜索增强

**Files:**
- Modify: `miniapp/pages/fighter/index.js`
- Create: `miniapp/pages/fighter/index.wxml`
- Create: `miniapp/pages/fighter/index.wxss`
- Modify: `miniapp/pages/search-fighter/index.js`
- Create: `miniapp/pages/search-fighter/index.wxml`
- Create: `miniapp/pages/search-fighter/index.wxss`
- Test: `miniapp/tests/fighter-pages.spec.js`

**Step 1: Write the failing test**

```js
test('search page navigates to fighter detail on select', () => {
  const ctx = createPageContext(searchPage)
  global.wx = { navigateTo: jest.fn() }

  searchPage.onSelectFighter.call(ctx, { currentTarget: { dataset: { id: 20 } } })

  expect(global.wx.navigateTo).toHaveBeenCalledWith({ url: '/pages/fighter/index?id=20' })
})
```

**Step 2: Run test to verify it fails**

Run: `cd miniapp && npm test -- fighter-pages.spec.js`
Expected: FAIL with handler mismatch

**Step 3: Write minimal implementation**

```js
function onSelectFighter(e) {
  const fighterId = e.currentTarget.dataset.id
  wx.navigateTo({ url: `/pages/fighter/index?id=${fighterId}` })
}
```

**Step 4: Run test to verify it passes**

Run: `cd miniapp && npm test -- fighter-pages.spec.js`
Expected: PASS

**Step 5: Commit**

```bash
git add miniapp/pages/fighter miniapp/pages/search-fighter miniapp/tests/fighter-pages.spec.js
git commit -m "feat: enhance fighter detail and search pages"
```

### Task 6: 回归与验收

**Files:**
- Modify: `miniapp/tests/support/miniapp-harness.js`
- Modify: `docs/release/mvp-checklist.md`

**Step 1: Write the failing regression test**

```js
test('news schedule and search routes are all reachable', async () => {
  const app = launchMiniApp()
  await app.open('/pages/news/index')
  await app.open('/pages/schedule/index')
  await app.open('/pages/search-fighter/index')
  expect(app.currentPage()).toBe('/pages/search-fighter/index')
})
```

**Step 2: Run test to verify it fails**

Run: `cd miniapp && npm test -- navigation.spec.js`
Expected: FAIL with unsupported route in harness

**Step 3: Write minimal implementation**

```js
async open(path) {
  current = path
}
```

**Step 4: Run full miniapp tests**

Run: `cd miniapp && npm test`
Expected: PASS

**Step 5: Run project verification**

Run: `cd /Users/weizhenguo/ai_coding_projects/w-mma && make test-e2e`
Expected: PASS

**Step 6: Commit**

```bash
git add miniapp/tests docs/release/mvp-checklist.md
git commit -m "test: complete miniapp enhanced verification"
```

## Execution Notes
- 全程遵循 `@test-driven-development`：先写失败测试，再写最小实现。
- 失败时遵循 `@systematic-debugging` 先复现再修复。
- 宣称完成前遵循 `@verification-before-completion`，必须给出命令输出证据。
