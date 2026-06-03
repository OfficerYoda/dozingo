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
	"sync"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/service"
	"golang.org/x/crypto/bcrypt"
)

var (
	testPool   *pgxpool.Pool
	testRouter *chi.Mux

	// userCookies maps a registered user's ID to the session_token cookie
	// the auth/register response set for them. Test factories use this to
	// replay the user's session on subsequent requests so handlers that read
	// the caller from middleware.SessionUserFromContext see the right user.
	// Cleared in cleanupTables alongside the sessions table truncation.
	userCookies = map[string]*http.Cookie{}

	// fakeMailer is the email.Sender implementation wired into the test
	// service.Auth. Tests inspect its recorded calls to assert that mail
	// was (or was not) sent and to capture issued tokens for end-to-end
	// flows. Reset between tests in cleanupTables.
	fakeMailer = &fakeEmailSender{}
)

// fakeEmailSender records every Send* call. Safe for concurrent use; tests
// run sequentially by default but a single test may make concurrent
// requests so guard the slices anyway.
type fakeEmailSender struct {
	mu           sync.Mutex
	resetCalls   []sentMail
	verifyCalls  []sentMail
	failNextSend error
}

type sentMail struct {
	To    string
	Token string
}

func (f *fakeEmailSender) SendResetPassword(to, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNextSend != nil {
		err := f.failNextSend
		f.failNextSend = nil
		return err
	}
	f.resetCalls = append(f.resetCalls, sentMail{To: to, Token: token})
	return nil
}

func (f *fakeEmailSender) SendEmailVerification(to, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNextSend != nil {
		err := f.failNextSend
		f.failNextSend = nil
		return err
	}
	f.verifyCalls = append(f.verifyCalls, sentMail{To: to, Token: token})
	return nil
}

func (f *fakeEmailSender) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resetCalls = nil
	f.verifyCalls = nil
	f.failNextSend = nil
}

func (f *fakeEmailSender) lastReset() (sentMail, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.resetCalls) == 0 {
		return sentMail{}, false
	}
	return f.resetCalls[len(f.resetCalls)-1], true
}

func (f *fakeEmailSender) lastVerify() (sentMail, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.verifyCalls) == 0 {
		return sentMail{}, false
	}
	return f.verifyCalls[len(f.verifyCalls)-1], true
}

func (f *fakeEmailSender) resetCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.resetCalls)
}

func (f *fakeEmailSender) verifyCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.verifyCalls)
}

// cookiesFor returns the session cookie associated with userID, or nil if no
// cookie was captured (e.g. the user was created outside the auth API).
func cookiesFor(userID string) []*http.Cookie {
	if c, ok := userCookies[userID]; ok && c != nil {
		return []*http.Cookie{c}
	}
	return nil
}

// cookiesForBoard looks up the board's author and returns their session
// cookie, so test factories that operate on a board (cells, etc.) can act as
// the author without callers having to thread the author ID through.
func cookiesForBoard(t *testing.T, boardID string) []*http.Cookie {
	t.Helper()
	q := generated.New(testPool)
	id := pgtype.UUID{}
	if err := id.Scan(boardID); err != nil {
		t.Fatalf("invalid board id %q: %v", boardID, err)
	}
	board, err := q.GetBoardByID(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to load board %s for cookie lookup: %v", boardID, err)
	}
	return cookiesFor(board.AuthorID.String())
}

// cookiesForGame looks up the game's player and returns their session cookie.
func cookiesForGame(t *testing.T, gameID string) []*http.Cookie {
	t.Helper()
	q := generated.New(testPool)
	id := pgtype.UUID{}
	if err := id.Scan(gameID); err != nil {
		t.Fatalf("invalid game id %q: %v", gameID, err)
	}
	game, err := q.GetGameByID(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to load game %s for cookie lookup: %v", gameID, err)
	}
	return cookiesFor(game.PlayerID.String())
}

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

	// Disable rate limiting in tests by swapping the package-level limiters
	// with permissive ones BEFORE handlers register and capture them. The
	// handlers store the *httprate.RateLimiter pointer at registration time,
	// so this must happen before any *.Register(apiGroup) call.
	disableRateLimitsForTests()

	// SessionUser middleware is required by /auth/* and any handler calling
	// RequireSessionCtx. Tests run over plain HTTP via httptest, so disable
	// the Secure cookie flag to let cookies round-trip.
	testCfg := &config.Config{SecureCookie: false}
	queries := generated.New(testPool)
	api.UseMiddleware(middleware.NewSessionMiddleware(testCfg, queries).Handler(api))

	apiGroup := huma.NewGroup(api, "/api")

	repos := repository.New(testPool)
	txRunner := repository.NewTxRunner(testPool)
	NewHealthHandler(testPool).Register(apiGroup)
	NewBoardsHandler(service.NewBoards(repos.Boards, queries)).Register(apiGroup)
	NewCellsHandler(service.NewCells(repos.Cells, repos.Boards, queries)).Register(apiGroup)
	NewGameCellsHandler(service.NewGameCells(repos.GameCells, repos.Games, queries)).Register(apiGroup)
	NewGamesHandler(service.NewGames(repos.Games, queries)).Register(apiGroup)
	votesSvc := service.NewVotes(repos.Votes, queries)
	NewVotesHandler(votesSvc).Register(apiGroup)
	NewAuthHandler(service.NewAuth(repos, fakeMailer, queries, txRunner)).Register(apiGroup)
	NewUsersHandler(service.NewUsers(repos, queries, fakeMailer, txRunner), votesSvc).Register(apiGroup)

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
		"TRUNCATE TABLE game_cells, games, votes, cells, boards, sessions, verification_tokens, user_passwords, users RESTART IDENTITY CASCADE")
}

// cleanupTables truncates all tables in the correct order (respecting foreign keys).
func cleanupTables(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"TRUNCATE TABLE game_cells, games, votes, cells, boards, sessions, verification_tokens, user_passwords, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to clean up tables: %v", err)
	}
	// Cookies reference session rows we just truncated; drop the cache so
	// the next test can't reuse a now-invalid token.
	for k := range userCookies {
		delete(userCookies, k)
	}
	// Fake mailer captures cross every test boundary, drop them too.
	fakeMailer.reset()
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

// assertJSONInt asserts that the JSON field at key equals the expected
// integer value. JSON numbers decode as float64 by default; this helper
// centralizes the cast and avoids repeating it in every test.
func assertJSONInt(t *testing.T, data map[string]any, key string, expected int64) {
	t.Helper()
	val, ok := data[key]
	if !ok {
		t.Errorf("expected field %q in response, but it was missing", key)
		return
	}
	f, ok := val.(float64)
	if !ok {
		t.Errorf("expected field %q to be a number, got %T", key, val)
		return
	}
	if int64(f) != expected {
		t.Errorf("expected %s = %d, got %d", key, expected, int64(f))
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
// Pass an empty string for email to omit it from the request. The session
// cookie returned by /auth/register is captured into userCookies so other
// factories (and tests) can replay it via cookiesFor.
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
	userID := user.ID.String()
	if c := extractSessionCookie(w); c != nil {
		userCookies[userID] = c
	}
	return userID
}

// createTestUserWithRegister registers a user via the auth API and returns the
// full response body (so callers can inspect cookies / id / email handling).
// Pass nil for email to omit it from the request. The session cookie set by
// /auth/register is also cached in userCookies for later replay.
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

	if c := extractSessionCookie(w); c != nil {
		if id, ok := resp["user_id"].(string); ok && id != "" {
			userCookies[id] = c
		}
	}
	return &resp
}

// mintAnonSession inserts an anonymous session row directly and returns a
// cookie that can be replayed against the test router. This is the only way
// to obtain a pre-existing anon session because no anon-friendly handler
// currently calls RequireSessionCtx.
func mintAnonSession(t *testing.T, ttl time.Duration) (token string, cookie *http.Cookie) {
	t.Helper()
	q := generated.New(testPool)
	tok := auth.GenerateToken()
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
// description to omit it from the request. The caller is authenticated as
// authorID via the captured session cookie; the server reads the author from
// the session, not from a body field.
func createTestBoard(t *testing.T, title string, size int, authorID string, description *string) string {
	t.Helper()
	body := map[string]any{
		"title": title,
		"size":  size,
	}
	if description != nil {
		body["description"] = *description
	}
	w := doRequestWithCookies(http.MethodPost, "/api/boards", body, cookiesFor(authorID))
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["board_id"].(string)
}

// createTestCell creates a cell on a board via the API and returns its ID.
// Acts as the board's author.
func createTestCell(t *testing.T, boardID, content string) string {
	t.Helper()
	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": content,
	}, cookiesForBoard(t, boardID))
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["cell_id"].(string)
}

// createTestGame creates a game via the API and returns its ID. The caller is
// authenticated as playerID via the captured session cookie; the server reads
// the player from the session, not from a query param.
func createTestGame(t *testing.T, playerID, boardID string) string {
	t.Helper()
	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/boards/%s/games", boardID), nil, cookiesFor(playerID))
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["game_id"].(string)
}

// createTestVote upserts a vote on a board for a user and returns the response
// body. The caller is authenticated as userID via the captured session cookie.
func createTestVote(t *testing.T, boardID, userID string, value int) map[string]any {
	t.Helper()
	w := doRequestWithCookies(
		http.MethodPut,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		map[string]any{"vote_value": value},
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp
}

// createTestGameCell creates a single game cell on a game and returns its ID.
// Use createTestGameCells for bulk creation when you need multiple cells.
// Acts as the game's player.
func createTestGameCell(t *testing.T, gameID, cellID, content string, position int) string {
	t.Helper()
	body := []map[string]any{
		{"cell_id": cellID, "content": content, "position": position},
	}
	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), body, cookiesForGame(t, gameID))
	assertStatus(t, w, http.StatusOK)
	var resp []map[string]any
	decodeJSON(t, w, &resp)
	if len(resp) == 0 {
		t.Fatalf("expected at least one game cell in response, got 0")
	}
	return resp[0]["game_cell_id"].(string)
}

// createAnonGame creates a game via the API using the supplied cookie
// directly (instead of looking it up in userCookies). Pass nil for cookie to
// send the request without any cookie at all -- the server will mint a fresh
// anonymous session and emit a Set-Cookie header. Returns the new game ID.
func createAnonGame(t *testing.T, cookie *http.Cookie, boardID string) string {
	t.Helper()
	var cookies []*http.Cookie
	if cookie != nil {
		cookies = []*http.Cookie{cookie}
	}
	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/boards/%s/games", boardID), nil, cookies)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["game_id"].(string)
}

// disableRateLimitsForTests swaps the package-level rate limiters with
// permissive limiters that will not fire during a test run. Must be called
// before handlers register operations, since handlers capture the
// *httprate.RateLimiter pointer at huma.Register time.
//
// Tests run many requests in rapid succession from the same RemoteAddr
// (httptest.NewRequest defaults to 192.0.2.1:1234) and share session ids
// across factories, so production limits would otherwise trip and turn the
// suite red regardless of correctness.
func disableRateLimitsForTests() {
	huge := func() *httprate.RateLimiter {
		return httprate.NewRateLimiter(1_000_000, time.Minute)
	}
	middleware.StrictAuthLimiter = huge()
	middleware.HeavyAuthLimiter = huge()
	middleware.WriteLimiter = huge()
	middleware.WriteHeavyLimiter = huge()
	middleware.ReadLimiter = huge()
	middleware.ReadListLimiter = huge()
	middleware.GameplayLimiter = huge()
	middleware.HealthLimiter = huge()
}
