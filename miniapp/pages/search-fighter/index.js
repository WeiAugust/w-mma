const defaultApi = require('../../services/api')

let api = defaultApi

const pageDef = {
  data: {
    keyword: '',
    loading: false,
    error: '',
    items: [],
  },

  onKeywordInput(event) {
    const keyword = event && event.detail ? event.detail.value || '' : ''
    this.setData({ keyword })
  },

  async onSearchTap() {
    await this.search(this.data.keyword)
  },

  async search(keyword) {
    this.setData({ loading: true, error: '' })

    try {
      const data = await api.searchFighters(keyword || '')
      const items = Array.isArray(data && data.items) ? data.items : []
      this.setData({
        loading: false,
        error: '',
        items,
      })
      return data
    } catch (err) {
      this.setData({
        loading: false,
        items: [],
        error: (err && err.message) || '搜索失败',
      })
      throw err
    }
  },

  onSelectFighter(eventOrID) {
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
