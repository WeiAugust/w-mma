const { searchFighters } = require('../../services/api')

async function search(keyword) {
  return searchFighters(keyword)
}

function onSelectFighter(fighterId) {
  wx.navigateTo({
    url: `/pages/fighter/index?id=${fighterId}`,
  })
}

module.exports = {
  search,
  onSelectFighter,
}
