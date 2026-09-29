package handler

import (
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// StatsSummary 仪表盘统计
// GET /api/admin/stats/summary
func (h *Handler) StatsSummary(c *gin.Context) {
	response.Success(c, h.StatsService.Summary())
}
