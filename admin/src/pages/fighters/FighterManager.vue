<template>
  <section>
    <h1>选手管理</h1>
    <p class="hint">支持手动录入选手基础资料，可附头像与介绍视频。</p>
    <div class="form-grid two">
      <input v-model.number="form.source_id" placeholder="source_id" data-test="source-id" />
      <input v-model="form.name" placeholder="姓名" data-test="name" />
      <input v-model="form.country" placeholder="国籍" data-test="country" />
      <input v-model="form.record" placeholder="战绩" data-test="record" />
      <input v-model="form.avatar_url" placeholder="头像 URL（可选）" />
      <input v-model="form.intro_video_url" placeholder="介绍视频 URL（可选）" />
    </div>
    <button data-test="create" @click="onCreate">手动录入</button>
    <button @click="onSearch">刷新检索结果</button>
    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <p class="section-title">检索结果（关键词：{{ keyword }}）</p>
    <ul>
      <li v-for="item in items" :key="item.id">
        <strong>{{ item.name }}</strong>
        <span>{{ item.country || '未知地区' }}</span>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { createManualFighter, searchFighters, type FighterItem } from '../../api/fighters'

const form = reactive({
  source_id: 0,
  name: '',
  country: '',
  record: '',
  avatar_url: '',
  intro_video_url: '',
})
const keyword = ref('Alex')
const items = ref<FighterItem[]>([])
const error = ref('')
const success = ref('')

onMounted(async () => {
  await onSearch()
})

async function onCreate() {
  error.value = ''
  success.value = ''
  try {
    await createManualFighter({ ...form })
    success.value = `已录入选手：${form.name}`
    if (form.name) {
      keyword.value = form.name
    }
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
