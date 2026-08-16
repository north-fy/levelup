package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/services"
)

type createRoadmapRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type updateRoadmapRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type addNodeRequest struct {
	Title         string           `json:"title" binding:"required"`
	Description   string           `json:"description"`
	Type          domain.QuestType `json:"type" binding:"required"`
	RewardXP      int              `json:"reward_xp"`
	RewardGold    int              `json:"reward_gold"`
	DurationHours int              `json:"duration_hours"`
	Dependencies  []uint           `json:"dependencies"`
}

type updateNodeRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	RewardXP      *int    `json:"reward_xp"`
	RewardGold    *int    `json:"reward_gold"`
	DurationHours *int    `json:"duration_hours"`
	Dependencies  *[]uint `json:"dependencies"`
}

// RoadmapHandler exposes the roadmap endpoints.
type RoadmapHandler struct {
	roadmaps *services.RoadmapService
}

// NewRoadmapHandler creates the roadmap handler.
func NewRoadmapHandler(roadmaps *services.RoadmapService) *RoadmapHandler {
	return &RoadmapHandler{roadmaps: roadmaps}
}

// Create godoc
//
//	@Summary	Create a personal roadmap
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	createRoadmapRequest	true	"Roadmap payload"
//	@Success	201	{object}	domain.Roadmap
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/roadmaps [post]
func (h *RoadmapHandler) Create(c *gin.Context) {
	var req createRoadmapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	roadmap, err := h.roadmaps.Create(c.Request.Context(), middleware.UserID(c), services.CreateRoadmapInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, roadmap)
}

// List godoc
//
//	@Summary	List the user's roadmaps
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	domain.Roadmap
//	@Failure	401	{object}	ErrorResponse
//	@Router		/roadmaps [get]
func (h *RoadmapHandler) List(c *gin.Context) {
	roadmaps, err := h.roadmaps.List(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, roadmaps)
}

// Get godoc
//
//	@Summary	Get a roadmap with its graph
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Roadmap id"
//	@Success	200	{object}	domain.RoadmapDetail
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/roadmaps/{id} [get]
func (h *RoadmapHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	detail, err := h.roadmaps.Get(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

// Update godoc
//
//	@Summary	Update a roadmap
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	uint					true	"Roadmap id"
//	@Param		body	body	updateRoadmapRequest	true	"Roadmap fields"
//	@Success	200	{object}	domain.Roadmap
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/roadmaps/{id} [patch]
func (h *RoadmapHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateRoadmapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	roadmap, err := h.roadmaps.Update(c.Request.Context(), middleware.UserID(c), id, services.UpdateRoadmapInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, roadmap)
}

// Delete godoc
//
//	@Summary	Delete a roadmap
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Param		id	path	uint	true	"Roadmap id"
//	@Success	204
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/roadmaps/{id} [delete]
func (h *RoadmapHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.roadmaps.Delete(c.Request.Context(), middleware.UserID(c), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AddNode godoc
//
//	@Summary	Add a node to a roadmap
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	uint			true	"Roadmap id"
//	@Param		body	body	addNodeRequest	true	"Node payload"
//	@Success	201	{object}	domain.RoadmapNode
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/roadmaps/{id}/nodes [post]
func (h *RoadmapHandler) AddNode(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req addNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	node, err := h.roadmaps.AddNode(c.Request.Context(), middleware.UserID(c), id, services.AddNodeInput{
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		RewardXP:      req.RewardXP,
		RewardGold:    req.RewardGold,
		DurationHours: req.DurationHours,
		Dependencies:  req.Dependencies,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, node)
}

// UpdateNode godoc
//
//	@Summary	Update a roadmap node
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	uint				true	"Roadmap id"
//	@Param		nodeId	path	uint				true	"Node id"
//	@Param		body	body	updateNodeRequest	true	"Node fields"
//	@Success	200	{object}	domain.RoadmapNode
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/roadmaps/{id}/nodes/{nodeId} [patch]
func (h *RoadmapHandler) UpdateNode(c *gin.Context) {
	roadmapID, ok := parseID(c)
	if !ok {
		return
	}
	nodeID, ok := parseIDParam(c, "nodeId")
	if !ok {
		return
	}

	var req updateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	node, err := h.roadmaps.UpdateNode(c.Request.Context(), middleware.UserID(c), roadmapID, nodeID, services.UpdateNodeInput{
		Title:         req.Title,
		Description:   req.Description,
		RewardXP:      req.RewardXP,
		RewardGold:    req.RewardGold,
		DurationHours: req.DurationHours,
		Dependencies:  req.Dependencies,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, node)
}

// CompleteNode godoc
//
//	@Summary	Complete a roadmap node once its dependencies are done
//	@Tags		roadmaps
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id		path	uint	true	"Roadmap id"
//	@Param		nodeId	path	uint	true	"Node id"
//	@Success	200	{object}	domain.RoadmapNode
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Failure	409	{object}	ErrorResponse
//	@Router		/roadmaps/{id}/nodes/{nodeId}/complete [post]
func (h *RoadmapHandler) CompleteNode(c *gin.Context) {
	roadmapID, ok := parseID(c)
	if !ok {
		return
	}
	nodeID, ok := parseIDParam(c, "nodeId")
	if !ok {
		return
	}
	node, err := h.roadmaps.CompleteNode(c.Request.Context(), middleware.UserID(c), roadmapID, nodeID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, node)
}
