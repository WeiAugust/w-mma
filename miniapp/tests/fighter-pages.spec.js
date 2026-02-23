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
    expect(ctx.data.statItems).toEqual([{ label: '有效击打命中', value: '5.45' }])
    expect(ctx.data.recordItems).toEqual([{ label: 'KO/TKO获胜', value: '8' }])
    expect(ctx.data.loading).toBe(false)
  })

  test('fighter page formats bilingual country and sanitizes nickname', async () => {
    const ctx = createPageContext(fighterPage)
    fighterPage.__setApi({
      getFighterDetail: jest.fn().mockResolvedValue({
        id: 20,
        name: 'Alex Pereira',
        nickname: '&quot;Poatan&quot;',
        country: 'Brazil',
        country_zh: '巴西',
        stats: { Height: `6' 4"` },
      }),
    })

    await fighterPage.onLoad.call(ctx, { id: '20' })

    expect(ctx.data.countryDisplay).toBe('巴西 / Brazil')
    expect(ctx.data.fighter.nickname).toBe('Poatan')
  })

  test('fighter page translates key labels and exposes hero tags', async () => {
    const ctx = createPageContext(fighterPage)
    fighterPage.__setApi({
      getFighterDetail: jest.fn().mockResolvedValue({
        id: 20,
        name: 'Alex Pereira',
        weight_class: 'Light Heavyweight',
        stats: {
          'PFP Rank': '#5',
          'Athlete Status': 'Active',
          'Title Status': 'Title Holder',
          Height: '76.00',
          'Fighting style': 'Kickboxer',
        },
        updates: ['2025-10-04 · 胜 · KO/TKO终结 · 第1回合 1:20'],
      }),
    })

    await fighterPage.onLoad.call(ctx, { id: '20' })

    expect(ctx.data.weightClassDisplay).toBe('轻重量级')
    expect(ctx.data.pfpTag).toBe('#5 P4P')
    expect(ctx.data.statusTag).toBe('现役')
    expect(ctx.data.titleTag).toBe('现任冠军')
    expect(ctx.data.statItems).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ label: '身高', value: '193.0 厘米（76.00 英寸）' }),
        expect.objectContaining({ label: '技术风格', value: '踢拳' }),
      ]),
    )
    expect(ctx.data.historyItems).toEqual([
      {
        raw: '2025-10-04 · 胜 · KO/TKO终结 · 第1回合 1:20',
        display: '2025-10-04 · KO/TKO终结 · 第1回合 1:20',
        result: '胜',
        resultClass: 'result-win',
      },
    ])
  })

  test('fighter page marks China fighters for themed hero background', async () => {
    const ctx = createPageContext(fighterPage)
    fighterPage.__setApi({
      getFighterDetail: jest.fn().mockResolvedValue({
        id: 30,
        name: 'Zhang Weili',
        country: 'China',
        country_zh: '中国',
        stats: {},
        records: {},
      }),
    })

    await fighterPage.onLoad.call(ctx, { id: '30' })

    expect(ctx.data.isChinaFighter).toBe(true)
  })

  test('search page navigates to fighter detail on select', () => {
    const ctx = createPageContext(searchPage)
    global.wx = { navigateTo: jest.fn() }

    searchPage.onSelectFighter.call(ctx, { currentTarget: { dataset: { id: 20 } } })

    expect(global.wx.navigateTo).toHaveBeenCalledWith({ url: '/pages/fighter/index?id=20' })
  })

  test('search page annotates country with bilingual display', async () => {
    const ctx = createPageContext(searchPage)
    searchPage.__setApi({
      searchFighters: jest.fn().mockResolvedValue({
        items: [{ id: 20, name: 'Alex Pereira', country: 'Brazil', country_zh: '巴西' }],
      }),
    })

    await searchPage.search.call(ctx, 'alex')

    expect(ctx.data.items[0].country_display).toBe('巴西 / Brazil')
  })
})
