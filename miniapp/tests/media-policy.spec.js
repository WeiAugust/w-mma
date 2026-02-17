const { createPageContext } = require('./support/page-context')
const newsPage = require('../pages/news/index')

describe('media policy', () => {
  test('news card falls back when can_play is false', async () => {
    const ctx = createPageContext(newsPage)
    newsPage.__setApi({
      listArticles: jest.fn().mockResolvedValue({
        items: [
          {
            id: 1,
            title: 'news-a',
            summary: 'summary-a',
            can_play: false,
            video_url: 'https://video.example.com/a.mp4',
          },
        ],
      }),
    })

    await newsPage.loadArticles.call(ctx)

    expect(ctx.data.items).toHaveLength(1)
    expect(ctx.data.items[0].can_play).toBe(false)
    expect(ctx.data.items[0].video_url).toBe('')
  })
})
