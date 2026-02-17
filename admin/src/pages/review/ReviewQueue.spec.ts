import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ReviewQueue from './ReviewQueue.vue'
import { approvePending, listPending } from '../../api/review'

vi.mock('../../api/review', () => ({
  listPending: vi.fn(),
  approvePending: vi.fn(),
}))

describe('ReviewQueue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows pending items and approves one item', async () => {
    vi.mocked(listPending).mockResolvedValue([{ id: 1, title: 'news-a' }])

    const wrapper = mount(ReviewQueue)
    await flushPromises()

    expect(wrapper.text()).toContain('news-a')

    await wrapper.get('[data-test="approve-1"]').trigger('click')
    expect(approvePending).toHaveBeenCalledWith(1)
  })
})
