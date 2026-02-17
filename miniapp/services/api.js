const API_BASE_URL = process.env.MINIAPP_API_BASE_URL || 'http://localhost:8080'

function request(path) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}${path}`,
      method: 'GET',
      success(res) {
        resolve(res.data)
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

function listEvents() {
  return request('/api/events')
}

function getEventCard(eventId) {
  return request(`/api/events/${eventId}`)
}

function searchFighters(keyword) {
  return request(`/api/fighters/search?q=${encodeURIComponent(keyword)}`)
}

function getFighterDetail(fighterId) {
  return request(`/api/fighters/${fighterId}`)
}

module.exports = {
  listEvents,
  getEventCard,
  searchFighters,
  getFighterDetail,
}
