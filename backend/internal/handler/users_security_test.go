package handler

import (
	"net/http"
	"testing"
	"time"
)

/// ===== GET /users/me/security =====

func TestSecurityInformation_NoCookie_401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/users/me/security", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestSecurityInformation_AnonSession_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me/security", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestSecurityInformation_Authenticated(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "secuser", "pw12345678", nil)
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me/security", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var body map[string]any
	decodeJSON(t, w, &body)

	// All three fields must be present.
	for _, field := range []string{"password_last_changed_at", "active_sessions", "last_login_at"} {
		if _, ok := body[field]; !ok {
			t.Errorf("expected field %q in response, but it was missing", field)
		}
	}

	// password_last_changed_at must be a non-zero timestamp string.
	if ts, ok := body["password_last_changed_at"].(string); !ok || ts == "" || ts == "0001-01-01T00:00:00Z" {
		t.Errorf("expected non-zero password_last_changed_at, got %v", body["password_last_changed_at"])
	}

	// last_login_at must be a non-zero timestamp string (current session exists).
	if ts, ok := body["last_login_at"].(string); !ok || ts == "" || ts == "0001-01-01T00:00:00Z" {
		t.Errorf("expected non-zero last_login_at, got %v", body["last_login_at"])
	}

	// Registration creates exactly 1 session.
	assertJSONInt(t, body, "active_sessions", 1)
}

func TestSecurityInformation_ActiveSessionsCount(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "secmulti", "pw12345678", nil)
	userID := (*resp)["user_id"].(string)

	// Open a second session via login.
	loginResp := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "secmulti",
		"password": "pw12345678",
	})
	assertStatus(t, loginResp, http.StatusOK)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me/security", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var body map[string]any
	decodeJSON(t, w, &body)

	assertJSONInt(t, body, "active_sessions", 2)
}

func TestSecurityInformation_ActiveSessionsCount_ExcludesExpired(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "secexpired", "pw12345678", nil)
	userID := (*resp)["user_id"].(string)
	cookie := cookiesFor(userID)

	// Open a second session via login, then log it out immediately.
	loginResp := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "secexpired",
		"password": "pw12345678",
	})
	assertStatus(t, loginResp, http.StatusOK)
	cookieB := extractSessionCookie(loginResp)

	logout := doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, []*http.Cookie{cookieB})
	assertStatus(t, logout, http.StatusNoContent)

	// Only the original session from registration should remain.
	w := doRequestWithCookies(http.MethodGet, "/api/users/me/security", nil, cookie)
	assertStatus(t, w, http.StatusOK)

	var body map[string]any
	decodeJSON(t, w, &body)

	assertJSONInt(t, body, "active_sessions", 1)
}
