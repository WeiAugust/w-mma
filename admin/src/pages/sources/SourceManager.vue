<template>
  <section class="source-console">
    <header class="page-head">
      <div>
        <h1>数据源管理中心</h1>
        <p class="hint">按业务流程完成配置、启停、同步、删除与恢复，形成闭环运营。</p>
      </div>
      <button class="primary" data-test="open-create" type="button" @click="showCreate = true">新建数据源</button>
    </header>

    <div class="kpi-grid">
      <article class="kpi-card">
        <p class="kpi-label">总数据源</p>
        <p class="kpi-value">{{ stats.total }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">启用中</p>
        <p class="kpi-value">{{ stats.enabled }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">内置源</p>
        <p class="kpi-value">{{ stats.builtin }}</p>
      </article>
      <article class="kpi-card">
        <p class="kpi-label">已删除</p>
        <p class="kpi-value">{{ stats.deleted }}</p>
      </article>
    </div>

    <div class="filters">
      <div class="filters-grid">
        <label>
          平台
          <input v-model.trim="filters.platform" data-test="filter-platform" placeholder="ufc / one / wbc..." />
        </label>
        <label>
          类型
          <select v-model="filters.sourceType" data-test="filter-source-type">
            <option value="">全部</option>
            <option value="news">news</option>
            <option value="schedule">schedule</option>
            <option value="fighter">fighter</option>
          </select>
        </label>
        <label>
          启停状态
          <select v-model="filters.enabled" data-test="filter-enabled">
            <option value="">全部</option>
            <option value="true">启用</option>
            <option value="false">停用</option>
          </select>
        </label>
        <label>
          内置源
          <select v-model="filters.builtin" data-test="filter-builtin">
            <option value="">全部</option>
            <option value="true">是</option>
            <option value="false">否</option>
          </select>
        </label>
      </div>
      <div class="filters-actions">
        <label class="checkbox">
          <input v-model="filters.includeDeleted" data-test="include-deleted" type="checkbox" />
          显示已删除
        </label>
        <button class="ghost" data-test="reset-filters" type="button" @click="onResetFilters">重置</button>
        <button class="primary" data-test="apply-filters" type="button" @click="loadItems">应用筛选</button>
      </div>
    </div>

    <p v-if="error" class="status-error">{{ error }}</p>
    <p v-if="success" class="status-success">{{ success }}</p>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>平台</th>
            <th>解析器</th>
            <th>状态</th>
            <th>授权</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id" :class="{ deleted: !!item.deleted_at }">
            <td>
              <div class="name-cell">
                <strong>{{ item.name }}</strong>
                <span class="muted">{{ item.source_url }}</span>
              </div>
            </td>
            <td>{{ item.source_type }}</td>
            <td>{{ item.platform }}</td>
            <td>{{ item.parser_kind }}</td>
            <td>
              <span class="tag" :class="item.enabled ? 'ok' : 'off'">{{ item.enabled ? '启用' : '停用' }}</span>
              <span v-if="item.deleted_at" class="tag warn">已删除</span>
              <span v-if="item.is_builtin" class="tag info">内置</span>
            </td>
            <td>
              <div class="rights">
                <label class="checkbox inline">
                  <input
                    :checked="item.rights_playback"
                    :disabled="busyRows[item.id] || !!item.deleted_at"
                    :data-test="`playback-${item.id}`"
                    type="checkbox"
                    @change="onPlaybackChange(item.id, ($event.target as HTMLInputElement).checked)"
                  />
                  播放
                </label>
              </div>
            </td>
            <td>
              <div class="row-actions">
                <button class="ghost" :data-test="`edit-${item.id}`" :disabled="busyRows[item.id]" @click="onOpenEdit(item)">
                  编辑
                </button>
                <button class="ghost" :data-test="`toggle-${item.id}`" :disabled="busyRows[item.id] || !!item.deleted_at" @click="onToggle(item.id)">
                  启停
                </button>
                <button class="primary small" :data-test="`sync-${item.id}`" :disabled="busyRows[item.id] || !!item.deleted_at" @click="onSync(item.id)">
                  同步
                </button>
                <button
                  v-if="!item.deleted_at"
                  class="danger"
                  :data-test="`delete-${item.id}`"
                  :disabled="busyRows[item.id]"
                  @click="onDelete(item.id)"
                >
                  删除
                </button>
                <button
                  v-else
                  class="ghost"
                  :data-test="`restore-${item.id}`"
                  :disabled="busyRows[item.id]"
                  @click="onRestore(item.id)"
                >
                  恢复
                </button>
              </div>
              <p v-if="syncHints[item.id]" class="sync-hint">{{ syncHints[item.id] }}</p>
            </td>
          </tr>
          <tr v-if="items.length === 0">
            <td class="empty" colspan="7">暂无符合条件的数据源</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showCreate" class="drawer-mask">
      <section class="drawer">
        <header>
          <h2>新建数据源</h2>
          <button class="ghost" type="button" @click="showCreate = false">关闭</button>
        </header>
        <div class="form-grid two">
          <label>
            名称
            <input v-model="createDraft.name" data-test="create-name" />
          </label>
          <label>
            类型
            <select v-model="createDraft.source_type" data-test="create-source-type">
              <option value="news">news</option>
              <option value="schedule">schedule</option>
              <option value="fighter">fighter</option>
            </select>
          </label>
          <label>
            平台
            <input v-model="createDraft.platform" data-test="create-platform" />
          </label>
          <label>
            来源 URL
            <input v-model="createDraft.source_url" data-test="create-source-url" />
          </label>
          <label>
            解析器
            <input v-model="createDraft.parser_kind" />
          </label>
          <label>
            账号 ID
            <input v-model="createDraft.account_id" />
          </label>
        </div>
        <div class="form-checks">
          <label class="checkbox inline"><input v-model="createDraft.enabled" type="checkbox" />启用</label>
          <label class="checkbox inline"><input v-model="createDraft.is_builtin" type="checkbox" />内置</label>
          <label class="checkbox inline"><input v-model="createDraft.rights_display" type="checkbox" />展示授权</label>
          <label class="checkbox inline"><input v-model="createDraft.rights_playback" type="checkbox" />播放授权</label>
          <label class="checkbox inline"><input v-model="createDraft.rights_ai_summary" type="checkbox" />AI 摘要授权</label>
        </div>
        <footer>
          <button class="primary" data-test="submit-create" type="button" @click="onCreate">保存并创建</button>
        </footer>
      </section>
    </div>

    <div v-if="editingId !== null" class="drawer-mask">
      <section class="drawer">
        <header>
          <h2>编辑数据源</h2>
          <button class="ghost" type="button" @click="closeEdit">关闭</button>
        </header>
        <div class="form-grid two">
          <label>
            名称
            <input v-model="editDraft.name" data-test="edit-name" />
          </label>
          <label>
            平台
            <input v-model="editDraft.platform" />
          </label>
          <label>
            来源 URL
            <input v-model="editDraft.source_url" />
          </label>
          <label>
            解析器
            <input v-model="editDraft.parser_kind" />
          </label>
        </div>
        <div class="form-checks">
          <label class="checkbox inline">
            <input v-model="editDraft.rights_display" type="checkbox" />
            展示授权
          </label>
          <label class="checkbox inline">
            <input v-model="editDraft.rights_playback" data-test="edit-rights-playback" type="checkbox" />
            播放授权
          </label>
          <label class="checkbox inline">
            <input v-model="editDraft.rights_ai_summary" type="checkbox" />
            AI 摘要授权
          </label>
        </div>
        <footer>
          <button class="primary" data-test="save-edit" type="button" @click="onSaveEdit">保存修改</button>
        </footer>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import {
  createSource,
  deleteSource,
  listSources,
  restoreSource,
  toggleSource,
  triggerIngestFetch,
  updateSource,
  type SourceItem,
  type SourcePayload,
} from '../../api/sources'

const items = ref<SourceItem[]>([])
const error = ref('')
const success = ref('')
const busyRows = ref<Record<number, boolean>>({})
const syncHints = ref<Record<number, string>>({})

const filters = reactive({
  includeDeleted: false,
  platform: '',
  sourceType: '' as '' | 'news' | 'schedule' | 'fighter',
  enabled: '' as '' | 'true' | 'false',
  builtin: '' as '' | 'true' | 'false',
})

const showCreate = ref(false)
const createDraft = reactive(createDefaultDraft())
const editingId = ref<number | null>(null)
const editDraft = reactive(createDefaultDraft())

const stats = computed(() => {
  const total = items.value.length
  let enabled = 0
  let builtin = 0
  let deleted = 0
  for (const item of items.value) {
    if (item.enabled) enabled++
    if (item.is_builtin) builtin++
    if (item.deleted_at) deleted++
  }
  return { total, enabled, builtin, deleted }
})

onMounted(loadItems)

function createDefaultDraft(): SourcePayload {
  return {
    name: '',
    source_type: 'news',
    platform: '',
    account_id: '',
    source_url: '',
    parser_kind: 'generic',
    enabled: true,
    is_builtin: false,
    rights_display: true,
    rights_playback: false,
    rights_ai_summary: false,
    rights_expires_at: '',
    rights_proof_url: '',
  }
}

function syncDraft(target: SourcePayload, source: SourcePayload) {
  target.name = source.name
  target.source_type = source.source_type
  target.platform = source.platform
  target.account_id = source.account_id
  target.source_url = source.source_url
  target.parser_kind = source.parser_kind
  target.enabled = source.enabled
  target.is_builtin = source.is_builtin
  target.rights_display = source.rights_display
  target.rights_playback = source.rights_playback
  target.rights_ai_summary = source.rights_ai_summary
  target.rights_expires_at = source.rights_expires_at || ''
  target.rights_proof_url = source.rights_proof_url || ''
}

function toBoolean(value: '' | 'true' | 'false'): boolean | undefined {
  if (value === '') return undefined
  return value === 'true'
}

async function loadItems() {
  error.value = ''
  try {
    items.value = await listSources({
      include_deleted: filters.includeDeleted,
      platform: filters.platform,
      source_type: filters.sourceType || undefined,
      enabled: toBoolean(filters.enabled),
      is_builtin: toBoolean(filters.builtin),
    })
  } catch (err) {
    error.value = (err as Error).message || '加载数据源失败'
  }
}

function onResetFilters() {
  filters.platform = ''
  filters.sourceType = ''
  filters.enabled = ''
  filters.builtin = ''
  filters.includeDeleted = false
  loadItems()
}

async function onCreate() {
  error.value = ''
  success.value = ''
  try {
    await createSource({ ...createDraft })
    success.value = `已新增数据源：${createDraft.name}`
    syncDraft(createDraft, createDefaultDraft())
    showCreate.value = false
    await loadItems()
  } catch (err) {
    error.value = (err as Error).message || '新增失败'
  }
}

function onOpenEdit(item: SourceItem) {
  editingId.value = item.id
  syncDraft(editDraft, {
    name: item.name,
    source_type: item.source_type,
    platform: item.platform,
    account_id: item.account_id || '',
    source_url: item.source_url,
    parser_kind: item.parser_kind,
    enabled: item.enabled,
    is_builtin: item.is_builtin,
    rights_display: item.rights_display,
    rights_playback: item.rights_playback,
    rights_ai_summary: item.rights_ai_summary,
    rights_expires_at: item.rights_expires_at || '',
    rights_proof_url: item.rights_proof_url || '',
  })
}

function closeEdit() {
  editingId.value = null
}

async function onSaveEdit() {
  if (editingId.value === null) return
  error.value = ''
  success.value = ''
  try {
    await updateSource(editingId.value, { ...editDraft })
    success.value = `数据源 #${editingId.value} 已更新`
    closeEdit()
    await loadItems()
  } catch (err) {
    error.value = (err as Error).message || '更新失败'
  }
}

async function runRowAction(sourceID: number, action: () => Promise<void>) {
  busyRows.value[sourceID] = true
  try {
    await action()
  } finally {
    busyRows.value[sourceID] = false
  }
}

async function onToggle(sourceID: number) {
  error.value = ''
  success.value = ''
  await runRowAction(sourceID, async () => {
    await toggleSource(sourceID)
    success.value = `已切换数据源 #${sourceID} 启停状态`
    await loadItems()
  })
}

async function onSync(sourceID: number) {
  error.value = ''
  success.value = ''
  await runRowAction(sourceID, async () => {
    await triggerIngestFetch(sourceID)
    const hint = '同步任务已触发，请刷新列表查看最新结果'
    syncHints.value[sourceID] = hint
    success.value = hint
  })
}

async function onDelete(sourceID: number) {
  error.value = ''
  success.value = ''
  await runRowAction(sourceID, async () => {
    await deleteSource(sourceID)
    success.value = `已删除数据源 #${sourceID}`
    await loadItems()
  })
}

async function onRestore(sourceID: number) {
  error.value = ''
  success.value = ''
  await runRowAction(sourceID, async () => {
    await restoreSource(sourceID)
    success.value = `已恢复数据源 #${sourceID}`
    await loadItems()
  })
}

async function onPlaybackChange(sourceID: number, checked: boolean) {
  error.value = ''
  success.value = ''
  await runRowAction(sourceID, async () => {
    await updateSource(sourceID, { rights_playback: checked })
    success.value = `已更新数据源 #${sourceID} 播放权限`
    const hit = items.value.find((item) => item.id === sourceID)
    if (hit) hit.rights_playback = checked
  })
}
</script>

<style scoped>
.source-console {
  display: grid;
  gap: 14px;
}

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.page-head h1 {
  margin: 0;
  font-size: 24px;
  letter-spacing: 0.2px;
}

.kpi-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.kpi-card {
  border: 1px solid rgba(88, 121, 164, 0.32);
  border-radius: 12px;
  padding: 12px;
  background:
    linear-gradient(130deg, rgba(16, 37, 59, 0.92), rgba(9, 22, 38, 0.9)),
    radial-gradient(200px 80px at 90% -20%, rgba(68, 213, 185, 0.24), transparent 60%);
}

.kpi-label {
  margin: 0;
  color: var(--text-muted);
  font-size: 12px;
}

.kpi-value {
  margin: 6px 0 0;
  font-size: 26px;
  font-weight: 700;
  color: #f3fbff;
}

.filters {
  border: 1px solid rgba(88, 121, 164, 0.3);
  border-radius: 12px;
  padding: 12px;
  background: rgba(10, 25, 42, 0.75);
}

.filters-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.filters-actions {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.filters-actions .checkbox {
  margin-right: auto;
}

.table-wrap {
  overflow: auto;
  border: 1px solid rgba(88, 121, 164, 0.28);
  border-radius: 12px;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 980px;
}

th,
td {
  text-align: left;
  padding: 10px 11px;
  border-bottom: 1px solid rgba(88, 121, 164, 0.22);
  vertical-align: top;
}

thead th {
  color: #b9d1ee;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  background: rgba(8, 18, 34, 0.84);
}

tbody tr.deleted {
  opacity: 0.76;
}

.name-cell {
  display: grid;
  gap: 4px;
}

.muted {
  color: var(--text-muted);
  font-size: 12px;
  max-width: 340px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tag {
  display: inline-block;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 600;
  margin-right: 6px;
}

.tag.ok {
  background: rgba(47, 197, 148, 0.2);
  color: #70f2c8;
}

.tag.off {
  background: rgba(255, 179, 95, 0.2);
  color: #ffc27d;
}

.tag.warn {
  background: rgba(255, 128, 159, 0.2);
  color: #ff94b3;
}

.tag.info {
  background: rgba(102, 152, 255, 0.2);
  color: #99c0ff;
}

.rights {
  display: flex;
  align-items: center;
}

.checkbox {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--text-muted);
  margin: 0;
}

.checkbox.inline {
  margin-right: 10px;
}

.row-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.sync-hint {
  margin: 6px 0 0;
  color: #9de7d7;
  font-size: 12px;
}

.empty {
  text-align: center;
  color: var(--text-muted);
  padding: 16px;
}

.primary,
.ghost,
.danger {
  border-radius: 10px;
  padding: 7px 11px;
  font-size: 13px;
  cursor: pointer;
}

.primary {
  border: 1px solid #5ce4c7;
  background: linear-gradient(135deg, #56ddbe, #7ef0d4);
  color: #03211a;
  font-weight: 700;
}

.primary.small {
  padding: 7px 10px;
}

.ghost {
  border: 1px solid rgba(117, 153, 196, 0.45);
  background: rgba(8, 19, 34, 0.65);
  color: #dcecff;
}

.danger {
  border: 1px solid rgba(255, 135, 165, 0.55);
  background: rgba(76, 20, 35, 0.8);
  color: #ffd5df;
}

.drawer-mask {
  position: fixed;
  inset: 0;
  background: rgba(3, 8, 16, 0.74);
  display: grid;
  place-items: center;
  padding: 14px;
  z-index: 1200;
}

.drawer {
  width: min(780px, 100%);
  border: 1px solid rgba(117, 153, 196, 0.42);
  border-radius: 14px;
  background:
    radial-gradient(360px 110px at 100% 0, rgba(86, 221, 190, 0.15), transparent 65%),
    rgba(8, 20, 34, 0.95);
  padding: 14px;
  display: grid;
  gap: 12px;
}

.drawer header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer h2 {
  margin: 0;
}

.form-grid {
  display: grid;
  gap: 10px;
}

.form-grid.two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.form-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.drawer footer {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 1000px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .filters-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .page-head {
    flex-direction: column;
  }

  .kpi-grid,
  .filters-grid,
  .form-grid.two {
    grid-template-columns: 1fr;
  }
}
</style>
