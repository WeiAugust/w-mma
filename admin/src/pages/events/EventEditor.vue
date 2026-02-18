<template>
  <section class="event-console">
    <header class="page-head">
      <div>
        <h1>赛程控制台</h1>
        <p class="hint">统一维护赛事名称与状态，确保小程序赛程展示一致。</p>
      </div>
      <button class="ghost" type="button" @click="loadEvents">刷新赛事</button>
    </header>

    <div class="kpi-grid">
      <article class="kpi-card">
        <p class="kpi-label">赛事总数</p>
        <p class="kpi-value">{{ events.length }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">Live 场次</p>
        <p class="kpi-value">{{ liveCount }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">筛选后</p>
        <p class="kpi-value">{{ filtered.length }}</p>
      </article>
    </div>

    <div class="filters">
      <label>
        赛事组织
        <input v-model.trim="draftOrg" data-test="filter-org" placeholder="如 UFC / ONE / PFL" />
      </label>
      <label>
        状态
        <select v-model="statusFilter">
          <option value="all">全部</option>
          <option value="scheduled">scheduled</option>
          <option value="live">live</option>
          <option value="completed">completed</option>
        </select>
      </label>
      <button class="primary" data-test="apply-filter" type="button" @click="applyFilter">应用筛选</button>
    </div>

    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>组织</th>
            <th>赛事名</th>
            <th>状态</th>
            <th>开始时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="event in filtered" :key="event.id">
            <td>#{{ event.id }}</td>
            <td>{{ event.org }}</td>
            <td>{{ event.name }}</td>
            <td><span class="tag">{{ event.status }}</span></td>
            <td>{{ event.starts_at }}</td>
            <td>
              <button class="primary" :data-test="`edit-${event.id}`" @click="openEdit(event)">编辑</button>
            </td>
          </tr>
          <tr v-if="filtered.length === 0">
            <td class="empty" colspan="6">暂无赛事数据。</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="editingId !== null" class="drawer-mask">
      <section class="drawer">
        <header>
          <h2>编辑赛事</h2>
          <button class="ghost" type="button" @click="editingId = null">关闭</button>
        </header>
        <div class="form-grid">
          <label>
            赛事名称
            <input v-model="editName" data-test="edit-name" />
          </label>
          <label>
            赛事状态
            <select v-model="editStatus" data-test="edit-status">
              <option value="scheduled">scheduled</option>
              <option value="live">live</option>
              <option value="completed">completed</option>
            </select>
          </label>
        </div>
        <footer>
          <button class="primary" data-test="save-edit" type="button" @click="saveEdit">保存修改</button>
        </footer>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { listEvents, updateEvent, type EventItem } from '../../api/events'

const events = ref<EventItem[]>([])
const error = ref('')
const success = ref('')
const draftOrg = ref('')
const activeOrg = ref('')
const statusFilter = ref<'all' | EventItem['status']>('all')
const editingId = ref<number | null>(null)
const editName = ref('')
const editStatus = ref<EventItem['status']>('scheduled')

const liveCount = computed(() => events.value.filter((event) => event.status === 'live').length)

const filtered = computed(() =>
  events.value.filter((event) => {
    if (activeOrg.value && !event.org.toLowerCase().includes(activeOrg.value.toLowerCase())) return false
    if (statusFilter.value !== 'all' && event.status !== statusFilter.value) return false
    return true
  }),
)

onMounted(loadEvents)

function applyFilter() {
  activeOrg.value = draftOrg.value
}

function openEdit(event: EventItem) {
  editingId.value = event.id
  editName.value = event.name
  editStatus.value = event.status
}

async function saveEdit() {
  if (editingId.value === null) return
  error.value = ''
  success.value = ''
  try {
    await updateEvent(editingId.value, { name: editName.value, status: editStatus.value })
    success.value = `已更新赛事 #${editingId.value}`
    editingId.value = null
    await loadEvents()
  } catch (err) {
    error.value = (err as Error).message || '保存失败'
  }
}

async function loadEvents() {
  error.value = ''
  success.value = ''
  try {
    events.value = await listEvents()
  } catch (err) {
    error.value = (err as Error).message || '加载赛事失败'
  }
}
</script>

<style scoped>
.event-console {
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
  padding: 12px;
  background: rgba(7, 17, 31, 0.72);
  display: grid;
  gap: 10px;
  grid-template-columns: 1fr 1fr auto;
}
.table-wrap {
  overflow: auto;
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
}
table {
  width: 100%;
  border-collapse: collapse;
  min-width: 860px;
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
.tag {
  display: inline-block;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 12px;
  background: rgba(90, 147, 255, 0.2);
  color: #a6cbff;
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
  width: min(640px, 100%);
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
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
