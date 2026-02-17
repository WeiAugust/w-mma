<template>
  <section>
    <h1>数据源管理</h1>
    <p class="hint">配置赛程、资讯、选手来源，并控制展示/播放/AI 摘要授权位。</p>

    <div class="form-grid two">
      <input v-model="draft.name" placeholder="名称" data-test="name" />
      <select v-model="draft.source_type" data-test="source-type">
        <option value="news">news</option>
        <option value="schedule">schedule</option>
        <option value="fighter">fighter</option>
      </select>
      <input v-model="draft.platform" placeholder="平台" data-test="platform" />
      <input v-model="draft.source_url" placeholder="来源 URL" data-test="source-url" />
      <input v-model="draft.account_id" placeholder="账号ID(可选)" />
      <input v-model="draft.parser_kind" placeholder="解析器(generic)" />
      <label>
        <input v-model="draft.rights_display" type="checkbox" />
        rights_display
      </label>
      <label>
        <input v-model="draft.rights_playback" type="checkbox" />
        rights_playback
      </label>
      <label>
        <input v-model="draft.rights_ai_summary" type="checkbox" />
        rights_ai_summary
      </label>
      <input v-model="draft.rights_proof_url" placeholder="授权证明链接(可选)" />
    </div>
    <button data-test="create" @click="onCreate">新增</button>
    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <p class="section-title">已有数据源</p>
    <ul>
      <li v-for="item in items" :key="item.id">
        <strong>{{ item.name }}</strong>
        <span>({{ item.source_type }}/{{ item.platform }})</span>
        <span>{{ item.enabled ? '启用中' : '已停用' }}</span>
        <input
          type="checkbox"
          :checked="item.rights_playback"
          :data-test="`playback-${item.id}`"
          @change="onPlaybackChange(item.id, ($event.target as HTMLInputElement).checked)"
        />
        <span>播放授权</span>
        <button :data-test="`toggle-${item.id}`" @click="onToggle(item.id)">启停</button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { createSource, listSources, toggleSource, updateSource, type SourceItem, type SourcePayload } from '../../api/sources'

const items = ref<SourceItem[]>([])
const error = ref('')
const success = ref('')
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
  await loadItems()
})

async function onCreate() {
  error.value = ''
  success.value = ''
  try {
    const created = await createSource({ ...draft })
    items.value.push(created)
    success.value = `已新增数据源：${created.name}`
  } catch (err) {
    error.value = (err as Error).message || '新增失败'
  }
}

async function onToggle(id: number) {
  error.value = ''
  success.value = ''
  try {
    await toggleSource(id)
    await loadItems()
    success.value = `已切换数据源 #${id} 启停状态`
  } catch (err) {
    error.value = (err as Error).message || '切换失败'
  }
}

async function onPlaybackChange(id: number, checked: boolean) {
  error.value = ''
  success.value = ''
  try {
    await updateSource(id, { rights_playback: checked })
    const target = items.value.find((item) => item.id === id)
    if (target) {
      target.rights_playback = checked
    }
    success.value = `已更新数据源 #${id} 播放权限`
  } catch (err) {
    error.value = (err as Error).message || '更新失败'
  }
}

async function loadItems() {
  try {
    items.value = await listSources()
  } catch (err) {
    error.value = (err as Error).message || '加载数据源失败'
  }
}
</script>
