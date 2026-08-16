package domain

import "errors"

// Sentinel errors shared across layers.
var (
	ErrNotFound             = errors.New("not found")
	ErrForbidden            = errors.New("forbidden")
	ErrEmailAlreadyUsed     = errors.New("email already in use")
	ErrGitHubAlreadyUsed    = errors.New("github account already linked")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenRevoked         = errors.New("token revoked")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidState         = errors.New("invalid oauth state")
	ErrQuestAlreadyDone     = errors.New("quest already completed")
	ErrQuestAlreadyStarted  = errors.New("quest already started")
	ErrQuestNotInProgress   = errors.New("quest is not in progress")
	ErrActiveTimerConflict  = errors.New("another quest is already in progress")
	ErrTimedQuestIncomplete = errors.New("timed quest requires start/stop")
	ErrCannotBuyOwnItem     = errors.New("cannot buy your own item")
	ErrItemNotActive        = errors.New("item is not active")
	ErrNotEnoughGold        = errors.New("not enough gold")
)

// ValidationError indicates invalid client input.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError builds a ValidationError with the given message.
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}
