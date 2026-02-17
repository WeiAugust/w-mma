function createPageContext(pageDef, initialData = {}) {
  return {
    data: {
      ...(pageDef.data || {}),
      ...initialData,
    },
    setData(patch) {
      this.data = {
        ...this.data,
        ...patch,
      }
    },
  }
}

module.exports = {
  createPageContext,
}
