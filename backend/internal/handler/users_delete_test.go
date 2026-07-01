package handler

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

/// ===== DELETE /users/me =====

func TestDeleteUser_Unauthenticated_401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodDelete, "/api/users/me", map[string]any{
		"password": "somepassword1",
	})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestDeleteUser_AnonymousSessionOnly_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodDelete, "/api/users/me", map[string]any{
		"password": "somepassword1",
	}, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestDeleteUser_Success(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "delme", "correctpw1", stringPtr("delme@example.com"))
	userID := (*resp)["user_id"].(string)
	cookie := cookiesFor(userID)

	w := doRequestWithCookies(http.MethodDelete, "/api/users/me", map[string]any{
		"password": "correctpw1",
	}, cookie)
	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Fatalf("expected 200/204 for successful delete, got %d (body: %s)", w.Code, w.Body.String())
	}

	// The user must no longer be findable by ID.
	get := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", userID), nil)
	assertStatus(t, get, http.StatusNotFound)

	// The session must be invalidated — /me must return 401.
	me := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, cookie)
	assertStatus(t, me, http.StatusUnauthorized)
}

func TestDeleteUser_WrongPassword_401(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "delwrong", "correctpw1", stringPtr("delwrong@example.com"))
	userID := (*resp)["user_id"].(string)
	cookie := cookiesFor(userID)

	w := doRequestWithCookies(http.MethodDelete, "/api/users/me", map[string]any{
		"password": "wrongpassword1",
	}, cookie)
	assertStatus(t, w, http.StatusUnauthorized)

	// The user must still exist.
	get := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", userID), nil)
	assertStatus(t, get, http.StatusOK)
}

func TestDeleteUser_PasswordTooShort_422(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "delshort", "correctpw1", stringPtr("delshort@example.com"))
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodDelete, "/api/users/me", map[string]any{
		"password": "short",
	}, cookiesFor(userID))
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for password below minLength, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestDeleteUser_PasswordTooLong_422(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "dellong", "correctpw1", stringPtr("dellong@example.com"))
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodDelete, "/api/users/me", map[string]any{
		"password": strings.Repeat("a", 73),
	}, cookiesFor(userID))
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for password above maxLength, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestDeleteUser_InvalidatesAllSessions(t *testing.T) {
	setupTest(t)

	// Register produces session A.
	resp := createTestUserWithRegister(t, "delmulti", "correctpw1", stringPtr("delmulti@example.com"))
	userID := (*resp)["user_id"].(string)
	cookieA := cookiesFor(userID)[0]

	// A second login produces session B.
	loginResp := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"identifier": "delmulti",
		"password":   "correctpw1",
	})
	assertStatus(t, loginResp, http.StatusOK)
	cookieB := extractSessionCookie(loginResp)
	if cookieA == nil || cookieB == nil || cookieA.Value == cookieB.Value {
		t.Fatalf("expected two distinct session cookies, got A=%v B=%v", cookieA, cookieB)
	}

	// Delete the account using session B.
	w := doRequestWithCookies(http.MethodDelete, "/api/users/me", map[string]any{
		"password": "correctpw1",
	}, []*http.Cookie{cookieB})
	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Fatalf("expected 200/204 for successful delete, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Both sessions must be dead.
	for label, c := range map[string]*http.Cookie{"A": cookieA, "B": cookieB} {
		me := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{c})
		if me.Code != http.StatusUnauthorized {
			t.Errorf("session %s: expected 401 after account deletion, got %d (body: %s)", label, me.Code, me.Body.String())
		}
	}
}

func TestDeleteUser_DoesNotLeakDBDetails(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "delleak", "correctpw1", stringPtr("delleak@example.com"))
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodDelete, "/api/users/me", map[string]any{
		"password": "wrongpassword1",
	}, cookiesFor(userID))
	assertResponseDoesNotLeak(t, w.Body.String(),
		"user_passwords", "constraint", "23505", "pgx", "sql:",
		"correctpw1", "wrongpassword1")
}
