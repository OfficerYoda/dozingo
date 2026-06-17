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
// given user. Fails the test if confirm does not return 200 with recovery codes.
func confirm2FA(t *testing.T, userID, secret string) {
	t.Helper()
	confirm2FAWithCodes(t, userID, secret)
}

// confirm2FAWithCodes is like confirm2FA but returns the plaintext recovery
// codes issued by the server, so callers can use them in subsequent test steps.
func confirm2FAWithCodes(t *testing.T, userID, secret string) []string {
	t.Helper()
	code := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/confirm",
		map[string]any{"code": code}, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	raw, ok := resp["recovery_codes"].([]any)
	if !ok || len(raw) == 0 {
		t.Fatalf("confirm2FAWithCodes: expected non-empty recovery_codes in response, got %v", resp)
	}
	codes := make([]string, len(raw))
	for i, v := range raw {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("confirm2FAWithCodes: expected string recovery code at index %d, got %T", i, v)
		}
		codes[i] = s
	}
	return codes
}

// loginWith2FA logs out the given user, does a fresh login and returns the
// pending session token and cookie. The caller must still call verify-recovery
// or verify to unlock the session.
func loginWith2FA(t *testing.T, username string) (tok string, cookie *http.Cookie) {
	t.Helper()
	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": username, "password": "testpassword123"},
		[]*http.Cookie{anonCookie})
	assertStatus(t, w, http.StatusOK)
	return tok, &http.Cookie{Name: "session_token", Value: tok}
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

	// Second setup call while session is still pending — must succeed because
	// Setup now uses requiresAuthenticatedSession (pending-agnostic) and
	// Upsert instead of Create.
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

// ===== Confirm =====

func TestConfirm2FA_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "confirmuser", "confirm@example.com")
	secret := setup2FA(t, userID)

	code := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/confirm",
		map[string]any{"code": code}, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)
	codes, ok := resp["recovery_codes"].([]any)
	if !ok || len(codes) == 0 {
		t.Error("expected non-empty recovery_codes in confirm response")
	}

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

func TestLogin_With2FAEnabled_ReturnsTwoFAPending(t *testing.T) {
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

	// Login succeeds (200) but signals that 2FA is required
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	pending, _ := resp["two_fa_pending"].(bool)
	if !pending {
		t.Errorf("expected two_fa_pending=true in login response, got %v", resp["two_fa_pending"])
	}
	// No user data should be present until 2FA is verified
	if resp["user_id"] != nil && resp["user_id"] != "" {
		t.Errorf("expected user_id to be absent when two_fa_pending=true, got %v", resp["user_id"])
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

	var resp map[string]any
	decodeJSON(t, w, &resp)

	// two_fa_pending must be absent or false
	if pending, _ := resp["two_fa_pending"].(bool); pending {
		t.Error("expected two_fa_pending=false for user without 2FA")
	}
	// User data must be present
	if resp["username"] == nil || resp["username"] == "" {
		t.Error("expected username in login response for user without 2FA")
	}

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
	// login (200 with two_fa_pending=true) → verify (204) → protected endpoint works.
	userID := createTestUser(t, "fullflowuser", "fullflow@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))

	// Login returns 200 with two_fa_pending=true and marks session pending
	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	loginW := doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": "fullflowuser", "password": "testpassword123"},
		[]*http.Cookie{anonCookie})
	assertStatus(t, loginW, http.StatusOK)

	var loginResp map[string]any
	decodeJSON(t, loginW, &loginResp)
	if pending, _ := loginResp["two_fa_pending"].(bool); !pending {
		t.Error("expected two_fa_pending=true in login response")
	}

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

// ===== Pending session access control =====

// TestPendingSession_BlocksProtectedEndpoints verifies that a session marked
// two_fa_pending=true is rejected by endpoints guarded by requiresVerifiedSession.
// This is the server-side guarantee that mirrors the client-side two_fa_pending
// flag: even if the client ignores the flag, the backend will not serve
// protected resources until the 2FA challenge is completed.
func TestPendingSession_BlocksProtectedEndpoints(t *testing.T) {
	setupTest(t)

	// Set up a user with verified 2FA, then simulate a fresh login that
	// leaves the session pending.
	userID := createTestUser(t, "pendingblockuser", "pendingblock@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))

	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	loginW := doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": "pendingblockuser", "password": "testpassword123"},
		[]*http.Cookie{anonCookie})
	assertStatus(t, loginW, http.StatusOK)

	pendingCookie := &http.Cookie{Name: "session_token", Value: tok}

	// All endpoints guarded by requiresVerifiedSession must return 401.
	protectedEndpoints := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodGet, "/api/users/me", nil},
		{http.MethodGet, "/api/users/me/security", nil},
		{http.MethodPost, "/api/boards", map[string]any{"title": "test", "size": 5}},
	}

	for _, ep := range protectedEndpoints {
		w := doRequestWithCookies(ep.method, ep.path, ep.body, []*http.Cookie{pendingCookie})
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401 for pending session, got %d (body: %s)",
				ep.method, ep.path, w.Code, w.Body.String())
		}
	}
}

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

// ===== Verify Recovery =====

func TestVerifyRecovery_Success(t *testing.T) {
	setupTest(t)

	// Set up 2FA and capture the recovery codes issued at confirm time.
	userID := createTestUser(t, "recvuser", "recv@example.com")
	secret := setup2FA(t, userID)
	codes := confirm2FAWithCodes(t, userID, secret)

	// Log out so the next login goes through the 2FA pending flow.
	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))

	tok, pendingCookie := loginWith2FA(t, "recvuser")

	// Session must be pending before we use the recovery code.
	session, ok := loadSessionByToken(t, tok)
	if !ok {
		t.Fatal("pending session not found after login")
	}
	if !session.TwoFaPending {
		t.Fatal("expected session to be pending after login with 2FA user")
	}

	// Use the first recovery code to unlock the session.
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": codes[0]}, []*http.Cookie{pendingCookie})
	assertStatus(t, w, http.StatusNoContent)

	// Session must be unlocked.
	session, ok = loadSessionByToken(t, tok)
	if !ok {
		t.Fatal("session not found after verify-recovery")
	}
	if session.TwoFaPending {
		t.Error("expected session two_fa_pending=false after verify-recovery")
	}

	// Protected endpoint must now work.
	me := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{pendingCookie})
	assertStatus(t, me, http.StatusOK)
}

func TestVerifyRecovery_CodeConsumed_CannotReuse(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "recvonce", "recvonce@example.com")
	secret := setup2FA(t, userID)
	codes := confirm2FAWithCodes(t, userID, secret)

	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))

	// First login: use the code to unlock.
	tok1, pendingCookie1 := loginWith2FA(t, "recvonce")
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": codes[0]}, []*http.Cookie{pendingCookie1})
	assertStatus(t, w, http.StatusNoContent)

	// Log out again and log back in to get a second pending session.
	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil,
		[]*http.Cookie{{Name: "session_token", Value: tok1}})

	_, pendingCookie2 := loginWith2FA(t, "recvonce")

	// Reusing the already-consumed code must fail.
	w2 := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": codes[0]}, []*http.Cookie{pendingCookie2})
	assertStatus(t, w2, http.StatusBadRequest)
}

func TestVerifyRecovery_InvalidCode_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "recvinvalid", "recvinvalid@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))
	_, pendingCookie := loginWith2FA(t, "recvinvalid")

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": "DEADBEEF-DEADBEEF"}, []*http.Cookie{pendingCookie})
	assertStatus(t, w, http.StatusBadRequest)
}

func TestVerifyRecovery_NotPending_Returns401(t *testing.T) {
	setupTest(t)

	// A fully-authenticated (non-pending) session cannot call verify-recovery.
	userID := createTestUser(t, "recvnotpending", "recvnotpending@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": "DEADBEEF-DEADBEEF"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestVerifyRecovery_TotpNotVerified_Returns403(t *testing.T) {
	setupTest(t)

	// Setup was called but confirm was never completed — no recovery codes
	// exist and totp_verified_at is NULL. Must return 403.
	userID := createTestUser(t, "recvunverified", "recvunverified@example.com")
	setup2FA(t, userID) // session now pending, but totp not confirmed

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": "DEADBEEF-DEADBEEF"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusForbidden)
}

func TestVerifyRecovery_No2FA_Returns404(t *testing.T) {
	setupTest(t)

	// No 2FA row at all — verify-recovery must return 404.
	userID := createTestUser(t, "recvno2fa", "recvno2fa@example.com")
	setSessionPending(t, userCookies[userID].Value)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": "DEADBEEF-DEADBEEF"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusNotFound)
}

// ===== Regenerate Codes =====

func TestRegenerateCodes_WithTOTP_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "regentotp", "regentotp@example.com")
	secret := setup2FA(t, userID)
	oldCodes := confirm2FAWithCodes(t, userID, secret)

	totpCode := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "testpassword123", "code": totpCode},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)
	raw, ok := resp["recovery_codes"].([]any)
	if !ok || len(raw) == 0 {
		t.Fatalf("expected non-empty recovery_codes in regenerate response, got %v", resp)
	}

	// New codes must differ from the old ones.
	newCode0, _ := raw[0].(string)
	for _, old := range oldCodes {
		if newCode0 == old {
			t.Error("expected new recovery codes to differ from old ones")
		}
	}

	// Old codes must now be invalid — log out, log back in, try an old code.
	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))
	_, pendingCookie := loginWith2FA(t, "regentotp")
	w2 := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/verify-recovery",
		map[string]any{"code": oldCodes[0]}, []*http.Cookie{pendingCookie})
	assertStatus(t, w2, http.StatusBadRequest)
}

func TestRegenerateCodes_WithRecoveryCode_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "regenrecv", "regenrecv@example.com")
	secret := setup2FA(t, userID)
	oldCodes := confirm2FAWithCodes(t, userID, secret)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "testpassword123", "recovery_code": oldCodes[0]},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)
	raw, ok := resp["recovery_codes"].([]any)
	if !ok || len(raw) == 0 {
		t.Fatalf("expected non-empty recovery_codes in regenerate response, got %v", resp)
	}
}

func TestRegenerateCodes_WrongPassword_Returns401(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "regenwrongpw", "regenwrongpw@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	totpCode := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "wrongpassword1", "code": totpCode},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestRegenerateCodes_InvalidTOTP_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "regeninvalidtotp", "regeninvalidtotp@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "testpassword123", "code": "000000"},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestRegenerateCodes_NoAuthMethod_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "regennoauth", "regennoauth@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "testpassword123"},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestRegenerateCodes_BothAuthMethods_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "regenboth", "regenboth@example.com")
	secret := setup2FA(t, userID)
	codes := confirm2FAWithCodes(t, userID, secret)

	totpCode := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "testpassword123", "code": totpCode, "recovery_code": codes[0]},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestRegenerateCodes_Unauthenticated_Returns401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "testpassword123", "code": "123456"})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestRegenerateCodes_No2FA_Returns403(t *testing.T) {
	setupTest(t)

	// User has no 2FA set up — regenerate must return 403.
	userID := createTestUser(t, "regenno2fa", "regenno2fa@example.com")

	w := doRequestWithCookies(http.MethodPost, "/api/auth/2fa/regenerate-codes",
		map[string]any{"password": "testpassword123", "code": "123456"},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusForbidden)
}

// ===== Disable 2FA =====

func TestDisable2FA_WithTOTP_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "disabletotp", "disabletotp@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	totpCode := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "testpassword123", "code": totpCode},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	// 2FA row must be gone.
	uid := userIDFromString(t, userID)
	if _, ok := load2FARow(t, uid); ok {
		t.Error("expected user_two_factors row to be deleted after disable")
	}

	// A subsequent login must NOT require 2FA.
	doRequestWithCookies(http.MethodPost, "/api/auth/logout", nil, cookiesFor(userID))
	tok, anonCookie := mintAnonSession(t, 24*time.Hour)
	loginW := doRequestWithCookies(http.MethodPost, "/api/auth/login",
		map[string]any{"username": "disabletotp", "password": "testpassword123"},
		[]*http.Cookie{anonCookie})
	assertStatus(t, loginW, http.StatusOK)

	var loginResp map[string]any
	decodeJSON(t, loginW, &loginResp)
	if pending, _ := loginResp["two_fa_pending"].(bool); pending {
		t.Error("expected two_fa_pending=false after disabling 2FA")
	}

	session, ok := loadSessionByToken(t, tok)
	if !ok {
		t.Fatal("session not found after login post-disable")
	}
	if session.TwoFaPending {
		t.Error("expected session two_fa_pending=false after disabling 2FA")
	}
}

func TestDisable2FA_WithRecoveryCode_Success(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "disablerecv", "disablerecv@example.com")
	secret := setup2FA(t, userID)
	codes := confirm2FAWithCodes(t, userID, secret)

	w := doRequestWithCookies(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "testpassword123", "recovery_code": codes[0]},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	uid := userIDFromString(t, userID)
	if _, ok := load2FARow(t, uid); ok {
		t.Error("expected user_two_factors row to be deleted after disable with recovery code")
	}
}

func TestDisable2FA_WrongPassword_Returns401(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "disablewrongpw", "disablewrongpw@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	totpCode := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "wrongpassword1", "code": totpCode},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestDisable2FA_InvalidTOTP_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "disableinvalidtotp", "disableinvalidtotp@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	w := doRequestWithCookies(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "testpassword123", "code": "000000"},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestDisable2FA_NoAuthMethod_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "disablenoauth", "disablenoauth@example.com")
	secret := setup2FA(t, userID)
	confirm2FA(t, userID, secret)

	w := doRequestWithCookies(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "testpassword123"},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestDisable2FA_BothAuthMethods_Returns400(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "disableboth", "disableboth@example.com")
	secret := setup2FA(t, userID)
	codes := confirm2FAWithCodes(t, userID, secret)

	totpCode := generateTOTPCode(t, secret)
	w := doRequestWithCookies(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "testpassword123", "code": totpCode, "recovery_code": codes[0]},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

func TestDisable2FA_Unauthenticated_Returns401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "testpassword123", "code": "123456"})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestDisable2FA_No2FA_Returns403(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "disableno2fa", "disableno2fa@example.com")

	w := doRequestWithCookies(http.MethodDelete, "/api/auth/2fa",
		map[string]any{"password": "testpassword123", "code": "123456"},
		cookiesFor(userID))
	assertStatus(t, w, http.StatusForbidden)
}
