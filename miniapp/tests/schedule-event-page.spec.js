const { createPageContext } = require('./support/page-context')
const schedulePage = require('../pages/schedule/index')
const eventDetailPage = require('../pages/event-detail/index')

describe('schedule and event detail pages', () => {
  test('schedule page filters by org', async () => {
    const ctx = createPageContext(schedulePage)
    schedulePage.__setApi({
      listEvents: jest.fn().mockResolvedValue({
        items: [
          { id: 1, org: 'UFC', name: 'UFC A' },
          { id: 2, org: 'ONE', name: 'ONE B' },
        ],
      }),
    })

    await schedulePage.loadEvents.call(ctx)
    schedulePage.onOrgChange.call(ctx, { detail: { value: 'UFC' } })

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
})
