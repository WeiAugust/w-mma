package takedown

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type createRequest struct {
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	Reason      string `json:"reason"`
	Complainant string `json:"complainant"`
	EvidenceURL string `json:"evidence_url"`
}

type resolveRequest struct {
	Action string `json:"action"`
}

func RegisterAdminTakedownRoutes(r *gin.Engine, svc *Service) {
	r.POST("/admin/takedowns", func(c *gin.Context) {
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.TargetType == "" || req.TargetID <= 0 || req.Reason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target_type, target_id and reason are required"})
			return
		}

		item, err := svc.Create(c.Request.Context(), CreateInput{
			TargetType:  req.TargetType,
			TargetID:    req.TargetID,
			Reason:      req.Reason,
			Complainant: req.Complainant,
			EvidenceURL: req.EvidenceURL,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	})

	r.POST("/admin/takedowns/:id/resolve", func(c *gin.Context) {
		ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid takedown id"})
			return
		}

		var req resolveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Action != ActionOfflined && req.Action != ActionRejected {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
			return
		}

		if err := svc.Resolve(c.Request.Context(), ticketID, req.Action); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
