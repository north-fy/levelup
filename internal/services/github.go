package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/north-fy/levelup/internal/config"
)

//nolint:gosec // URLs, not credentials.
const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubAPIURL       = "https://api.github.com"
)

// GitHubUser is the subset of the GitHub API profile we use.
type GitHubUser struct {
	ID        string
	Login     string
	Email     string
	AvatarURL string
}

// GitHubClient talks to the GitHub OAuth API.
type GitHubClient struct {
	clientID     string
	clientSecret string
	redirectURL  string
	http         *http.Client
}

// NewGitHubClient creates the client from configuration.
func NewGitHubClient(cfg config.GitHub) *GitHubClient {
	return &GitHubClient{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURL:  cfg.RedirectURL,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
}

// AuthorizeURL builds the URL the user is sent to start the OAuth flow.
func (c *GitHubClient) AuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", c.redirectURL)
	params.Set("scope", "read:user user:email")
	params.Set("state", state)
	return githubAuthorizeURL + "?" + params.Encode()
}

// FetchUser exchanges an authorization code for the user's profile.
func (c *GitHubClient) FetchUser(ctx context.Context, code string) (*GitHubUser, error) {
	accessToken, err := c.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	profile, err := c.fetchProfile(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (c *GitHubClient) exchangeCode(ctx context.Context, code string) (string, error) {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("client_secret", c.clientSecret)
	params.Set("code", code)
	params.Set("redirect_uri", c.redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github token exchange failed: status %d", resp.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.AccessToken == "" {
		return "", errors.New("github token exchange returned no token")
	}
	return body.AccessToken, nil
}

func (c *GitHubClient) fetchProfile(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL+"/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github profile fetch failed: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	user := &GitHubUser{
		ID:        fmt.Sprintf("%d", payload.ID),
		Login:     payload.Login,
		Email:     payload.Email,
		AvatarURL: payload.AvatarURL,
	}

	if user.Email == "" {
		user.Email = c.fetchPrimaryEmail(ctx, accessToken)
	}

	return user, nil
}

func (c *GitHubClient) fetchPrimaryEmail(ctx context.Context, accessToken string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL+"/user/emails", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	return ""
}
