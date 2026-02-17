const defaultApi = require('../../services/api')

let api = defaultApi

function formatUpdatedAt(date = new Date()) {
  const pad = (value) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const pageDef = {
  data: {
    loading: false,
    error: '',
    items: [],
    updatedAtText: '',
  },

  async onLoad() {
    await this.loadArticles()
  },

  async onPullDownRefresh() {
    await this.loadArticles()
    if (typeof wx.stopPullDownRefresh === 'function') {
      wx.stopPullDownRefresh()
    }
  },

  async onRetryTap() {
    await this.loadArticles()
  },

  async loadArticles() {
    this.setData({ loading: true, error: '' })

    try {
      const data = await api.listArticles()
      const items = Array.isArray(data && data.items) ? data.items : []
      this.setData({
        loading: false,
        error: '',
        items,
        updatedAtText: formatUpdatedAt(),
      })
    } catch (err) {
      this.setData({
        loading: false,
        items: [],
        error: (err && err.message) || '资讯加载失败',
      })
    }
  },

  onCopySourceTap(event) {
    const url = event && event.currentTarget && event.currentTarget.dataset && event.currentTarget.dataset.url
    if (!url || typeof wx.setClipboardData !== 'function') {
      return
    }

    wx.setClipboardData({ data: url })
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
