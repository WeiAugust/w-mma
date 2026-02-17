<template>
  <section>
    <h1>合规下架</h1>
    <input v-model="form.target_type" placeholder="target_type" data-test="target-type" />
    <input v-model.number="form.target_id" placeholder="target_id" data-test="target-id" />
    <textarea v-model="form.reason" placeholder="reason" data-test="reason" />
    <button data-test="create" @click="onCreate">创建工单</button>

    <input v-model.number="resolveID" placeholder="takedown_id" data-test="resolve-id" />
    <select v-model="action" data-test="action">
      <option value="offlined">offlined</option>
      <option value="rejected">rejected</option>
    </select>
    <button data-test="resolve" @click="onResolve">处理工单</button>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { createTakedown, resolveTakedown } from '../../api/takedowns'

const form = ref({
  target_type: 'article',
  target_id: 0,
  reason: '',
})
const resolveID = ref(0)
const action = ref<'offlined' | 'rejected'>('offlined')

async function onCreate() {
  await createTakedown({ ...form.value })
}

async function onResolve() {
  await resolveTakedown(resolveID.value, action.value)
}
</script>
