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

启动后台 Web 页面：

```bash
cd admin
pnpm dev --host 0.0.0.0 --port 5173
```

浏览器访问：`http://localhost:5173`

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

## 10. UFC 赛程与战卡联调（主赛/副赛）
在仓库根目录执行：

```bash
./ops/verify_ufc_event_card_flow.sh
```

该脚本会自动：
- 登录后台并获取 JWT
- 定位 UFC 赛程 source 并触发同步
- 读取公开赛事接口并抽取 UFC 赛事
- 自动选择有对阵数据的 UFC 赛事
- 校验 `/api/events/:id` 中 `main_card/prelims` 与选手字段（国家/姓名/排名/量级/头像）是否完整
- 校验赛程条目使用开赛时间（非抓取时间），并且状态稳定映射为 `scheduled/completed`
- 校验海报/选手头像均为本地镜像地址（`/media-cache/ufc/*`），避免小程序直连三方图片 403
- 校验已完赛对阵包含 `result/method/round/time_sec` 赛果字段

若输出 `NO_BOUT_DATA`，通常表示当前网络环境访问 `ufc.com` 被重定向到不可用站点（如 `ufc.cn` 404/403），导致无法抓取对阵详情。此时需在可访问 `ufc.com` 赛事页的网络环境下重试。

可选环境变量：

```bash
API_BASE_URL=http://localhost:8080 \
ADMIN_USERNAME=admin \
ADMIN_PASSWORD=admin123456 \
./ops/verify_ufc_event_card_flow.sh
```

若你本机通过本地代理可访问 `ufc.com`（例如 Clash）：

```bash
LOCAL_PROXY_URL=http://127.0.0.1:7890 ./ops/start_local_api_with_proxy.sh
./ops/verify_ufc_event_card_flow.sh
./ops/stop_local_api.sh
```

`start_local_api_with_proxy.sh` 会：
- 仅保留 MySQL/Redis 容器运行
- 在宿主机启动 API，并注入 `HTTP_PROXY/HTTPS_PROXY/NO_PROXY`
- 使用宿主机端口 `localhost:23306`（MySQL）和 `localhost:26379`（Redis）
- 同步 UFC 数据时将海报/头像镜像到 `MEDIA_CACHE_DIR`（默认 `.worktrees/media-cache`），并通过 `http://localhost:8080/media-cache/ufc/*` 对外提供

## 11. 停止服务
```bash
docker compose down
```
