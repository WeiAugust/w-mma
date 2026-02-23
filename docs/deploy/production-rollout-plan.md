# 生产上线计划（云：caddy/api/mysql/redis，本地：worker/admin）

更新时间：2026-02-23

## 1. 范围与目标

### 1.1 本次上线范围（In Scope）

- 云服务器部署：`caddy`、`api`、`mysql`、`redis`
- 本地部署：`worker`、`admin`
- 小程序请求域名切换到生产 API 域名

### 1.2 非范围（Out of Scope）

- `worker` 不部署到云服务器
- `admin` 不部署到云服务器
- 其他运营系统上线不在本计划内

### 1.3 上线成功标准

- `https://<API_DOMAIN>/healthz` 返回 `{"status":"ok"}`
- 小程序在真机可正常访问赛程/赛事详情/选手
- 本地 `worker` 可持续同步 UFC 数据到云 MySQL

## 2. 目标架构

```
微信小程序 -> HTTPS -> Caddy(云) -> API(云) -> MySQL/Redis(云)
                           ^
                           |
                    Admin(本地浏览器)

Worker(本地) -> UFC官网
Worker(本地) -> MySQL/Redis(云)
```

## 3. 依赖与网络要求

### 3.1 云侧端口

- 对公网开放：`80/443`（给 Caddy）
- 不建议公网开放：`3306/6379`

### 3.2 本地 Worker 连云 MySQL/Redis

二选一：

1. 直连模式（简单，但风险更高）
   - 云防火墙只放行你的本地公网 IP 到 `3306/6379`
2. 隧道模式（推荐）
   - 使用 VPN/SSH Tunnel，让本地 Worker 访问云内网 MySQL/Redis

## 4. 上线前准备（T-7 ~ T-1）

### 4.1 云服务器准备

1. 安装 Docker + Docker Compose
2. 域名 `A` 记录指向云服务器公网 IP
3. 仓库部署到服务器（例如 `/srv/w-mma`）

### 4.2 云侧配置

```bash
cd /srv/w-mma
cp .env.prod.example .env.prod
```

`.env.prod` 必填：

- `API_DOMAIN`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_PASSWORD`
- `MYSQL_DSN`（云侧 `api` 连接云侧 `mysql`）
- `REDIS_ADDR=redis:6379`
- `PUBLIC_BASE_URL=https://<API_DOMAIN>`
- `ADMIN_USERNAME`
- `ADMIN_PASSWORD_HASH`
- `ADMIN_JWT_SECRET`

生成管理员密码哈希：

```bash
HASH=$(docker run --rm httpd:2.4-alpine \
  htpasswd -nbBC 10 "" "YourStrongPassword" | tr -d ':\n')
echo "$HASH" | sed 's/\$/$$/g'
```

### 4.3 本地 Worker 配置

本地运行 `worker` 时，环境变量必须改为“连云”：

- `MYSQL_DSN` -> 指向云 MySQL 地址
- `REDIS_ADDR` -> 指向云 Redis 地址
- `PUBLIC_BASE_URL=https://<API_DOMAIN>`
- `MEDIA_CACHE_DIR` -> 本地可写目录

本地 `admin` 仅用于运营操作，指向云 API：

- `VITE_API_BASE_URL=https://<API_DOMAIN>`

### 4.4 小程序配置

1. 修改 `miniapp/config/runtime.js` 的 `PROD_API_BASE_URL` 为 `https://<API_DOMAIN>`
2. 微信公众平台配置合法域名：
   - `request`：`https://<API_DOMAIN>`
   - `download`：`https://<API_DOMAIN>`

## 5. 上线执行（T 日）

### 5.1 代码预检

```bash
cd /path/to/w-mma
cd backend && go test ./internal/ufc ./internal/live
cd ../miniapp && npm ci && npm test
```

### 5.2 云侧启动（仅 4 个服务）

```bash
cd /srv/w-mma
git fetch --all --tags
git checkout <release-tag>
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build mysql redis api caddy
```

校验云侧：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml ps
docker compose --env-file .env.prod -f docker-compose.prod.yml logs --tail=200 api
curl https://<API_DOMAIN>/healthz
curl https://<API_DOMAIN>/api/events
```

### 5.3 本地启动 Worker

示例（本地 shell）：

```bash
cd /path/to/w-mma/backend
export MYSQL_DSN='app:***@tcp(<CLOUD_MYSQL_HOST>:3306)/bajiaozhi?parseTime=true&multiStatements=true'
export REDIS_ADDR='<CLOUD_REDIS_HOST>:6379'
export REDIS_PASSWORD=''
export REDIS_DB='0'
export PUBLIC_BASE_URL='https://<API_DOMAIN>'
export MEDIA_CACHE_DIR='/tmp/w-mma-media-cache'
go run ./cmd/worker
```

校验本地 Worker：

- 启动后无 panic
- 能持续执行同步
- 云侧 `api/events` 数据在刷新

### 5.4 本地启动 Admin（可选）

```bash
cd /path/to/w-mma/admin
export VITE_API_BASE_URL='https://<API_DOMAIN>'
npm ci
npm run dev -- --host 0.0.0.0 --port 5173
```

说明：`admin` 只在本地使用，不上云。

### 5.5 小程序发版

1. 上传体验版
2. 真机验收（见第 7 节）
3. 提交审核并发布

## 6. 上线后巡检（T+0 ~ T+7）

### 6.1 云侧巡检

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml ps
docker compose --env-file .env.prod -f docker-compose.prod.yml logs --tail=200 api
curl -s https://<API_DOMAIN>/api/events | jq '.items | length'
```

### 6.2 本地 Worker 巡检

- 进程是否持续在线
- 日志是否出现大量抓取错误
- 赛事状态是否正常推进（未开赛 -> 进行中 -> 已结束）

## 7. 验收清单（小程序真机）

1. 赛程页可打开，且显示“北京时间”
2. 赛事详情胜负背景与赛果文案正常
3. 选手搜索与详情正常
4. 图片（海报/头像）加载正常
5. `UFC 323` 等已结束赛事状态正确

## 8. 回滚方案

### 8.1 回滚触发条件

- `healthz` 持续异常
- 小程序核心页面不可用
- 本地 Worker 无法稳定同步

### 8.2 云侧回滚

```bash
cd /srv/w-mma
git fetch --all --tags
git checkout <previous-stable-tag>
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build mysql redis api caddy
```

### 8.3 本地 Worker 回滚

- 切回上一个稳定 tag
- 按第 5.3 节环境变量重新启动

### 8.4 数据保护

禁止执行 `docker compose down -v`，避免删库删缓存卷。

## 9. 职责分工（建议）

- 云侧发布负责人：管理 `caddy/api/mysql/redis`
- 本地服务负责人：管理 `worker/admin`
- 验收负责人：小程序真机验收
- 兜底负责人：触发并执行回滚

## 10. 关键文件

- 生产编排：`docker-compose.prod.yml`
- 云侧配置模板：`.env.prod.example`
- HTTPS 配置：`ops/prod/Caddyfile`
- 小程序生产 API 配置：`miniapp/config/runtime.js`
