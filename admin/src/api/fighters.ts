import { request } from './request'

export type ManualFighterPayload = {
  source_id: number
  name: string
  country?: string
  record?: string
  avatar_url?: string
  intro_video_url?: string
}

export async function createManualFighter(payload: ManualFighterPayload): Promise<void> {
  await request('/admin/fighters/manual', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
