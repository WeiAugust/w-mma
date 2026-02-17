import { request } from './request'

export type ArticleItem = {
  id: number
  source_id: number
  title: string
  summary: string
  source_url: string
  can_play: boolean
}

export type ManualArticlePayload = {
  source_id: number
  title: string
  summary: string
  source_url: string
  cover_url?: string
  video_url?: string
}

export async function createManualArticle(payload: ManualArticlePayload): Promise<void> {
  await request('/admin/articles/manual', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function createSummaryJob(articleID: number, sourceID: number): Promise<void> {
  await request(`/admin/articles/${articleID}/summarize`, {
    method: 'POST',
    body: JSON.stringify({ source_id: sourceID }),
  })
}

export async function listPublishedArticles(): Promise<ArticleItem[]> {
  const data = await request<{ items: ArticleItem[] }>('/api/articles')
  return data.items || []
}
