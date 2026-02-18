<template>
  <section class="takedown-console">
    <header class="page-head">
      <div>
        <h1>合规处置中心</h1>
        <p class="hint">统一处理投诉工单，支持一键下架或驳回，并保留最近处理轨迹。</p>
      </div>
    </header>

    <div class="panel-grid">
      <article class="panel">
        <h2>创建投诉工单</h2>
        <div class="form-grid two">
          <label>
            目标类型
            <select v-model="form.target_type" data-test="target-type">
              <option value="article">article</option>
              <option value="event">event</option>
              <option value="fighter">fighter</option>
            </select>
          </label>
          <label>
            target_id
            <input v-model.number="form.target_id" data-test="target-id" />
          </label>
          <label class="full">
            投诉原因
            <textarea v-model="form.reason" data-test="reason" placeholder="版权投诉、授权过期、错误信息等" />
          </label>
          <label>
            投诉方
            <input v-model="form.complainant" />
          </label>
          <label>
            证据链接
            <input v-model="form.evidence_url" />
          </label>
        </div>
        <footer>
          <button class="primary" data-test="create" type="button" @click="onCreate">创建工单</button>
        </footer>
      </article>

      <article class="panel">
        <h2>处理工单</h2>
        <div class="form-grid">
          <label>
            工单 ID
            <input v-model.number="resolveID" data-test="resolve-id" />
          </label>
          <label>
            处置动作
            <select v-model="action" data-test="action">
              <option value="offlined">offlined</option>
              <option value="rejected">rejected</option>
            </select>
          </label>
        </div>
        <footer>
          <button class="danger" data-test="resolve" type="button" @click="onResolve">执行处置</button>
        </footer>
      </article>
    </div>

    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <article class="history">
      <h2>最近处理记录</h2>
      <ul>
        <li v-for="item in activity" :key="item.id">
          <span>{{ item.message }}</span>
          <time>{{ item.at }}</time>
        </li>
        <li v-if="activity.length === 0" class="empty">暂无处理记录</li>
      </ul>
    </article>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { createTakedown, resolveTakedown } from '../../api/takedowns'

type ActivityItem = {
  id: number
  message: string
  at: string
}

const form = ref({
  target_type: 'article',
  target_id: 0,
  reason: '',
  complainant: '',
  evidence_url: '',
})
const resolveID = ref(0)
const action = ref<'offlined' | 'rejected'>('offlined')
const error = ref('')
const success = ref('')
const activity = ref<ActivityItem[]>([])
let nextActivityID = 1

function addActivity(message: string) {
  activity.value.unshift({
    id: nextActivityID++,
    message,
    at: new Date().toLocaleString(),
  })
  activity.value = activity.value.slice(0, 8)
}

async function onCreate() {
  error.value = ''
  success.value = ''
  try {
    const payload: {
      target_type: 'article' | 'event' | 'fighter'
      target_id: number
      reason: string
      complainant?: string
      evidence_url?: string
    } = {
      target_type: form.value.target_type as 'article' | 'event' | 'fighter',
      target_id: form.value.target_id,
      reason: form.value.reason,
    }
    if (form.value.complainant) payload.complainant = form.value.complainant
    if (form.value.evidence_url) payload.evidence_url = form.value.evidence_url

    await createTakedown(payload)
    success.value = '工单创建成功'
    addActivity(`已创建 ${payload.target_type}#${payload.target_id} 投诉工单`)
  } catch (err) {
    error.value = (err as Error).message || '创建工单失败'
  }
}

async function onResolve() {
  error.value = ''
  success.value = ''
  try {
    await resolveTakedown(resolveID.value, action.value)
    success.value = `工单 #${resolveID.value} 已处理为 ${action.value}`
    addActivity(`工单 #${resolveID.value} 处理为 ${action.value}`)
  } catch (err) {
    error.value = (err as Error).message || '处理工单失败'
  }
}
</script>

<style scoped>
.takedown-console {
  display: grid;
  gap: 14px;
}
.page-head h1 {
  margin: 0;
}
.panel-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.panel {
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
  background: rgba(8, 20, 36, 0.78);
  padding: 12px;
  display: grid;
  gap: 10px;
}
.panel h2,
.history h2 {
  margin: 0;
  font-size: 18px;
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
.panel footer {
  display: flex;
  justify-content: flex-end;
}
.history {
  border: 1px solid rgba(100, 145, 194, 0.28);
  border-radius: 12px;
  background: rgba(7, 17, 31, 0.72);
  padding: 12px;
}
.history ul {
  list-style: none;
  margin: 10px 0 0;
  padding: 0;
  display: grid;
  gap: 8px;
}
.history li {
  border: 1px solid rgba(100, 145, 194, 0.2);
  border-radius: 10px;
  padding: 9px 10px;
  display: flex;
  justify-content: space-between;
  gap: 10px;
}
.history time {
  color: var(--text-muted);
  font-size: 12px;
}
.empty {
  color: var(--text-muted);
}
.primary,
.danger,
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
.danger {
  border: 1px solid rgba(255, 135, 165, 0.6);
  background: rgba(80, 25, 40, 0.8);
  color: #ffd9e3;
}
.ghost {
  border: 1px solid rgba(111, 156, 200, 0.5);
  background: rgba(8, 19, 34, 0.65);
  color: #d7e8ff;
}
@media (max-width: 920px) {
  .panel-grid,
  .form-grid.two {
    grid-template-columns: 1fr;
  }
  .form-grid .full {
    grid-column: auto;
  }
}
</style>
