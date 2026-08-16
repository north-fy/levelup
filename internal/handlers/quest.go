package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/services"
)

type createQuestRequest struct {
	Title         string           `json:"title" binding:"required"`
	Description   string           `json:"description"`
	Type          domain.QuestType `json:"type" binding:"required"`
	RewardXP      int              `json:"reward_xp"`
	RewardGold    int              `json:"reward_gold"`
	DurationHours int              `json:"duration_hours"`
}

type updateQuestRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	RewardXP      *int    `json:"reward_xp"`
	RewardGold    *int    `json:"reward_gold"`
	DurationHours *int    `json:"duration_hours"`
}

// QuestHandler exposes the quest endpoints.
type QuestHandler struct {
	quests *services.QuestService
}

// NewQuestHandler creates the quest handler.
func NewQuestHandler(quests *services.QuestService) *QuestHandler {
	return &QuestHandler{quests: quests}
}

// Create godoc
//
//	@Summary	Create a quest in a branch
//	@Tags		quests
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		branch_id	path	uint				true	"Branch id"
//	@Param		body		body	createQuestRequest	true	"Quest payload"
//	@Success	201	{object}	domain.Quest
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/branches/{branch_id}/quests [post]
func (h *QuestHandler) Create(c *gin.Context) {
	branchID, ok := parseIDParam(c, "branch_id")
	if !ok {
		return
	}

	var req createQuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	quest, err := h.quests.Create(c.Request.Context(), middleware.UserID(c), branchID, services.CreateQuestInput{
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		RewardXP:      req.RewardXP,
		RewardGold:    req.RewardGold,
		DurationHours: req.DurationHours,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, quest)
}

// List godoc
//
//	@Summary	List quests in a branch
//	@Tags		quests
//	@Security	BearerAuth
//	@Produce	json
//	@Param		branch_id	path	uint	true	"Branch id"
//	@Success	200	{array}	domain.Quest
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/branches/{branch_id}/quests [get]
func (h *QuestHandler) List(c *gin.Context) {
	branchID, ok := parseIDParam(c, "branch_id")
	if !ok {
		return
	}
	quests, err := h.quests.List(c.Request.Context(), middleware.UserID(c), branchID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quests)
}

// Get godoc
//
//	@Summary	Get a quest by id
//	@Tags		quests
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Quest id"
//	@Success	200	{object}	domain.Quest
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/quests/{id} [get]
func (h *QuestHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	quest, err := h.quests.Get(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quest)
}

// Update godoc
//
//	@Summary	Update a quest
//	@Tags		quests
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	uint				true	"Quest id"
//	@Param		body	body	updateQuestRequest	true	"Quest fields"
//	@Success	200	{object}	domain.Quest
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/quests/{id} [patch]
func (h *QuestHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateQuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	quest, err := h.quests.Update(c.Request.Context(), middleware.UserID(c), id, services.UpdateQuestInput{
		Title:         req.Title,
		Description:   req.Description,
		RewardXP:      req.RewardXP,
		RewardGold:    req.RewardGold,
		DurationHours: req.DurationHours,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quest)
}

// Delete godoc
//
//	@Summary	Delete a quest
//	@Tags		quests
//	@Security	BearerAuth
//	@Param		id	path	uint	true	"Quest id"
//	@Success	204
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/quests/{id} [delete]
func (h *QuestHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.quests.Delete(c.Request.Context(), middleware.UserID(c), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Complete godoc
//
//	@Summary	Complete a simple quest and claim its reward
//	@Tags		quests
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Quest id"
//	@Success	200	{object}	domain.Quest
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/quests/{id}/complete [post]
func (h *QuestHandler) Complete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	quest, err := h.quests.Complete(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quest)
}

// Start godoc
//
//	@Summary	Start the time tracker of a timed quest
//	@Tags		quests
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Quest id"
//	@Success	200	{object}	domain.Quest
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/quests/{id}/start [post]
func (h *QuestHandler) Start(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	quest, err := h.quests.Start(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quest)
}

// Stop godoc
//
//	@Summary	Stop the time tracker and complete the timed quest
//	@Tags		quests
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Quest id"
//	@Success	200	{object}	domain.Quest
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/quests/{id}/stop [post]
func (h *QuestHandler) Stop(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	quest, err := h.quests.Stop(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quest)
}
