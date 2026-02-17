<template>
  <section>
    <h1>资讯管理</h1>
    <input v-model.number="form.source_id" placeholder="source_id" data-test="source-id" />
    <input v-model="form.title" placeholder="标题" data-test="title" />
    <textarea v-model="form.summary" placeholder="摘要" data-test="summary" />
    <input v-model="form.source_url" placeholder="来源链接" data-test="source-url" />
    <button data-test="create" @click="onCreate">手动录入</button>
  </section>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

import { createManualArticle } from '../../api/articles'

const form = reactive({
  source_id: 0,
  title: '',
  summary: '',
  source_url: '',
})

async function onCreate() {
  await createManualArticle({ ...form })
}
</script>
