# 八角志（w-mma）

面向格斗资讯与赛事数据的小程序 + 后台 + 后端一体化项目。

## 功能概览
- 资讯抓取入队（Redis Stream）与审核发布
- 后台账号密码登录 + JWT 鉴权（`/admin/*`）
- 数据源管理（资讯/赛程/选手，含展示/播放/AI 摘要授权位）
- 资讯手动录入、可选 AI 总结任务（无 key 自动降级人工）
- 合规投诉与一键下架（下架后公开接口不可见）
- 赛事列表、战卡详情、选手搜索与详情
- live 赛果 30 秒更新（幂等写入）
- MySQL 持久化（资讯/审核/赛事/战卡/选手）
- 小程序读接口 Redis 缓存加速（Cache-Aside）
- 启动自动迁移 + `schema_migrations` 版本记录（支持重复启动与并发启动）

## 技术栈
- 后端：Go + Gin + GORM + MySQL + Redis
- 后台：admin（Vue 3 + Vite + Element Plus + Vitest）
- 小程序：原生微信小程序 + Jest

## 目录结构
- `backend/` 后端 API、Worker、测试
- `admin/` 管理后台
- `miniapp/` 微信小程序
- `ops/` 环境变量样例

## 快速开始
```bash
docker compose up -d --build
```

服务默认地址：
- API: `http://localhost:8080`
- Admin Web: `http://localhost:5173`
- MySQL: `localhost:3306`
- Redis: `localhost:6379`

详细启动、联调、微信开发者工具接入步骤见：`GETTING_START.md`

## 常用命令
```bash
make test
make test-e2e
cd admin && pnpm dev --host 0.0.0.0 --port 5173
```

## 核心联调接口
```bash
# 后台登录（默认账号：admin / admin123456）
curl -X POST http://localhost:8080/admin/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123456"}'

# 触发真实 URL 抓取
curl -X POST http://localhost:8080/admin/ingest/fetch \
  -H "Authorization: Bearer <ADMIN_JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"source_id":1,"url":"https://www.ufc.com"}'

# 查看待审核
curl -H "Authorization: Bearer <ADMIN_JWT>" http://localhost:8080/admin/review/pending

# 审核通过（示例 id=1）
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" "http://localhost:8080/admin/review/1/approve?reviewer_id=9001"

# 创建下架工单并执行下架（示例：文章 id=1）
curl -X POST http://localhost:8080/admin/takedowns \
  -H "Authorization: Bearer <ADMIN_JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"target_type":"article","target_id":1,"reason":"copyright complaint"}'
curl -X POST http://localhost:8080/admin/takedowns/1/resolve \
  -H "Authorization: Bearer <ADMIN_JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"action":"offlined"}'

# 小程序资讯接口
curl http://localhost:8080/api/articles
```
