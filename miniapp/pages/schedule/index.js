function onEventTap(eventId) {
  wx.navigateTo({
    url: `/pages/event-detail/index?id=${eventId}`,
  })
}

module.exports = {
  onEventTap,
}
