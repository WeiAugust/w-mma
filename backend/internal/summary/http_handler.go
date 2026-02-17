package summary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type summarizeRequest struct {
	SourceID int64 `json:"source_id"`
}

func RegisterAdminSummaryRoutes(r *gin.Engine, svc *Service) {
	r.POST("/admin/articles/:id/summarize", func(c *gin.Context) {
		articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
			return
		}

		var req summarizeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.SourceID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source_id is required"})
			return
		}

		job, err := svc.CreateArticleJob(c.Request.Context(), req.SourceID, articleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, job)
	})

	r.GET("/admin/summary-jobs", func(c *gin.Context) {
		items, err := svc.ListJobs(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})
}
