const { createPageContext } = require('./support/page-context')
const fighterPage = require('../pages/fighter/index')
const searchPage = require('../pages/search-fighter/index')

describe('fighter pages', () => {
  test('fighter page loads detail by id', async () => {
    const ctx = createPageContext(fighterPage)
    fighterPage.__setApi({
      getFighterDetail: jest.fn().mockResolvedValue({
        id: 20,
        name: 'Alex Pereira',
        stats: { 'Sig. Str. Landed': '5.45' },
        records: { 'Wins by Knockout': '8' },
      }),
    })

    await fighterPage.onLoad.call(ctx, { id: '20' })

    expect(ctx.data.fighter.name).toBe('Alex Pereira')
    expect(ctx.data.statItems).toEqual([{ label: 'Sig. Str. Landed', value: '5.45' }])
    expect(ctx.data.recordItems).toEqual([{ label: 'Wins by Knockout', value: '8' }])
    expect(ctx.data.loading).toBe(false)
  })

  test('search page navigates to fighter detail on select', () => {
    const ctx = createPageContext(searchPage)
    global.wx = { navigateTo: jest.fn() }

    searchPage.onSelectFighter.call(ctx, { currentTarget: { dataset: { id: 20 } } })

    expect(global.wx.navigateTo).toHaveBeenCalledWith({ url: '/pages/fighter/index?id=20' })
  })
})
