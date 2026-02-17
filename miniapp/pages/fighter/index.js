const defaultApi = require('../../services/api')

let api = defaultApi

function toNumber(value) {
  const n = Number(value)
  return Number.isNaN(n) ? 0 : n
}

async function loadFighterWithContext(ctx, fighterID) {
  ctx.setData({ loading: true, error: '' })

  try {
    const fighter = await api.getFighterDetail(fighterID)
    ctx.setData({
      loading: false,
      error: '',
      fighter,
    })
  } catch (err) {
    ctx.setData({
      loading: false,
      fighter: null,
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
