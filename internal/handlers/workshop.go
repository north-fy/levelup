package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/services"
)

type createWorkshopRequest struct {
	RoadmapID   uint   `json:"roadmap_id" binding:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateWorkshopRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	IsPublished *bool   `json:"is_published"`
}

// WorkshopHandler exposes the workshop endpoints.
type WorkshopHandler struct {
	workshop *services.WorkshopService
}

// NewWorkshopHandler creates the workshop handler.
func NewWorkshopHandler(workshop *services.WorkshopService) *WorkshopHandler {
	return &WorkshopHandler{workshop: workshop}
}

// Create godoc
//
//	@Summary	Publish a roadmap to the workshop
//	@Tags		workshop
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	createWorkshopRequest	true	"Workshop payload"
//	@Success	201	{object}	domain.WorkshopRoadmap
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/workshop/roadmaps [post]
func (h *WorkshopHandler) Create(c *gin.Context) {
	var req createWorkshopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	workshop, err := h.workshop.Create(c.Request.Context(), middleware.UserID(c), services.CreateWorkshopInput{
		RoadmapID:   req.RoadmapID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, workshop)
}

// List godoc
//
//	@Summary	List published workshop roadmaps
//	@Tags		workshop
//	@Security	BearerAuth
//	@Produce	json
//	@Param		mine	query	bool	false	"List only the current user's workshops"
//	@Success	200	{array}	domain.WorkshopRoadmap
//	@Failure	401	{object}	ErrorResponse
//	@Router		/workshop/roadmaps [get]
func (h *WorkshopHandler) List(c *gin.Context) {
	mine := c.Query("mine") == "true"
	workshops, err := h.workshop.List(c.Request.Context(), middleware.UserID(c), mine)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, workshops)
}

// Update godoc
//
//	@Summary	Update an authored workshop roadmap
//	@Tags		workshop
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	uint					true	"Workshop id"
//	@Param		body	body	updateWorkshopRequest	true	"Workshop fields"
//	@Success	200	{object}	domain.WorkshopRoadmap
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/workshop/roadmaps/{id} [patch]
func (h *WorkshopHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateWorkshopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	workshop, err := h.workshop.Update(c.Request.Context(), middleware.UserID(c), id, services.UpdateWorkshopInput{
		Title:       req.Title,
		Description: req.Description,
		IsPublished: req.IsPublished,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, workshop)
}

// Install godoc
//
//	@Summary	Install a published workshop roadmap
//	@Tags		workshop
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Workshop id"
//	@Success	200	{object}	domain.RoadmapDetail
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/workshop/roadmaps/{id}/install [post]
func (h *WorkshopHandler) Install(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	detail, err := h.workshop.Install(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}
