describe('miniapp api client', () => {
  function loadApi() {
    jest.resetModules()
    return require('../services/api')
  }

  beforeEach(() => {
    global.wx = undefined
  })

  test('listArticles requests backend articles endpoint', async () => {
    const { listArticles } = loadApi()
    global.wx = {
      request: jest.fn(({ success }) => success({ data: { items: [] } })),
    }

    await listArticles()

    expect(global.wx.request).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'https://localhost:8443/api/articles', method: 'GET' }),
    )
  })

  test('searchFighters encodes query string', async () => {
    const { searchFighters } = loadApi()
    global.wx = {
      request: jest.fn(({ success }) => success({ data: { items: [] } })),
    }

    await searchFighters('Alex Pereira')

    expect(global.wx.request).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8443/api/fighters/search?q=Alex%20Pereira',
      }),
    )
  })

  test('release build uses production api base url', async () => {
    global.wx = {
      getAccountInfoSync: jest.fn(() => ({ miniProgram: { envVersion: 'release' } })),
      request: jest.fn(({ success }) => success({ data: { items: [] } })),
    }
    const { listEvents } = loadApi()

    await listEvents()

    expect(global.wx.request).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'https://api.example.com/api/events', method: 'GET' }),
    )
  })

  test('exports all required api methods', () => {
    const { listEvents, getEventCard, getFighterDetail } = loadApi()
    expect(typeof listEvents).toBe('function')
    expect(typeof getEventCard).toBe('function')
    expect(typeof getFighterDetail).toBe('function')
  })
})
