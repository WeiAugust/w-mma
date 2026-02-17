<template>
  <section>
    <h1>数据源管理</h1>

    <div class="create">
      <input v-model="draft.name" placeholder="名称" data-test="name" />
      <select v-model="draft.source_type" data-test="source-type">
        <option value="news">news</option>
        <option value="schedule">schedule</option>
        <option value="fighter">fighter</option>
      </select>
      <input v-model="draft.platform" placeholder="平台" data-test="platform" />
      <input v-model="draft.source_url" placeholder="来源 URL" data-test="source-url" />
      <button data-test="create" @click="onCreate">新增</button>
    </div>

    <ul>
      <li v-for="item in items" :key="item.id">
        <span>{{ item.name }}</span>
        <span>({{ item.source_type }})</span>
        <input
          type="checkbox"
          :checked="item.rights_playback"
          :data-test="`playback-${item.id}`"
          @change="onPlaybackChange(item.id, ($event.target as HTMLInputElement).checked)"
        />
        <button :data-test="`toggle-${item.id}`" @click="onToggle(item.id)">启停</button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { createSource, listSources, toggleSource, updateSource, type SourceItem, type SourcePayload } from '../../api/sources'

const items = ref<SourceItem[]>([])
const draft = reactive<SourcePayload>({
  name: '',
  source_type: 'news',
  platform: '',
  account_id: '',
  source_url: '',
  parser_kind: 'generic',
  enabled: true,
  rights_display: true,
  rights_playback: false,
  rights_ai_summary: false,
  rights_expires_at: '',
  rights_proof_url: '',
})

onMounted(async () => {
  items.value = await listSources()
})

async function onCreate() {
  const created = await createSource({ ...draft })
  items.value.push(created)
}

async function onToggle(id: number) {
  await toggleSource(id)
  items.value = await listSources()
}

async function onPlaybackChange(id: number, checked: boolean) {
  await updateSource(id, { rights_playback: checked })
  const target = items.value.find((item) => item.id === id)
  if (target) {
    target.rights_playback = checked
  }
}
</script>
