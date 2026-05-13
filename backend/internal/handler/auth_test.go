package handler

import (
	"net/http"
	"testing"
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
