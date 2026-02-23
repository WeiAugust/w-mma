const defaultApi = require('../../services/api')

let api = defaultApi
const BEIJING_OFFSET_MS = 8 * 60 * 60 * 1000

function formatUpdatedAt(date = new Date()) {
  const pad = (value) => String(value).padStart(2, '0')
  return `${date.getMonth() + 1}/${date.getDate()} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function formatLocalDateTime(raw) {
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return '--'
  }

  const beijingDate = new Date(date.getTime() + BEIJING_OFFSET_MS)
  const pad = (value) => String(value).padStart(2, '0')
  return `${beijingDate.getUTCFullYear()}-${pad(beijingDate.getUTCMonth() + 1)}-${pad(beijingDate.getUTCDate())} ${pad(beijingDate.getUTCHours())}:${pad(beijingDate.getUTCMinutes())}:${pad(beijingDate.getUTCSeconds())}`
}

function mapStatus(status) {
  const normalized = String(status || '').toLowerCase()
  if (normalized === 'scheduled' || normalized === 'upcoming') {
    return { text: '未开赛', css: 'scheduled' }
  }
  if (normalized === 'completed') {
    return { text: '已结束', css: 'completed' }
  }
  if (normalized === 'live') {
    return { text: '进行中', css: 'live' }
  }
  return { text: '待定', css: 'unknown' }
}

function normalizeEventItem(item = {}) {
  const mappedStatus = mapStatus(item.status)
  return {
    ...item,
    status_text: mappedStatus.text,
    status_class: mappedStatus.css,
    starts_at_text: formatLocalDateTime(item.starts_at),
  }
}

function hasPoster(item = {}) {
  return Boolean(String(item.poster_url || '').trim())
}

function hasCompleteMatchups(card = {}) {
  const mainCard = Array.isArray(card.main_card) ? card.main_card : []
  const prelims = Array.isArray(card.prelims) ? card.prelims : []
  const detailBouts = mainCard.concat(prelims)
  if (detailBouts.length > 0) {
    return detailBouts.every((bout) => {
      const red = bout && bout.red_fighter
      const blue = bout && bout.blue_fighter
      return Boolean(red && red.id && String(red.name || '').trim() && blue && blue.id && String(blue.name || '').trim())
    })
  }
  const legacyBouts = Array.isArray(card.bouts) ? card.bouts : []
  return legacyBouts.length > 0
}

async function filterDisplayableEvents(items = []) {
  const withPoster = items.filter((item) => hasPoster(item))
  if (typeof api.getEventCard !== 'function') {
    return withPoster
  }

  const checks = await Promise.all(
    withPoster.map(async (item) => {
      try {
        const card = await api.getEventCard(item.id)
        if (!hasCompleteMatchups(card)) {
          return null
        }
        return item
      } catch (err) {
        return null
      }
    }),
  )
  return checks.filter(Boolean)
}

function applyOrgFilter(items, org) {
  if (!org || org === 'ALL') {
    return items
  }
  return items.filter((item) => item.org === org)
}

const pageDef = {
  data: {
    loading: false,
    error: '',
    items: [],
    filteredItems: [],
    selectedOrg: 'ALL',
    orgOptions: ['ALL', 'UFC', 'ONE', 'PFL', 'Bellator'],
    updatedAtText: '',
  },

  async onLoad() {
    await this.loadEvents()
  },

  async onPullDownRefresh() {
    await this.loadEvents()
    if (typeof wx.stopPullDownRefresh === 'function') {
      wx.stopPullDownRefresh()
    }
  },

  async onRetryTap() {
    await this.loadEvents()
  },

  async loadEvents() {
    this.setData({ loading: true, error: '' })

    try {
      const data = await api.listEvents()
      const rawItems = Array.isArray(data && data.items) ? data.items : []
      const normalized = rawItems.map((item) => normalizeEventItem(item))
      const items = await filterDisplayableEvents(normalized)
      const filteredItems = applyOrgFilter(items, this.data.selectedOrg)
      this.setData({
        loading: false,
        error: '',
        items,
        filteredItems,
        updatedAtText: formatUpdatedAt(),
      })
    } catch (err) {
      this.setData({
        loading: false,
        error: (err && err.message) || '赛程加载失败',
        items: [],
        filteredItems: [],
      })
    }
  },

  onOrgChange(event) {
    const value = event && event.detail ? event.detail.value : 'ALL'
    let selectedOrg = value

    if (typeof value === 'number') {
      selectedOrg = this.data.orgOptions[value] || 'ALL'
    } else if (typeof value === 'string' && /^\d+$/.test(value)) {
      const idx = Number(value)
      selectedOrg = this.data.orgOptions[idx] || 'ALL'
    }

    const filteredItems = applyOrgFilter(this.data.items, selectedOrg)
    this.setData({ selectedOrg, filteredItems })
  },

  onEventTap(eventOrID) {
    const eventID =
      typeof eventOrID === 'number'
        ? eventOrID
        : eventOrID && eventOrID.currentTarget && eventOrID.currentTarget.dataset && eventOrID.currentTarget.dataset.id

    if (!eventID) {
      return
    }

    wx.navigateTo({
      url: `/pages/event-detail/index?id=${eventID}`,
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
