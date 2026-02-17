<template>
  <div class="console-bg">
    <header class="console-header reveal">
      <div>
        <p class="eyebrow">BajiaoZhi Admin Console</p>
        <h1>八角志管理后台</h1>
      </div>
      <div class="header-meta">
        <p class="api-hint">API: {{ apiBaseUrl }}</p>
        <button v-if="isAuthed" class="ghost-btn" type="button" @click="onLogout">退出登录</button>
      </div>
    </header>

    <main class="console-main reveal">
      <aside class="console-nav">
        <button
          v-for="tab in visibleTabs"
          :key="tab.key"
          :class="['nav-item', { active: tab.key === activeTab }]"
          type="button"
          @click="activeTab = tab.key"
        >
          <span>{{ tab.label }}</span>
        </button>
      </aside>

      <section class="console-content">
        <component :is="currentComponent" @login-success="onLoginSuccess" />
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { API_BASE_URL, getAuthToken } from './api/request'
import { logout } from './api/auth'
import ArticleManager from './pages/articles/ArticleManager.vue'
import Login from './pages/auth/Login.vue'
import TakedownManager from './pages/compliance/TakedownManager.vue'
import EventEditor from './pages/events/EventEditor.vue'
import FighterManager from './pages/fighters/FighterManager.vue'
import ReviewQueue from './pages/review/ReviewQueue.vue'
import SourceManager from './pages/sources/SourceManager.vue'

type TabKey =
  | 'login'
  | 'sources'
  | 'review'
  | 'articles'
  | 'events'
  | 'fighters'
  | 'compliance'

type TabItem = {
  key: TabKey
  label: string
  component: unknown
  authRequired: boolean
}

const tabs: TabItem[] = [
  { key: 'login', label: '登录', component: Login, authRequired: false },
  { key: 'sources', label: '数据源', component: SourceManager, authRequired: true },
  { key: 'review', label: '审核', component: ReviewQueue, authRequired: true },
  { key: 'articles', label: '资讯', component: ArticleManager, authRequired: true },
  { key: 'events', label: '赛程', component: EventEditor, authRequired: true },
  { key: 'fighters', label: '选手', component: FighterManager, authRequired: true },
  { key: 'compliance', label: '下架', component: TakedownManager, authRequired: true },
]

const apiBaseUrl = API_BASE_URL
const isAuthed = ref(Boolean(getAuthToken()))
const activeTab = ref<TabKey>(isAuthed.value ? 'sources' : 'login')

const visibleTabs = computed(() =>
  tabs.filter((item) => {
    if (!item.authRequired) {
      return true
    }
    return isAuthed.value
  }),
)

const currentComponent = computed(() => {
  const fallback = isAuthed.value ? 'sources' : 'login'
  const key = isAuthed.value ? activeTab.value : 'login'
  const matched = tabs.find((item) => item.key === key) || tabs.find((item) => item.key === fallback)
  return matched?.component || Login
})

function onLoginSuccess() {
  isAuthed.value = true
  activeTab.value = 'sources'
}

function onLogout() {
  logout()
  isAuthed.value = false
  activeTab.value = 'login'
}
</script>
