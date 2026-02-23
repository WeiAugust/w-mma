<template>
  <section class="fighter-console">
    <header class="page-head">
      <div>
        <h1>选手资料管理</h1>
        <p class="hint">支持按关键词检索选手，同时可手动补录选手资料与媒体链接。</p>
      </div>
      <button class="primary" data-test="open-create" type="button" @click="showCreate = true">新建选手</button>
    </header>

    <div class="kpi-grid">
      <article class="kpi-card">
        <p class="kpi-label">选手库数量</p>
        <p class="kpi-value">{{ items.length }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">有头像</p>
        <p class="kpi-value">{{ withAvatar }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">有战绩</p>
        <p class="kpi-value">{{ withRecord }}</p>
      </article>
    </div>

    <div class="search-bar">
      <label>
        检索关键词
        <input v-model.trim="keyword" data-test="search-keyword" placeholder="如 Alex / 亚历克斯 / Poatan" />
      </label>
      <button class="ghost" data-test="run-search" type="button" @click="onSearch">查询</button>
    </div>

    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>姓名</th>
            <th>外号</th>
            <th>国籍</th>
            <th>战绩</th>
            <th>头像</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>#{{ item.id }}</td>
            <td>{{ item.name }}<span v-if="item.name_zh"> / {{ item.name_zh }}</span></td>
            <td>{{ item.nickname || '-' }}</td>
            <td>{{ item.country || '未知' }}</td>
            <td>{{ item.record || '未录入' }}</td>
            <td>{{ item.avatar_url ? '已配置' : '未配置' }}</td>
          </tr>
          <tr v-if="items.length === 0">
            <td class="empty" colspan="6">暂无匹配选手</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showCreate" class="drawer-mask">
      <section class="drawer">
        <header>
          <h2>手动录入选手</h2>
          <button class="ghost" type="button" @click="showCreate = false">关闭</button>
        </header>
        <div class="form-grid two">
          <label>
            source_id
            <input v-model.number="form.source_id" data-test="create-source-id" />
          </label>
          <label>
            姓名
            <input v-model="form.name" data-test="create-name" />
          </label>
          <label>
            中文名
            <input v-model="form.name_zh" />
          </label>
          <label>
            外号
            <input v-model="form.nickname" />
          </label>
          <label>
            国籍
            <input v-model="form.country" data-test="create-country" />
          </label>
          <label>
            战绩
            <input v-model="form.record" data-test="create-record" />
          </label>
          <label>
            量级
            <input v-model="form.weight_class" />
          </label>
          <label>
            头像 URL
            <input v-model="form.avatar_url" />
          </label>
          <label>
            介绍视频 URL
            <input v-model="form.intro_video_url" />
          </label>
        </div>
        <footer>
          <button class="primary" data-test="submit-create" type="button" @click="onCreate">保存选手</button>
        </footer>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import { createManualFighter, searchFighters, type FighterItem } from '../../api/fighters'

const form = reactive({
  source_id: 0,
  name: '',
  name_zh: '',
  nickname: '',
  country: '',
  record: '',
  weight_class: '',
  avatar_url: '',
  intro_video_url: '',
})
const keyword = ref('Alex')
const items = ref<FighterItem[]>([])
const error = ref('')
const success = ref('')
const showCreate = ref(false)

const withAvatar = computed(() => items.value.filter((item) => item.avatar_url).length)
const withRecord = computed(() => items.value.filter((item) => item.record).length)

onMounted(onSearch)

async function onCreate() {
  error.value = ''
  success.value = ''
  try {
    await createManualFighter({ ...form })
    success.value = `已录入选手：${form.name}`
    if (form.name) keyword.value = form.name
    showCreate.value = false
    await onSearch()
  } catch (err) {
    error.value = (err as Error).message || '录入失败'
  }
}

async function onSearch() {
  error.value = ''
  try {
    items.value = await searchFighters(keyword.value)
  } catch (err) {
    error.value = (err as Error).message || '检索失败'
  }
}
</script>

<style scoped>
.fighter-console {
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
.search-bar {
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
  min-width: 700px;
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
.drawer-mask {
  position: fixed;
  inset: 0;
  background: rgba(3, 8, 16, 0.75);
  display: grid;
  place-items: center;
  padding: 14px;
}
.drawer {
  width: min(740px, 100%);
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
  .search-bar,
  .form-grid.two {
    grid-template-columns: 1fr;
  }
}
</style>
