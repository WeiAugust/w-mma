<template>
  <section class="article-console">
    <header class="page-head">
      <div>
        <h1>资讯管理中心</h1>
        <p class="hint">支持手动录入并预览公开资讯效果，快速检查视频可播权限。</p>
      </div>
      <button class="primary" data-test="open-create" type="button" @click="showCreate = true">新建资讯</button>
    </header>

    <div class="kpi-grid">
      <article class="kpi-card">
        <p class="kpi-label">公开资讯总数</p>
        <p class="kpi-value">{{ published.length }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">可播视频</p>
        <p class="kpi-value">{{ playableCount }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">仅来源跳转</p>
        <p class="kpi-value">{{ published.length - playableCount }}</p>
      </article>
    </div>

    <div class="filters">
      <label>
        关键词
        <input v-model.trim="draftKeyword" data-test="filter-keyword" placeholder="标题关键词" />
      </label>
      <label>
        视频能力
        <select v-model="playFilter">
          <option value="all">全部</option>
          <option value="playable">可播放</option>
          <option value="source_only">仅来源跳转</option>
        </select>
      </label>
      <button class="ghost" type="button" @click="loadPublished">刷新</button>
      <button class="primary" data-test="apply-filter" type="button" @click="applyFilter">应用筛选</button>
    </div>

    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>标题</th>
            <th>来源</th>
            <th>视频策略</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filtered" :key="item.id">
            <td>#{{ item.id }}</td>
            <td>{{ item.title }}</td>
            <td><a :href="item.source_url" target="_blank" rel="noreferrer">{{ item.source_url }}</a></td>
            <td>
              <span class="tag" :class="item.can_play ? 'ok' : 'warn'">
                {{ item.can_play ? '站内可播' : '仅来源跳转' }}
              </span>
            </td>
          </tr>
          <tr v-if="filtered.length === 0">
            <td class="empty" colspan="4">暂无符合条件的资讯</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showCreate" class="drawer-mask">
      <section class="drawer">
        <header>
          <h2>手动录入资讯</h2>
          <button class="ghost" type="button" @click="showCreate = false">关闭</button>
        </header>
        <div class="form-grid two">
          <label>
            source_id
            <input v-model.number="form.source_id" data-test="create-source-id" />
          </label>
          <label>
            标题
            <input v-model="form.title" data-test="create-title" />
          </label>
          <label class="full">
            摘要
            <textarea v-model="form.summary" data-test="create-summary" />
          </label>
          <label class="full">
            来源 URL
            <input v-model="form.source_url" data-test="create-source-url" />
          </label>
          <label>
            封面 URL
            <input v-model="form.cover_url" />
          </label>
          <label>
            视频 URL
            <input v-model="form.video_url" />
          </label>
        </div>
        <footer>
          <button class="primary" data-test="submit-create" type="button" @click="onCreate">提交录入</button>
        </footer>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import { createManualArticle, listPublishedArticles, type ArticleItem } from '../../api/articles'

const form = reactive({
  source_id: 0,
  title: '',
  summary: '',
  source_url: '',
  cover_url: '',
  video_url: '',
})
const published = ref<ArticleItem[]>([])
const error = ref('')
const success = ref('')
const showCreate = ref(false)
const draftKeyword = ref('')
const activeKeyword = ref('')
const playFilter = ref<'all' | 'playable' | 'source_only'>('all')

const playableCount = computed(() => published.value.filter((item) => item.can_play).length)

const filtered = computed(() =>
  published.value.filter((item) => {
    if (activeKeyword.value && !item.title.toLowerCase().includes(activeKeyword.value.toLowerCase())) {
      return false
    }
    if (playFilter.value === 'playable' && !item.can_play) {
      return false
    }
    if (playFilter.value === 'source_only' && item.can_play) {
      return false
    }
    return true
  }),
)

onMounted(loadPublished)

function applyFilter() {
  activeKeyword.value = draftKeyword.value
}

async function onCreate() {
  error.value = ''
  success.value = ''
  try {
    await createManualArticle({ ...form })
    success.value = '录入成功，已进入待审核队列'
    showCreate.value = false
    await loadPublished()
  } catch (err) {
    error.value = (err as Error).message || '录入失败'
  }
}

async function loadPublished() {
  error.value = ''
  try {
    published.value = await listPublishedArticles()
  } catch (err) {
    error.value = (err as Error).message || '加载资讯失败'
  }
}
</script>

<style scoped>
.article-console {
  display: grid;
  gap: 14px;
}
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
}
.page-head h1 {
  margin: 0;
}
.kpi-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}
.kpi-card {
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
  padding: 12px;
  background: rgba(8, 20, 36, 0.78);
}
.kpi-label {
  margin: 0;
  color: var(--text-muted);
}
.kpi-value {
  margin: 6px 0 0;
  font-size: 24px;
  font-weight: 700;
}
.filters {
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
  background: rgba(7, 17, 31, 0.72);
  padding: 12px;
  display: grid;
  gap: 10px;
  grid-template-columns: 1.2fr 0.8fr auto auto;
}
.table-wrap {
  overflow: auto;
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
}
table {
  width: 100%;
  border-collapse: collapse;
  min-width: 760px;
}
th,
td {
  padding: 10px;
  border-bottom: 1px solid rgba(100, 145, 194, 0.2);
  text-align: left;
}
thead th {
  background: rgba(9, 18, 32, 0.9);
  color: #bad0ec;
  font-size: 12px;
  text-transform: uppercase;
}
a {
  color: #9fd0ff;
}
.tag {
  display: inline-block;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 600;
}
.tag.ok {
  background: rgba(88, 230, 186, 0.2);
  color: #84ffd7;
}
.tag.warn {
  background: rgba(255, 170, 110, 0.2);
  color: #ffc992;
}
.empty {
  text-align: center;
  color: var(--text-muted);
}
.drawer-mask {
  position: fixed;
  inset: 0;
  background: rgba(3, 8, 16, 0.75);
  display: grid;
  place-items: center;
  padding: 14px;
}
.drawer {
  width: min(760px, 100%);
  border: 1px solid rgba(111, 156, 200, 0.34);
  border-radius: 14px;
  background: rgba(8, 20, 36, 0.95);
  padding: 14px;
  display: grid;
  gap: 12px;
}
.drawer header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.form-grid {
  display: grid;
  gap: 10px;
}
.form-grid.two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.form-grid .full {
  grid-column: span 2;
}
.drawer footer {
  display: flex;
  justify-content: flex-end;
}
.primary,
.ghost {
  border-radius: 10px;
  padding: 8px 12px;
  cursor: pointer;
}
.primary {
  border: 1px solid #58e6ba;
  background: linear-gradient(135deg, #56ddbe, #7ef0d4);
  color: #052019;
  font-weight: 700;
}
.ghost {
  border: 1px solid rgba(111, 156, 200, 0.5);
  background: rgba(8, 19, 34, 0.65);
  color: #d7e8ff;
}
@media (max-width: 920px) {
  .kpi-grid,
  .filters,
  .form-grid.two {
    grid-template-columns: 1fr;
  }
  .form-grid .full {
    grid-column: auto;
  }
}
</style>
