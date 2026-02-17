<template>
  <section>
    <h1>赛事管理</h1>
    <ul>
      <li v-for="event in events" :key="event.id">
        <span>{{ event.name }}</span>
        <select v-model="event.status" :data-test="`status-${event.id}`">
          <option value="scheduled">scheduled</option>
          <option value="live">live</option>
          <option value="completed">completed</option>
        </select>
        <button :data-test="`save-${event.id}`" @click="saveEvent(event.id, event.status)">保存</button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { listEvents, updateEvent, type EventItem } from '../../api/events'

const events = ref<EventItem[]>([])

onMounted(async () => {
  events.value = await listEvents()
})

async function saveEvent(id: number, status: EventItem['status']) {
  await updateEvent(id, { status })
}
</script>
