const defaultApi = require('../../services/api')

let api = defaultApi

function toNumber(value) {
  const n = Number(value)
  return Number.isNaN(n) ? 0 : n
}

function toKeyValueItems(raw) {
  if (!raw || typeof raw !== 'object') {
    return []
  }
  return Object.keys(raw)
    .filter((key) => key && raw[key] !== undefined && raw[key] !== null && String(raw[key]).trim() !== '')
    .map((key) => ({
      label: key,
      value: String(raw[key]).trim(),
    }))
}

async function loadFighterWithContext(ctx, fighterID) {
  ctx.setData({ loading: true, error: '' })

  try {
    const fighter = await api.getFighterDetail(fighterID)
    const statItems = toKeyValueItems(fighter && fighter.stats)
    const recordItems = toKeyValueItems(fighter && fighter.records)
    ctx.setData({
      loading: false,
      error: '',
      fighter,
      statItems,
      recordItems,
    })
  } catch (err) {
    ctx.setData({
      loading: false,
      fighter: null,
      statItems: [],
      recordItems: [],
      error: (err && err.message) || '选手信息加载失败',
    })
  }
}

const pageDef = {
  data: {
    loading: false,
    error: '',
    fighterID: 0,
    fighter: null,
    statItems: [],
    recordItems: [],
  },

  async onLoad(options = {}) {
    const fighterID = toNumber(options.id)
    this.setData({ fighterID })
    if (!fighterID) {
      this.setData({ error: '无效选手 ID' })
      return
    }

    await loadFighterWithContext(this, fighterID)
  },

  async onPullDownRefresh() {
    if (this.data.fighterID) {
      await loadFighterWithContext(this, this.data.fighterID)
    }

    if (typeof wx.stopPullDownRefresh === 'function') {
      wx.stopPullDownRefresh()
    }
  },

  async onRetryTap() {
    if (!this.data.fighterID) {
      return
    }
    await loadFighterWithContext(this, this.data.fighterID)
  },

  async loadFighter(fighterID) {
    await loadFighterWithContext(this, fighterID)
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
