<template>
  <section>
    <h1>选手管理</h1>
    <input v-model.number="form.source_id" placeholder="source_id" data-test="source-id" />
    <input v-model="form.name" placeholder="姓名" data-test="name" />
    <input v-model="form.country" placeholder="国籍" data-test="country" />
    <input v-model="form.record" placeholder="战绩" data-test="record" />
    <button data-test="create" @click="onCreate">手动录入</button>
  </section>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

import { createManualFighter } from '../../api/fighters'

const form = reactive({
  source_id: 0,
  name: '',
  country: '',
  record: '',
})

async function onCreate() {
  await createManualFighter({ ...form })
}
</script>
