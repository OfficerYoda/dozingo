package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
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

	// Email must be stripped from the public-by-id response so other users
	// can't enumerate addresses; see users.go userByID handler.
	if got["email"] != nil {
		t.Errorf("expected email to be null on public GET /users/{id}, got %v", got["email"])
	}

	// Register auto-generated an avatar for this user, so avatar_url is
	// always set on the public-by-id response too.
	if got["avatar_url"] == nil {
		t.Errorf("expected avatar_url to be set on public GET /users/{id} for user without custom avatar, got nil")
	}
}

func TestUserByID_OtherAuthenticatedUser_EmailHidden(t *testing.T) {
	setupTest(t)

	// userA's row gets fetched by userB while userB has a session.
	a := createTestUserWithRegister(t, "byidA", "mypassword123", stringPtr("a@example.com"))
	b := createTestUserWithRegister(t, "byidB", "mypassword123", stringPtr("b@example.com"))
	aID := (*a)["user_id"].(string)
	bID := (*b)["user_id"].(string)

	w := doRequestWithCookies(http.MethodGet, fmt.Sprintf("/api/users/%s", aID), nil, cookiesFor(bID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)
	assertJSONField(t, got, "user_id", aID)
	if got["email"] != nil {
		t.Errorf("expected another user's email to be null, got %v", got["email"])
	}
}

func TestUserByID_Self_EmailVisible(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "selfemail", "mypassword123", stringPtr("self@example.com"))
	userID := (*resp)["user_id"].(string)

	// Same user requesting their own row should see their email, mirroring
	// /api/users/me. The hide-email-from-strangers protection is targeted
	// at *other* users.
	w := doRequestWithCookies(http.MethodGet, fmt.Sprintf("/api/users/%s", userID), nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)
	assertJSONField(t, got, "email", "self@example.com")
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

/// ===== PATCH /users/me =====

// loadUserByID reads a user row directly via the test pool. Used as ground
// truth for assertions that need to peek past the API surface (e.g. that
// email_verified_at really did get cleared).
func loadUserByID(t *testing.T, userID string) generated.User {
	t.Helper()
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		t.Fatalf("invalid user id %q: %v", userID, err)
	}
	q := generated.New(testPool)
	user, err := q.GetUserByID(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to load user %s: %v", userID, err)
	}
	return user
}

// PATCH /api/users/{user_id} was removed for security reasons (only the
// session-scoped /api/users/me variant remains). Confirm the path-version
// route is gone so a future regression doesn't silently bring it back.
func TestUpdateUserByID_RouteRemoved(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "noupdatebyid", "mypassword123", stringPtr("nbi@example.com"))
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodPatch, fmt.Sprintf("/api/users/%s", userID),
		map[string]any{"username": "renamed"}, cookiesFor(userID))

	if w.Code == http.StatusOK {
		t.Fatalf("PATCH /api/users/{user_id} must not be routed; got 200")
	}
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 404 or 405 for removed route, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestUpdateMe_Success_Username(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "merename", "mypassword123", stringPtr("merename@example.com"))
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{"username": "merenamed"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)
	assertJSONField(t, got, "user_id", userID)
	assertJSONField(t, got, "username", "merenamed")
	assertJSONField(t, got, "email", "merename@example.com")

	if fakeMailer.verifyCount() != 0 {
		t.Errorf("expected 0 verification mails for a pure username change, got %d", fakeMailer.verifyCount())
	}
}

func TestUpdateMe_Success_Email_SendsVerificationAndClearsVerifiedAt(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "meemail", "mypassword123", stringPtr("meold@example.com"))
	userID := (*resp)["user_id"].(string)

	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		t.Fatalf("invalid user id: %v", err)
	}
	now := time.Now()
	q := generated.New(testPool)
	if _, err := q.SetUserEmailVerifiedAt(context.Background(), generated.SetUserEmailVerifiedAtParams{
		UserID:          id,
		EmailVerifiedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}); err != nil {
		t.Fatalf("seed email_verified_at: %v", err)
	}

	w := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{"email": "menew@example.com"}, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)
	assertJSONField(t, got, "email", "menew@example.com")

	user := loadUserByID(t, userID)
	if !user.Email.Valid || user.Email.String != "menew@example.com" {
		t.Errorf("expected stored email 'menew@example.com', got Valid=%v String=%q", user.Email.Valid, user.Email.String)
	}
	if user.EmailVerifiedAt.Valid {
		t.Errorf("expected email_verified_at to be NULL after email change, got %v", user.EmailVerifiedAt.Time)
	}

	if fakeMailer.verifyCount() != 1 {
		t.Fatalf("expected exactly 1 verification mail, got %d", fakeMailer.verifyCount())
	}
	last, _ := fakeMailer.lastVerify()
	if last.To != "menew@example.com" {
		t.Errorf("expected mail to 'menew@example.com', got %q", last.To)
	}
}

func TestUpdateMe_ClearEmail(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "meclear", "mypassword123", stringPtr("meclear@example.com"))
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{"email": ""}, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)
	if got["email"] != nil {
		t.Errorf("expected email to be null in response, got %v", got["email"])
	}

	user := loadUserByID(t, userID)
	if user.Email.Valid {
		t.Errorf("expected stored email to be NULL after clear, got %q", user.Email.String)
	}

	if fakeMailer.verifyCount() != 0 {
		t.Errorf("expected 0 verification mails for an email clear, got %d", fakeMailer.verifyCount())
	}
}

func TestUpdateMe_NoOpEmptyBody(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "menoop", "mypassword123", stringPtr("menoop@example.com"))
	userID := (*resp)["user_id"].(string)
	preEmail := (*resp)["email"]
	preUsername := (*resp)["username"]

	w := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{}, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var got map[string]any
	decodeJSON(t, w, &got)
	if got["username"] != preUsername {
		t.Errorf("expected username unchanged %v, got %v", preUsername, got["username"])
	}
	if got["email"] != preEmail {
		t.Errorf("expected email unchanged %v, got %v", preEmail, got["email"])
	}

	if fakeMailer.verifyCount() != 0 {
		t.Errorf("expected 0 verification mails for a no-op patch, got %d", fakeMailer.verifyCount())
	}
}

func TestUpdateMe_NotLoggedIn_401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodPatch, "/api/users/me", map[string]any{"username": "x"})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUpdateMe_AnonymousSessionOnly_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{"username": "x"}, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUpdateMe_DuplicateUsername_409(t *testing.T) {
	setupTest(t)

	createTestUserWithRegister(t, "metaken", "mypassword123", stringPtr("metaken@example.com"))
	b := createTestUserWithRegister(t, "metryingto", "mypassword123", stringPtr("metryingto@example.com"))
	bID := (*b)["user_id"].(string)

	w := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{"username": "metaken"}, cookiesFor(bID))
	assertStatus(t, w, http.StatusConflict)

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

	var got map[string]any
	decodeJSON(t, w, &got)
	assertJSONField(t, got, "detail", "conflict")

	user := loadUserByID(t, bID)
	if user.Username != "metryingto" {
		t.Errorf("expected username unchanged 'metryingto', got %q", user.Username)
	}
}

func TestUpdateMe_InvalidEmailFormat_422(t *testing.T) {
	setupTest(t)

	resp := createTestUserWithRegister(t, "mebademail", "mypassword123", nil)
	userID := (*resp)["user_id"].(string)

	w := doRequestWithCookies(http.MethodPatch, "/api/users/me",
		map[string]any{"email": "not-an-email"}, cookiesFor(userID))
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for malformed email, got %d (body: %s)", w.Code, w.Body.String())
	}
}

/// ===== GET /users/{user_id}/votes =====

// decodeJSONArray decodes a JSON array response body into a slice of maps.
func decodeJSONArray(t *testing.T, w *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	var out []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode JSON array: %v\nbody: %s", err, w.Body.String())
	}
	return out
}

func TestListVotesFromUser_Success_SingleVote(t *testing.T) {
	setupTest(t)

	authorID := createTestUser(t, "boardauthor", "ba@example.com")
	desc := "a description"
	boardID := createTestBoard(t, "Title One", 5, authorID, &desc)

	voterID := createTestUser(t, "singlevoter", "sv@example.com")
	voteResp := createTestVote(t, boardID, voterID, 1)
	voteID := voteResp["vote_id"].(string)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s/votes", voterID), nil)
	assertStatus(t, w, http.StatusOK)

	got := decodeJSONArray(t, w)
	if len(got) != 1 {
		t.Fatalf("expected 1 vote, got %d: %v", len(got), got)
	}
	row := got[0]
	assertJSONField(t, row, "vote_id", voteID)
	assertJSONField(t, row, "board_id", boardID)
	assertJSONField(t, row, "title", "Title One")
	assertJSONField(t, row, "description", "a description")
	assertJSONField(t, row, "board_author_id", authorID)
	assertJSONInt(t, row, "vote_value", 1)
	assertJSONInt(t, row, "size", 5)
	assertJSONInt(t, row, "vote_score", 1)
	assertJSONInt(t, row, "vote_count", 1)
	assertJSONInt(t, row, "play_count", 0)
}

func TestListVotesFromUser_Success_MultipleBoardsAndAggregates(t *testing.T) {
	setupTest(t)

	authorID := createTestUser(t, "multiauthor", "ma@example.com")
	board1 := createTestBoard(t, "Board One", 5, authorID, nil)
	board2 := createTestBoard(t, "Board Two", 4, authorID, nil)

	userA := createTestUser(t, "userA_multi", "a_multi@example.com")
	userB := createTestUser(t, "userB_multi", "b_multi@example.com")

	// userA votes on both boards.
	createTestVote(t, board1, userA, 1)
	createTestVote(t, board2, userA, -1)
	// userB also votes on board1 (so aggregates on board1 should include both).
	createTestVote(t, board1, userB, 1)

	// Create a game on board1 to bump play_count to 1.
	createTestGame(t, userA, board1)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s/votes", userA), nil)
	assertStatus(t, w, http.StatusOK)

	got := decodeJSONArray(t, w)
	if len(got) != 2 {
		t.Fatalf("expected 2 votes for userA, got %d: %v", len(got), got)
	}

	byBoard := map[string]map[string]any{}
	for _, row := range got {
		byBoard[row["board_id"].(string)] = row
	}

	row1, ok := byBoard[board1]
	if !ok {
		t.Fatalf("expected a row for board1=%s, got %v", board1, got)
	}
	// board1 aggregates: 1 (userA) + 1 (userB) = score 2, vote_count 2, play_count 1.
	assertJSONInt(t, row1, "vote_value", 1)
	assertJSONInt(t, row1, "vote_score", 2)
	assertJSONInt(t, row1, "vote_count", 2)
	assertJSONInt(t, row1, "play_count", 1)

	row2, ok := byBoard[board2]
	if !ok {
		t.Fatalf("expected a row for board2=%s, got %v", board2, got)
	}
	// board2 aggregates: only userA's -1 vote, no games.
	assertJSONInt(t, row2, "vote_value", -1)
	assertJSONInt(t, row2, "vote_score", -1)
	assertJSONInt(t, row2, "vote_count", 1)
	assertJSONInt(t, row2, "play_count", 0)
}

func TestListVotesFromUser_NotFoundUser_EmptyList(t *testing.T) {
	setupTest(t)

	missing := uuid.NewString()
	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s/votes", missing), nil)
	assertStatus(t, w, http.StatusOK)

	got := decodeJSONArray(t, w)
	if len(got) != 0 {
		t.Errorf("expected empty list for unknown user, got %d entries: %v", len(got), got)
	}
}

func TestListVotesFromUser_InvalidUUID_422(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/users/not-a-uuid/votes", nil)
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

/// ===== GET /users/me/votes =====

func TestListVotesFromMe_NoCookie_401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/users/me/votes", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestListVotesFromMe_AnonymousSessionOnly_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me/votes", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestListVotesFromMe_Success_Empty(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "menovotes", "menovotes@example.com")

	w := doRequestWithCookies(http.MethodGet, "/api/users/me/votes", nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	got := decodeJSONArray(t, w)
	if len(got) != 0 {
		t.Errorf("expected 0 votes, got %d: %v", len(got), got)
	}
}

func TestListVotesFromMe_Success_OnlyOwnVotes(t *testing.T) {
	setupTest(t)

	authorID := createTestUser(t, "meauthor", "meauth@example.com")
	boardID := createTestBoard(t, "Shared Board", 5, authorID, nil)

	userA := createTestUser(t, "meownA", "meA@example.com")
	userB := createTestUser(t, "meownB", "meB@example.com")

	voteA := createTestVote(t, boardID, userA, 1)
	createTestVote(t, boardID, userB, -1)
	voteAID := voteA["vote_id"].(string)

	w := doRequestWithCookies(http.MethodGet, "/api/users/me/votes", nil, cookiesFor(userA))
	assertStatus(t, w, http.StatusOK)

	got := decodeJSONArray(t, w)
	if len(got) != 1 {
		t.Fatalf("expected 1 vote (userA's only), got %d: %v", len(got), got)
	}
	assertJSONField(t, got[0], "vote_id", voteAID)
	assertJSONField(t, got[0], "board_id", boardID)
	// Aggregates over the board still see both votes.
	assertJSONInt(t, got[0], "vote_count", 2)
}
