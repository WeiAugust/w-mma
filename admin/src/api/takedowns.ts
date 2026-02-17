import { request } from './request'

export type CreateTakedownPayload = {
  target_type: 'article' | 'event' | 'fighter'
  target_id: number
  reason: string
  complainant?: string
  evidence_url?: string
}

export async function createTakedown(payload: CreateTakedownPayload): Promise<void> {
  await request('/admin/takedowns', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function resolveTakedown(takedownID: number, action: 'offlined' | 'rejected'): Promise<void> {
  await request(`/admin/takedowns/${takedownID}/resolve`, {
    method: 'POST',
    body: JSON.stringify({ action }),
  })
}
