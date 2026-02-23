const schedulePage = require('../../pages/schedule/index')
const eventDetailPage = require('../../pages/event-detail/index')

const knownRoutes = new Set([
  '/pages/schedule/index',
  '/pages/event-detail/index',
  '/pages/fighter/index',
  '/pages/search-fighter/index',
])

function normalizeRoute(path) {
  return String(path || '').split('?')[0]
}

function launchMiniApp() {
  let current = ''

  global.wx = {
    navigateTo({ url }) {
      current = url
    },
    request() {},
  }

  return {
    async open(path) {
      const normalized = normalizeRoute(path)
      if (!knownRoutes.has(normalized)) {
        throw new Error(`unknown page route: ${path}`)
      }
      current = path
    },
    async tap(selector) {
      if (current === '/pages/schedule/index' && selector === '[data-test="event-10"]') {
        schedulePage.onEventTap(10)
        return
      }

      if (current.startsWith('/pages/event-detail/index') && selector === '[data-test="fighter-20"]') {
        eventDetailPage.onFighterTap(20)
      }
    },
    currentPage() {
      return current
    },
  }
}

module.exports = {
  launchMiniApp,
}
