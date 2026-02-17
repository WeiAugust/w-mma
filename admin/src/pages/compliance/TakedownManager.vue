<template>
  <section>
    <h1>合规下架</h1>
    <p class="hint">创建投诉工单后可执行 offlined/rejected 处理动作。</p>

    <div class="form-grid two">
      <select v-model="form.target_type" data-test="target-type">
        <option value="article">article</option>
        <option value="event">event</option>
        <option value="fighter">fighter</option>
      </select>
      <input v-model.number="form.target_id" placeholder="target_id" data-test="target-id" />
      <textarea v-model="form.reason" placeholder="reason" data-test="reason" />
      <input v-model="form.complainant" placeholder="complainant (可选)" />
      <input v-model="form.evidence_url" placeholder="evidence_url (可选)" />
    </div>
    <button data-test="create" @click="onCreate">创建工单</button>

    <p class="section-title">处理工单</p>
    <input v-model.number="resolveID" placeholder="takedown_id" data-test="resolve-id" />
    <select v-model="action" data-test="action">
      <option value="offlined">offlined</option>
      <option value="rejected">rejected</option>
    </select>
    <button data-test="resolve" @click="onResolve">处理工单</button>
    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { createTakedown, resolveTakedown } from '../../api/takedowns'

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
    if (form.value.complainant) {
      payload.complainant = form.value.complainant
    }
    if (form.value.evidence_url) {
      payload.evidence_url = form.value.evidence_url
    }

    await createTakedown(payload)
    success.value = '工单创建成功'
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
  } catch (err) {
    error.value = (err as Error).message || '处理工单失败'
  }
}
</script>
