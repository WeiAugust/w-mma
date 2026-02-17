package fighter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterFighterRoutes(r *gin.Engine, svc *Service) {
	r.GET("/api/fighters/search", func(c *gin.Context) {
		items, err := svc.Search(c.Request.Context(), c.Query("q"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})

	r.GET("/api/fighters/:id", func(c *gin.Context) {
		fighterID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fighter id"})
			return
		}
		fighter, err := svc.Get(c.Request.Context(), fighterID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, fighter)
	})
}
