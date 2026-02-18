import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ArticleManager from './ArticleManager.vue'
import { createManualArticle, listPublishedArticles } from '../../api/articles'

vi.mock('../../api/articles', () => ({
  createManualArticle: vi.fn(),
  listPublishedArticles: vi.fn(),
}))

describe('ArticleManager', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('supports create + refresh + keyword filter workflow', async () => {
    vi.mocked(listPublishedArticles).mockResolvedValue([
      {
        id: 1,
        source_id: 6,
        title: 'UFC 314 战卡公布',
        summary: '...',
        source_url: 'https://www.ufc.com/event/ufc-314',
        can_play: false,
      },
      {
        id: 2,
        source_id: 6,
        title: 'UFC 315 预热',
        summary: '...',
        source_url: 'https://www.ufc.com/event/ufc-315',
        can_play: true,
      },
    ])
    vi.mocked(createManualArticle).mockResolvedValue()

    const wrapper = mount(ArticleManager)
    await flushPromises()

    expect(wrapper.text()).toContain('公开资讯总数')
    expect(wrapper.text()).toContain('UFC 314 战卡公布')

    await wrapper.get('[data-test="open-create"]').trigger('click')
    await wrapper.get('[data-test="create-source-id"]').setValue('6')
    await wrapper.get('[data-test="create-title"]').setValue('UFC 316 速览')
    await wrapper.get('[data-test="create-summary"]').setValue('手动录入摘要')
    await wrapper.get('[data-test="create-source-url"]').setValue('https://www.ufc.com/event/ufc-316')
    await wrapper.get('[data-test="submit-create"]').trigger('click')
    await flushPromises()

    expect(createManualArticle).toHaveBeenCalledWith(
      expect.objectContaining({
        source_id: 6,
        title: 'UFC 316 速览',
      }),
    )

    await wrapper.get('[data-test="filter-keyword"]').setValue('315')
    await wrapper.get('[data-test="apply-filter"]').trigger('click')
    expect(wrapper.text()).toContain('UFC 315 预热')
    expect(wrapper.text()).not.toContain('UFC 314 战卡公布')
  })
})
