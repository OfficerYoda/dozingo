package handler

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

/// ===== /users/me =====

func TestMe_NoCookie_401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/users/me", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestMe_AnonymousSessionOnly_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestMe_Authenticated(t *testing.T) {
	setupTest(t)

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

	w := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	var me map[string]any
	decodeJSON(t, w, &me)

	assertJSONField(t, me, "username", "meuser")
	assertJSONField(t, me, "email", "me@example.com")
	if me["user_id"] != registered["user_id"] {
		t.Errorf("expected /me id %v to match register id %v", me["user_id"], registered["user_id"])
	}
}

func TestMe_AfterLogout_401(t *testing.T) {
	setupTest(t)

	r := doRequest(http.MethodPost, "/api/auth/register", map[string]any{
		"username": "logmeout",
		"password": "pw12345678",
	})
	assertStatus(t, r, http.StatusOK)
	cookie := extractSessionCookie(r)

	logout := doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, []*http.Cookie{cookie})
	assertStatus(t, logout, http.StatusNoContent)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

/// ===== GET /users/{user_id} =====

func TestUserByID_Success(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "byiduser", "mypassword123", stringPtr("byid@example.com"))
	userID, ok := (*resp)["user_id"].(string)
	if !ok || userID == "" {
		t.Fatalf("expected non-empty user_id from register, got %v", (*resp)["user_id"])
	}

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", userID), nil)
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)

	assertJSONField(t, got, "user_id", userID)
	assertJSONField(t, got, "username", "byiduser")
	assertJSONField(t, got, "email", "byid@example.com")
}

func TestUserByID_NotFound(t *testing.T) {
	setupTest(t)

	missing := uuid.NewString()

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", missing), nil)
	assertStatus(t, w, http.StatusNotFound)

	var got map[string]any
	decodeJSON(t, w, &got)
	assertJSONField(t, got, "detail", "not found")
}

func TestUserByID_InvalidUUID_Rejected(t *testing.T) {
	setupTest(t)

	// Huma's format:"uuid" validator must reject non-UUID path segments
	// before the handler runs.
	w := doRequest(http.MethodGet, "/api/users/not-a-uuid", nil)
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestUserByID_DoesNotLeakInternalDetails(t *testing.T) {
	setupTest(t)

	missing := uuid.NewString()
	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", missing), nil)
	assertStatus(t, w, http.StatusNotFound)

	rawBody := w.Body.String()
	for _, leak := range []string{
		"sql",
		"pgx",
		"no rows",
		"constraint",
		"SQLSTATE",
	} {
		if strings.Contains(strings.ToLower(rawBody), strings.ToLower(leak)) {
			t.Errorf("response body must not leak %q, got: %s", leak, rawBody)
		}
	}
}
