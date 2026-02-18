import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import FighterManager from './FighterManager.vue'
import { createManualFighter, searchFighters } from '../../api/fighters'

vi.mock('../../api/fighters', () => ({
  createManualFighter: vi.fn(),
  searchFighters: vi.fn(),
}))

describe('FighterManager', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('supports search + create + quick query workflow', async () => {
    vi.mocked(searchFighters)
      .mockResolvedValueOnce([
        { id: 1, name: 'Alex Pereira', country: 'Brazil', record: '10-2-0' },
        { id: 2, name: 'Asu Almabayev', country: 'Kazakhstan', record: '21-2-0' },
      ])
      .mockResolvedValue([
        { id: 3, name: 'Yadong Song', country: 'China', record: '21-8-1' },
      ])
    vi.mocked(createManualFighter).mockResolvedValue()

    const wrapper = mount(FighterManager)
    await flushPromises()
    expect(wrapper.text()).toContain('选手库数量')
    expect(wrapper.text()).toContain('Alex Pereira')

    await wrapper.get('[data-test="search-keyword"]').setValue('song')
    await wrapper.get('[data-test="run-search"]').trigger('click')
    await flushPromises()
    expect(searchFighters).toHaveBeenLastCalledWith('song')
    expect(wrapper.text()).toContain('Yadong Song')

    await wrapper.get('[data-test="open-create"]').trigger('click')
    await wrapper.get('[data-test="create-source-id"]').setValue('6')
    await wrapper.get('[data-test="create-name"]').setValue('Su Mudaerji')
    await wrapper.get('[data-test="create-country"]').setValue('China')
    await wrapper.get('[data-test="create-record"]').setValue('17-6-0')
    await wrapper.get('[data-test="submit-create"]').trigger('click')
    await flushPromises()
    expect(createManualFighter).toHaveBeenCalledWith(expect.objectContaining({ name: 'Su Mudaerji' }))
  })
})
