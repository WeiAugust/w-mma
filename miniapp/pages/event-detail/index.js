const defaultApi = require('../../services/api')

let api = defaultApi

function toNumber(value) {
  const n = Number(value)
  return Number.isNaN(n) ? 0 : n
}

async function loadEventCardWithContext(ctx, eventID) {
  ctx.setData({ loading: true, error: '' })

  try {
    const event = await api.getEventCard(eventID)
    const bouts = Array.isArray(event && event.bouts) ? event.bouts : []
    ctx.setData({
      loading: false,
      error: '',
      event,
      bouts,
    })
  } catch (err) {
    ctx.setData({
      loading: false,
      event: null,
      bouts: [],
      error: (err && err.message) || '战卡加载失败',
    })
  }
}

const pageDef = {
  data: {
    loading: false,
    error: '',
    eventID: 0,
    event: null,
    bouts: [],
  },

  async onLoad(options = {}) {
    const eventID = toNumber(options.id)
    this.setData({ eventID })
    if (!eventID) {
      this.setData({ error: '无效赛事 ID' })
      return
    }
    await loadEventCardWithContext(this, eventID)
  },

  async onPullDownRefresh() {
    if (this.data.eventID) {
      await loadEventCardWithContext(this, this.data.eventID)
    }
    if (typeof wx.stopPullDownRefresh === 'function') {
      wx.stopPullDownRefresh()
    }
  },

  async onRetryTap() {
    if (!this.data.eventID) {
      return
    }
    await loadEventCardWithContext(this, this.data.eventID)
  },

  async loadEventCard(eventID) {
    await loadEventCardWithContext(this, eventID)
  },

  onFighterTap(eventOrID) {
    const fighterID =
      typeof eventOrID === 'number'
        ? eventOrID
        : eventOrID && eventOrID.currentTarget && eventOrID.currentTarget.dataset && eventOrID.currentTarget.dataset.id

    if (!fighterID) {
      return
    }

    wx.navigateTo({
      url: `/pages/fighter/index?id=${fighterID}`,
    })
  },

  __setApi(nextApi) {
    api = nextApi
  },

  __resetApi() {
    api = defaultApi
  },
}

if (typeof Page === 'function') {
  Page(pageDef)
}

module.exports = pageDef
