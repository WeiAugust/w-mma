const DEV_API_BASE_URL = 'https://localhost:8443'
const PROD_API_BASE_URL = 'https://api.example.com'

function miniProgramEnvVersion() {
  if (typeof wx === 'undefined' || typeof wx.getAccountInfoSync !== 'function') {
    return ''
  }
  try {
    const info = wx.getAccountInfoSync()
    return String((info && info.miniProgram && info.miniProgram.envVersion) || '').toLowerCase()
  } catch (err) {
    return ''
  }
}

function resolveApiBaseUrl() {
  const envVersion = miniProgramEnvVersion()
  if (envVersion === 'release' || envVersion === 'trial') {
    return PROD_API_BASE_URL
  }
  return DEV_API_BASE_URL
}

module.exports = {
  DEV_API_BASE_URL,
  PROD_API_BASE_URL,
  resolveApiBaseUrl,
}
