package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/services"
)

// StatsHandler exposes the statistics endpoints.
type StatsHandler struct {
	stats *services.StatsService
}

// NewStatsHandler creates the statistics handler.
func NewStatsHandler(stats *services.StatsService) *StatsHandler {
	return &StatsHandler{stats: stats}
}

// Overview godoc
//
//	@Summary	Get the user's overall statistics
//	@Tags		stats
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{object}	services.OverviewStats
//	@Failure	401	{object}	ErrorResponse
//	@Router		/stats/overview [get]
func (h *StatsHandler) Overview(c *gin.Context) {
	stats, err := h.stats.Overview(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Branches godoc
//
//	@Summary	Get per-branch statistics
//	@Tags		stats
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	services.BranchStat
//	@Failure	401	{object}	ErrorResponse
//	@Router		/stats/branches [get]
func (h *StatsHandler) Branches(c *gin.Context) {
	stats, err := h.stats.Branches(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Roadmaps godoc
//
//	@Summary	Get per-roadmap statistics
//	@Tags		stats
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	services.RoadmapStat
//	@Failure	401	{object}	ErrorResponse
//	@Router		/stats/roadmaps [get]
func (h *StatsHandler) Roadmaps(c *gin.Context) {
	stats, err := h.stats.Roadmaps(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Quests godoc
//
//	@Summary	Get daily quest statistics for a period
//	@Tags		stats
//	@Security	BearerAuth
//	@Produce	json
//	@Param		period	query	string	false	"day, week or month"	Enums(day, week, month)
//	@Success	200	{array}	services.QuestStat
//	@Failure	401	{object}	ErrorResponse
//	@Router		/stats/quests [get]
func (h *StatsHandler) Quests(c *gin.Context) {
	period := c.DefaultQuery("period", "day")
	stats, err := h.stats.Quests(c.Request.Context(), middleware.UserID(c), period)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
