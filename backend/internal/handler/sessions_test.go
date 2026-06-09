package handler

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
)

// insertSessionWithExpiry inserts a session row directly with the given
// expires_at offset relative to now. UserID is NULL (anonymous) -- that's all
// the middleware tests need.
//
// Returns the plaintext token (the cookie value). The DB row stores the
// SHA-256 hex digest so the middleware can find it.
func insertSessionWithExpiry(t *testing.T, expiresIn time.Duration) (token string) {
	t.Helper()
	q := generated.New(testPool)
	tok := auth.GenerateToken()
	_, err := q.CreateSession(context.Background(), generated.CreateSessionParams{
		UserID: pgtype.UUID{Valid: false},
		Token:  auth.HashToken(tok),
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(expiresIn),
			Valid: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}
	return tok
}

// fetchSessionExpiry returns the expires_at for a session token (plaintext,
// the cookie value), or false if the row is gone (or already expired --
// GetSessionUserByToken filters those).
func fetchSessionExpiry(t *testing.T, token string) (time.Time, bool) {
	t.Helper()
	q := generated.New(testPool)
	row, err := q.GetSessionUserByToken(context.Background(), auth.HashToken(token))
	if err != nil {
		return time.Time{}, false
	}
	return row.ExpiresAt.Time, true
}

/// ===== Cookie minting & clearing =====

func TestSession_NoCookieOnAnonRequest(t *testing.T) {
	setupTest(t)

	// /api/health is registered directly on the bare api (not apiGroup), so
	// it bypasses session middleware entirely and must never set a cookie.
	w := doRequest(http.MethodGet, "/api/health", nil)
	if c := extractSessionCookie(w); c != nil {
		t.Errorf("expected no session cookie on anon request, got %+v", c)
	}
}

func TestSession_InvalidCookie_ClearedAndRequestStillWorks(t *testing.T) {
	setupTest(t)

	// Use an endpoint that lives on apiGroup (where the session middleware
	// runs) but does not itself require auth. /api/health is registered
	// directly on the bare api to bypass middleware, so it would not
	// exercise the cookie-clearing path.
	junk := &http.Cookie{Name: "session_token", Value: "this-token-does-not-exist"}
	w := doRequestWithCookies(http.MethodGet, "/api/boards", nil, []*http.Cookie{junk})
	assertStatus(t, w, http.StatusOK)

	c := extractSessionCookie(w)
	if c == nil {
		t.Fatal("middleware should clear an unknown session_token cookie")
	}
	if c.MaxAge >= 0 {
		t.Errorf("expected MaxAge<0 (cookie cleared), got %d", c.MaxAge)
	}
}

func TestSession_ExpiredCookie_TreatedAsUnknown(t *testing.T) {
	setupTest(t)

	// Insert a row that's already expired. GetSessionUserByToken filters by
	// expires_at > now(), so the middleware sees pgx.ErrNoRows and clears
	// the cookie.
	tok := insertSessionWithExpiry(t, -1*time.Hour)
	cookie := &http.Cookie{Name: "session_token", Value: tok}

	w := doRequestWithCookies(http.MethodGet, "/api/users/me", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)

	c := extractSessionCookie(w)
	if c == nil {
		t.Fatal("expected expired cookie to be cleared")
	}
	if c.MaxAge >= 0 {
		t.Errorf("expected MaxAge<0 (cookie cleared), got %d", c.MaxAge)
	}
}

/// ===== Session extension =====

func TestSession_NearExpiry_GetsExtended(t *testing.T) {
	setupTest(t)

	// 1 day < 7-day extension threshold => should be bumped on use.
	tok := insertSessionWithExpiry(t, 1*24*time.Hour)
	originalExpiry, ok := fetchSessionExpiry(t, tok)
	if !ok {
		t.Fatal("test setup: session not visible")
	}

	cookie := &http.Cookie{Name: "session_token", Value: tok}
	// /api/boards lives on apiGroup so it goes through session middleware
	// (which is what we're testing here). /api/health bypasses it.
	w := doRequestWithCookies(http.MethodGet, "/api/boards", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	newExpiry, ok := fetchSessionExpiry(t, tok)
	if !ok {
		t.Fatal("session disappeared after request")
	}
	if !newExpiry.After(originalExpiry) {
		t.Errorf("expected expiry to be extended past %v, got %v", originalExpiry, newExpiry)
	}
	// Should be roughly now() + 30d.
	delta := time.Until(newExpiry)
	if delta < 25*24*time.Hour || delta > 31*24*time.Hour {
		t.Errorf("extended expiry not close to 30d from now (delta=%v)", delta)
	}
}

func TestSession_FreshSession_NotExtended(t *testing.T) {
	setupTest(t)

	// 20 days from now is well outside the 7-day extension threshold.
	tok := insertSessionWithExpiry(t, 20*24*time.Hour)
	originalExpiry, ok := fetchSessionExpiry(t, tok)
	if !ok {
		t.Fatal("test setup: session not visible")
	}

	cookie := &http.Cookie{Name: "session_token", Value: tok}
	w := doRequestWithCookies(http.MethodGet, "/api/boards", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	newExpiry, ok := fetchSessionExpiry(t, tok)
	if !ok {
		t.Fatal("session disappeared after request")
	}
	if !newExpiry.Equal(originalExpiry) {
		t.Errorf("fresh session should not be extended; original=%v new=%v", originalExpiry, newExpiry)
	}
}
