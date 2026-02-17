const {
  listArticles,
  listEvents,
  getEventCard,
  searchFighters,
  getFighterDetail,
} = require('../services/api')

describe('miniapp api client', () => {
  test('listArticles requests backend articles endpoint', async () => {
    global.wx = {
      request: jest.fn(({ success }) => success({ data: { items: [] } })),
    }

    await listArticles()

    expect(global.wx.request).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'http://localhost:8080/api/articles', method: 'GET' }),
    )
  })

  test('searchFighters encodes query string', async () => {
    global.wx = {
      request: jest.fn(({ success }) => success({ data: { items: [] } })),
    }

    await searchFighters('Alex Pereira')

    expect(global.wx.request).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'http://localhost:8080/api/fighters/search?q=Alex%20Pereira',
      }),
    )
  })

  test('exports all required api methods', () => {
    expect(typeof listEvents).toBe('function')
    expect(typeof getEventCard).toBe('function')
    expect(typeof getFighterDetail).toBe('function')
  })
})
