import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import SourceManager from './SourceManager.vue'
import {
  createSource,
  deleteSource,
  listSources,
  restoreSource,
  toggleSource,
  triggerIngestFetch,
  updateSource,
  type SourceItem,
  type SourceListQuery,
} from '../../api/sources'

vi.mock('../../api/sources', () => ({
  listSources: vi.fn(),
  createSource: vi.fn(),
  updateSource: vi.fn(),
  toggleSource: vi.fn(),
  deleteSource: vi.fn(),
  restoreSource: vi.fn(),
  triggerIngestFetch: vi.fn(),
}))

const ACTIVE_SOURCE: SourceItem = {
  id: 1,
  name: 'UFC 官方赛程',
  source_type: 'schedule',
  platform: 'ufc',
  account_id: '',
  source_url: 'https://www.ufc.com/events',
  parser_kind: 'ufc_schedule',
  enabled: true,
  is_builtin: true,
  rights_display: true,
  rights_playback: false,
  rights_ai_summary: true,
  rights_expires_at: '',
  rights_proof_url: '',
  deleted_at: '',
}

const DELETED_SOURCE: SourceItem = {
  id: 2,
  name: 'WBC 官方赛程',
  source_type: 'schedule',
  platform: 'wbc',
  account_id: '',
  source_url: 'https://www.wbcboxing.com/',
  parser_kind: 'wbc_schedule',
  enabled: true,
  is_builtin: true,
  rights_display: true,
  rights_playback: false,
  rights_ai_summary: true,
  rights_expires_at: '',
  rights_proof_url: '',
  deleted_at: '2026-02-17T10:00:00Z',
}

describe('SourceManager', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('supports operator-friendly full lifecycle workflow', async () => {
    vi.mocked(listSources).mockImplementation(async (query?: SourceListQuery) => {
      if (query?.include_deleted) {
        return [ACTIVE_SOURCE, DELETED_SOURCE]
      }
      return [ACTIVE_SOURCE]
    })
    vi.mocked(createSource).mockResolvedValue({
      ...ACTIVE_SOURCE,
      id: 3,
      name: 'ONE 官方赛程',
      platform: 'one',
      parser_kind: 'one_schedule',
      source_url: 'https://www.onefc.com/events/',
    })
    vi.mocked(updateSource).mockResolvedValue()
    vi.mocked(toggleSource).mockResolvedValue()
    vi.mocked(deleteSource).mockResolvedValue()
    vi.mocked(restoreSource).mockResolvedValue()
    vi.mocked(triggerIngestFetch).mockResolvedValue()

    const wrapper = mount(SourceManager)
    await flushPromises()

    expect(wrapper.text()).toContain('总数据源')
    expect(listSources).toHaveBeenCalledWith({
      include_deleted: false,
      platform: '',
      source_type: undefined,
      enabled: undefined,
      is_builtin: undefined,
    })

    await wrapper.get('[data-test="filter-platform"]').setValue('ufc')
    await wrapper.get('[data-test="apply-filters"]').trigger('click')
    expect(listSources).toHaveBeenLastCalledWith({
      include_deleted: false,
      platform: 'ufc',
      source_type: undefined,
      enabled: undefined,
      is_builtin: undefined,
    })

    await wrapper.get('[data-test="edit-1"]').trigger('click')
    await wrapper.get('[data-test="edit-rights-playback"]').setValue(true)
    await wrapper.get('[data-test="save-edit"]').trigger('click')
    expect(updateSource).toHaveBeenCalledWith(1, expect.objectContaining({ rights_playback: true }))

    await wrapper.get('[data-test="toggle-1"]').trigger('click')
    expect(toggleSource).toHaveBeenCalledWith(1)
    await flushPromises()

    await wrapper.get('[data-test="sync-1"]').trigger('click')
    expect(triggerIngestFetch).toHaveBeenCalledWith(1)
    expect(wrapper.text()).toContain('同步任务已触发')
    await flushPromises()

    await wrapper.get('[data-test="delete-1"]').trigger('click')
    expect(deleteSource).toHaveBeenCalledWith(1)

    await wrapper.get('[data-test="include-deleted"]').setValue(true)
    await wrapper.get('[data-test="apply-filters"]').trigger('click')
    expect(listSources).toHaveBeenLastCalledWith({
      include_deleted: true,
      platform: 'ufc',
      source_type: undefined,
      enabled: undefined,
      is_builtin: undefined,
    })
    await flushPromises()
    await wrapper.get('[data-test="restore-2"]').trigger('click')
    expect(restoreSource).toHaveBeenCalledWith(2)

    await wrapper.get('[data-test="open-create"]').trigger('click')
    await wrapper.get('[data-test="create-name"]').setValue('ONE 官方赛程')
    await wrapper.get('[data-test="create-source-type"]').setValue('schedule')
    await wrapper.get('[data-test="create-platform"]').setValue('one')
    await wrapper.get('[data-test="create-source-url"]').setValue('https://www.onefc.com/events/')
    await wrapper.get('[data-test="submit-create"]').trigger('click')
    expect(createSource).toHaveBeenCalled()
  })
})
