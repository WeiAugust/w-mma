function onFighterTap(fighterId) {
  wx.navigateTo({
    url: `/pages/fighter/index?id=${fighterId}`,
  })
}

module.exports = {
  onFighterTap,
}
