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
  rights_display: boolean
  rights_playback: boolean
  rights_ai_summary: boolean
  rights_expires_at?: string
  rights_proof_url?: string
}

export type SourcePayload = Omit<SourceItem, 'id'>

export async function listSources(): Promise<SourceItem[]> {
  const data = await request<{ items: SourceItem[] }>('/admin/sources')
  return data.items || []
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
