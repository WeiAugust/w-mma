const DEFAULT_API_BASE_URL = 'http://localhost:8080'

let apiBaseUrl = DEFAULT_API_BASE_URL

function setApiBaseUrl(url) {
  apiBaseUrl = url || DEFAULT_API_BASE_URL
}

function request(path, options = {}) {
  const method = options.method || 'GET'
  const data = options.data || undefined
  const timeout = options.timeout || 8000

  return new Promise((resolve, reject) => {
    wx.request({
      url: `${apiBaseUrl}${path}`,
      method,
      data,
      timeout,
      success(res) {
        if (res.statusCode && res.statusCode >= 400) {
          reject(new Error(`request failed: ${res.statusCode}`))
          return
        }
        resolve(res.data)
      },
      fail(err) {
        reject(new Error((err && err.errMsg) || 'request failed'))
      },
    })
  })
}

function listArticles() {
  return request('/api/articles')
}

function listEvents() {
  return request('/api/events')
}

function getEventCard(eventId) {
  return request(`/api/events/${eventId}`)
}

function searchFighters(keyword) {
  return request(`/api/fighters/search?q=${encodeURIComponent(keyword || '')}`)
}

function getFighterDetail(fighterId) {
  return request(`/api/fighters/${fighterId}`)
}

module.exports = {
  request,
  setApiBaseUrl,
  listArticles,
  listEvents,
  getEventCard,
  searchFighters,
  getFighterDetail,
}
