<template>
  <section class="review-console">
    <header class="page-head">
      <div>
        <h1>审核工作台</h1>
        <p class="hint">快速过滤待审内容，逐条通过并追踪操作结果。</p>
      </div>
      <button class="ghost" type="button" @click="loadItems">刷新列表</button>
    </header>

    <div class="kpi-grid">
      <article class="kpi-card">
        <p class="kpi-label">待审核总数</p>
        <p class="kpi-value">{{ items.length }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">筛选后数量</p>
        <p class="kpi-value">{{ filtered.length }}</p>
      </article>
    </div>

    <div class="filter-bar">
      <label>
        标题关键词
        <input v-model.trim="draftKeyword" data-test="keyword" placeholder="输入标题关键词" />
      </label>
      <button class="primary" data-test="apply-filter" type="button" @click="activeKeyword = draftKeyword">
        应用筛选
      </button>
    </div>

    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>标题</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filtered" :key="item.id">
            <td>#{{ item.id }}</td>
            <td>{{ item.title }}</td>
            <td>
              <button class="primary" :data-test="`approve-${item.id}`" @click="onApprove(item.id)">通过</button>
            </td>
          </tr>
          <tr v-if="filtered.length === 0">
            <td class="empty" colspan="3">当前没有待审核内容。</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { approvePending, listPending, type PendingItem } from '../../api/review'

const items = ref<PendingItem[]>([])
const error = ref('')
const success = ref('')
const draftKeyword = ref('')
const activeKeyword = ref('')

const filtered = computed(() =>
  items.value.filter((item) => {
    if (!activeKeyword.value) return true
    return item.title.toLowerCase().includes(activeKeyword.value.toLowerCase())
  }),
)

onMounted(loadItems)

async function onApprove(id: number) {
  error.value = ''
  success.value = ''
  try {
    await approvePending(id)
    items.value = items.value.filter((item) => item.id !== id)
    success.value = `已通过待审核内容 #${id}`
  } catch (err) {
    error.value = (err as Error).message || '审核失败'
  }
}

async function loadItems() {
  error.value = ''
  success.value = ''
  try {
    items.value = await listPending()
  } catch (err) {
    error.value = (err as Error).message || '加载失败'
  }
}
</script>

<style scoped>
.review-console {
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
  grid-template-columns: repeat(2, minmax(0, 1fr));
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
.filter-bar {
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
  padding: 12px;
  background: rgba(7, 17, 31, 0.72);
  display: grid;
  gap: 10px;
  grid-template-columns: 1fr auto;
}
.table-wrap {
  overflow: auto;
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
}
table {
  width: 100%;
  border-collapse: collapse;
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
.empty {
  text-align: center;
  color: var(--text-muted);
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
  .filter-bar {
    grid-template-columns: 1fr;
  }
}
</style>
