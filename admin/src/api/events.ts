import { request } from './request'

export type EventItem = {
  id: number
  org: string
  name: string
  status: 'scheduled' | 'live' | 'completed'
  starts_at: string
}

export async function listEvents(): Promise<EventItem[]> {
  const data = await request<{ items: EventItem[] }>('/admin/events')
  return data.items || []
}

export async function updateEvent(
  id: number,
  patch: Partial<Pick<EventItem, 'status' | 'name'>>,
): Promise<void> {
  await request(`/admin/events/${id}`, {
    method: 'PUT',
    body: JSON.stringify(patch),
  })
}
