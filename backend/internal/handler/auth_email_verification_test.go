package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

/// ===== /auth/send-email-verification =====

func TestSendEmailVerification_NoCookie_401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/send-email-verification", nil)
	assertStatus(t, w, http.StatusUnauthorized)
	if fakeMailer.verifyCount() != 0 {
		t.Errorf("expected no verification mail to be sent, got %d", fakeMailer.verifyCount())
	}
}

func TestSendEmailVerification_AnonSession_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/send-email-verification", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestSendEmailVerification_NoEmailOnAccount_401(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "noemailver", "pw12345678", nil)
	cookie := userCookies[(*resp)["user_id"].(string)]

	w := doRequestWithCookies(http.MethodPost, "/api/auth/send-email-verification", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
	if fakeMailer.verifyCount() != 0 {
		t.Errorf("expected no mail when user has no email, got %d", fakeMailer.verifyCount())
	}
}

func TestSendEmailVerification_AlreadyVerified_409(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "alreadyver", "pw12345678", stringPtr("av@example.com"))
	userIDStr := (*resp)["user_id"].(string)
	userID := userIDFromString(t, userIDStr)
	cookie := userCookies[userIDStr]

	// Stamp email_verified_at directly.
	q := generated.New(testPool)
	now := time.Now()
	if _, err := q.SetUserEmailVerifiedAt(context.Background(), generated.SetUserEmailVerifiedAtParams{
		UserID:          userID,
		EmailVerifiedAt: pgmap.PgTimestamptzFromTime(&now),
	}); err != nil {
		t.Fatalf("set email_verified_at: %v", err)
	}

	// Registration sends a verification mail as a side effect; reset so we
	// can assert only on what the endpoint under test does.
	fakeMailer.reset()

	w := doRequestWithCookies(http.MethodPost, "/api/auth/send-email-verification", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusConflict)
	if fakeMailer.verifyCount() != 0 {
		t.Errorf("expected no mail for already-verified user, got %d", fakeMailer.verifyCount())
	}
}

func TestSendEmailVerification_Success(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "sendver", "pw12345678", stringPtr("sv@example.com"))
	userIDStr := (*resp)["user_id"].(string)
	userID := userIDFromString(t, userIDStr)
	cookie := userCookies[userIDStr]

	// Registration sends a verification mail as a side effect; reset so we
	// can assert only on what the endpoint under test does.
	fakeMailer.reset()

	w := doRequestWithCookies(http.MethodPost, "/api/auth/send-email-verification", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	if fakeMailer.verifyCount() != 1 {
		t.Fatalf("expected exactly 1 verification mail, got %d", fakeMailer.verifyCount())
	}
	last, _ := fakeMailer.lastVerify()
	if last.To != "sv@example.com" {
		t.Errorf("expected mail to=sv@example.com, got %q", last.To)
	}
	if last.Token == "" {
		t.Error("expected non-empty token in verification mail")
	}

	if got := countValidTokens(t, userID, generated.TokenTypeEmailVerification); got != 1 {
		t.Errorf("expected 1 valid email_verification token in DB, got %d", got)
	}

	var body map[string]any
	decodeJSON(t, w, &body)
	assertJSONField(t, body, "status", "verification email sent")
}

func TestSendEmailVerification_RotatesExistingToken(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "verrot", "pw12345678", stringPtr("vrot@example.com"))
	userIDStr := (*resp)["user_id"].(string)
	userID := userIDFromString(t, userIDStr)
	cookie := userCookies[userIDStr]

	// Registration sends a verification mail as a side effect; reset so we
	// can assert only on what the endpoint under test does.
	fakeMailer.reset()

	for range 2 {
		w := doRequestWithCookies(http.MethodPost, "/api/auth/send-email-verification", nil, []*http.Cookie{cookie})
		assertStatus(t, w, http.StatusOK)
	}

	if got := countValidTokens(t, userID, generated.TokenTypeEmailVerification); got != 1 {
		t.Errorf("expected exactly 1 valid email_verification token after rotation, got %d", got)
	}
	if fakeMailer.verifyCount() != 2 {
		t.Errorf("expected 2 verification mails recorded, got %d", fakeMailer.verifyCount())
	}
}

/// ===== /auth/verify-email =====

func TestVerifyEmail_Success(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "veuser", "pw12345678", stringPtr("ve@example.com"))
	userIDStr := (*resp)["user_id"].(string)
	userID := userIDFromString(t, userIDStr)
	cookie := userCookies[userIDStr]

	// Mint a real token via /auth/send-email-verification so the test
	// flows like a real client. If that endpoint is buggy the test will
	// fail here, which is the desired behavior.
	send := doRequestWithCookies(http.MethodPost, "/api/auth/send-email-verification", nil, []*http.Cookie{cookie})
	assertStatus(t, send, http.StatusOK)
	mail, ok := fakeMailer.lastVerify()
	if !ok {
		t.Fatal("send-email-verification did not record a mail")
	}

	w := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": mail.Token,
	})
	assertStatus(t, w, http.StatusOK)

	var body map[string]any
	decodeJSON(t, w, &body)
	assertJSONField(t, body, "username", "veuser")
	assertJSONField(t, body, "email", "ve@example.com")

	// users.email_verified_at must be populated.
	q := generated.New(testPool)
	user, err := q.GetUserByUsername(context.Background(), "veuser")
	if err != nil {
		t.Fatalf("look up user: %v", err)
	}
	if !user.EmailVerifiedAt.Valid {
		t.Error("expected email_verified_at to be set after verify-email")
	}

	// Token row must be deleted (single use).
	if got := countValidTokens(t, userID, generated.TokenTypeEmailVerification); got != 0 {
		t.Errorf("expected token deleted after verify-email, %d remaining", got)
	}

	// Existing session must still work.
	me := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{cookie})
	assertStatus(t, me, http.StatusOK)
}

func TestVerifyEmail_InvalidToken_404(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": "definitely-not-a-real-token",
	})
	assertStatus(t, w, http.StatusNotFound)
}

func TestVerifyEmail_ExpiredToken_RejectsAndDoesNotVerify(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "veexp", "pw12345678", stringPtr("veexp@example.com"))
	userID := userIDFromString(t, (*resp)["user_id"].(string))

	tok := insertVerificationToken(t, userID, generated.TokenTypeEmailVerification, stringPtr("veexp@example.com"), -1*time.Minute)

	w := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": tok,
	})
	// Either 404 (filtered out by GetVerificationTokenByToken) or 410 Gone
	// (caught by the explicit ExpiresAt check) is acceptable -- both
	// guarantee the user isn't accidentally verified.
	if w.Code != http.StatusGone && w.Code != http.StatusNotFound {
		t.Errorf("expected 410 Gone or 404 NotFound, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Crucial: email must NOT have been marked verified.
	q := generated.New(testPool)
	user, err := q.GetUserByUsername(context.Background(), "veexp")
	if err != nil {
		t.Fatalf("look up user: %v", err)
	}
	if user.EmailVerifiedAt.Valid {
		t.Error("expired token must not verify the email")
	}
}

func TestVerifyEmail_WrongTokenType_BadRequest(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "vewrong", "pw12345678", stringPtr("vewrong@example.com"))
	userID := userIDFromString(t, (*resp)["user_id"].(string))

	// Insert a password_reset token; verify-email must refuse it.
	tok := insertVerificationToken(t, userID, generated.TokenTypePasswordReset, nil, time.Hour)

	w := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": tok,
	})
	assertStatus(t, w, http.StatusBadRequest)

	q := generated.New(testPool)
	user, err := q.GetUserByUsername(context.Background(), "vewrong")
	if err != nil {
		t.Fatalf("look up user: %v", err)
	}
	if user.EmailVerifiedAt.Valid {
		t.Error("wrong-type token must not verify the email")
	}
	if got := countValidTokens(t, userID, generated.TokenTypePasswordReset); got != 1 {
		t.Errorf("wrong-type token must not be consumed, %d remaining", got)
	}
}

func TestVerifyEmail_TokenIsSingleUse(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "vesingle", "pw12345678", stringPtr("vesingle@example.com"))
	userID := userIDFromString(t, (*resp)["user_id"].(string))

	tok := insertVerificationToken(t, userID, generated.TokenTypeEmailVerification, stringPtr("vesingle@example.com"), time.Hour)

	first := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": tok,
	})
	assertStatus(t, first, http.StatusOK)

	second := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": tok,
	})
	assertStatus(t, second, http.StatusNotFound)
}

func TestVerifyEmail_StaleTokenAfterEmailChange_Gone(t *testing.T) {
	setupTest(t)

	// Register with emailA; this also mints a token for emailA.
	resp := createTestUserWithRegister(t, "vestale", "pw12345678", stringPtr("stale-a@example.com"))
	userIDStr := (*resp)["user_id"].(string)
	userID := userIDFromString(t, userIDStr)
	cookie := userCookies[userIDStr]

	// Capture the token that was sent to emailA during registration.
	fakeMailer.reset() // reset counts; we'll re-request explicitly below
	send := doRequestWithCookies(http.MethodPost, "/api/auth/send-email-verification", nil, []*http.Cookie{cookie})
	assertStatus(t, send, http.StatusOK)
	mailA, ok := fakeMailer.lastVerify()
	if !ok {
		t.Fatal("expected a verification mail for emailA")
	}

	// Change the email to emailB — this rotates the token (old one deleted,
	// new one for emailB issued).
	update := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{"email": "stale-b@example.com"}, cookiesFor(userIDStr))
	assertStatus(t, update, http.StatusOK)

	// Verify that the DB no longer has a valid token for userID
	// associated with emailA (the rotation deleted it).
	if got := countValidTokens(t, userID, generated.TokenTypeEmailVerification); got != 1 {
		t.Fatalf("expected exactly 1 valid token (for emailB) after email change, got %d", got)
	}

	// Now try to use the stale emailA token — it must be rejected.
	w := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": mailA.Token,
	})
	// Token was deleted by rotation, so the DB lookup returns not found.
	if w.Code != http.StatusGone && w.Code != http.StatusNotFound {
		t.Errorf("expected 410 Gone or 404 NotFound for stale token, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Crucial: email must NOT have been verified.
	q := generated.New(testPool)
	user, err := q.GetUserByUsername(context.Background(), "vestale")
	if err != nil {
		t.Fatalf("look up user: %v", err)
	}
	if user.EmailVerifiedAt.Valid {
		t.Error("stale token from old email must not verify the current email")
	}
}

func TestVerifyEmail_DoesNotLeakDBDetails(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/verify-email", map[string]any{
		"token": "abc123-bogus",
	})
	body := w.Body.String()
	assertResponseDoesNotLeak(t, body,
		"verification_tokens", "constraint", "users_email",
		"23505", "pgx", "sql:", "abc123-bogus")
	// also make sure we don't have a giant traceback shape leaking in
	if strings.Contains(body, "pq:") {
		t.Errorf("response body must not leak pq error prefix, got: %s", body)
	}
}
