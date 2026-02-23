const defaultApi = require('../../services/api')

let api = defaultApi

function sanitizeText(value) {
  if (value === undefined || value === null) {
    return ''
  }
  return String(value)
    .replace(/&quot;|&#34;/gi, '"')
    .replace(/&#39;|&apos;/gi, "'")
    .replace(/&amp;/gi, '&')
    .replace(/&nbsp;/gi, ' ')
    .trim()
}

function stripWrappingQuotes(value) {
  return sanitizeText(value).replace(/^[\s"'“”‘’]+|[\s"'“”‘’]+$/g, '').trim()
}

function countryDisplay(countryZH, countryEN) {
  const zh = sanitizeText(countryZH)
  const en = sanitizeText(countryEN)
  if (zh && en) {
    return `${zh} / ${en}`
  }
  return zh || en
}

function normalizeItem(raw) {
  if (!raw || typeof raw !== 'object') {
    return null
  }
  return {
    ...raw,
    name: sanitizeText(raw.name),
    name_zh: sanitizeText(raw.name_zh),
    nickname: stripWrappingQuotes(raw.nickname),
    country: sanitizeText(raw.country),
    country_zh: sanitizeText(raw.country_zh),
    record: sanitizeText(raw.record),
    country_display: countryDisplay(raw.country_zh, raw.country),
  }
}

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
      const items = Array.isArray(data && data.items) ? data.items.map((item) => normalizeItem(item)).filter((item) => item) : []
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
