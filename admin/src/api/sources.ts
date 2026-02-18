import { request } from './request'

export type SourceType = 'news' | 'schedule' | 'fighter'

export type SourceItem = {
  id: number
  name: string
  source_type: SourceType
  platform: string
  account_id?: string
  source_url: string
  parser_kind: string
  enabled: boolean
  is_builtin: boolean
  rights_display: boolean
  rights_playback: boolean
  rights_ai_summary: boolean
  rights_expires_at?: string
  rights_proof_url?: string
  deleted_at?: string
}

export type SourcePayload = Omit<SourceItem, 'id' | 'deleted_at'>

export type SourceListQuery = {
  include_deleted?: boolean
  source_type?: SourceType
  platform?: string
  enabled?: boolean
  is_builtin?: boolean
}

function toQueryString(query?: SourceListQuery): string {
  if (!query) {
    return ''
  }

  const params = new URLSearchParams()
  if (typeof query.include_deleted === 'boolean') {
    params.set('include_deleted', String(query.include_deleted))
  }
  if (query.source_type) {
    params.set('source_type', query.source_type)
  }
  if (query.platform) {
    params.set('platform', query.platform)
  }
  if (typeof query.enabled === 'boolean') {
    params.set('enabled', String(query.enabled))
  }
  if (typeof query.is_builtin === 'boolean') {
    params.set('is_builtin', String(query.is_builtin))
  }

  const text = params.toString()
  return text ? `?${text}` : ''
}

export async function listSources(query?: SourceListQuery): Promise<SourceItem[]> {
  const data = await request<{ items: SourceItem[] }>(`/admin/sources${toQueryString(query)}`)
  return data.items || []
}

export async function getSource(id: number, includeDeleted = false): Promise<SourceItem> {
  const suffix = includeDeleted ? '?include_deleted=true' : ''
  return request<SourceItem>(`/admin/sources/${id}${suffix}`)
}

export async function createSource(payload: SourcePayload): Promise<SourceItem> {
  return request<SourceItem>('/admin/sources', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function updateSource(id: number, patch: Partial<SourcePayload>): Promise<void> {
  await request(`/admin/sources/${id}`, {
    method: 'PUT',
    body: JSON.stringify(patch),
  })
}

export async function toggleSource(id: number): Promise<void> {
  await request(`/admin/sources/${id}/toggle`, {
    method: 'POST',
  })
}

export async function deleteSource(id: number): Promise<void> {
  await request(`/admin/sources/${id}`, {
    method: 'DELETE',
  })
}

export async function restoreSource(id: number): Promise<void> {
  await request(`/admin/sources/${id}/restore`, {
    method: 'POST',
  })
}

export async function triggerIngestFetch(sourceID: number): Promise<void> {
  await request(`/admin/sources/${sourceID}/sync`, {
    method: 'POST',
  })
}
