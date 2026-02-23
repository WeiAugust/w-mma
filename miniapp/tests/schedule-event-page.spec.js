const { createPageContext } = require('./support/page-context')
const schedulePage = require('../pages/schedule/index')
const eventDetailPage = require('../pages/event-detail/index')

describe('schedule and event detail pages', () => {
  test('schedule page filters by org', async () => {
    const ctx = createPageContext(schedulePage)
    schedulePage.__setApi({
      listEvents: jest.fn().mockResolvedValue({
        items: [
          { id: 1, org: 'UFC', name: 'UFC A', poster_url: 'https://img.test/a.jpg' },
          { id: 2, org: 'ONE', name: 'ONE B', poster_url: 'https://img.test/b.jpg' },
        ],
      }),
      getEventCard: jest.fn().mockResolvedValue({ bouts: [{ id: 1001 }] }),
    })

    await schedulePage.loadEvents.call(ctx)
    schedulePage.onOrgChange.call(ctx, { detail: { value: 'UFC' } })

    expect(ctx.data.filteredItems).toHaveLength(1)
    expect(ctx.data.filteredItems[0].org).toBe('UFC')
  })

  test('schedule page filters by picker index string', async () => {
    const ctx = createPageContext(schedulePage)
    schedulePage.__setApi({
      listEvents: jest.fn().mockResolvedValue({
        items: [
          { id: 1, org: 'UFC', name: 'UFC A', poster_url: 'https://img.test/a.jpg' },
          { id: 2, org: 'ONE', name: 'ONE B', poster_url: 'https://img.test/b.jpg' },
        ],
      }),
      getEventCard: jest.fn().mockResolvedValue({ bouts: [{ id: 1001 }] }),
    })

    await schedulePage.loadEvents.call(ctx)
    schedulePage.onOrgChange.call(ctx, { detail: { value: '1' } })

    expect(ctx.data.selectedOrg).toBe('UFC')
    expect(ctx.data.filteredItems).toHaveLength(1)
    expect(ctx.data.filteredItems[0].org).toBe('UFC')
  })

  test('event detail loads card by route id', async () => {
    const ctx = createPageContext(eventDetailPage)
    eventDetailPage.__setApi({
      getEventCard: jest.fn().mockResolvedValue({
        id: 10,
        bouts: [{ id: 1001, red_fighter_id: 20, blue_fighter_id: 21 }],
      }),
    })

    await eventDetailPage.onLoad.call(ctx, { id: '10' })

    expect(ctx.data.event.id).toBe(10)
    expect(ctx.data.bouts).toHaveLength(1)
  })

  test('schedule page localizes status and formats starts_at', async () => {
    const ctx = createPageContext(schedulePage, { selectedOrg: 'ALL' })
    schedulePage.__setApi({
      listEvents: jest.fn().mockResolvedValue({
        items: [
          {
            id: 1,
            org: 'UFC',
            name: 'UFC A',
            status: 'scheduled',
            starts_at: '2026-02-21T20:30:40Z',
            poster_url: 'https://img.test/a.jpg',
          },
        ],
      }),
      getEventCard: jest.fn().mockResolvedValue({ bouts: [{ id: 1 }] }),
    })

    await schedulePage.loadEvents.call(ctx)

    expect(ctx.data.filteredItems[0].status_text).toBe('未开赛')
    expect(ctx.data.filteredItems[0].status_class).toBe('scheduled')
    expect(ctx.data.filteredItems[0].starts_at_text).toBe('2026-02-22 04:30:40')
  })

  test('schedule page hides events without poster or complete matchups', async () => {
    const ctx = createPageContext(schedulePage, { selectedOrg: 'ALL' })
    const getEventCard = jest.fn(async (eventID) => {
      if (eventID === 1) {
        return {
          main_card: [{ id: 1001, red_fighter: { id: 20, name: 'A' }, blue_fighter: { id: 21, name: 'B' } }],
          prelims: [{ id: 1002, red_fighter: { id: 22, name: 'C' }, blue_fighter: { id: 23, name: 'D' } }],
        }
      }
      if (eventID === 3) {
        return { main_card: [], prelims: [], bouts: [] }
      }
      return { main_card: [], prelims: [], bouts: [{ id: 1 }] }
    })
    schedulePage.__setApi({
      listEvents: jest.fn().mockResolvedValue({
        items: [
          { id: 1, org: 'UFC', name: 'keep', poster_url: 'https://img.test/1.jpg', status: 'scheduled', starts_at: '2026-02-21T20:30:40Z' },
          { id: 2, org: 'UFC', name: 'no-poster', poster_url: '', status: 'scheduled', starts_at: '2026-02-21T20:30:40Z' },
          { id: 3, org: 'UFC', name: 'no-bouts', poster_url: 'https://img.test/3.jpg', status: 'scheduled', starts_at: '2026-02-21T20:30:40Z' },
        ],
      }),
      getEventCard,
    })

    await schedulePage.loadEvents.call(ctx)

    expect(ctx.data.items).toHaveLength(1)
    expect(ctx.data.items[0].id).toBe(1)
    expect(getEventCard).toHaveBeenCalledTimes(2)
  })

  test('event detail groups main card and prelims with placeholders', async () => {
    const ctx = createPageContext(eventDetailPage)
    eventDetailPage.__setApi({
      getEventCard: jest.fn().mockResolvedValue({
        id: 10,
        main_card: [
          {
            id: 1001,
            weight_class: 'Flyweight',
            red_fighter: { id: 20, name: 'Manel Kape', country: 'Angola', rank: '#6', avatar_url: '' },
            blue_fighter: { id: 21, name: 'Asu Almabayev', country: '', rank: '', avatar_url: '' },
          },
        ],
        prelims: [
          {
            id: 1002,
            weight_class: '',
            red_fighter: { id: 22, name: '', country: '', rank: '', avatar_url: '' },
            blue_fighter: { id: 23, name: 'Julian Marquez', country: 'USA', rank: '#14', avatar_url: '' },
          },
        ],
      }),
    })

    await eventDetailPage.onLoad.call(ctx, { id: '10' })

    expect(ctx.data.mainCard).toHaveLength(1)
    expect(ctx.data.prelims).toHaveLength(1)
    expect(ctx.data.mainCard[0].red_fighter.rank_text).toBe('#6')
    expect(ctx.data.mainCard[0].blue_fighter.country_text).toBe('--')
    expect(ctx.data.prelims[0].weight_class_text).toBe('--')
    expect(ctx.data.prelims[0].red_fighter.name_text).toBe('--')
  })

  test('event detail localizes status weight class and result', async () => {
    const ctx = createPageContext(eventDetailPage)
    eventDetailPage.__setApi({
      getEventCard: jest.fn().mockResolvedValue({
        id: 10,
        status: 'completed',
        main_card: [
          {
            id: 1001,
            weight_class: 'Middleweight',
            result: 'KO/TKO R2 1:40',
            method: 'KO/TKO',
            round: 2,
            time_sec: 100,
            winner_id: 20,
            red_fighter: { id: 20, name: 'Gabriel Bonfim', country: 'Brazil', rank: '#14', avatar_url: '' },
            blue_fighter: { id: 21, name: 'Randy Brown', country: 'Jamaica', rank: '', avatar_url: '' },
          },
        ],
        prelims: [],
      }),
    })

    await eventDetailPage.onLoad.call(ctx, { id: '10' })

    expect(ctx.data.event.status_text).toBe('已结束')
    expect(ctx.data.mainCard[0].weight_class_text).toBe('中量级')
    expect(ctx.data.mainCard[0].result_text).toBe('终结：Gabriel Bonfim获胜（KO/TKO） · 第2回合 · 1:40')
    expect(ctx.data.mainCard[0].red_fighter_state_class).toBe('fighter--winner')
    expect(ctx.data.mainCard[0].blue_fighter_state_class).toBe('fighter--loser')
    expect(ctx.data.mainCard[0].red_fighter_state_label).toBe('胜')
    expect(ctx.data.mainCard[0].blue_fighter_state_label).toBe('负')
  })

  test('event detail renders judge decision winner text', async () => {
    const ctx = createPageContext(eventDetailPage)
    eventDetailPage.__setApi({
      getEventCard: jest.fn().mockResolvedValue({
        id: 10,
        status: 'completed',
        main_card: [
          {
            id: 1001,
            method: 'Unanimous Decision',
            round: 3,
            time_sec: 300,
            winner_id: 21,
            red_fighter: { id: 20, name: 'Fighter A', country: 'USA', rank: '#2', avatar_url: '' },
            blue_fighter: { id: 21, name: 'Fighter B', country: 'Brazil', rank: '#1', avatar_url: '' },
          },
        ],
        prelims: [],
      }),
    })

    await eventDetailPage.onLoad.call(ctx, { id: '10' })

    expect(ctx.data.mainCard[0].result_text).toBe('裁判判定：Fighter B获胜（一致判定）')
  })

  test('event detail explains special case when completed bout has no winner', async () => {
    const ctx = createPageContext(eventDetailPage)
    eventDetailPage.__setApi({
      getEventCard: jest.fn().mockResolvedValue({
        id: 10,
        status: 'completed',
        main_card: [
          {
            id: 1001,
            method: 'No Contest',
            round: 1,
            time_sec: 90,
            winner_id: 0,
            red_fighter: { id: 20, name: 'Fighter A', country: 'USA', rank: '#2', avatar_url: '' },
            blue_fighter: { id: 21, name: 'Fighter B', country: 'Brazil', rank: '#1', avatar_url: '' },
          },
        ],
        prelims: [],
      }),
    })

    await eventDetailPage.onLoad.call(ctx, { id: '10' })

    expect(ctx.data.mainCard[0].result_text).toBe('特殊情况：无结果（No Contest）')
  })
})
