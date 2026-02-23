const appConfig = require('../app.json')

test('registers all mvp pages in app config', () => {
  expect(appConfig.pages).toEqual([
    'pages/schedule/index',
    'pages/event-detail/index',
    'pages/fighter/index',
    'pages/search-fighter/index',
  ])
})
