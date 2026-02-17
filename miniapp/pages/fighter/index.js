const { getFighterDetail } = require('../../services/api')

async function loadFighter(fighterId) {
  return getFighterDetail(fighterId)
}

module.exports = {
  loadFighter,
}
