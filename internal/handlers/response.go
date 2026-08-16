package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/domain"
)

// ErrorResponse is the standard error body returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondError maps domain errors to HTTP statuses.
func respondError(c *gin.Context, err error) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
	case errors.Is(err, domain.ErrEmailAlreadyUsed), errors.Is(err, domain.ErrGitHubAlreadyUsed):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
	case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrTokenRevoked):
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidState):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid oauth state"})
	case errors.Is(err, domain.ErrQuestAlreadyDone),
		errors.Is(err, domain.ErrQuestAlreadyStarted),
		errors.Is(err, domain.ErrQuestNotInProgress),
		errors.Is(err, domain.ErrActiveTimerConflict),
		errors.Is(err, domain.ErrCannotBuyOwnItem),
		errors.Is(err, domain.ErrItemNotActive),
		errors.Is(err, domain.ErrNotEnoughGold):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrTimedQuestIncomplete):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
	}
}
