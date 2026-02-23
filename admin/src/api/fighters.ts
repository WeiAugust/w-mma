import { request } from './request'

export type FighterItem = {
  id: number
  name: string
  name_zh?: string
  nickname?: string
  country?: string
  record?: string
  weight_class?: string
  avatar_url?: string
  intro_video_url?: string
  stats?: Record<string, string>
  records?: Record<string, string>
  updates?: string[]
}

export type ManualFighterPayload = {
  source_id: number
  name: string
  name_zh?: string
  nickname?: string
  country?: string
  record?: string
  weight_class?: string
  avatar_url?: string
  intro_video_url?: string
}

export async function createManualFighter(payload: ManualFighterPayload): Promise<void> {
  await request('/admin/fighters/manual', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function searchFighters(keyword: string): Promise<FighterItem[]> {
  const data = await request<{ items: FighterItem[] }>(
    `/api/fighters/search?q=${encodeURIComponent(keyword || '')}`,
  )
  return data.items || []
}
