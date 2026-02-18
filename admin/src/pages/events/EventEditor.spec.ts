import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import EventEditor from './EventEditor.vue'
import { listEvents, updateEvent } from '../../api/events'

vi.mock('../../api/events', () => ({
  listEvents: vi.fn(),
  updateEvent: vi.fn(),
}))

describe('EventEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('supports schedule console filtering and save', async () => {
    vi.mocked(listEvents).mockResolvedValue([
      { id: 1, org: 'UFC', name: 'UFC Fight Night', status: 'scheduled', starts_at: '2026-03-01T00:00:00Z' },
      { id: 2, org: 'ONE', name: 'ONE 173', status: 'live', starts_at: '2026-03-02T00:00:00Z' },
    ])
    vi.mocked(updateEvent).mockResolvedValue()

    const wrapper = mount(EventEditor)
    await flushPromises()

    expect(wrapper.text()).toContain('赛程控制台')
    expect(wrapper.text()).toContain('赛事总数')

    await wrapper.get('[data-test="filter-org"]').setValue('UFC')
    await wrapper.get('[data-test="apply-filter"]').trigger('click')
    expect(wrapper.text()).toContain('UFC Fight Night')
    expect(wrapper.text()).not.toContain('ONE 173')

    await wrapper.get('[data-test="edit-1"]').trigger('click')
    await wrapper.get('[data-test="edit-name"]').setValue('UFC Fight Night 200')
    await wrapper.get('[data-test="edit-status"]').setValue('live')
    await wrapper.get('[data-test="save-edit"]').trigger('click')

    expect(updateEvent).toHaveBeenCalledWith(1, { name: 'UFC Fight Night 200', status: 'live' })
  })
})
