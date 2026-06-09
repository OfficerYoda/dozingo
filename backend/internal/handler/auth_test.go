package handler

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRegister_Success(t *testing.T) {
	setupTest(t)

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

	// Register auto-generates a profile picture (best-effort) and
	// uploads it before responding, so avatar_url is always set on a
	// successful registration.
	if resp["avatar_url"] == nil {
		t.Errorf("expected avatar_url to be set on fresh register, got nil")
	}
}

func TestRegister_WithoutEmail(t *testing.T) {
	setupTest(t)

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
	setupTest(t)

	createTestUserWithRegister(t, "duplicate", "password1", stringPtr("dup1@example.com"))

	// Try registering with the same username
	body := map[string]any{
		"username": "duplicate",
		"password": "password2",
		"email":    "dup2@example.com",
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusConflict)

	// Response must not leak internal DB details (constraint name, SQL
	// state, table name, etc.). Only the generic "conflict" detail.
	rawBody := w.Body.String()
	for _, leak := range []string{
		"users_username_key",
		"_key",
		"constraint",
		"users_username",
		"23505",
	} {
		if strings.Contains(rawBody, leak) {
			t.Errorf("response body must not leak %q, got: %s", leak, rawBody)
		}
	}

	var resp map[string]any
	decodeJSON(t, w, &resp)
	assertJSONField(t, resp, "detail", "conflict")
}

func TestRegister_DuplicateEmail(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "user1", "password1", stringPtr("same@example.com"))

	// Try registering with different username but same email
	body := map[string]any{
		"username": "user2",
		"password": "password2",
		"email":    "same@example.com",
	}

	w := doRequest(http.MethodPost, "/api/auth/register", body)
	assertStatus(t, w, http.StatusConflict)

	// Response must not leak internal DB details.
	rawBody := w.Body.String()
	for _, leak := range []string{
		"users_email_key",
		"_key",
		"constraint",
		"users_email",
		"23505",
	} {
		if strings.Contains(rawBody, leak) {
			t.Errorf("response body must not leak %q, got: %s", leak, rawBody)
		}
	}

	var resp map[string]any
	decodeJSON(t, w, &resp)
	assertJSONField(t, resp, "detail", "conflict")
}

func TestLogin_Success(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "loginuser", "correctpassword", stringPtr("login@example.com"))

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

	// Login response shape mirrors register; the user got an
	// auto-generated avatar at registration time, so avatar_url is set.
	if resp["avatar_url"] == nil {
		t.Errorf("expected avatar_url to be set on login for user without custom avatar, got nil")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "wrongpwuser", "correctpassword", stringPtr("wrongpw@example.com"))

	body := map[string]any{
		"username": "wrongpwuser",
		"password": "wrongpassword",
	}

	w := doRequest(http.MethodPost, "/api/auth/login", body)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestLogin_NonexistentUser(t *testing.T) {
	setupTest(t)

	body := map[string]any{
		"username": "doesnotexist",
		"password": "somepassword",
	}

	w := doRequest(http.MethodPost, "/api/auth/login", body)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestLogin_WithoutEmail(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "noemaillogin", "mypassword", nil)

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

/// ===== Register: cookies & ID =====

func TestRegister_ReturnsID(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "iduser", "mypassword123", stringPtr("iduser@example.com"))

	id, ok := (*resp)["user_id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty id field, got %v", (*resp)["user_id"])
	}
}

func TestRegister_SetsSessionCookie(t *testing.T) {
	setupTest(t)

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
	setupTest(t)

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
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "longpw",
		"password": strings.Repeat("a", 73),
	})
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for >72 byte password, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestRegister_EmailWhitespaceOnly_Rejected(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "wsemail",
		"password": "mypassword123",
		"email":    "   ",
	})
	// Whitespace-only is not a valid RFC 5322 address; Huma's
	// format:"email" validator must reject it before the handler runs.
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for whitespace-only email, got %d (body: %s)", w.Code, w.Body.String())
	}
}

/// ===== Login: cookies & anon-attach =====

func TestLogin_SetsSessionCookie(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "logincookie", "pw12345678", stringPtr("lc@example.com"))

	w := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "logincookie",
		"password": "pw12345678",
	})
	assertStatus(t, w, http.StatusOK)

	if c := extractSessionCookie(w); c == nil || c.Value == "" {
		t.Fatal("expected login to set a session_token cookie")
	}
}

func TestLogin_AttachesExistingAnonSession(t *testing.T) {
	setupTest(t)

	// Register a user first (this will create its own session, but we don't
	// reuse it here -- we want a fresh anon cookie below).
	createTestUserWithRegister(t, "anonloginer", "pw12345678", nil)

	token, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "anonloginer",
		"password": "pw12345678",
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

/// ===== /auth/logout =====

func TestLogout_NoCookie_NoOp(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/logout", nil)
	assertStatus(t, w, http.StatusNoContent)
}

func TestLogout_AuthenticatedDeletesSessionAndClearsCookie(t *testing.T) {
	setupTest(t)

	r := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "logoutuser",
		"password": "pw12345678",
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
	setupTest(t)

	staleCookie := &http.Cookie{Name: "session_token", Value: "definitely-not-a-real-token"}
	w := doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, []*http.Cookie{staleCookie})
	assertStatus(t, w, http.StatusNoContent)
}
