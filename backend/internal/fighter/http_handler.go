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

func RegisterAdminFighterRoutes(r *gin.Engine, svc *Service) {
	r.POST("/admin/fighters/manual", func(c *gin.Context) {
		var req struct {
			SourceID      int64  `json:"source_id"`
			Name          string `json:"name"`
			NameZH        string `json:"name_zh"`
			Nickname      string `json:"nickname"`
			Country       string `json:"country"`
			Record        string `json:"record"`
			WeightClass   string `json:"weight_class"`
			AvatarURL     string `json:"avatar_url"`
			IntroVideoURL string `json:"intro_video_url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.SourceID <= 0 || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source_id and name are required"})
			return
		}

		item, err := svc.CreateManual(c.Request.Context(), CreateManualInput{
			SourceID:      req.SourceID,
			Name:          req.Name,
			NameZH:        req.NameZH,
			Nickname:      req.Nickname,
			Country:       req.Country,
			Record:        req.Record,
			WeightClass:   req.WeightClass,
			AvatarURL:     req.AvatarURL,
			IntroVideoURL: req.IntroVideoURL,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	})
}
