package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/service"
	"golang.org/x/crypto/bcrypt"
)

var (
	testPool   *pgxpool.Pool
	testRouter *chi.Mux
)

// TestMain sets up the test database connection and router once for all tests.
func TestMain(m *testing.M) {
	// Use minimal bcrypt cost in tests for faster execution
	auth.PasswordCost = bcrypt.MinCost

	// Try loading .env from project root (relative to this package: internal/handler/)
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		slog.Error("TEST_DATABASE_URL is not set. Ensure .env is configured and Docker postgres is running.")
		os.Exit(1)
	}

	var err error
	testPool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		slog.Error("failed to create test pool", "error", err)
		os.Exit(1)
	}

	if err := testPool.Ping(context.Background()); err != nil {
		slog.Error("failed to ping test database", "error", err)
		os.Exit(1)
	}

	// Set up the router with all handlers registered
	testRouter = chi.NewMux()
	api := humachi.New(testRouter, huma.DefaultConfig("Dozingo Test API", "0.1.0"))

	// SessionUser middleware is required by /auth/* and any handler calling
	// RequireSessionCtx. Tests run over plain HTTP via httptest, so disable
	// the Secure cookie flag to let cookies round-trip.
	middleware.SetCookieSecureForTesting(false)
	queries := generated.New(testPool)
	api.UseMiddleware(middleware.SessionUser(api, queries))

	apiGroup := huma.NewGroup(api, "/api")

	RegisterHealth(apiGroup)

	// New layering: repositories -> services -> handlers.
	repos := repository.New(testPool)
	NewBoardsHandler(service.NewBoards(repos.Boards)).Register(apiGroup)

	// Legacy registrars (to be migrated).
	RegisterCells(apiGroup, testPool)
	RegisterVotes(apiGroup, testPool)
	RegisterGames(apiGroup, testPool)
	RegisterGameCells(apiGroup, testPool)
	RegisterAuth(apiGroup, testPool)

	// Clean tables before running tests to ensure a fresh state
	truncateAllTables()

	code := m.Run()

	testPool.Close()
	os.Exit(code)
}

/// ===== Per-test setup =====

// setupTest registers a t.Cleanup that truncates all tables when the test
// finishes. Call this as the first line of every test that touches the DB:
//
//	func TestSomething(t *testing.T) {
//	    setupTest(t)
//	    ...
//	}
func setupTest(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { cleanupTables(t) })
}

// truncateAllTables truncates every table touched by the test suite in FK-safe
// order. Used by both TestMain (for a clean baseline) and cleanupTables.
func truncateAllTables() {
	_, _ = testPool.Exec(context.Background(),
		"TRUNCATE TABLE game_cells, games, votes, cells, boards, sessions, user_passwords, users RESTART IDENTITY CASCADE")
}

// cleanupTables truncates all tables in the correct order (respecting foreign keys).
func cleanupTables(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"TRUNCATE TABLE game_cells, games, votes, cells, boards, sessions, user_passwords, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to clean up tables: %v", err)
	}
}

/// ===== HTTP helpers =====

// doRequest performs an HTTP request against the test router and returns the response.
func doRequest(method, path string, body any) *httptest.ResponseRecorder {
	return doRequestWithCookies(method, path, body, nil)
}

// doRequestWithCookies performs an HTTP request with the given cookies attached.
func doRequestWithCookies(method, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal request body: %v", err))
		}
		req = httptest.NewRequest(method, path, strings.NewReader(string(b)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

// extractSessionCookie looks for a Set-Cookie header carrying session_token
// and returns it. Returns nil if the response did not set or clear that
// cookie. The cookie's MaxAge / Expires fields determine whether it was set
// or cleared (MaxAge < 0 means cleared).
func extractSessionCookie(w *httptest.ResponseRecorder) *http.Cookie {
	resp := http.Response{Header: w.Header()}
	for _, c := range resp.Cookies() {
		if c.Name == "session_token" {
			return c
		}
	}
	return nil
}

// decodeJSON decodes a JSON response body into the given target.
func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(target); err != nil {
		t.Fatalf("failed to decode response body: %v\nbody: %s", err, w.Body.String())
	}
}

// assertStatus checks that the response status code matches the expected one.
func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("expected status %d, got %d\nbody: %s", expected, w.Code, w.Body.String())
	}
}

// assertJSONField checks that a specific field in the JSON response has the expected string value.
func assertJSONField(t *testing.T, data map[string]any, key string, expected string) {
	t.Helper()
	val, ok := data[key]
	if !ok {
		t.Errorf("expected field %q in response, but it was missing", key)
		return
	}
	str, ok := val.(string)
	if !ok {
		t.Errorf("expected field %q to be a string, got %T", key, val)
		return
	}
	if str != expected {
		t.Errorf("expected %q = %q, got %q", key, expected, str)
	}
}

/// ===== Misc helpers =====

// stringPtr returns a pointer to s. Convenient for optional string fields in
// request bodies (e.g. email).
func stringPtr(s string) *string {
	return &s
}

/// ===== Factories: users & sessions =====

// createTestUser creates a user via the auth register API and returns its ID.
// Pass an empty string for email to omit it from the request.
func createTestUser(t *testing.T, username, email string) string {
	t.Helper()

	body := map[string]any{
		"username": username,
		"password": "testpassword123",
	}
	if email != "" {
		body["email"] = email
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusOK)

	// Look up the user ID from the database since the auth endpoint doesn't return it
	queries := generated.New(testPool)
	user, err := queries.GetUserByUsername(context.Background(), username)
	if err != nil {
		t.Fatalf("failed to look up user %q after registration: %v", username, err)
	}
	return user.ID.String()
}

// createTestUserWithRegister registers a user via the auth API and returns the
// full response body (so callers can inspect cookies / id / email handling).
// Pass nil for email to omit it from the request.
func createTestUserWithRegister(t *testing.T, username, password string, email *string) *map[string]any {
	t.Helper()

	body := map[string]any{
		"username": username,
		"password": password,
	}
	if email != nil {
		body["email"] = *email
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)
	return &resp
}

// mintAnonSession inserts an anonymous session row directly and returns a
// cookie that can be replayed against the test router. This is the only way
// to obtain a pre-existing anon session because no anon-friendly handler
// currently calls RequireSessionCtx.
func mintAnonSession(t *testing.T, ttl time.Duration) (token string, cookie *http.Cookie) {
	t.Helper()
	q := generated.New(testPool)
	tok := auth.GenerateSessionToken()
	_, err := q.CreateSession(context.Background(), generated.CreateSessionParams{
		UserID: pgtype.UUID{Valid: false},
		Token:  tok,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(ttl),
			Valid: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to insert anon session: %v", err)
	}
	return tok, &http.Cookie{Name: "session_token", Value: tok}
}

// loadSessionByToken fetches the session row (joined with the user) for the
// given token. The bool is false when no row exists.
func loadSessionByToken(t *testing.T, token string) (generated.GetSessionUserByTokenRow, bool) {
	t.Helper()
	q := generated.New(testPool)
	row, err := q.GetSessionUserByToken(context.Background(), token)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, false
	}
	return row, true
}

/// ===== Factories: domain entities =====

// createTestBoard creates a board via the API and returns its ID. Pass nil for
// description to omit it from the request.
func createTestBoard(t *testing.T, title string, size int, authorID string, description *string) string {
	t.Helper()
	body := map[string]any{
		"title":     title,
		"size":      size,
		"author_id": authorID,
	}
	if description != nil {
		body["description"] = *description
	}
	w := doRequest(http.MethodPost, "/api/boards", body)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["board_id"].(string)
}

// createTestCell creates a cell on a board via the API and returns its ID.
func createTestCell(t *testing.T, boardID, content string) string {
	t.Helper()
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": content,
	})
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["cell_id"].(string)
}

// createTestGame creates a game via the API and returns its ID.
func createTestGame(t *testing.T, playerID, boardID string) string {
	t.Helper()
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/games?player_id=%s", boardID, playerID), nil)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["game_id"].(string)
}

// createTestVote upserts a vote on a board for a user and returns the response
// body. Most callers can ignore the return value.
func createTestVote(t *testing.T, boardID, userID string, value int) map[string]any {
	t.Helper()
	w := doRequest(http.MethodPut,
		fmt.Sprintf("/api/boards/%s/vote?user_id=%s", boardID, userID),
		map[string]any{"vote_value": value},
	)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp
}

// createTestGameCell creates a single game cell on a game and returns its ID.
// Use createTestGameCells for bulk creation when you need multiple cells.
func createTestGameCell(t *testing.T, gameID, cellID, content string, position int) string {
	t.Helper()
	body := []map[string]any{
		{"cell_id": cellID, "content": content, "position": position},
	}
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), body)
	assertStatus(t, w, http.StatusOK)
	var resp []map[string]any
	decodeJSON(t, w, &resp)
	if len(resp) == 0 {
		t.Fatalf("expected at least one game cell in response, got 0")
	}
	return resp[0]["game_cell_id"].(string)
}
