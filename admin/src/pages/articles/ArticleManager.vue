<template>
  <section>
    <h1>资讯管理</h1>
    <p class="hint">手动录入后进入待审核，审核通过后出现在公开资讯。</p>
    <div class="form-grid two">
      <input v-model.number="form.source_id" placeholder="source_id" data-test="source-id" />
      <input v-model="form.title" placeholder="标题" data-test="title" />
      <textarea v-model="form.summary" placeholder="摘要" data-test="summary" />
      <input v-model="form.source_url" placeholder="来源链接" data-test="source-url" />
      <input v-model="form.cover_url" placeholder="封面图 URL（可选）" />
      <input v-model="form.video_url" placeholder="视频 URL（可选）" />
    </div>
    <button data-test="create" @click="onCreate">手动录入</button>
    <button @click="loadPublished">刷新公开资讯</button>
    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <p class="section-title">公开资讯预览</p>
    <ul>
      <li v-for="item in published" :key="item.id">
        <strong>{{ item.title }}</strong>
        <span>#{{ item.id }}</span>
        <span>{{ item.can_play ? '可播放视频' : '仅可跳转源链接' }}</span>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

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

onMounted(async () => {
  await loadPublished()
})

async function onCreate() {
  error.value = ''
  success.value = ''
  try {
    await createManualArticle({ ...form })
    success.value = '录入成功，已进入待审核队列'
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
