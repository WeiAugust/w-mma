import { flushPromises, mount } from '@vue/test-utils'
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

  it('supports review workspace flow', async () => {
    vi.mocked(listPending).mockResolvedValue([
      { id: 1, title: 'UFC 314 战卡更新' },
      { id: 2, title: 'ONE 172 赛程变更' },
    ])
    vi.mocked(approvePending).mockResolvedValue()

    const wrapper = mount(ReviewQueue)
    await flushPromises()

    expect(wrapper.text()).toContain('审核工作台')
    expect(wrapper.text()).toContain('待审核总数')

    await wrapper.get('[data-test="keyword"]').setValue('ONE')
    await wrapper.get('[data-test="apply-filter"]').trigger('click')
    expect(wrapper.text()).toContain('ONE 172 赛程变更')
    expect(wrapper.text()).not.toContain('UFC 314 战卡更新')

    await wrapper.get('[data-test="approve-2"]').trigger('click')
    expect(approvePending).toHaveBeenCalledWith(2)
    await flushPromises()
    expect(wrapper.text()).toContain('已通过待审核内容 #2')
  })
})
