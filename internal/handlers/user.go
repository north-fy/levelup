package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/services"
)

type updateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	Status    *string `json:"status"`
	AvatarURL *string `json:"avatar_url"`
}

// UserHandler exposes the profile endpoints.
type UserHandler struct {
	users *services.UserService
}

// NewUserHandler creates the user handler.
func NewUserHandler(users *services.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// Me godoc
//
//	@Summary	Get the authenticated user's profile
//	@Tags		users
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	domain.User
//	@Failure	401	{object}	ErrorResponse
//	@Router		/users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	user, err := h.users.GetByID(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateMe godoc
//
//	@Summary	Update the authenticated user's profile
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body	updateProfileRequest	true	"Profile fields"
//	@Success	200	{object}	domain.User
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/users/me [patch]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	input := services.UpdateProfileInput{
		Nickname:  req.Nickname,
		Status:    req.Status,
		AvatarURL: req.AvatarURL,
	}

	user, err := h.users.Update(c.Request.Context(), middleware.UserID(c), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}
