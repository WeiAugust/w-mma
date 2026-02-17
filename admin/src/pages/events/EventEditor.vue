<template>
  <section>
    <h1>赛事管理</h1>
    <p class="hint">用于更新赛事名称与状态（scheduled/live/completed）。</p>
    <button @click="loadEvents">刷新赛事</button>
    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <ul>
      <li v-for="event in events" :key="event.id">
        <input v-model="event.name" :data-test="`name-${event.id}`" />
        <select v-model="event.status" :data-test="`status-${event.id}`">
          <option value="scheduled">scheduled</option>
          <option value="live">live</option>
          <option value="completed">completed</option>
        </select>
        <button :data-test="`save-${event.id}`" @click="saveEvent(event.id, event.name, event.status)">保存</button>
      </li>
    </ul>
    <p v-if="events.length === 0" class="hint">暂无赛事数据。</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { listEvents, updateEvent, type EventItem } from '../../api/events'

const events = ref<EventItem[]>([])
const error = ref('')
const success = ref('')

onMounted(async () => {
  await loadEvents()
})

async function saveEvent(id: number, name: string, status: EventItem['status']) {
  error.value = ''
  success.value = ''
  try {
    await updateEvent(id, { name, status })
    success.value = `已更新赛事 #${id}`
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
