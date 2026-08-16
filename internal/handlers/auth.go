package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/services"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Nickname string `json:"nickname" binding:"required,min=2,max=64"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	AccessToken      string      `json:"access_token"`
	RefreshToken     string      `json:"refresh_token"`
	AccessExpiresAt  int64       `json:"access_expires_at"`
	RefreshExpiresAt int64       `json:"refresh_expires_at"`
	User             domain.User `json:"user"`
}

// AuthHandler exposes the authentication endpoints.
type AuthHandler struct {
	auth *services.AuthService
}

// NewAuthHandler creates the authentication handler.
func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register godoc
//
//	@Summary	Register a new account
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	registerRequest	true	"Registration payload"
//	@Success	201	{object}	tokenResponse
//	@Failure	400	{object}	ErrorResponse
//	@Failure	409	{object}	ErrorResponse
//	@Router		/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, pair, err := h.auth.Register(c.Request.Context(), req.Email, req.Password, req.Nickname)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newTokenResponse(user, pair))
}

// Login godoc
//
//	@Summary	Log in with email and password
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	loginRequest	true	"Login payload"
//	@Success	200	{object}	tokenResponse
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, pair, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newTokenResponse(user, pair))
}

// Refresh godoc
//
//	@Summary	Rotate a refresh token into a new token pair
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	refreshRequest	true	"Refresh payload"
//	@Success	200	{object}	tokenResponse
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, pair, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newTokenResponse(user, pair))
}

// Logout godoc
//
//	@Summary	Revoke the current tokens
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	logoutRequest	true	"Tokens to revoke"
//	@Success	204
//	@Failure	400	{object}	ErrorResponse
//	@Router		/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.auth.Logout(c.Request.Context(), req.AccessToken, req.RefreshToken); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GitHubRedirect godoc
//
//	@Summary	Redirect the user to GitHub for OAuth
//	@Tags		auth
//	@Produce	plain
//	@Success	302
//	@Router		/auth/github/redirect [get]
func (h *AuthHandler) GitHubRedirect(c *gin.Context) {
	state := h.auth.NewOAuthState()
	if err := h.auth.SaveOAuthState(c.Request.Context(), state); err != nil {
		respondError(c, err)
		return
	}
	c.Redirect(http.StatusFound, h.auth.GitHubRedirectURL(state))
}

// GitHubCallback godoc
//
//	@Summary	Handle the GitHub OAuth callback
//	@Tags		auth
//	@Produce	json
//	@Param		code	query	string	true	"Authorization code"
//	@Param		state	query	string	true	"OAuth state"
//	@Success	200	{object}	tokenResponse
//	@Failure	400	{object}	ErrorResponse
//	@Router		/auth/github/callback [get]
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
		return
	}

	if err := h.auth.ValidateOAuthState(c.Request.Context(), state); err != nil {
		respondError(c, err)
		return
	}

	user, pair, err := h.auth.GitHubCallback(c.Request.Context(), code)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newTokenResponse(user, pair))
}

func newTokenResponse(user *domain.User, pair services.TokenPair) tokenResponse {
	return tokenResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt.Unix(),
		RefreshExpiresAt: pair.RefreshExpiresAt.Unix(),
		User:             *user,
	}
}
