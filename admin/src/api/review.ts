import { request } from './request'

export type PendingItem = {
  id: number
  title: string
}

export async function listPending(): Promise<PendingItem[]> {
  const data = await request<{ items: PendingItem[] }>('/admin/review/pending')
  return data.items || []
}

export async function approvePending(id: number): Promise<void> {
  await request(`/admin/review/${id}/approve?reviewer_id=9001`, { method: 'POST' })
}
