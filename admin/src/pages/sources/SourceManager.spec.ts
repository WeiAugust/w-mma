import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import SourceManager from './SourceManager.vue'
import { createSource, listSources, toggleSource, updateSource } from '../../api/sources'

vi.mock('../../api/sources', () => ({
  listSources: vi.fn(),
  createSource: vi.fn(),
  updateSource: vi.fn(),
  toggleSource: vi.fn(),
}))

describe('SourceManager', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('creates source and updates rights flags', async () => {
    vi.mocked(listSources).mockResolvedValue([
      {
        id: 1,
        name: 'douyin-1',
        source_type: 'news',
        platform: 'douyin',
        account_id: 'a',
        source_url: 'https://www.douyin.com/a',
        parser_kind: 'generic',
        enabled: true,
        rights_display: true,
        rights_playback: false,
        rights_ai_summary: true,
        rights_expires_at: '',
        rights_proof_url: '',
      },
    ])
    vi.mocked(createSource).mockResolvedValue({
      id: 2,
      name: 'ufc-events',
      source_type: 'schedule',
      platform: 'ufc',
      account_id: '',
      source_url: 'https://www.ufc.com/events',
      parser_kind: 'generic',
      enabled: true,
      rights_display: true,
      rights_playback: false,
      rights_ai_summary: false,
      rights_expires_at: '',
      rights_proof_url: '',
    })
    vi.mocked(updateSource).mockResolvedValue()
    vi.mocked(toggleSource).mockResolvedValue()

    const wrapper = mount(SourceManager)
    await flushPromises()

    await wrapper.get('[data-test="name"]').setValue('ufc-events')
    await wrapper.get('[data-test="source-type"]').setValue('schedule')
    await wrapper.get('[data-test="platform"]').setValue('ufc')
    await wrapper.get('[data-test="source-url"]').setValue('https://www.ufc.com/events')
    await wrapper.get('[data-test="create"]').trigger('click')

    expect(createSource).toHaveBeenCalled()
    expect(wrapper.text()).toContain('ufc-events')

    await wrapper.get('[data-test="playback-1"]').setValue(true)
    expect(updateSource).toHaveBeenCalledWith(1, { rights_playback: true })
  })
})
