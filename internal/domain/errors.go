package domain

import "errors"

// Sentinel errors shared across layers.
var (
	ErrNotFound           = errors.New("not found")
	ErrEmailAlreadyUsed   = errors.New("email already in use")
	ErrGitHubAlreadyUsed  = errors.New("github account already linked")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidState       = errors.New("invalid oauth state")
)
