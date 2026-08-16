package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/pkg/jwt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestManager() *jwt.Manager {
	return jwt.New(config.JWT{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
	})
}

func performWithToken(token string) *httptest.ResponseRecorder {
	mgr := newTestManager()
	router := gin.New()
	router.Use(Authenticate(mgr, nil))
	router.GET("/protected", func(c *gin.Context) {
		id := UserID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": id})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAuthenticateValidToken(t *testing.T) {
	mgr := newTestManager()
	token, _, err := mgr.IssueAccess(99, "access-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	rec := performWithToken(token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); body != `{"user_id":99}` {
		t.Errorf("body = %s", body)
	}
}

func TestAuthenticateMissingToken(t *testing.T) {
	rec := performWithToken("")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	rec := performWithToken("garbage-token")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthenticateRevokedToken(t *testing.T) {
	mgr := newTestManager()
	token, _, err := mgr.IssueAccess(1, "revoked-access")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	router := gin.New()
	router.Use(Authenticate(mgr, func(_ *gin.Context, tokenID string) (bool, error) {
		return tokenID == "revoked-access", nil
	}))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
