package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
)

func registerTestUser(t *testing.T, username, password string, email *string) *map[string]any {
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

func stringPtr(s string) *string {
	return &s
}

func TestRegister_Success(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	body := map[string]any{
		"username": "newuser",
		"password": "mypassword123",
		"email":    "newuser@example.com",
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "username", "newuser")
	assertJSONField(t, resp, "email", "newuser@example.com")
}

func TestRegister_WithoutEmail(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	body := map[string]any{
		"username": "noemailuser",
		"password": "mypassword123",
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "username", "noemailuser")

	// Email should be null/nil when not provided
	if resp["email"] != nil {
		t.Errorf("expected email to be null, got %v", resp["email"])
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	registerTestUser(t, "duplicate", "password1", stringPtr("dup1@example.com"))

	// Try registering with the same username
	body := map[string]any{
		"username": "duplicate",
		"password": "password2",
		"email":    "dup2@example.com",
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusConflict)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	registerTestUser(t, "user1", "password1", stringPtr("same@example.com"))

	// Try registering with different username but same email
	body := map[string]any{
		"username": "user2",
		"password": "password2",
		"email":    "same@example.com",
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusConflict)
}

func TestLogin_Success(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	registerTestUser(t, "loginuser", "correctpassword", stringPtr("login@example.com"))

	body := map[string]any{
		"username": "loginuser",
		"password": "correctpassword",
	}

	w := doRequest(http.MethodPost, "/api/auth/login", body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "username", "loginuser")
	assertJSONField(t, resp, "email", "login@example.com")
}

func TestLogin_WrongPassword(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	registerTestUser(t, "wrongpwuser", "correctpassword", stringPtr("wrongpw@example.com"))

	body := map[string]any{
		"username": "wrongpwuser",
		"password": "wrongpassword",
	}

	w := doRequest(http.MethodPost, "/api/auth/login", body)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestLogin_NonexistentUser(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	body := map[string]any{
		"username": "doesnotexist",
		"password": "somepassword",
	}

	w := doRequest(http.MethodPost, "/api/auth/login", body)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestLogin_WithoutEmail(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	registerTestUser(t, "noemaillogin", "mypassword", nil)

	body := map[string]any{
		"username": "noemaillogin",
		"password": "mypassword",
	}

	w := doRequest(http.MethodPost, "/api/auth/login", body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "username", "noemaillogin")

	// Email should be null for user registered without email
	if resp["email"] != nil {
		t.Errorf("expected email to be null, got %v", resp["email"])
	}
}

/// ===== Helpers for cookie-driven tests =====

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

func loadSessionByToken(t *testing.T, token string) (generated.GetSessionUserByTokenRow, bool) {
	t.Helper()
	q := generated.New(testPool)
	row, err := q.GetSessionUserByToken(context.Background(), token)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, false
	}
	return row, true
}

/// ===== Register: cookies & ID =====

func TestRegister_ReturnsID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	resp := registerTestUser(t, "iduser", "mypassword123", stringPtr("iduser@example.com"))

	id, ok := (*resp)["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id field, got %v", (*resp)["id"])
	}
	// Must be parseable as a UUID
	if _, err := uuidFromString(id); err != nil {
		t.Errorf("returned id %q is not a valid UUID: %v", id, err)
	}
}

func TestRegister_SetsSessionCookie(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "cookieuser",
		"password": "mypassword123",
	})
	assertStatus(t, w, http.StatusOK)

	c := extractSessionCookie(w)
	if c == nil {
		t.Fatal("expected Set-Cookie: session_token=...; missing")
	}
	if c.Value == "" {
		t.Error("expected non-empty session token in cookie")
	}
	if !c.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if c.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite=Strict, got %v", c.SameSite)
	}
	if c.Path != "/" {
		t.Errorf("expected Path=/, got %q", c.Path)
	}
	// Test setup disabled SecureCookie so it should be false here.
	if c.Secure {
		t.Error("Secure flag must not be set in test mode")
	}
}

func TestRegister_AttachesExistingAnonSession(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	token, cookie := mintAnonSession(t, 30*24*time.Hour)

	// Confirm pre-state: anon row, no user_id.
	pre, ok := loadSessionByToken(t, token)
	if !ok {
		t.Fatal("anon session not visible before register")
	}
	if pre.UserID.Valid {
		t.Fatal("expected anon session to start with NULL user_id")
	}

	w := doRequestWithCookies(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "anonattach",
		"password": "anonattach123",
	}, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	// The same token should now be bound to the new user.
	post, ok := loadSessionByToken(t, token)
	if !ok {
		t.Fatal("session disappeared after register")
	}
	if !post.UserID.Valid {
		t.Fatal("expected session to be bound to a user after register")
	}
	if post.Username.String != "anonattach" {
		t.Errorf("expected session username 'anonattach', got %q", post.Username.String)
	}
}

func TestRegister_PasswordTooLong_Rejected(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "longpw",
		"password": strings.Repeat("a", 73),
	})
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for >72 byte password, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestRegister_EmailWhitespaceTreatedAsNull(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "wsemail",
		"password": "mypassword123",
		"email":    "   ",
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)
	if resp["email"] != nil {
		t.Errorf("expected whitespace-only email to be normalised to null, got %v", resp["email"])
	}
}

/// ===== Login: cookies & anon-attach =====

func TestLogin_SetsSessionCookie(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	registerTestUser(t, "logincookie", "pw123", stringPtr("lc@example.com"))

	w := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "logincookie",
		"password": "pw123",
	})
	assertStatus(t, w, http.StatusOK)

	if c := extractSessionCookie(w); c == nil || c.Value == "" {
		t.Fatal("expected login to set a session_token cookie")
	}
}

func TestLogin_AttachesExistingAnonSession(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	// Register a user first (this will create its own session, but we don't
	// reuse it here -- we want a fresh anon cookie below).
	registerTestUser(t, "anonloginer", "pw123", nil)

	token, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "anonloginer",
		"password": "pw123",
	}, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	post, ok := loadSessionByToken(t, token)
	if !ok {
		t.Fatal("anon session disappeared after login")
	}
	if !post.UserID.Valid {
		t.Fatal("expected anon session to become user-bound after login")
	}
	if post.Username.String != "anonloginer" {
		t.Errorf("expected session username 'anonloginer', got %q", post.Username.String)
	}
}

/// ===== /auth/me =====

func TestMe_NoCookie_401(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodGet, "/api/auth/me", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestMe_AnonymousSessionOnly_401(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodGet, "/api/auth/me", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestMe_Authenticated(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	registerResp := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "meuser",
		"password": "mypw1234",
		"email":    "me@example.com",
	})
	assertStatus(t, registerResp, http.StatusOK)
	cookie := extractSessionCookie(registerResp)
	if cookie == nil {
		t.Fatal("register should have set a session cookie")
	}

	var registered map[string]any
	decodeJSON(t, registerResp, &registered)

	w := doRequestWithCookies(http.MethodGet, "/api/auth/me", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	var me map[string]any
	decodeJSON(t, w, &me)

	assertJSONField(t, me, "username", "meuser")
	assertJSONField(t, me, "email", "me@example.com")
	if me["id"] != registered["id"] {
		t.Errorf("expected /me id %v to match register id %v", me["id"], registered["id"])
	}
}

func TestMe_AfterLogout_401(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	r := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "logmeout",
		"password": "pw12345",
	})
	assertStatus(t, r, http.StatusOK)
	cookie := extractSessionCookie(r)

	logout := doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, []*http.Cookie{cookie})
	assertStatus(t, logout, http.StatusNoContent)

	w := doRequestWithCookies(http.MethodGet, "/api/auth/me", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

/// ===== /auth/logout =====

func TestLogout_NoCookie_NoOp(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodPost, "/api/auth/logout", nil)
	assertStatus(t, w, http.StatusNoContent)
}

func TestLogout_AuthenticatedDeletesSessionAndClearsCookie(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	r := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "logoutuser",
		"password": "pw12345",
	})
	assertStatus(t, r, http.StatusOK)
	cookie := extractSessionCookie(r)
	if cookie == nil {
		t.Fatal("register should have set a session cookie")
	}
	token := cookie.Value

	// Sanity: row exists and is bound to the user.
	if pre, ok := loadSessionByToken(t, token); !ok {
		t.Fatal("session row not found before logout")
	} else if !pre.UserID.Valid {
		t.Fatal("session row must be user-bound after register")
	}

	w := doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusNoContent)

	// Cookie should be cleared
	clearing := extractSessionCookie(w)
	if clearing == nil {
		t.Fatal("logout should emit a Set-Cookie clearing session_token")
	}
	if clearing.MaxAge >= 0 {
		t.Errorf("expected MaxAge<0 on clearing cookie, got %d", clearing.MaxAge)
	}

	// DB row should be gone
	if _, ok := loadSessionByToken(t, token); ok {
		t.Fatal("session row should be deleted after logout")
	}
}

func TestLogout_StaleCookie_NoOp(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	staleCookie := &http.Cookie{Name: "session_token", Value: "definitely-not-a-real-token"}
	w := doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, []*http.Cookie{staleCookie})
	assertStatus(t, w, http.StatusNoContent)
}
