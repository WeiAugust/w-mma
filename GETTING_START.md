# GETTING START

本文档用于本地部署、联调和验收。

## 1. 环境要求
- Docker Desktop（推荐直接使用 `docker compose`）
- Node.js 18+
- pnpm（用于 `admin`）
- 微信开发者工具（用于 `miniapp`）

## 2. 一键启动后端依赖与服务
在仓库根目录执行：

```bash
docker compose up -d --build
```

检查状态：

```bash
docker compose ps
curl http://localhost:8080/healthz
```

健康返回应为：

```json
{"status":"ok"}
```

## 3. 数据持久化与迁移说明
- MySQL 数据持久化在 Docker volume：`mysql_data`
- Redis 数据持久化在 Docker volume：`redis_data`
- 服务启动时会自动执行迁移，并写入 `schema_migrations`（避免重复执行导致启动失败）
- 可通过重启验证幂等性：

```bash
docker compose restart api worker
docker compose ps
```

## 4. 后台登录（admin API 需要 JWT）
默认开发账号：
- 用户名：`admin`
- 密码：`admin123456`

```bash
curl -X POST http://localhost:8080/admin/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123456"}'
```

响应中的 `token` 用于后续 `Authorization: Bearer <token>`。

## 5. 配置数据源并触发真实抓取
先创建数据源（示例：赛程源）：

```bash
curl -X POST http://localhost:8080/admin/sources \
  -H "Authorization: Bearer <ADMIN_JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"name":"UFC Official","source_type":"schedule","platform":"official","source_url":"https://www.ufc.com/events","parser_kind":"generic","enabled":true,"rights_display":true,"rights_playback":false,"rights_ai_summary":true}'
```

触发抓取：

```bash
curl -X POST http://localhost:8080/admin/ingest/fetch \
  -H "Authorization: Bearer <ADMIN_JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"source_id":1,"url":"https://www.ufc.com"}'
```

查看待审核：

```bash
curl -H "Authorization: Bearer <ADMIN_JWT>" http://localhost:8080/admin/review/pending
```

审核通过（示例 pending id=1）：

```bash
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" "http://localhost:8080/admin/review/1/approve?reviewer_id=9001"
```

查看发布资讯：

```bash
curl http://localhost:8080/api/articles
```

## 6. 合规下架联调（示例）
创建工单并下架文章：

```bash
curl -X POST http://localhost:8080/admin/takedowns \
  -H "Authorization: Bearer <ADMIN_JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"target_type":"article","target_id":1,"reason":"copyright complaint"}'

curl -X POST http://localhost:8080/admin/takedowns/1/resolve \
  -H "Authorization: Bearer <ADMIN_JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"action":"offlined"}'
```

再次检查：

```bash
curl http://localhost:8080/api/articles
```

## 7. 前端本地测试
后台：

```bash
cd admin
pnpm install
pnpm test
```

小程序：

```bash
cd miniapp
npm install
npm test -- navigation.spec.js
```

## 8. 微信开发者工具启动小程序
1. 打开微信开发者工具，选择“导入项目”
2. 项目目录选择仓库下 `miniapp/`
3. AppID 使用你自己的小程序 AppID（测试可用测试号）
4. 打开后点击“编译”

说明：
- 小程序真机/预览请求必须使用“request 合法域名”
- `http://localhost:8080` 不在合法域名中，默认会报错
- 本地调试可在开发者工具中临时勾选“不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书”
- 若后台已更新域名白名单，需在微信开发者工具执行“详情 -> 域名信息 -> 刷新项目配置”，然后重新编译

## 9. 一键验收
在仓库根目录执行：

```bash
make test-e2e
```

## 10. 停止服务
```bash
docker compose down
```
