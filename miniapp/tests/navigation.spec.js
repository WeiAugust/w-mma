const { launchMiniApp } = require('./support/miniapp-harness')

test('from schedule to event card to fighter detail', async () => {
  const app = launchMiniApp()
  await app.open('/pages/schedule/index')
  await app.tap('[data-test="event-10"]')
  await app.tap('[data-test="fighter-20"]')
  expect(app.currentPage()).toBe('/pages/fighter/index?id=20')
})

test('core routes are reachable in miniapp harness', async () => {
  const app = launchMiniApp()
  await app.open('/pages/news/index')
  await app.open('/pages/schedule/index')
  await app.open('/pages/search-fighter/index')
  expect(app.currentPage()).toBe('/pages/search-fighter/index')
})

test('opening unknown route throws error', async () => {
  const app = launchMiniApp()
  await expect(app.open('/pages/unknown/index')).rejects.toThrow('unknown page route')
})
