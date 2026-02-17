package http

import "github.com/gin-gonic/gin"

func NewServer() *gin.Engine {
	r := gin.New()
	RegisterRoutes(r)
	return r
}
