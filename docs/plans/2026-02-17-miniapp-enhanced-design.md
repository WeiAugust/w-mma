# 八角志小程序增强版设计（直连后端）

## 1. 目标
在现有 MVP 链路基础上，将小程序补齐到可本地验收的增强版：
- 资讯浏览（直连后端 `/api/articles`）
- 赛程浏览（`/api/events`）
- 战卡详情（`/api/events/:id`）
- 选手详情（`/api/fighters/:id`）
- 选手搜索（`/api/fighters/search`）
- 页面级加载态/空态/错误态、下拉刷新、重试

## 2. 信息架构与路由
- 页面清单：
  - `pages/news/index`
  - `pages/schedule/index`
  - `pages/event-detail/index`
  - `pages/fighter/index`
  - `pages/search-fighter/index`
- 主链路：
  1. 赛程 -> 赛事详情 -> 选手详情
  2. 选手搜索 -> 选手详情
  3. 首页资讯独立浏览

## 3. 数据与状态设计
统一状态模型：
- `loading: boolean`
- `error: string`
- `items/detail`
- `isEmpty: boolean`（由数据长度推导）
- `updatedAtText: string`

状态流：
1. 进入页面置 `loading=true,error=''`
2. API 成功：更新数据与 `updatedAtText`
3. API 失败：设置 `error` 并展示重试
4. 下拉刷新：复用同一加载函数

## 4. 接口与请求封装
- `services/api.js` 提供：
  - `listArticles`
  - `listEvents`
  - `getEventCard`
  - `searchFighters`
  - `getFighterDetail`
- 请求层能力：
  - 统一 base URL（默认 `http://localhost:8080`）
  - 超时控制
  - 统一错误对象 `{ message }`

## 5. 页面交互细节
- 资讯页：列表 + 来源链接提示 + 下拉刷新 + 错误重试
- 赛程页：组织筛选（全部/UFC/ONE/PFL/Bellator）+ 列表点击跳详情
- 战卡页：赛事信息 + bout 列表，点选手跳详情
- 选手页：基础资料 + 最近动态 + 失败重试
- 搜索页：关键词输入、显式触发搜索、空结果提示、点击结果跳详情

## 6. 测试策略
在保留现有导航测试基础上新增：
- API 封装测试（路径与查询参数）
- 资讯页逻辑测试（成功/失败状态）
- 搜索页逻辑测试（搜索与跳转）

## 7. 非目标
本轮不做：
- 离线 mock 自动降级
- 用户系统/收藏/推送
- 拼音与别名检索
