<template>
  <section>
    <h1>待审核</h1>
    <ul>
      <li v-for="item in items" :key="item.id">
        <span>{{ item.title }}</span>
        <button :data-test="`approve-${item.id}`" @click="onApprove(item.id)">通过</button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { approvePending, listPending, type PendingItem } from '../../api/review'

const items = ref<PendingItem[]>([])

onMounted(async () => {
  items.value = await listPending()
})

async function onApprove(id: number) {
  await approvePending(id)
  items.value = items.value.filter((item) => item.id !== id)
}
</script>
