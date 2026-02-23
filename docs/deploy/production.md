# Production Deployment

当前部署策略：

- 云服务器：`caddy`、`api`、`mysql`、`redis`
- 本地机器：`worker`、`admin`

完整上线 Runbook（含执行顺序、验收、回滚）见：

- `docs/deploy/production-rollout-plan.md`

快速启动（云侧）：

```bash
cp .env.prod.example .env.prod
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build mysql redis api caddy
curl https://<API_DOMAIN>/healthz
```
