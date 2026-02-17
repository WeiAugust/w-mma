# GETTING START

本文档用于本地部署、联调和验收。

## 1. 环境要求
- Docker Desktop（推荐直接用 `docker compose`）
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

## 3. 真实数据抓取与发布联调
触发抓取：

```bash
curl -X POST http://localhost:8080/admin/ingest/fetch \
  -H 'Content-Type: application/json' \
  -d '{"source_id":1,"url":"https://www.ufc.com"}'
```

查看待审核：

```bash
curl http://localhost:8080/admin/review/pending
```

审核通过（示例 pending id=1）：

```bash
curl -X POST "http://localhost:8080/admin/review/1/approve?reviewer_id=9001"
```

查看发布资讯：

```bash
curl http://localhost:8080/api/articles
```

## 4. 前端本地测试
后台：

```bash
cd admin
pnpm install
pnpm vitest run e2e/review_publish.spec.ts
```

小程序：

```bash
cd miniapp
npm install
npm test -- navigation.spec.js
```

## 5. 微信开发者工具启动小程序
1. 打开微信开发者工具，选择“导入项目”
2. 项目目录选择仓库下 `miniapp/`
3. AppID 使用你自己的小程序 AppID（测试可用测试号）
4. 打开后点击“编译”

说明：
- 小程序真机/预览请求必须使用“request 合法域名”
- `http://localhost:8080` 不属于合法域名，不能直接用于线上环境请求
- 本地调试可在开发者工具中按需放开域名校验，或通过可访问的 HTTPS 网关转发到本地服务

## 6. 一键验证
在仓库根目录执行：

```bash
make test-e2e
```

## 7. 停止服务
```bash
docker compose down
```
