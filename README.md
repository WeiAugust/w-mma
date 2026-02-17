# 八角志（w-mma）

面向格斗资讯与赛事数据的小程序 + 后台 + 后端一体化项目。

## 功能概览
- 资讯抓取入队（Redis Stream）与审核发布
- 赛事列表、战卡详情、选手搜索与详情
- live 赛果 30 秒更新（幂等写入）
- MySQL 持久化（资讯/审核/赛事/战卡/选手）
- 小程序读接口 Redis 缓存加速（Cache-Aside）

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
- MySQL: `localhost:3306`
- Redis: `localhost:6379`

详细启动、联调、微信开发者工具接入步骤见：`GETTING_start.md`

## 常用命令
```bash
make test
make test-e2e
```

## 核心联调接口
```bash
# 触发真实 URL 抓取
curl -X POST http://localhost:8080/admin/ingest/fetch \
  -H 'Content-Type: application/json' \
  -d '{"source_id":1,"url":"https://www.ufc.com"}'

# 查看待审核
curl http://localhost:8080/admin/review/pending

# 审核通过（示例 id=1）
curl -X POST "http://localhost:8080/admin/review/1/approve?reviewer_id=9001"

# 小程序资讯接口
curl http://localhost:8080/api/articles
```
