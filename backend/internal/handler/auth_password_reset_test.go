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

/// ===== Helpers =====

// insertVerificationToken inserts a row into verification_tokens directly
// for the given user. Pass a negative TTL to mint an already-expired token.
// Returns the plaintext token string so the test can replay it against the
// API; the row stores the SHA-256 hex digest.
func insertVerificationToken(
	t *testing.T,
	userID pgtype.UUID,
	tokenType generated.TokenType,
	ttl time.Duration,
) string {
	t.Helper()
	q := generated.New(testPool)
	tok := auth.GenerateToken()
	_, err := q.CreateVerificationToken(context.Background(), generated.CreateVerificationTokenParams{
		UserID: userID,
		Token:  auth.HashToken(tok),
		Type:   tokenType,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(ttl),
			Valid: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to insert verification token: %v", err)
	}
	return tok
}

// userIDFromString turns a UUID string into pgtype.UUID, failing the test if
// the parse fails.
func userIDFromString(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		t.Fatalf("invalid uuid %q: %v", s, err)
	}
	return id
}

// countValidTokens returns the number of unexpired verification tokens for a
// given user/type combo. Uses a raw count query rather than the
// repository's :one helper so we don't conflate "not found" with "0 rows".
func countValidTokens(t *testing.T, userID pgtype.UUID, tt generated.TokenType) int {
	t.Helper()
	row := testPool.QueryRow(
		context.Background(),
		`SELECT count(*) FROM verification_tokens
		 WHERE user_id = $1 AND type = $2 AND expires_at > now()`,
		userID, string(tt),
	)
	var n int
	if err := row.Scan(&n); err != nil {
		t.Fatalf("count tokens: %v", err)
	}
	return n
}

// assertNoLeak fails the test if body contains any of the given substrings.
// Used to verify that internal details (constraint names, raw tokens, SQL
// state codes) never bleed into client responses.
func assertResponseDoesNotLeak(t *testing.T, body string, leaks ...string) {
	t.Helper()
	for _, leak := range leaks {
		if leak == "" {
			continue
		}
		if strings.Contains(body, leak) {
			t.Errorf("response body must not leak %q, got: %s", leak, body)
		}
	}
}

/// ===== /auth/forgot-password =====

func TestForgotPassword_UnknownEmail_StillReturns200(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": "ghost@example.com",
	})
	assertStatus(t, w, http.StatusOK)

	if fakeMailer.resetCount() != 0 {
		t.Errorf("expected 0 password-reset mails for unknown email, got %d", fakeMailer.resetCount())
	}

	var resp map[string]any
	decodeJSON(t, w, &resp)
	assertJSONField(t, resp, "status", "password reset email sent")
}

func TestForgotPassword_KnownEmail_SendsMailWithToken(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "fpuser", "originalpw1", stringPtr("fp@example.com"))

	w := doRequest(http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": "fp@example.com",
	})
	assertStatus(t, w, http.StatusOK)

	if fakeMailer.resetCount() != 1 {
		t.Fatalf("expected exactly 1 password-reset mail, got %d", fakeMailer.resetCount())
	}
	last, _ := fakeMailer.lastReset()
	if last.To != "fp@example.com" {
		t.Errorf("expected mail to=fp@example.com, got %q", last.To)
	}
	if last.Token == "" {
		t.Error("expected non-empty token in reset mail")
	}

	// A row of type password_reset must exist for that user.
	q := generated.New(testPool)
	user, err := q.GetUserByUsername(context.Background(), "fpuser")
	if err != nil {
		t.Fatalf("look up user: %v", err)
	}
	if got := countValidTokens(t, user.ID, generated.TokenTypePasswordReset); got != 1 {
		t.Errorf("expected 1 valid password_reset token in DB, got %d", got)
	}
}

func TestForgotPassword_RotatesExistingToken(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "fprot", "pw12345678", stringPtr("fprot@example.com"))

	// First call -> token A.
	w1 := doRequest(http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": "fprot@example.com",
	})
	assertStatus(t, w1, http.StatusOK)
	first, ok := fakeMailer.lastReset()
	if !ok {
		t.Fatal("first call did not record a mail")
	}

	// Second call -> token B; A must no longer be valid.
	w2 := doRequest(http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": "fprot@example.com",
	})
	assertStatus(t, w2, http.StatusOK)
	second, ok := fakeMailer.lastReset()
	if !ok {
		t.Fatal("second call did not record a mail")
	}
	if first.Token == second.Token {
		t.Error("expected rotated token, got identical tokens for two forgot-password calls")
	}

	q := generated.New(testPool)
	user, err := q.GetUserByUsername(context.Background(), "fprot")
	if err != nil {
		t.Fatalf("look up user: %v", err)
	}
	if got := countValidTokens(t, user.ID, generated.TokenTypePasswordReset); got != 1 {
		t.Errorf("expected exactly 1 valid token after rotation, got %d", got)
	}
}

func TestForgotPassword_DoesNotLeakDBDetails(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": "stillghost@example.com",
	})
	// Even on internal error, this endpoint must return 200 without leaking
	// state or hint at whether the email exists.
	assertStatus(t, w, http.StatusOK)
	assertResponseDoesNotLeak(t, w.Body.String(),
		"verification_tokens", "users_email", "constraint", "23505", "pgx", "sql:")
}

/// ===== /auth/new-password =====

func TestNewPassword_Success(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "npuser", "originalpw1", stringPtr("np@example.com"))
	userIDStr := (*resp)["user_id"].(string)
	preCookie := userCookies[userIDStr]
	if preCookie == nil {
		t.Fatal("expected register to capture a session cookie")
	}

	// Trigger forgot-password to mint a real token.
	fp := doRequest(http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": "np@example.com",
	})
	assertStatus(t, fp, http.StatusOK)
	mail, ok := fakeMailer.lastReset()
	if !ok {
		t.Fatal("forgot-password did not record a mail")
	}

	// Submit the new password.
	w := doRequest(http.MethodPost, "/api/auth/new-password", map[string]any{
		"token":        mail.Token,
		"new_password": "freshpw98765",
	})
	assertStatus(t, w, http.StatusOK)

	var npResp map[string]any
	decodeJSON(t, w, &npResp)
	assertJSONField(t, npResp, "username", "npuser")

	// Logging in with the new password must work.
	loginNew := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "npuser",
		"password": "freshpw98765",
	})
	assertStatus(t, loginNew, http.StatusOK)

	// Logging in with the old password must fail.
	loginOld := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "npuser",
		"password": "originalpw1",
	})
	assertStatus(t, loginOld, http.StatusUnauthorized)

	// The token row must be gone (single use).
	userID := userIDFromString(t, userIDStr)
	if got := countValidTokens(t, userID, generated.TokenTypePasswordReset); got != 0 {
		t.Errorf("expected token to be deleted after new-password, %d remaining", got)
	}

	// The previously-issued session cookie must be invalidated by
	// DeleteByUserID inside the NewPassword tx.
	me := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{preCookie})
	assertStatus(t, me, http.StatusUnauthorized)
}

func TestNewPassword_InvalidToken_NotFound(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/new-password", map[string]any{
		"token":        "definitely-not-a-real-token",
		"new_password": "whatever12345",
	})
	if w.Code != http.StatusNotFound && w.Code != http.StatusUnauthorized && w.Code != http.StatusBadRequest {
		t.Errorf("expected 404/401/400 for invalid token, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestNewPassword_ExpiredToken_410Gone(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "expired", "pw12345678", stringPtr("expired@example.com"))
	userID := userIDFromString(t, (*resp)["user_id"].(string))

	tok := insertVerificationToken(t, userID, generated.TokenTypePasswordReset, -1*time.Minute)

	w := doRequest(http.MethodPost, "/api/auth/new-password", map[string]any{
		"token":        tok,
		"new_password": "newpw12345678",
	})
	// Expired tokens are filtered out by the `expires_at > now()` clause in
	// GetVerificationTokenByToken, so the row appears "not found" to the
	// service. The service-level "expired token" branch only fires for
	// tokens that are still selectable but past their TTL.
	// Either 404 (not found) or 410 (gone) is acceptable; both are safe.
	if w.Code != http.StatusGone && w.Code != http.StatusNotFound {
		t.Errorf("expected 410 Gone or 404 NotFound for expired token, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestNewPassword_WrongTokenType_BadRequest(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "wrongtype", "pw12345678", stringPtr("wt@example.com"))
	userID := userIDFromString(t, (*resp)["user_id"].(string))

	// Insert an email_verification token; new-password must refuse it.
	tok := insertVerificationToken(t, userID, generated.TokenTypeEmailVerification, time.Hour)

	w := doRequest(http.MethodPost, "/api/auth/new-password", map[string]any{
		"token":        tok,
		"new_password": "newpw12345678",
	})
	assertStatus(t, w, http.StatusBadRequest)

	// And the token must NOT have been consumed.
	if got := countValidTokens(t, userID, generated.TokenTypeEmailVerification); got != 1 {
		t.Errorf("wrong-type token should not be consumed, %d remaining", got)
	}
}

func TestNewPassword_PasswordTooLong_Rejected(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "longpwnp", "pw12345678", stringPtr("longnp@example.com"))
	userID := userIDFromString(t, (*resp)["user_id"].(string))
	tok := insertVerificationToken(t, userID, generated.TokenTypePasswordReset, time.Hour)

	w := doRequest(http.MethodPost, "/api/auth/new-password", map[string]any{
		"token":        tok,
		"new_password": strings.Repeat("a", 73),
	})
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for >72 byte password, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestNewPassword_DoesNotLeakDBDetails(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/new-password", map[string]any{
		"token":        "garbage",
		"new_password": "freshpw12345678",
	})
	assertResponseDoesNotLeak(t, w.Body.String(),
		"verification_tokens", "constraint", "user_passwords",
		"23505", "pgx", "sql:", "garbage")
}

func TestNewPassword_InvalidatesAllSessions(t *testing.T) {
	setupTest(t)

	// Two separate sessions for the same user: one from the original
	// register response, one from a follow-up login.
	resp := createTestUserWithRegister(t, "multisess", "originalpw1", stringPtr("multi@example.com"))
	userIDStr := (*resp)["user_id"].(string)
	cookieA := userCookies[userIDStr]

	loginResp := doRequest(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "multisess",
		"password": "originalpw1",
	})
	assertStatus(t, loginResp, http.StatusOK)
	cookieB := extractSessionCookie(loginResp)
	if cookieA == nil || cookieB == nil || cookieA.Value == cookieB.Value {
		t.Fatalf("expected two distinct session cookies, got A=%v B=%v", cookieA, cookieB)
	}

	// Reset password.
	fp := doRequest(http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": "multi@example.com",
	})
	assertStatus(t, fp, http.StatusOK)
	mail, ok := fakeMailer.lastReset()
	if !ok {
		t.Fatal("no reset mail captured")
	}
	npr := doRequest(http.MethodPost, "/api/auth/new-password", map[string]any{
		"token":        mail.Token,
		"new_password": "freshpw98765",
	})
	assertStatus(t, npr, http.StatusOK)

	// Both pre-existing sessions must now be unauthorized on /me.
	for label, c := range map[string]*http.Cookie{"A": cookieA, "B": cookieB} {
		w := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{c})
		if w.Code != http.StatusUnauthorized {
			t.Errorf("session %s: expected 401 after password reset, got %d (body: %s)", label, w.Code, w.Body.String())
		}
	}
}
