const defaultApi = require('../../services/api')

let api = defaultApi

const WEIGHT_CLASS_MAP = {
  "women's strawweight": '女子草量级',
  "women’s strawweight": '女子草量级',
  "women's flyweight": '女子蝇量级',
  "women’s flyweight": '女子蝇量级',
  "women's bantamweight": '女子雏量级',
  "women’s bantamweight": '女子雏量级',
  "women's featherweight": '女子羽量级',
  "women’s featherweight": '女子羽量级',
  strawweight: '草量级',
  flyweight: '蝇量级',
  bantamweight: '雏量级',
  featherweight: '羽量级',
  lightweight: '轻量级',
  welterweight: '次中量级',
  middleweight: '中量级',
  'light heavyweight': '轻重量级',
  heavyweight: '重量级',
  catchweight: '协议体重',
}

function toNumber(value) {
  const n = Number(value)
  return Number.isNaN(n) ? 0 : n
}

function withPlaceholder(value) {
  const text = String(value || '').trim()
  return text || '--'
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

function mapWeightClass(value) {
  const normalized = String(value || '').trim().toLowerCase()
  if (!normalized) {
    return '--'
  }
  return WEIGHT_CLASS_MAP[normalized] || withPlaceholder(value)
}

function formatFightClock(timeSec) {
  const n = toNumber(timeSec)
  if (n <= 0) {
    return ''
  }
  const minute = Math.floor(n / 60)
  const second = n % 60
  return `${minute}:${String(second).padStart(2, '0')}`
}

function buildResultText(bout = {}, winnerName = '') {
  const parts = []
  if (winnerName) {
    parts.push(`胜者 ${winnerName}`)
  }
  const method = String(bout.method || '').trim()
  if (method) {
    parts.push(method)
  }
  const round = toNumber(bout.round)
  if (round > 0) {
    parts.push(`第${round}回合`)
  }
  const clock = formatFightClock(bout.time_sec)
  if (clock) {
    parts.push(clock)
  }
  if (parts.length > 0) {
    return parts.join(' · ')
  }
  const rawResult = String(bout.result || '').trim()
  if (rawResult && rawResult.toLowerCase() !== 'pending') {
    return rawResult
  }
  return '结果待更新'
}

function normalizeFighter(fighter = {}) {
  return {
    id: fighter.id || 0,
    name: fighter.name || '',
    country: fighter.country || '',
    rank: fighter.rank || '',
    weight_class: fighter.weight_class || '',
    avatar_url: fighter.avatar_url || '',
    name_text: withPlaceholder(fighter.name),
    country_text: withPlaceholder(fighter.country),
    rank_text: withPlaceholder(fighter.rank),
    weight_class_text: mapWeightClass(fighter.weight_class),
  }
}

function normalizeBout(bout = {}) {
  const redFighter = normalizeFighter(bout.red_fighter || {})
  const blueFighter = normalizeFighter(bout.blue_fighter || {})
  const winnerID = toNumber(bout.winner_id)
  let winnerName = ''
  if (winnerID && winnerID === redFighter.id) {
    winnerName = redFighter.name
  } else if (winnerID && winnerID === blueFighter.id) {
    winnerName = blueFighter.name
  }
  return {
    ...bout,
    winner_name: winnerName || '--',
    weight_class_text: mapWeightClass(bout.weight_class),
    result_text: buildResultText(bout, winnerName),
    red_fighter: redFighter,
    blue_fighter: blueFighter,
  }
}

function normalizeLegacyBout(bout = {}) {
  return normalizeBout({
    ...bout,
    red_fighter: { id: bout.red_fighter_id },
    blue_fighter: { id: bout.blue_fighter_id },
  })
}

function normalizeBoutsGroup(list) {
  if (!Array.isArray(list)) {
    return []
  }
  return list.map((item) => normalizeBout(item))
}

function buildLegacyGroups(bouts = []) {
  const normalized = bouts.map((item) => normalizeLegacyBout(item))
  return {
    mainCard: normalized.slice(0, 5),
    prelims: normalized.slice(5),
  }
}

function normalizeEventPayload(event = {}) {
  const bouts = Array.isArray(event.bouts) ? event.bouts : []
  const mainCard = normalizeBoutsGroup(event.main_card)
  const prelims = normalizeBoutsGroup(event.prelims)

  if (mainCard.length || prelims.length) {
    return { bouts, mainCard, prelims }
  }

  const legacy = buildLegacyGroups(bouts)
  return {
    bouts,
    mainCard: legacy.mainCard,
    prelims: legacy.prelims,
  }
}

function normalizeEvent(event = {}) {
  const status = mapStatus(event.status)
  return {
    ...event,
    status_text: status.text,
    status_class: status.css,
  }
}

async function loadEventCardWithContext(ctx, eventID) {
  ctx.setData({ loading: true, error: '' })

  try {
    const event = await api.getEventCard(eventID)
    const normalized = normalizeEventPayload(event)
    ctx.setData({
      loading: false,
      error: '',
      event: normalizeEvent(event),
      bouts: normalized.bouts,
      mainCard: normalized.mainCard,
      prelims: normalized.prelims,
    })
  } catch (err) {
    ctx.setData({
      loading: false,
      event: null,
      bouts: [],
      mainCard: [],
      prelims: [],
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
    mainCard: [],
    prelims: [],
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
