package review

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterAdminReviewRoutes(r *gin.Engine, svc *Service) {
	r.GET("/admin/review/pending", func(c *gin.Context) {
		items, err := svc.ListPending(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})

	r.POST("/admin/review/:id/approve", func(c *gin.Context) {
		pendingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pending id"})
			return
		}
		reviewerID, _ := strconv.ParseInt(c.Query("reviewer_id"), 10, 64)

		if err := svc.Approve(c.Request.Context(), pendingID, reviewerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
