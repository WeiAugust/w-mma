<template>
  <section>
    <h1>待审核</h1>
    <p class="hint">审核通过后，内容将进入公开资讯列表。</p>
    <button @click="loadItems">刷新列表</button>
    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <ul>
      <li v-for="item in items" :key="item.id">
        <strong>{{ item.title }}</strong>
        <span>#{{ item.id }}</span>
        <button :data-test="`approve-${item.id}`" @click="onApprove(item.id)">通过</button>
      </li>
    </ul>
    <p v-if="items.length === 0" class="hint">当前没有待审核内容。</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { approvePending, listPending, type PendingItem } from '../../api/review'

const items = ref<PendingItem[]>([])
const error = ref('')
const success = ref('')

onMounted(async () => {
  await loadItems()
})

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
