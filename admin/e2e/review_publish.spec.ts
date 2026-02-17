import { describe, expect, it, vi } from 'vitest'

import { approvePending, listPending } from '../src/api/review'

describe('review publish flow', () => {
  it('lists pending then approves one item', async () => {
    const pending = [{ id: 1, title: 'news-a' }]

    globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)
      if (url.endsWith('/admin/review/pending')) {
        return new Response(JSON.stringify({ items: pending }), { status: 200 })
      }

      if (url.includes('/admin/review/1/approve') && init?.method === 'POST') {
        pending.splice(0, 1)
        return new Response(JSON.stringify({ ok: true }), { status: 200 })
      }

      return new Response(JSON.stringify({ error: 'not found' }), { status: 404 })
    }) as unknown as typeof fetch

    const before = await listPending()
    expect(before).toHaveLength(1)

    await approvePending(1)

    const after = await listPending()
    expect(after).toHaveLength(0)
  })
})
