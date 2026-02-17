const { launchMiniApp } = require('./support/miniapp-harness')

test('from schedule to event card to fighter detail', async () => {
  const app = launchMiniApp()
  await app.open('/pages/schedule/index')
  await app.tap('[data-test="event-10"]')
  await app.tap('[data-test="fighter-20"]')
  expect(app.currentPage()).toBe('/pages/fighter/index?id=20')
})
