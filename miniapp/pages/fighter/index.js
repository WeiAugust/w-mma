const defaultApi = require('../../services/api')

let api = defaultApi

const LABEL_TRANSLATIONS = {
  Age: '年龄',
  Height: '身高',
  Weight: '体重',
  Reach: '臂展',
  'Leg reach': '腿长',
  'Fighting style': '技术风格',
  'Place of Birth': '出生地',
  'Trains at': '训练馆',
  'Octagon Debut': 'UFC 首秀',
  'PFP Rank': 'PFP 排名',
  'Athlete Status': '状态',
  'Title Status': '头衔状态',
  Status: '状态',
  'Per Min': '每分钟',
  'Per 15 Min': '每15分钟',
  'Sig. Str. Landed': '有效击打命中',
  'Sig. Str. Absorbed': '有效击打承受',
  'Submission avg': '降服尝试',
  'Takedown avg': '抱摔',
  'Sig. Str. Defense': '有效击打防御',
  'Takedown Defense': '抱摔防御',
  'Knockdown Avg': '击倒均值',
  'Average fight time': '平均作战时长',
  'Professional Record': '职业战绩',
  'Wins by Knockout': 'KO/TKO获胜',
  'Wins by Submission': '降服获胜',
  'First Round Finishes': '首回合终结',
}

const HERO_STAT_KEYS = new Set(['PFP Rank', 'Athlete Status', 'Title Status', 'Status'])

const WEIGHT_CLASS_TRANSLATIONS = {
  Strawweight: '草量级',
  Flyweight: '蝇量级',
  Bantamweight: '雏量级',
  Featherweight: '羽量级',
  Lightweight: '轻量级',
  Welterweight: '次中量级',
  Middleweight: '中量级',
  'Light Heavyweight': '轻重量级',
  Heavyweight: '重量级',
  Catchweight: '协议体重',
  "Women's Strawweight": '女子草量级',
  "Women's Flyweight": '女子蝇量级',
  "Women's Bantamweight": '女子雏量级',
  "Women's Featherweight": '女子羽量级',
}

function toNumber(value) {
  const n = Number(value)
  return Number.isNaN(n) ? 0 : n
}

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

function toCountryDisplay(countryZH, countryEN) {
  const zh = sanitizeText(countryZH)
  const en = sanitizeText(countryEN)
  if (zh && en) {
    return `${zh} / ${en}`
  }
  return zh || en
}

function isChinaFighterCountry(countryZH, countryEN) {
  const zh = sanitizeText(countryZH)
  if (zh.includes('中国')) {
    return true
  }
  const en = sanitizeText(countryEN).toLowerCase()
  if (!en) {
    return false
  }
  if (/\bchina\b/.test(en)) {
    return true
  }
  return en.includes('people') && en.includes('republic') && en.includes('china')
}

function translateWeightClass(value) {
  const clean = sanitizeText(value)
  if (!clean) {
    return ''
  }
  return WEIGHT_CLASS_TRANSLATIONS[clean] || clean
}

function translateStatus(value) {
  const clean = sanitizeText(value).toLowerCase()
  if (clean === 'active') {
    return '现役'
  }
  if (clean === 'inactive') {
    return '非现役'
  }
  if (clean === 'retired') {
    return '已退役'
  }
  return sanitizeText(value)
}

function translateTitleStatus(value) {
  const clean = sanitizeText(value).toLowerCase()
  if (clean === 'title holder') {
    return '现任冠军'
  }
  return sanitizeText(value)
}

function translateFightingStyle(value) {
  const clean = sanitizeText(value).toLowerCase()
  if (clean === 'kickboxer') {
    return '踢拳'
  }
  if (clean === 'wrestler') {
    return '摔跤'
  }
  if (clean === 'jiu-jitsu' || clean === 'bjj') {
    return '巴西柔术'
  }
  return sanitizeText(value)
}

function formatMeasurement(label, value) {
  const n = Number(value)
  if (Number.isNaN(n)) {
    return sanitizeText(value)
  }
  if (label === 'Height' || label === 'Reach' || label === 'Leg reach') {
    const cm = (n * 2.54).toFixed(1)
    return `${cm} 厘米（${sanitizeText(value)} 英寸）`
  }
  if (label === 'Weight') {
    const kg = (n * 0.45359237).toFixed(1)
    return `${kg} 公斤（${sanitizeText(value)} 磅）`
  }
  return sanitizeText(value)
}

function translateLabel(label) {
  const clean = sanitizeText(label)
  return LABEL_TRANSLATIONS[clean] || clean
}

function translateValue(label, value) {
  const cleanLabel = sanitizeText(label)
  const cleanValue = sanitizeText(value)
  if (!cleanValue) {
    return ''
  }
  if (cleanLabel === 'Height' || cleanLabel === 'Reach' || cleanLabel === 'Leg reach' || cleanLabel === 'Weight') {
    return formatMeasurement(cleanLabel, cleanValue)
  }
  if (cleanLabel === 'Fighting style') {
    return translateFightingStyle(cleanValue)
  }
  if (cleanLabel === 'Athlete Status' || cleanLabel === 'Status') {
    return translateStatus(cleanValue)
  }
  if (cleanLabel === 'Title Status') {
    return translateTitleStatus(cleanValue)
  }
  return cleanValue
}

function sanitizeMap(raw) {
  if (!raw || typeof raw !== 'object') {
    return {}
  }
  const out = {}
  Object.keys(raw).forEach((key) => {
    const cleanKey = sanitizeText(key)
    const cleanValue = sanitizeText(raw[key])
    if (!cleanKey || !cleanValue) {
      return
    }
    out[cleanKey] = cleanValue
  })
  return out
}

function sanitizeFighter(raw) {
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
    weight_class: sanitizeText(raw.weight_class),
    avatar_url: sanitizeText(raw.avatar_url),
    intro_video_url: sanitizeText(raw.intro_video_url),
    updates: Array.isArray(raw.updates) ? raw.updates.map((item) => sanitizeText(item)).filter((item) => item) : [],
    stats: sanitizeMap(raw.stats),
    records: sanitizeMap(raw.records),
  }
}

function toKeyValueItems(raw, options = {}) {
  if (!raw || typeof raw !== 'object') {
    return []
  }
  const excludeKeys = options.excludeKeys || new Set()
  return Object.keys(raw)
    .filter((key) => key && !excludeKeys.has(key) && raw[key] !== undefined && raw[key] !== null && sanitizeText(raw[key]) !== '')
    .sort((a, b) => String(a).localeCompare(String(b)))
    .map((key) => ({
      label: translateLabel(key),
      value: translateValue(key, raw[key]),
    }))
}

function parseHistoryResult(content) {
  const clean = sanitizeText(content)
  const hit = clean.match(/·\s*(胜|负|平|无结果)\s*·/)
  if (hit) {
    return hit[1]
  }
  return ''
}

function historyResultClass(result) {
  if (result === '胜') {
    return 'result-win'
  }
  if (result === '负') {
    return 'result-loss'
  }
  if (result === '平') {
    return 'result-draw'
  }
  if (result === '无结果') {
    return 'result-nc'
  }
  return 'result-unknown'
}

function toHistoryItems(updates) {
  if (!Array.isArray(updates)) {
    return []
  }
  return updates
    .map((item) => sanitizeText(item))
    .filter((item) => item)
    .map((item) => {
      const result = parseHistoryResult(item)
      const display = result ? item.replace(` · ${result} · `, ' · ') : item
      return {
        raw: item,
        display,
        result,
        resultClass: historyResultClass(result),
      }
    })
}

async function loadFighterWithContext(ctx, fighterID) {
  ctx.setData({ loading: true, error: '' })

  try {
    const fighter = sanitizeFighter(await api.getFighterDetail(fighterID))
    const statItems = toKeyValueItems(fighter && fighter.stats, { excludeKeys: HERO_STAT_KEYS })
    const recordItems = toKeyValueItems(fighter && fighter.records)
    const statusValue = fighter && fighter.stats ? fighter.stats['Athlete Status'] || fighter.stats.Status : ''
    const titleValue = fighter && fighter.stats ? fighter.stats['Title Status'] : ''
    const pfpValue = fighter && fighter.stats ? sanitizeText(fighter.stats['PFP Rank']) : ''
    const pfpTag = pfpValue ? `${pfpValue} P4P` : ''
    const historyItems = fighter ? toHistoryItems(fighter.updates) : []
    const isChinaFighter = fighter ? isChinaFighterCountry(fighter.country_zh, fighter.country) : false
    ctx.setData({
      loading: false,
      error: '',
      fighter,
      statItems,
      recordItems,
      countryDisplay: fighter ? toCountryDisplay(fighter.country_zh, fighter.country) : '',
      weightClassDisplay: fighter ? translateWeightClass(fighter.weight_class) : '',
      pfpTag,
      statusTag: translateStatus(statusValue),
      titleTag: translateTitleStatus(titleValue),
      historyItems,
      isChinaFighter,
    })
  } catch (err) {
    ctx.setData({
      loading: false,
      fighter: null,
      statItems: [],
      recordItems: [],
      countryDisplay: '',
      weightClassDisplay: '',
      pfpTag: '',
      statusTag: '',
      titleTag: '',
      historyItems: [],
      isChinaFighter: false,
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
    countryDisplay: '',
    weightClassDisplay: '',
    pfpTag: '',
    statusTag: '',
    titleTag: '',
    historyItems: [],
    isChinaFighter: false,
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
