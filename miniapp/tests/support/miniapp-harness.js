const schedulePage = require('../../pages/schedule/index')
const eventDetailPage = require('../../pages/event-detail/index')

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
