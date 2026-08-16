package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/services"
)

type createBranchRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

type updateBranchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
}

// BranchHandler exposes the branch endpoints.
type BranchHandler struct {
	branches *services.BranchService
}

// NewBranchHandler creates the branch handler.
func NewBranchHandler(branches *services.BranchService) *BranchHandler {
	return &BranchHandler{branches: branches}
}

// Create godoc
//
//	@Summary	Create a branch
//	@Tags		branches
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	createBranchRequest	true	"Branch payload"
//	@Success	201	{object}	domain.Branch
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/branches [post]
func (h *BranchHandler) Create(c *gin.Context) {
	var req createBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	branch, err := h.branches.Create(c.Request.Context(), middleware.UserID(c), services.CreateBranchInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, branch)
}

// List godoc
//
//	@Summary	List the user's branches
//	@Tags		branches
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	domain.Branch
//	@Failure	401	{object}	ErrorResponse
//	@Router		/branches [get]
func (h *BranchHandler) List(c *gin.Context) {
	branches, err := h.branches.List(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, branches)
}

// Get godoc
//
//	@Summary	Get a branch by id
//	@Tags		branches
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Branch id"
//	@Success	200	{object}	domain.Branch
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/branches/{id} [get]
func (h *BranchHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	branch, err := h.branches.Get(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, branch)
}

// Update godoc
//
//	@Summary	Update a branch
//	@Tags		branches
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	uint					true	"Branch id"
//	@Param		body	body	updateBranchRequest	true	"Branch fields"
//	@Success	200	{object}	domain.Branch
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/branches/{id} [patch]
func (h *BranchHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	branch, err := h.branches.Update(c.Request.Context(), middleware.UserID(c), id, services.UpdateBranchInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, branch)
}

// Delete godoc
//
//	@Summary	Delete a branch
//	@Tags		branches
//	@Security	BearerAuth
//	@Param		id	path	uint	true	"Branch id"
//	@Success	204
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/branches/{id} [delete]
func (h *BranchHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.branches.Delete(c.Request.Context(), middleware.UserID(c), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uint, bool) {
	return parseIDParam(c, "id")
}

func parseIDParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return 0, false
	}
	return uint(id), true
}
