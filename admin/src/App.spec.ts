import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import App from './App.vue'

vi.mock('./api/request', () => ({
  API_BASE_URL: 'http://localhost:8080',
  getAuthToken: vi.fn(() => 'token-for-test'),
}))

vi.mock('./api/auth', () => ({
  logout: vi.fn(),
}))

vi.mock('./pages/auth/Login.vue', () => ({
  default: { name: 'Login', template: '<div>Login</div>' },
}))

vi.mock('./pages/compliance/TakedownManager.vue', () => ({
  default: { name: 'TakedownManager', template: '<div>TakedownManager</div>' },
}))

vi.mock('./pages/events/EventEditor.vue', () => ({
  default: { name: 'EventEditor', template: '<div>EventEditor</div>' },
}))

vi.mock('./pages/fighters/FighterManager.vue', () => ({
  default: { name: 'FighterManager', template: '<div>FighterManager</div>' },
}))

vi.mock('./pages/sources/SourceManager.vue', () => ({
  default: { name: 'SourceManager', template: '<div>SourceManager</div>' },
}))

describe('App nav', () => {
  it('does not expose news-related tabs', () => {
    const wrapper = mount(App)
    const labels = wrapper.findAll('.nav-item').map((node) => node.text())

    expect(labels).not.toContain('资讯')
    expect(labels).not.toContain('审核')
  })
})
