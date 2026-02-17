const { createPageContext } = require('./support/page-context')
const newsPage = require('../pages/news/index')

describe('news page', () => {
  test('loadArticles writes items on success', async () => {
    const ctx = createPageContext(newsPage)
    newsPage.__setApi({
      listArticles: jest.fn().mockResolvedValue({
        items: [{ id: 1, title: 'news-a', summary: 'summary-a', can_play: true, video_url: 'https://video.example.com/a.mp4' }],
      }),
    })

    await newsPage.loadArticles.call(ctx)

    expect(ctx.data.items).toHaveLength(1)
    expect(ctx.data.items[0].can_play).toBe(true)
    expect(ctx.data.error).toBe('')
    expect(ctx.data.loading).toBe(false)
  })

  test('loadArticles writes error on failure', async () => {
    const ctx = createPageContext(newsPage)
    newsPage.__setApi({
      listArticles: jest.fn().mockRejectedValue(new Error('network down')),
    })

    await newsPage.loadArticles.call(ctx)

    expect(ctx.data.items).toHaveLength(0)
    expect(ctx.data.error).toBe('network down')
    expect(ctx.data.loading).toBe(false)
  })
})
