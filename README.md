# 八角志 MVP

八角志是面向格斗资讯与赛事数据的 MVP，包含：
- Go 后端 API + Worker 模块
- Vue 3 后台管理端
- 微信小程序页面脚本与联调测试

## 已实现能力
- 资讯抓取入队与解析（支持真实 URL 抓取）
- 待审核队列、审核通过发布
- 赛事列表、战卡详情、选手搜索与详情
- live 赛果更新器（30 秒轮询 + 幂等更新）
- 后台赛事管理与审核操作
- 小程序赛程 -> 战卡 -> 选手详情导航链路

## 项目结构
- `backend`: Go API、Worker、模块测试、E2E 测试
- `admin`: Vue 3 + Vitest 后台
- `miniapp`: 小程序页面脚本 + Jest 测试

## 本地运行
### 后端
```bash
cd backend
GOPROXY=https://goproxy.cn,direct GOSUMDB=off go run ./cmd/api
```

### 后台测试
```bash
cd admin
pnpm install
pnpm vitest run src/pages/review/ReviewQueue.spec.ts
```

### 小程序测试
```bash
cd miniapp
npm install
npm test -- navigation.spec.js
```

## 联动示例（真实数据）
1. 抓取来源页面并入审核池：
```bash
curl -X POST http://localhost:8080/admin/ingest/fetch \
  -H 'Content-Type: application/json' \
  -d '{"source_id":1,"url":"https://www.ufc.com"}'
```
2. 审核通过：
```bash
curl -X POST "http://localhost:8080/admin/review/1/approve?reviewer_id=9001"
```
3. 查看小程序资讯接口：
```bash
curl http://localhost:8080/api/articles
```

## 验证
```bash
make test-e2e
```
