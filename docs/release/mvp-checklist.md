# 八角志 MVP Release Checklist

## 功能验收
- [x] 小程序赛程页可打开赛事详情页
- [x] 赛事详情页可点击选手进入详情页
- [x] 小程序支持按名称搜索选手
- [x] 小程序选手搜索支持英文名/中文名/外号（name/name_zh/nickname）
- [x] 后台可查看待审核列表并执行通过
- [x] 后台可维护赛事状态（scheduled/live/completed）
- [x] 后端提供赛事、战卡、选手搜索、审核发布 API
- [x] live 赛果更新具备幂等写入逻辑
- [x] 资讯/审核/赛事/战卡/选手数据落 MySQL 持久化
- [x] 抓取任务通过 Redis Stream 持久化入队
- [x] 小程序读取接口支持 Redis 缓存加速
- [x] 后台登录鉴权（账号密码 + JWT）
- [x] 后台可管理数据源（news/schedule/fighter）
- [x] 资讯支持手动录入与可选 AI 总结任务（无 key 降级人工）
- [x] 合规投诉与下架流程可用（offlined 后公开接口隐藏）
- [x] 小程序资讯/赛程/赛事/选手页面支持图片或视频展示
- [x] 小程序赛程页展示赛事海报，状态中文化并按状态柔和配色区分
- [x] 小程序赛事详情按主赛/副赛中文分组展示并包含选手国家/姓名/排名/量级/照片（缺失字段占位）
- [x] 小程序赛事详情状态中文化，量级中文化，赛果展示（胜者/方法/回合/时间）可用
- [x] 提供 UFC 同步联调脚本 `./ops/verify_ufc_event_card_flow.sh`，可自动验证 `main_card/prelims` 数据契约
- [x] UFC 同步链路打通：`/admin/sources/:id/sync` 可稳定写入 `bouts`，联调结果可达到 `events>0 && bouts>0`
- [x] UFC 选手目录同步链路打通：`parser_kind=ufc_athletes` 可写入 fighters 的 `stats/records` 扩展字段
- [x] 小程序选手详情页可展示 `stats/records` 分区数据

## 真实数据接入验收
- [x] 后端提供 `POST /admin/ingest/fetch`，可抓取真实 URL 标题并入待审核池
- [x] 审核通过后，`GET /api/articles` 可返回发布内容

## 自动化验证
- [x] `cd backend && go test ./...`
- [x] `cd admin && pnpm vitest run src/pages/review/ReviewQueue.spec.ts`
- [x] `cd miniapp && npm test -- navigation.spec.js`
- [x] `cd backend && GOCACHE=/tmp/go-build go test ./internal/ufc ./internal/event -v`
- [x] `cd backend && go test ./internal/fighter ./internal/ufc ./internal/storage`
- [x] `cd miniapp && npm test -- --runTestsByPath tests/fighter-pages.spec.js`
- [x] `cd miniapp && npm test`
- [x] `make test-e2e`
- [x] `cd backend && go test ./tests/e2e -run TestE2E_PersistedArticleSurvivesRepositoryRecreate -v`
- [x] `cd backend && go test ./tests/e2e -run TestE2E_TakedownOfflinesArticleAndPublicAPIHidesIt -v`

## 本地验收步骤
1. 一键启动：`docker compose up -d --build`
2. 后台登录（默认 `admin/admin123456`），保存 JWT token
3. 触发真实抓取：
   `curl -X POST http://localhost:8080/admin/ingest/fetch -H "Authorization: Bearer <ADMIN_JWT>" -H 'Content-Type: application/json' -d '{"source_id":1,"url":"https://www.ufc.com"}'`
4. 查看待审核：`curl -H "Authorization: Bearer <ADMIN_JWT>" http://localhost:8080/admin/review/pending`
5. 审核通过：`curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" "http://localhost:8080/admin/review/1/approve?reviewer_id=9001"`
6. 创建下架工单并处理：
   `curl -X POST http://localhost:8080/admin/takedowns -H "Authorization: Bearer <ADMIN_JWT>" -H 'Content-Type: application/json' -d '{"target_type":"article","target_id":1,"reason":"copyright complaint"}'`
   `curl -X POST http://localhost:8080/admin/takedowns/1/resolve -H "Authorization: Bearer <ADMIN_JWT>" -H 'Content-Type: application/json' -d '{"action":"offlined"}'`
7. 下架后公开接口检查：`curl http://localhost:8080/api/articles`
8. 重启 API 容器后再次检查：`docker compose restart api && curl http://localhost:8080/api/articles`（验证持久化）
