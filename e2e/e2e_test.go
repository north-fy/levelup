//go:build e2e

// Package e2e contains end-to-end tests that spin up the full docker-compose
// stack (PostgreSQL, Redis, ClickHouse, Prometheus, Grafana and the app) and
// exercise the public HTTP API.
//
// Run with:
//
//	go test -tags e2e ./e2e/ -v -count=1
//
// The test requires Docker and builds the app image on first run. The stack is
// brought up once for the whole suite and torn down afterwards.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	baseURL    = "http://localhost:8080/api/v1"
	appURL     = "http://localhost:8080"
	composeDir = "../deploy"
	envFile    = "../.env"

	pgDSN = "postgres://levelup:levelup@localhost:5432/levelup?sslmode=disable"
	chDSN = "clickhouse://default:@localhost:9000?database=levelup&x-multi-statement=true"
)

var defaultEnvTemplate = []byte(`APP_ENV=development
APP_PORT=8080
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=levelup
POSTGRES_PASSWORD=levelup
POSTGRES_DB=levelup
POSTGRES_SSLMODE=disable
CLICKHOUSE_HOST=clickhouse
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=levelup
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
JWT_ACCESS_SECRET=e2e-access-secret
JWT_REFRESH_SECRET=e2e-refresh-secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_REDIRECT_URL=
RATE_LIMIT_GLOBAL=2000
RATE_LIMIT_PER_USER=120
RATE_LIMIT_WINDOW=1m
`)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Println("e2e skipped: docker not found on PATH")
		return 0
	}
	if out, err := exec.Command("docker", "info").CombinedOutput(); err != nil {
		fmt.Printf("e2e skipped: docker daemon not reachable (%v)\n%s", err, out)
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if err := stackUp(ctx); err != nil {
		fmt.Printf("e2e setup failed: %v\n", err)
		return 1
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		if err := stackDown(ctx); err != nil {
			fmt.Printf("e2e teardown failed: %v\n", err)
		}
	}()

	return m.Run()
}

func stackUp(ctx context.Context) error {
	if _, err := os.Stat(envFile); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(envFile, defaultEnvTemplate, 0o600); err != nil {
			return fmt.Errorf("write .env: %w", err)
		}
		fmt.Println("e2e: created .env for the compose stack")
	} else {
		fmt.Println("e2e: reusing existing .env (ensure it points to compose service names: postgres, redis, clickhouse)")
	}

	if out, err := compose(ctx, "up", "-d", "--build"); err != nil {
		return fmt.Errorf("compose up: %w\n%s", err, out)
	}
	fmt.Println("e2e: compose stack is up")

	for _, addr := range []string{"localhost:5432", "localhost:6379", "localhost:9000", "localhost:8123"} {
		if err := waitTCP(ctx, addr, 2*time.Minute); err != nil {
			return err
		}
	}

	if err := runMigrations(ctx); err != nil {
		return err
	}

	if err := waitURL(ctx, appURL+"/readyz", 2*time.Minute); err != nil {
		return fmt.Errorf("app not ready: %w", err)
	}
	fmt.Println("e2e: app is ready")

	return nil
}

func stackDown(ctx context.Context) error {
	if out, err := compose(ctx, "down"); err != nil {
		return fmt.Errorf("compose down: %w\n%s", err, out)
	}
	return nil
}

// compose runs the docker-compose CLI (v2 "docker compose" with v1 fallback).
func compose(ctx context.Context, args ...string) ([]byte, error) {
	if _, err := exec.LookPath("docker-compose"); err == nil {
		full := append([]string{"-f", composeDir + "/docker-compose.dev.yml"}, args...)
		return exec.CommandContext(ctx, "docker-compose", full...).CombinedOutput()
	}
	full := append([]string{"compose", "-f", composeDir + "/docker-compose.dev.yml"}, args...)
	return exec.CommandContext(ctx, "docker", full...).CombinedOutput()
}

func waitTCP(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for %s", addr)
}

func waitURL(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for %s", url)
}

func runMigrations(ctx context.Context) error {
	for _, mc := range []struct{ source, dsn string }{
		{"file://../migrations/postgres", pgDSN},
		{"file://../migrations/clickhouse", chDSN},
	} {
		if err := migrateRetry(ctx, mc.source, mc.dsn); err != nil {
			return err
		}
	}
	fmt.Println("e2e: migrations applied")
	return nil
}

func migrateRetry(ctx context.Context, source, dsn string) error {
	deadline := time.Now().Add(2 * time.Minute)
	var lastErr error
	for time.Now().Before(deadline) {
		m, err := migrate.New(source, dsn)
		if err == nil {
			err = m.Up()
			_, _ = m.Close()
			if err == nil || errors.Is(err, migrate.ErrNoChange) {
				return nil
			}
			lastErr = err
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("migrate %s: %w", source, lastErr)
}

// --- API helpers ------------------------------------------------------------

type apiClient struct {
	base string
	hc   *http.Client
}

func newClient() *apiClient {
	return &apiClient{base: baseURL, hc: &http.Client{Timeout: 20 * time.Second}}
}

func (c *apiClient) do(t *testing.T, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, rd)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, data
}

func asMap(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode response: %v\nbody: %s", err, data)
	}
	return m
}

func str(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func num(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		}
	}
	return 0
}

func ids(data []byte) []float64 {
	var arr []map[string]any
	_ = json.Unmarshal(data, &arr)
	var out []float64
	for _, m := range arr {
		out = append(out, num(m, "id"))
	}
	return out
}

type user struct {
	token   string
	refresh string
	id      int
	email   string
}

func registerUser(t *testing.T, c *apiClient, prefix string) user {
	t.Helper()
	email := fmt.Sprintf("%s-%d@e2e.dev", prefix, time.Now().UnixNano())
	status, body := c.do(t, http.MethodPost, "/auth/register", "", map[string]any{
		"email":    email,
		"password": "password123",
		"nickname": prefix,
	})
	if status != http.StatusCreated {
		t.Fatalf("register: status %d body %s", status, body)
	}
	m := asMap(t, body)
	u := user{
		token:   str(m, "access_token"),
		refresh: str(m, "refresh_token"),
		email:   email,
	}
	um, ok := m["user"].(map[string]any)
	if !ok {
		t.Fatalf("register: no user object: %s", body)
	}
	u.id = int(num(um, "id"))
	if u.token == "" || u.id == 0 {
		t.Fatalf("register: missing token/user: %s", body)
	}
	return u
}

func requireStatus(t *testing.T, got, want int, body []byte, what string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: status %d, want %d, body %s", what, got, want, body)
	}
}

func waitFor(t *testing.T, what string, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("timeout waiting for %s", what)
}

// --- Tests -------------------------------------------------------------------

func TestE2EHealth(t *testing.T) {
	c := newClient()

	status, body := c.do(t, http.MethodGet, "/healthz", "", nil)
	requireStatus(t, status, http.StatusOK, body, "healthz")
	if str(asMap(t, body), "status") != "ok" {
		t.Fatalf("healthz body: %s", body)
	}
}

func TestE2EAuth(t *testing.T) {
	c := newClient()
	u := registerUser(t, c, "auth")

	status, body := c.do(t, http.MethodPost, "/auth/login", "", map[string]any{
		"email": u.email, "password": "password123",
	})
	requireStatus(t, status, http.StatusOK, body, "login")
	u.token = str(asMap(t, body), "access_token")

	status, body = c.do(t, http.MethodGet, "/users/me", u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "me")
	if str(asMap(t, body), "nickname") != "auth" {
		t.Fatalf("me nickname: %s", body)
	}

	status, body = c.do(t, http.MethodPatch, "/users/me", u.token, map[string]any{"nickname": "renamed"})
	requireStatus(t, status, http.StatusOK, body, "patch me")
	if str(asMap(t, body), "nickname") != "renamed" {
		t.Fatalf("patched nickname: %s", body)
	}

	status, body = c.do(t, http.MethodGet, "/users/me", u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "me after patch")
	if str(asMap(t, body), "nickname") != "renamed" {
		t.Fatalf("cached nickname should be invalidated: %s", body)
	}

	status, body = c.do(t, http.MethodPost, "/auth/refresh", "", map[string]any{"refresh_token": u.refresh})
	requireStatus(t, status, http.StatusOK, body, "refresh")

	status, body = c.do(t, http.MethodPost, "/auth/logout", "", map[string]any{
		"access_token": u.token, "refresh_token": u.refresh,
	})
	requireStatus(t, status, http.StatusNoContent, body, "logout")

	status, _ = c.do(t, http.MethodGet, "/users/me", u.token, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("me after logout: status %d, want 401", status)
	}
}

func TestE2EBranchesAndQuests(t *testing.T) {
	c := newClient()
	u := registerUser(t, c, "quests")

	status, body := c.do(t, http.MethodPost, "/branches", u.token, map[string]any{
		"name": "Finance", "description": "money", "color": "green",
	})
	requireStatus(t, status, http.StatusCreated, body, "create branch")
	branchID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/branches/%d/quests", branchID), u.token, map[string]any{
		"title": "Read a book", "type": "simple", "reward_xp": 100, "reward_gold": 50,
	})
	requireStatus(t, status, http.StatusCreated, body, "create quest")
	questID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodGet, fmt.Sprintf("/branches/%d/quests", branchID), u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "list quests")
	if len(ids(body)) != 1 {
		t.Fatalf("expected 1 quest, got %s", body)
	}

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/quests/%d/complete", questID), u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "complete quest")
	if str(asMap(t, body), "status") != "done" {
		t.Fatalf("quest status: %s", body)
	}

	status, body = c.do(t, http.MethodGet, "/users/me", u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "me")
	if num(asMap(t, body), "xp") != 100 {
		t.Fatalf("expected xp=100 after quest, got %s", body)
	}

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/branches/%d/quests", branchID), u.token, map[string]any{
		"title": "Study Go", "type": "timed", "reward_xp": 100, "reward_gold": 50, "duration_hours": 2,
	})
	requireStatus(t, status, http.StatusCreated, body, "create timed quest")
	timedID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/quests/%d/start", timedID), u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "start timed quest")
	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/quests/%d/stop", timedID), u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "stop timed quest")
	if str(asMap(t, body), "status") != "done" {
		t.Fatalf("timed quest status: %s", body)
	}
}

func TestE2EShop(t *testing.T) {
	c := newClient()
	buyer := registerUser(t, c, "buyer")
	seller := registerUser(t, c, "seller")

	status, body := c.do(t, http.MethodPost, "/branches", buyer.token, map[string]any{"name": "Earn"})
	requireStatus(t, status, http.StatusCreated, body, "buyer branch")
	branchID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/branches/%d/quests", branchID), buyer.token, map[string]any{
		"title": "Earn gold", "type": "simple", "reward_gold": 100,
	})
	requireStatus(t, status, http.StatusCreated, body, "gold quest")
	questID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/quests/%d/complete", questID), buyer.token, nil)
	requireStatus(t, status, http.StatusOK, body, "complete gold quest")

	status, body = c.do(t, http.MethodPost, "/shop/items", seller.token, map[string]any{
		"title": "Coffee", "description": "A cup", "price_gold": 30,
	})
	requireStatus(t, status, http.StatusCreated, body, "create item")
	itemID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodGet, "/shop/items", buyer.token, nil)
	requireStatus(t, status, http.StatusOK, body, "list shop")
	if len(ids(body)) != 1 {
		t.Fatalf("expected 1 shop item, got %s", body)
	}

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/shop/items/%d/buy", itemID), buyer.token, nil)
	requireStatus(t, status, http.StatusCreated, body, "buy item")

	status, body = c.do(t, http.MethodGet, "/users/me", buyer.token, nil)
	requireStatus(t, status, http.StatusOK, body, "buyer me")
	if gold := num(asMap(t, body), "gold"); gold != 70 {
		t.Fatalf("buyer gold = %v, want 70", gold)
	}

	status, body = c.do(t, http.MethodGet, "/shop/purchases", buyer.token, nil)
	requireStatus(t, status, http.StatusOK, body, "purchases")
	if len(ids(body)) != 1 {
		t.Fatalf("expected 1 purchase, got %s", body)
	}
}

func TestE2ERoadmapWorkshop(t *testing.T) {
	c := newClient()
	author := registerUser(t, c, "author")
	installer := registerUser(t, c, "installer")

	status, body := c.do(t, http.MethodPost, "/roadmaps", author.token, map[string]any{
		"title": "Go Developer", "description": "path",
	})
	requireStatus(t, status, http.StatusCreated, body, "create roadmap")
	roadmapID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/roadmaps/%d/nodes", roadmapID), author.token, map[string]any{
		"title": "Learn basics", "type": "simple", "reward_xp": 100,
	})
	requireStatus(t, status, http.StatusCreated, body, "add node 1")
	node1 := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/roadmaps/%d/nodes", roadmapID), author.token, map[string]any{
		"title": "Build a project", "type": "simple", "reward_xp": 200, "dependencies": []float64{float64(node1)},
	})
	requireStatus(t, status, http.StatusCreated, body, "add node 2")
	node2 := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/roadmaps/%d/nodes/%d/complete", roadmapID, node1), author.token, nil)
	requireStatus(t, status, http.StatusOK, body, "complete node 1")

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/roadmaps/%d/nodes/%d/complete", roadmapID, node2), author.token, nil)
	requireStatus(t, status, http.StatusOK, body, "complete node 2")

	status, body = c.do(t, http.MethodPost, "/workshop/roadmaps", author.token, map[string]any{
		"roadmap_id": roadmapID, "title": "Go Developer (workshop)",
	})
	requireStatus(t, status, http.StatusCreated, body, "publish workshop")
	workshopID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodGet, "/workshop/roadmaps", installer.token, nil)
	requireStatus(t, status, http.StatusOK, body, "list workshop")
	if len(ids(body)) != 1 {
		t.Fatalf("expected 1 workshop roadmap, got %s", body)
	}

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/workshop/roadmaps/%d/install", workshopID), installer.token, nil)
	requireStatus(t, status, http.StatusCreated, body, "install roadmap")

	status, body = c.do(t, http.MethodGet, "/roadmaps", installer.token, nil)
	requireStatus(t, status, http.StatusOK, body, "installed roadmaps")
	if len(ids(body)) != 1 {
		t.Fatalf("expected 1 installed roadmap, got %s", body)
	}
}

func TestE2EStats(t *testing.T) {
	c := newClient()
	u := registerUser(t, c, "stats")

	status, body := c.do(t, http.MethodPost, "/branches", u.token, map[string]any{"name": "Reading"})
	requireStatus(t, status, http.StatusCreated, body, "create branch")
	branchID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/branches/%d/quests", branchID), u.token, map[string]any{
		"title": "Finish a chapter", "type": "simple", "reward_xp": 50, "reward_gold": 25,
	})
	requireStatus(t, status, http.StatusCreated, body, "create quest")
	questID := int(num(asMap(t, body), "id"))

	status, body = c.do(t, http.MethodPost, fmt.Sprintf("/quests/%d/complete", questID), u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "complete quest")

	// The outbox flusher writes to ClickHouse every 5s, so poll for the data.
	waitFor(t, "stats flush to ClickHouse", 30*time.Second, func() bool {
		status, _ := c.do(t, http.MethodGet, "/stats/branches", u.token, nil)
		return status == http.StatusOK
	})

	status, body = c.do(t, http.MethodGet, "/stats/overview", u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "stats overview")
	m := asMap(t, body)
	if num(m, "xp") < 50 {
		t.Fatalf("overview xp = %v, want >= 50", num(m, "xp"))
	}

	status, body = c.do(t, http.MethodGet, "/stats/branches", u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "stats branches")
	if len(ids(body)) != 1 {
		t.Fatalf("expected 1 branch stat, got %s", body)
	}

	status, body = c.do(t, http.MethodGet, "/stats/quests?period=day", u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "stats quests")
	if len(ids(body)) < 1 {
		t.Fatalf("expected daily quest stats, got %s", body)
	}

	status, body = c.do(t, http.MethodGet, "/stats/roadmaps", u.token, nil)
	requireStatus(t, status, http.StatusOK, body, "stats roadmaps")
}
