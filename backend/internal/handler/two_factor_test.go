package handler

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"

	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
)

// ===== Helpers =====

// setup2FA calls POST /api/auth/2fa/setup as the given user and returns the
// TOTP secret from the response. Fails the test if setup does not return 200.
func setup2FA(t *testing.T, userID string) string {
	t.Helper()
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/setup", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	secret, ok := resp["secret"].(string)
	if !ok || secret == "" {
		t.Fatalf("setup2FA: expected non-empty secret in response, got %v", resp)
	}
	return secret
}

// generateTOTPCode produces a valid 6-digit TOTP code for the given secret
// at the current time. Fails the test if generation fails.
func generateTOTPCode(t *testing.T, secret string) string {
	t.Helper()
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("generateTOTPCode: failed to generate code: %v", err)
	}
	return code
}

// confirm2FA calls POST /api/auth/2fa/confirm with a valid TOTP code as the
// given user. Fails the test if confirm does not return 204.
func confirm2FA(t *testing.T, userID, secret string) {
	t.Helper()
	code := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/confirm",
		map[string]any{"code": code}, cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)
}

// load2FARow fetches the user_two_factors row for the given user UUID directly
// from the DB. Returns the row and true, or zero-value and false if not found.
func load2FARow(t *testing.T, userID pgtype.UUID) (generated.UserTwoFactor, bool) {
	t.Helper()
	q := generated.New(testPool)
	row, err := q.GetTwoFactorByUserID(context.Background(), userID)
	if err != nil {
		return generated.UserTwoFactor{}, false
	}
	return row, true
}

// clearSessionPending directly resets two_fa_pending on the session for the
// given plaintext token. Used to manufacture specific session states in tests
// that need to bypass the normal guard flow.
func clearSessionPending(t *testing.T, plaintextToken string) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"UPDATE sessions SET two_fa_pending = false WHERE token = $1",
		auth.HashToken(plaintextToken))
	if err != nil {
		t.Fatalf("clearSessionPending: %v", err)
	}
}

// setSessionPending directly sets two_fa_pending on the session for the
// given plaintext token.
func setSessionPending(t *testing.T, plaintextToken string) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"UPDATE sessions SET two_fa_pending = true WHERE token = $1",
		auth.HashToken(plaintextToken))
	if err != nil {
		t.Fatalf("setSessionPending: %v", err)
	}
}

// ===== Setup =====

func TestSetup2FA_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "setupuser", "setup@example.com")

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/setup", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	if resp["secret"] == nil || resp["secret"] == "" {
		t.Errorf("expected non-empty secret in response")
	}
	if resp["otp_auth_url"] == nil || resp["otp_auth_url"] == "" {
		t.Errorf("expected non-empty otp_auth_url in response")
	}

	// Session must be marked pending after setup
	session, ok := loadSessionByToken(t, userCookies[userID].Value)
	if !ok {
		t.Fatal("session not found after setup")
	}
	if !session.TwoFaPending {
		t.Error("expected session two_fa_pending=true after setup, got false")
	}
}

func TestSetup2FA_Retry_SucceedsWithUpsert(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "retryuser", "retry@example.com")

	// First setup call — marks the session pending.
	secret1 := setup2FA(t, userID)

	// Clear the pending state so setup can be called again (simulating a user
	// who wants to restart setup without confirming first).
	clearSessionPending(t, userCookies[userID].Value)

	// Second setup call before confirming — must succeed (not 409) because
	// Setup uses Upsert instead of Create.
	secret2 := setup2FA(t, userID)

	if secret1 == "" || secret2 == "" {
		t.Error("expected non-empty secrets from both setup calls")
	}

	// Exactly one row must exist for this user
	uid := userIDFromString(t, userID)
	row, ok := load2FARow(t, uid)
	if !ok {
		t.Fatal("expected user_two_factors row to exist after second setup")
	}
	// Row must not be verified
	if row.TotpVerifiedAt.Valid {
		t.Error("expected totp_verified_at to be NULL after setup retry, got non-null")
	}
}

func TestSetup2FA_AlreadyVerified_Returns409(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "verifieduser", "verified@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	// After full setup+confirm, trying to set up again must be rejected
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/setup", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusConflict)
}

func TestSetup2FA_Unauthenticated_Returns401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/2fa/setup", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestSetup2FA_PendingSession_Returns401(t *testing.T) {
	setupTest(t)

	// Setup marks the session as pending. A second call to setup while pending
	// must be rejected because requiresSessionUser blocks pending sessions.
	userID := createTestUser(t, "pendingsetup", "pendingsetup@example.com")
	setup2FA(t, userID) // marks session pending

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/setup", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusUnauthorized)
}

// ===== Confirm =====

func TestConfirm2FA_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "confirmuser", "confirm@example.com")
	secret := setup2FA(t, userID)

	code := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/confirm",
		map[string]any{"code": code}, cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	// Session must no longer be pending
	session, ok := loadSessionByToken(t, userCookies[userID].Value)
	if !ok {
		t.Fatal("session not found after confirm")
	}
	if session.TwoFaPending {
		t.Error("expected session two_fa_pending=false after confirm, got true")
	}

	// TOTP row must be marked verified
	uid := userIDFromString(t, userID)
	row, ok := load2FARow(t, uid)
	if !ok {
		t.Fatal("expected user_two_factors row after confirm")
	}
	if !row.TotpVerifiedAt.Valid {
		t.Error("expected totp_verified_at to be set after confirm")
	}
}

func TestConfirm2FA_InvalidCode_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "badcodeconfirm", "badcode@example.com")
	setup2FA(t, userID)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/confirm",
		map[string]any{"code": "000000"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestConfirm2FA_NotPending_Returns401(t *testing.T) {
	setupTest(t)

	// A non-pending session (fresh register, no setup called) cannot call confirm.
	userID := createTestUser(t, "notpendingconfirm", "notpending@example.com")

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/confirm",
		map[string]any{"code": "123456"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestConfirm2FA_AlreadyVerified_Returns409(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "alreadyverified", "alreadyverified@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	// After confirm, session is no longer pending. Setup is also blocked
	// (2FA already verified). We reach the already-verified confirm path by
	// manually marking the session pending again.
	setSessionPending(t, userCookies[userID].Value)

	uid := userIDFromString(t, userID)
	row, _ := load2FARow(t, uid)
	code := generateTOTPCode(t, row.TotpSecret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/confirm",
		map[string]any{"code": code}, cookiesFor(userID))
	assertStatus(t, w, http.StatusConflict)
}

// ===== Verify =====

func TestVerify2FA_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "verifyuser", "verify@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	// Simulate login: mark session pending as the login flow would do.
	setSessionPending(t, userCookies[userID].Value)

	code := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify",
		map[string]any{"code": code}, cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	// Session must be unlocked after verify
	session, ok := loadSessionByToken(t, userCookies[userID].Value)
	if !ok {
		t.Fatal("session not found after verify")
	}
	if session.TwoFaPending {
		t.Error("expected session two_fa_pending=false after verify, got true")
	}
}

func TestVerify2FA_InvalidCode_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "verifybadcode", "verifybad@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	setSessionPending(t, userCookies[userID].Value)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify",
		map[string]any{"code": "000000"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestVerify2FA_NotPending_Returns401(t *testing.T) {
	setupTest(t)

	// A fully-authenticated, non-pending session cannot call verify.
	userID := createTestUser(t, "notpendingverify", "notpendingverify@example.com")

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify",
		map[string]any{"code": "123456"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestVerify2FA_NotYetSetUp_Returns404(t *testing.T) {
	setupTest(t)

	// A user who never called setup has no totp row; verify must return 404.
	userID := createTestUser(t, "nototp", "nototp@example.com")

	// Manually mark the session pending to bypass requiresPendingSession guard.
	setSessionPending(t, userCookies[userID].Value)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify",
		map[string]any{"code": "123456"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusNotFound)
}

func TestVerify2FA_TotpNotYetVerified_Returns403(t *testing.T) {
	setupTest(t)

	// User has a totp row (setup was called) but never confirmed it.
	// Verify must reject with 403 because totp_verified_at is NULL.
	userID := createTestUser(t, "unverifiedtotp", "unverifiedtotp@example.com")
	setup2FA(t, userID) // session now pending, totp row inserted but not verified

	uid := userIDFromString(t, userID)
	row, ok := load2FARow(t, uid)
	if !ok {
		t.Fatal("expected totp row after setup")
	}

	code := generateTOTPCode(t, row.TotpSecret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify",
		map[string]any{"code": code}, cookiesFor(userID))
	assertStatus(t, w, http.StatusForbidden)
}

// ===== Login enforcement =====

func TestLogin_With2FAEnabled_Returns403(t *testing.T) {
	setupTest(t)

	// Register, complete 2FA setup+confirm, then log out and log back in.
	userID := createTestUser(t, "login2fauser", "login2fa@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))

	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": "login2fauser", "password": "testpassword123"},
		[]*http.Cookie{anonCookie})

	assertStatus(t, w, http.StatusForbidden)

	// Response body must say "two-factor authentication required"
	var resp map[string]any
	decodeJSON(t, w, &resp)
	detail, _ := resp["detail"].(string)
	if detail != msgTwoFARequired {
		t.Errorf("expected detail %q, got %q", msgTwoFARequired, detail)
	}

	// The session must be marked pending in the DB
	session, ok := loadSessionByToken(t, tok)
	if !ok {
		t.Fatal("session not found after login with 2FA")
	}
	if !session.TwoFaPending {
		t.Error("expected session two_fa_pending=true after login with 2FA enabled")
	}
}

func TestLogin_Without2FA_Succeeds(t *testing.T) {
	setupTest(t)

	createTestUser(t, "no2falogin", "no2fa@example.com")

	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": "no2falogin", "password": "testpassword123"},
		[]*http.Cookie{anonCookie})

	assertStatus(t, w, http.StatusOK)

	// Session must NOT be pending
	session, ok := loadSessionByToken(t, tok)
	if !ok {
		t.Fatal("session not found after login without 2FA")
	}
	if session.TwoFaPending {
		t.Error("expected session two_fa_pending=false for user without 2FA, got true")
	}
}

func TestLogin_With2FA_SessionMarkedPending(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "pendingloginuser", "pendinglogin@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))

	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": "pendingloginuser", "password": "testpassword123"},
		[]*http.Cookie{anonCookie})

	// Regardless of HTTP status, the session row in the DB must have pending=true
	session, ok := loadSessionByToken(t, tok)
	if !ok {
		t.Fatal("session not found")
	}
	if !session.TwoFaPending {
		t.Error("expected two_fa_pending=true on session after login with 2FA user")
	}
}

func TestLogin_With2FA_FullFlow(t *testing.T) {
	setupTest(t)

	// End-to-end: register → setup → confirm → logout →
	// login (403) → verify (204) → protected endpoint works.
	userID := createTestUser(t, "fullflowuser", "fullflow@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))

	// Login returns 403 and marks session pending
	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	loginW := doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": "fullflowuser", "password": "testpassword123"},
		[]*http.Cookie{anonCookie})
	assertStatus(t, loginW, http.StatusForbidden)

	// The user calls /verify with the pending session cookie
	pendingCookie := &http.Cookie{Name: "session_token", Value: tok}
	code := generateTOTPCode(t, secret)
	verifyW := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify",
		map[string]any{"code": code}, []*http.Cookie{pendingCookie})
	assertStatus(t, verifyW, http.StatusNoContent)

	// Session must now be fully unlocked
	session, ok := loadSessionByToken(t, tok)
	if !ok {
		t.Fatal("session not found after verify")
	}
	if session.TwoFaPending {
		t.Error("expected two_fa_pending=false after successful verify")
	}

	// Authenticated endpoint (/users/me) must now work
	meW := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{pendingCookie})
	assertStatus(t, meW, http.StatusOK)
}

// ===== Session scoping =====

func TestSetup2FA_OnlyCurrentSessionMarkedPending(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "multisessionuser", "multisession@example.com")
	uid := userIDFromString(t, userID)

	// Insert a second session for the same user directly into the DB
	q := generated.New(testPool)
	otherTok := auth.GenerateToken()
	_, err := q.CreateSession(context.Background(), generated.CreateSessionParams{
		UserID: uid,
		Token:  auth.HashToken(otherTok),
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to insert second session: %v", err)
	}

	// Call setup on the first (registered) session
	setup2FA(t, userID)

	// The first session must be pending
	firstSession, ok := loadSessionByToken(t, userCookies[userID].Value)
	if !ok {
		t.Fatal("first session not found")
	}
	if !firstSession.TwoFaPending {
		t.Error("expected first session two_fa_pending=true after setup")
	}

	// The second session must NOT be pending
	otherSession, ok := loadSessionByToken(t, otherTok)
	if !ok {
		t.Fatal("second session not found")
	}
	if otherSession.TwoFaPending {
		t.Error("expected second session two_fa_pending=false (setup must only affect current session)")
	}
}
