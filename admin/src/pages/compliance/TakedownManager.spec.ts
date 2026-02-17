import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import TakedownManager from './TakedownManager.vue'
import { createTakedown, resolveTakedown } from '../../api/takedowns'

vi.mock('../../api/takedowns', () => ({
  createTakedown: vi.fn(),
  resolveTakedown: vi.fn(),
}))

describe('TakedownManager', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('creates takedown ticket and resolves offlined', async () => {
    vi.mocked(createTakedown).mockResolvedValue()
    vi.mocked(resolveTakedown).mockResolvedValue()

    const wrapper = mount(TakedownManager)
    await wrapper.get('[data-test="target-type"]').setValue('article')
    await wrapper.get('[data-test="target-id"]').setValue('101')
    await wrapper.get('[data-test="reason"]').setValue('copyright complaint')
    await wrapper.get('[data-test="create"]').trigger('click')
    await flushPromises()

    expect(createTakedown).toHaveBeenCalledWith({
      target_type: 'article',
      target_id: 101,
      reason: 'copyright complaint',
    })

    await wrapper.get('[data-test="resolve-id"]').setValue('10')
    await wrapper.get('[data-test="action"]').setValue('offlined')
    await wrapper.get('[data-test="resolve"]').trigger('click')
    await flushPromises()

    expect(resolveTakedown).toHaveBeenCalledWith(10, 'offlined')
  })
})
