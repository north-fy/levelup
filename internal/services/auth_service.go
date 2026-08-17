package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/jwt"
	"github.com/north-fy/levelup/internal/pkg/metrics"
)

const oauthStateTTL = 10 * time.Minute

// TokenPair contains freshly issued access and refresh tokens.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// AuthService orchestrates registration, login and token lifecycle.
type AuthService struct {
	users  UserStore
	tokens TokenStore
	jwt    *jwt.Manager
	gh     *GitHubClient
}

// NewAuthService creates the authentication service.
func NewAuthService(users UserStore, tokens TokenStore, jwtMgr *jwt.Manager, cfg config.GitHub) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
		jwt:    jwtMgr,
		gh:     NewGitHubClient(cfg),
	}
}

// Register creates a new account and returns tokens.
func (s *AuthService) Register(ctx context.Context, email, password, nickname string) (*domain.User, TokenPair, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return nil, TokenPair{}, err
	}

	if _, err := s.users.GetByEmail(ctx, normalized); err == nil {
		return nil, TokenPair{}, domain.ErrEmailAlreadyUsed
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, TokenPair{}, err
	}

	user := &domain.User{
		Email:        normalized,
		PasswordHash: string(hash),
		Nickname:     nickname,
		Level:        1,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, TokenPair{}, err
	}
	metrics.UsersRegistered.Inc()

	pair, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

// Login verifies credentials and returns tokens.
func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, TokenPair, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		return nil, TokenPair{}, err
	}

	user, err := s.users.GetByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, TokenPair{}, domain.ErrInvalidCredentials
		}
		return nil, TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, TokenPair{}, domain.ErrInvalidCredentials
	}

	pair, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

// Refresh rotates a refresh token into a fresh token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*domain.User, TokenPair, error) {
	claims, err := s.jwt.ParseRefresh(refreshToken)
	if err != nil {
		return nil, TokenPair{}, domain.ErrInvalidToken
	}

	userID, err := s.tokens.GetRefreshUser(ctx, claims.ID)
	if err != nil {
		return nil, TokenPair{}, domain.ErrTokenRevoked
	}
	if userID != claims.UserID {
		return nil, TokenPair{}, domain.ErrTokenRevoked
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, TokenPair{}, err
	}

	if err := s.tokens.DeleteRefresh(ctx, claims.ID); err != nil {
		return nil, TokenPair{}, err
	}

	pair, err := s.issueTokens(ctx, userID)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

// Logout revokes both the access and refresh tokens.
func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	remaining := s.jwt.AccessTTL()

	if accessToken != "" {
		if claims, err := s.jwt.ParseAccess(accessToken); err == nil {
			if exp := claims.ExpiresAt; exp != nil {
				remaining = time.Until(exp.Time)
			}
			if remaining > 0 {
				if err := s.tokens.BlacklistAccess(ctx, claims.ID, remaining); err != nil {
					return err
				}
			}
		}
	}

	if refreshToken != "" {
		if claims, err := s.jwt.ParseRefresh(refreshToken); err == nil {
			if err := s.tokens.DeleteRefresh(ctx, claims.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// IsTokenBlacklisted reports whether an access token has been revoked.
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	return s.tokens.IsBlacklisted(ctx, tokenID)
}

// NewOAuthState generates a fresh, unguessable OAuth state value.
func (s *AuthService) NewOAuthState() string {
	return uuid.NewString()
}

// SaveOAuthState persists a freshly generated OAuth state value.
func (s *AuthService) SaveOAuthState(ctx context.Context, state string) error {
	return s.tokens.SaveOAuthState(ctx, state, oauthStateTTL)
}

// ValidateOAuthState verifies and consumes an OAuth state value.
func (s *AuthService) ValidateOAuthState(ctx context.Context, state string) error {
	return s.tokens.ValidateOAuthState(ctx, state)
}

// GitHubRedirectURL returns the URL a user should be redirected to for OAuth.
func (s *AuthService) GitHubRedirectURL(state string) string {
	return s.gh.AuthorizeURL(state)
}

// GitHubCallback exchanges an authorization code and returns the user and tokens.
func (s *AuthService) GitHubCallback(ctx context.Context, code string) (*domain.User, TokenPair, error) {
	ghUser, err := s.gh.FetchUser(ctx, code)
	if err != nil {
		return nil, TokenPair{}, err
	}

	user, err := s.users.GetByGitHubID(ctx, ghUser.ID)
	if err == nil {
		pair, err := s.issueTokens(ctx, user.ID)
		if err != nil {
			return nil, TokenPair{}, err
		}
		return user, pair, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, TokenPair{}, err
	}

	if ghUser.Email != "" {
		if existing, err := s.users.GetByEmail(ctx, ghUser.Email); err == nil {
			existing.GitHubID = ghUser.ID
			if existing.AvatarURL == "" {
				existing.AvatarURL = ghUser.AvatarURL
			}
			if err := s.users.Update(ctx, existing); err != nil {
				return nil, TokenPair{}, err
			}
			pair, err := s.issueTokens(ctx, existing.ID)
			if err != nil {
				return nil, TokenPair{}, err
			}
			return existing, pair, nil
		}
	}

	user = &domain.User{
		Email:     ghUser.Email,
		GitHubID:  ghUser.ID,
		Nickname:  ghUser.Login,
		AvatarURL: ghUser.AvatarURL,
		Level:     1,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, TokenPair{}, err
	}

	pair, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID uint) (TokenPair, error) {
	accessID := uuid.NewString()
	refreshID := uuid.NewString()

	accessToken, accessExp, err := s.jwt.IssueAccess(userID, accessID)
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, refreshExp, err := s.jwt.IssueRefresh(userID, refreshID)
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.tokens.SaveRefresh(ctx, refreshID, userID, s.jwt.RefreshTTL()); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func normalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(normalized, "@") {
		return "", errors.New("invalid email")
	}
	return normalized, nil
}
