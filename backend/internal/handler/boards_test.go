package handler

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/generated"
)

func TestCreateBoard(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "boardauthor", "boardauthor@example.com")

	body := map[string]any{
		"title": "Test Board",
		"size":  5,
	}

	w := doRequestWithCookies(http.MethodPost, "/api/boards", body, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "title", "Test Board")
	assertJSONField(t, resp, "author_id", userID)

	if _, ok := resp["board_id"]; !ok {
		t.Error("expected 'id' field in response")
	}

	// Check size (JSON numbers decode as float64)
	if size, ok := resp["size"].(float64); !ok || int(size) != 5 {
		t.Errorf("expected size = 5, got %v", resp["size"])
	}

	// Freshly created boards have no votes and no games yet.
	assertJSONInt(t, resp, "score", 0)
	assertJSONInt(t, resp, "vote_count", 0)
	assertJSONInt(t, resp, "play_count", 0)
}

func TestCreateBoard_WithDescription(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "descauthor", "descauthor@example.com")

	body := map[string]any{
		"title":       "Described Board",
		"description": "A board with a description",
		"size":        4,
	}

	w := doRequestWithCookies(http.MethodPost, "/api/boards", body, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "title", "Described Board")
	assertJSONField(t, resp, "description", "A board with a description")
	assertJSONField(t, resp, "author_id", userID)
}

func TestCreateBoard_WithoutDescription(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "nodescauthor", "nodescauthor@example.com")

	body := map[string]any{
		"title": "No Desc Board",
		"size":  5,
	}

	w := doRequestWithCookies(http.MethodPost, "/api/boards", body, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "title", "No Desc Board")

	// Description should be empty string when not provided
	if desc, ok := resp["description"].(string); ok && desc != "" {
		t.Errorf("expected empty description, got %q", desc)
	}
}

func TestCreateBoard_Unauthenticated(t *testing.T) {
	setupTest(t)

	body := map[string]any{
		"title": "Anon Board",
		"size":  5,
	}

	w := doRequest(http.MethodPost, "/api/boards", body)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetBoards(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "listauthor", "listauthor@example.com")

	createTestBoard(t, "Board A", 5, userID, nil)
	createTestBoard(t, "Board B", 3, userID, nil)

	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 boards, got %d", len(resp))
	}
}

func TestGetBoards_Empty(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 boards, got %d", len(resp))
	}
}

func TestGetBoards_FilterByAuthor(t *testing.T) {
	setupTest(t)

	user1 := createTestUser(t, "author1", "author1@example.com")
	user2 := createTestUser(t, "author2", "author2@example.com")

	createTestBoard(t, "User1 Board", 5, user1, nil)
	createTestBoard(t, "User2 Board", 5, user2, nil)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards?author_id=%s", user1), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 board, got %d", len(resp))
		return
	}
	assertJSONField(t, resp[0], "title", "User1 Board")
}

func TestGetBoards_FilterBySize(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "sizeauthor", "sizeauthor@example.com")

	createTestBoard(t, "Small Board", 3, userID, nil)
	createTestBoard(t, "Large Board", 7, userID, nil)

	w := doRequest(http.MethodGet, "/api/boards?size=3", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 board, got %d", len(resp))
		return
	}
	assertJSONField(t, resp[0], "title", "Small Board")
}

func TestGetBoardByID(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "getboardauthor", "getboard@example.com")
	boardID := createTestBoard(t, "GetMe Board", 5, userID, stringPtr("Test description"))

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "title", "GetMe Board")
	assertJSONField(t, resp, "description", "Test description")
	assertJSONField(t, resp, "author_id", userID)

	// No votes, no games on this board.
	assertJSONInt(t, resp, "score", 0)
	assertJSONInt(t, resp, "vote_count", 0)
	assertJSONInt(t, resp, "play_count", 0)
}

func TestGetBoardByID_NotFound(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/boards/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestDeleteBoard(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "delboardauthor", "delboard@example.com")
	boardID := createTestBoard(t, "DeleteMe Board", 5, userID, nil)

	// Delete the board
	w := doRequestWithCookies(http.MethodDelete, fmt.Sprintf("/api/boards/%s", boardID), nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	// Verify board is gone
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s", boardID), nil)
	assertStatus(t, getResp, http.StatusNotFound)
}

func TestDeleteBoard_NotFound(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "delnoboard", "delnoboard@example.com")

	w := doRequestWithCookies(
		http.MethodDelete,
		"/api/boards/00000000-0000-0000-0000-000000000000",
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestDeleteBoard_NotFound_Unauthenticated(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodDelete, "/api/boards/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetBoards_CombinedFilters(t *testing.T) {
	setupTest(t)

	user1 := createTestUser(t, "comboauthor1", "combo1@example.com")
	user2 := createTestUser(t, "comboauthor2", "combo2@example.com")

	createTestBoard(t, "Match", 5, user1, nil)
	createTestBoard(t, "Wrong Author", 5, user2, nil)
	createTestBoard(t, "Wrong Size", 3, user1, nil)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards?author_id=%s&size=5", user1), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 board, got %d", len(resp))
		return
	}
	assertJSONField(t, resp[0], "title", "Match")
}

/// ===== GET /boards sort & limit =====

// titlesOf extracts the "title" field of every board in a list response, in
// order. Useful for asserting sort ordering.
func titlesOf(t *testing.T, resp []map[string]any) []string {
	t.Helper()
	out := make([]string, 0, len(resp))
	for i, b := range resp {
		title, ok := b["title"].(string)
		if !ok {
			t.Fatalf("board[%d]: expected string title, got %T", i, b["title"])
		}
		out = append(out, title)
	}
	return out
}

func TestGetBoards_SortNewest(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "sortnew", "sortnew@example.com")

	// Created in order: First, Second, Third. Newest first ⇒ Third, Second, First.
	createTestBoard(t, "First", 5, userID, nil)
	time.Sleep(2 * time.Millisecond)
	createTestBoard(t, "Second", 5, userID, nil)
	time.Sleep(2 * time.Millisecond)
	createTestBoard(t, "Third", 5, userID, nil)

	w := doRequest(http.MethodGet, "/api/boards?sort=newest", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	got := titlesOf(t, resp)
	want := []string{"Third", "Second", "First"}
	if !equalStrSlice(got, want) {
		t.Errorf("sort=newest: expected %v, got %v", want, got)
	}
}

func TestGetBoards_SortOldest(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "sortold", "sortold@example.com")

	createTestBoard(t, "First", 5, userID, nil)
	time.Sleep(2 * time.Millisecond)
	createTestBoard(t, "Second", 5, userID, nil)
	time.Sleep(2 * time.Millisecond)
	createTestBoard(t, "Third", 5, userID, nil)

	w := doRequest(http.MethodGet, "/api/boards?sort=oldest", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	got := titlesOf(t, resp)
	want := []string{"First", "Second", "Third"}
	if !equalStrSlice(got, want) {
		t.Errorf("sort=oldest: expected %v, got %v", want, got)
	}
}

func TestGetBoards_SortMostLiked(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "mlauthor", "mlauthor@example.com")
	v1 := createTestUser(t, "mlvoter1", "mlvoter1@example.com")
	v2 := createTestUser(t, "mlvoter2", "mlvoter2@example.com")
	v3 := createTestUser(t, "mlvoter3", "mlvoter3@example.com")

	low := createTestBoard(t, "Low", 5, author, nil)
	mid := createTestBoard(t, "Mid", 5, author, nil)
	high := createTestBoard(t, "High", 5, author, nil)

	// high: +2 (two upvotes)
	createTestVote(t, high, v1, 1)
	createTestVote(t, high, v2, 1)
	// mid:  +1
	createTestVote(t, mid, v1, 1)
	// low:  -1
	createTestVote(t, low, v3, -1)

	w := doRequest(http.MethodGet, "/api/boards?sort=most-liked", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	got := titlesOf(t, resp)
	want := []string{"High", "Mid", "Low"}
	if !equalStrSlice(got, want) {
		t.Errorf("sort=most-liked: expected %v, got %v", want, got)
	}
}

func TestGetBoards_SortLeastLiked(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "llauthor", "llauthor@example.com")
	v1 := createTestUser(t, "llvoter1", "llvoter1@example.com")
	v2 := createTestUser(t, "llvoter2", "llvoter2@example.com")
	v3 := createTestUser(t, "llvoter3", "llvoter3@example.com")

	low := createTestBoard(t, "Low", 5, author, nil)
	mid := createTestBoard(t, "Mid", 5, author, nil)
	high := createTestBoard(t, "High", 5, author, nil)

	createTestVote(t, high, v1, 1)
	createTestVote(t, high, v2, 1)
	createTestVote(t, mid, v1, 1)
	createTestVote(t, low, v3, -1)

	w := doRequest(http.MethodGet, "/api/boards?sort=least-liked", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	got := titlesOf(t, resp)
	want := []string{"Low", "Mid", "High"}
	if !equalStrSlice(got, want) {
		t.Errorf("sort=least-liked: expected %v, got %v", want, got)
	}
}

func TestGetBoards_SortMostPlayed(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "mpauthor", "mpauthor@example.com")

	one := createTestBoard(t, "One", 5, author, nil)
	two := createTestBoard(t, "Two", 5, author, nil)
	three := createTestBoard(t, "Three", 5, author, nil)

	// three: 3 games, two: 2 games, one: 1 game.
	createTestGame(t, author, one)
	createTestGame(t, author, two)
	createTestGame(t, author, two)
	createTestGame(t, author, three)
	createTestGame(t, author, three)
	createTestGame(t, author, three)

	w := doRequest(http.MethodGet, "/api/boards?sort=most-played", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	got := titlesOf(t, resp)
	want := []string{"Three", "Two", "One"}
	if !equalStrSlice(got, want) {
		t.Errorf("sort=most-played: expected %v, got %v", want, got)
	}
}

func TestGetBoards_SortLeastPlayed(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "lpauthor", "lpauthor@example.com")

	one := createTestBoard(t, "One", 5, author, nil)
	two := createTestBoard(t, "Two", 5, author, nil)
	three := createTestBoard(t, "Three", 5, author, nil)

	createTestGame(t, author, one)
	createTestGame(t, author, two)
	createTestGame(t, author, two)
	createTestGame(t, author, three)
	createTestGame(t, author, three)
	createTestGame(t, author, three)

	w := doRequest(http.MethodGet, "/api/boards?sort=least-played", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	got := titlesOf(t, resp)
	want := []string{"One", "Two", "Three"}
	if !equalStrSlice(got, want) {
		t.Errorf("sort=least-played: expected %v, got %v", want, got)
	}
}

func TestGetBoards_Limit(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "limitauthor", "limit@example.com")
	for i := 0; i < 5; i++ {
		createTestBoard(t, fmt.Sprintf("Board %d", i), 5, userID, nil)
	}

	w := doRequest(http.MethodGet, "/api/boards?limit=2", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 boards (limit=2), got %d", len(resp))
	}
}

func TestGetBoards_LimitWithMostLiked(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "lmlauthor", "lml@example.com")
	v1 := createTestUser(t, "lmlvoter1", "lmlv1@example.com")
	v2 := createTestUser(t, "lmlvoter2", "lmlv2@example.com")

	a := createTestBoard(t, "A", 5, author, nil)
	b := createTestBoard(t, "B", 5, author, nil)
	c := createTestBoard(t, "C", 5, author, nil)

	createTestVote(t, a, v1, 1)
	createTestVote(t, a, v2, 1)
	createTestVote(t, b, v1, 1)
	_ = c // C has 0 votes

	w := doRequest(http.MethodGet, "/api/boards?sort=most-liked&limit=2", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(resp))
	}
	got := titlesOf(t, resp)
	want := []string{"A", "B"}
	if !equalStrSlice(got, want) {
		t.Errorf("sort=most-liked&limit=2: expected %v, got %v", want, got)
	}
}

func TestGetBoards_DefaultSortNewestApplied(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "defsort", "defsort@example.com")

	createTestBoard(t, "Older", 5, userID, nil)
	time.Sleep(2 * time.Millisecond)
	createTestBoard(t, "Newer", 5, userID, nil)

	// No sort param ⇒ Huma default "newest".
	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	got := titlesOf(t, resp)
	want := []string{"Newer", "Older"}
	if !equalStrSlice(got, want) {
		t.Errorf("default sort: expected %v, got %v", want, got)
	}
}

func TestGetBoards_InvalidSortRejected(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/boards?sort=bogus", nil)
	// Huma enforces the enum; expect a 422 validation error.
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

// equalStrSlice reports whether two string slices are element-wise equal.
func equalStrSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

/// ===== GET /boards stats fields (score, vote_count, play_count) =====

// findByID returns the first map in resp whose "board_id" field equals id.
// Fails the test if no match is found.
func findByID(t *testing.T, resp []map[string]any, id string) map[string]any {
	t.Helper()
	for _, b := range resp {
		if got, _ := b["board_id"].(string); got == id {
			return b
		}
	}
	t.Fatalf("board %s not found in response of %d boards", id, len(resp))
	return nil
}

func TestGetBoards_IncludesStats(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "statsauthor", "statsauthor@example.com")
	v1 := createTestUser(t, "statsv1", "statsv1@example.com")
	v2 := createTestUser(t, "statsv2", "statsv2@example.com")
	boardID := createTestBoard(t, "Stats Board", 5, author, nil)

	// One upvote (+1) and one downvote (-1) ⇒ score=0, vote_count=2.
	createTestVote(t, boardID, v1, 1)
	createTestVote(t, boardID, v2, -1)
	// Two games played.
	createTestGame(t, author, boardID)
	createTestGame(t, author, boardID)

	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	board := findByID(t, resp, boardID)
	assertJSONInt(t, board, "score", 0)
	assertJSONInt(t, board, "vote_count", 2)
	assertJSONInt(t, board, "play_count", 2)
}

func TestGetBoards_ZeroStatsForFreshBoard(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "freshstats", "freshstats@example.com")
	boardID := createTestBoard(t, "Fresh Board", 5, userID, nil)

	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	board := findByID(t, resp, boardID)
	assertJSONInt(t, board, "score", 0)
	assertJSONInt(t, board, "vote_count", 0)
	assertJSONInt(t, board, "play_count", 0)
}

func TestGetBoards_StatsAcrossMultipleBoards(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "multistats", "multistats@example.com")
	v1 := createTestUser(t, "msv1", "msv1@example.com")
	v2 := createTestUser(t, "msv2", "msv2@example.com")

	a := createTestBoard(t, "A", 5, author, nil)
	b := createTestBoard(t, "B", 5, author, nil)
	c := createTestBoard(t, "C", 5, author, nil)

	// A: 2 upvotes, 1 game.
	createTestVote(t, a, v1, 1)
	createTestVote(t, a, v2, 1)
	createTestGame(t, author, a)

	// B: 0 votes, 3 games.
	createTestGame(t, author, b)
	createTestGame(t, author, b)
	createTestGame(t, author, b)

	// C: 1 downvote, 0 games.
	createTestVote(t, c, v1, -1)

	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	bA := findByID(t, resp, a)
	assertJSONInt(t, bA, "score", 2)
	assertJSONInt(t, bA, "vote_count", 2)
	assertJSONInt(t, bA, "play_count", 1)

	bB := findByID(t, resp, b)
	assertJSONInt(t, bB, "score", 0)
	assertJSONInt(t, bB, "vote_count", 0)
	assertJSONInt(t, bB, "play_count", 3)

	bC := findByID(t, resp, c)
	assertJSONInt(t, bC, "score", -1)
	assertJSONInt(t, bC, "vote_count", 1)
	assertJSONInt(t, bC, "play_count", 0)
}

func TestGetBoardByID_IncludesStats(t *testing.T) {
	setupTest(t)

	author := createTestUser(t, "getstats", "getstats@example.com")
	v1 := createTestUser(t, "getstatsv1", "getstatsv1@example.com")
	v2 := createTestUser(t, "getstatsv2", "getstatsv2@example.com")
	boardID := createTestBoard(t, "Get Stats", 5, author, nil)

	createTestVote(t, boardID, v1, 1)
	createTestVote(t, boardID, v2, 1)
	createTestGame(t, author, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONInt(t, resp, "score", 2)
	assertJSONInt(t, resp, "vote_count", 2)
	assertJSONInt(t, resp, "play_count", 1)
}

func TestCreateBoard_ReturnsZeroStats(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "createstats", "createstats@example.com")

	body := map[string]any{
		"title": "Fresh",
		"size":  5,
	}
	w := doRequestWithCookies(http.MethodPost, "/api/boards", body, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONInt(t, resp, "score", 0)
	assertJSONInt(t, resp, "vote_count", 0)
	assertJSONInt(t, resp, "play_count", 0)
}

/// ===== GET /boards/{board_id}/total-played-games =====

// mustParseUUID parses a UUID string into pgtype.UUID, failing the test on
// invalid input. Used by tests that need to interact with the generated
// queries layer directly (e.g. seeding anon games).
func mustParseUUID(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		t.Fatalf("invalid uuid %q: %v", s, err)
	}
	return u
}

// totalGamesFor decodes the response body of GET total-played-games and
// returns the total_games count. Fails the test if the field is missing or
// not a number.
func totalGamesFor(t *testing.T, resp map[string]any) int {
	t.Helper()
	v, ok := resp["total_games"]
	if !ok {
		t.Fatalf("expected 'total_games' field in response, body: %#v", resp)
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("expected 'total_games' to be a number, got %T (%v)", v, v)
	}
	return int(f)
}

func TestGetTotalGamesPlayed_NoGames(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "tgp_nogames", "tgp_nogames@example.com")
	boardID := createTestBoard(t, "Empty Board", 5, userID, nil)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "board_title", "Empty Board")
	if got := totalGamesFor(t, resp); got != 0 {
		t.Errorf("expected total_games = 0, got %d", got)
	}
}

func TestGetTotalGamesPlayed_SingleGame(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "tgp_single", "tgp_single@example.com")
	boardID := createTestBoard(t, "Single Game Board", 5, userID, nil)
	createTestGame(t, userID, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "board_id", boardID)
	if got := totalGamesFor(t, resp); got != 1 {
		t.Errorf("expected total_games = 1, got %d", got)
	}
}

func TestGetTotalGamesPlayed_MultipleGames(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "tgp_multi", "tgp_multi@example.com")
	boardID := createTestBoard(t, "Multi Game Board", 5, userID, nil)
	createTestGame(t, userID, boardID)
	createTestGame(t, userID, boardID)
	createTestGame(t, userID, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	if got := totalGamesFor(t, resp); got != 3 {
		t.Errorf("expected total_games = 3, got %d", got)
	}
}

func TestGetTotalGamesPlayed_MultiplePlayers(t *testing.T) {
	setupTest(t)

	user1 := createTestUser(t, "tgp_player1", "tgp_player1@example.com")
	user2 := createTestUser(t, "tgp_player2", "tgp_player2@example.com")
	user3 := createTestUser(t, "tgp_player3", "tgp_player3@example.com")
	boardID := createTestBoard(t, "Shared Board", 5, user1, nil)

	createTestGame(t, user1, boardID)
	createTestGame(t, user1, boardID)
	createTestGame(t, user2, boardID)
	createTestGame(t, user3, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	// Counts every game, not unique players.
	if got := totalGamesFor(t, resp); got != 4 {
		t.Errorf("expected total_games = 4, got %d", got)
	}
}

func TestGetTotalGamesPlayed_IncludesAnonymousSessionGames(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "tgp_anonmix", "tgp_anonmix@example.com")
	boardID := createTestBoard(t, "Anon Mix Board", 5, userID, nil)

	// One regular player-bound game (via the public API).
	createTestGame(t, userID, boardID)

	// One anonymous (session-only) game inserted directly via the
	// generated queries layer, mirroring the seed data's approach. The
	// public handler API doesn't yet expose a way to create session-only
	// games, but the database (and the new total-played-games SQL) treats
	// them as first-class participants.
	ctx := context.Background()
	q := generated.New(testPool)
	sess, err := q.CreateSession(ctx, generated.CreateSessionParams{
		UserID: pgtype.UUID{Valid: false},
		Token:  auth.HashToken(auth.GenerateToken()),
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(30 * 24 * time.Hour),
			Valid: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to mint anon session for game: %v", err)
	}
	if _, err := q.CreateGame(ctx, generated.CreateGameParams{
		PlayerID:  pgtype.UUID{Valid: false},
		SessionID: sess.ID,
		BoardID:   mustParseUUID(t, boardID),
	}); err != nil {
		t.Fatalf("failed to insert anonymous game: %v", err)
	}

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	if got := totalGamesFor(t, resp); got != 2 {
		t.Errorf("expected total_games = 2 (1 player-bound + 1 anon session), got %d", got)
	}
}

func TestGetTotalGamesPlayed_OnlyCountsThisBoard(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "tgp_isolation", "tgp_isolation@example.com")
	boardA := createTestBoard(t, "Board A", 5, userID, nil)
	boardB := createTestBoard(t, "Board B", 5, userID, nil)

	// 2 games on A, 5 games on B
	createTestGame(t, userID, boardA)
	createTestGame(t, userID, boardA)
	for i := 0; i < 5; i++ {
		createTestGame(t, userID, boardB)
	}

	wA := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardA), nil)
	assertStatus(t, wA, http.StatusOK)
	var respA map[string]any
	decodeJSON(t, wA, &respA)
	if got := totalGamesFor(t, respA); got != 2 {
		t.Errorf("board A: expected total_games = 2, got %d", got)
	}

	wB := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardB), nil)
	assertStatus(t, wB, http.StatusOK)
	var respB map[string]any
	decodeJSON(t, wB, &respB)
	if got := totalGamesFor(t, respB); got != 5 {
		t.Errorf("board B: expected total_games = 5, got %d", got)
	}
}

func TestGetTotalGamesPlayed_NotFound(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet,
		"/api/boards/00000000-0000-0000-0000-000000000000/total-played-games", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetTotalGamesPlayed_InvalidBoardID(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/boards/not-a-uuid/total-played-games", nil)
	// huma validates the uuid format on path params and returns 422.
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestGetTotalGamesPlayed_AfterGameDeletion(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "tgp_delete", "tgp_delete@example.com")
	boardID := createTestBoard(t, "Delete Board", 5, userID, nil)

	createTestGame(t, userID, boardID)
	gameToDelete := createTestGame(t, userID, boardID)
	createTestGame(t, userID, boardID)

	// Sanity check: 3 games before deletion.
	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	if got := totalGamesFor(t, resp); got != 3 {
		t.Fatalf("pre-delete: expected total_games = 3, got %d", got)
	}

	// Delete one game.
	delResp := doRequestWithCookies(http.MethodDelete,
		fmt.Sprintf("/api/games/%s", gameToDelete), nil, cookiesFor(userID))
	assertStatus(t, delResp, http.StatusNoContent)

	// Count should now be 2.
	w2 := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w2, http.StatusOK)
	var resp2 map[string]any
	decodeJSON(t, w2, &resp2)
	if got := totalGamesFor(t, resp2); got != 2 {
		t.Errorf("post-delete: expected total_games = 2, got %d", got)
	}
}

func TestGetTotalGamesPlayed_BoardTitleEchoed(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "tgp_title", "tgp_title@example.com")
	boardID := createTestBoard(t, "Echo Title Board", 5, userID, nil)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/total-played-games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "board_title", "Echo Title Board")
}

/// ===== GET /boards search filter =====

func TestGetBoards_SearchByTitle(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "searchauthor", "searchauthor@example.com")
	createTestBoard(t, "Go Programming", 5, userID, nil)
	createTestBoard(t, "Vue Basics", 5, userID, nil)
	createTestBoard(t, "Go Advanced", 5, userID, nil)
	w := doRequest(http.MethodGet, "/api/boards?search=Go", nil)
	assertStatus(t, w, http.StatusOK)
	var resp []map[string]any
	decodeJSON(t, w, &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 boards matching 'Go', got %d", len(resp))
	}
}

func TestGetBoards_SearchTypoTolerant(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "fuzzauthor", "fuzz@example.com")
	createTestBoard(t, "Go Programming", 5, userID, nil)
	createTestBoard(t, "Vue Basics", 5, userID, nil)

	// 'programing' (one missing 'm') is not a substring of any title, so
	// the previous ILIKE-only implementation would have returned zero
	// results. Trigram similarity (sim=0.625, well above the 0.3 default
	// threshold) catches it.
	w := doRequest(http.MethodGet, "/api/boards?search=programing", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 fuzzy match for 'programing', got %d", len(resp))
	}
	assertJSONField(t, resp[0], "title", "Go Programming")
}

func TestGetBoards_SearchCaseInsensitive(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "caseauthor", "case@example.com")
	createTestBoard(t, "Go Programming", 5, userID, nil)

	// Trigram similarity is case-insensitive, and so is ILIKE; verify the
	// API surface honors that for both arms of the OR.
	for _, q := range []string{"GO", "go", "Go", "PROGRAMMING"} {
		w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards?search=%s", q), nil)
		assertStatus(t, w, http.StatusOK)

		var resp []map[string]any
		decodeJSON(t, w, &resp)
		if len(resp) != 1 {
			t.Errorf("search=%q: expected 1 result, got %d", q, len(resp))
		}
	}
}

func TestGetBoards_SearchRelevanceRanking(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "rankauthor", "rank@example.com")

	// All three contain 'Go', but their similarity to the query 'Go
	// Programming' differs:
	//   sim('Go Programming', 'Go Programming') = 1.00 (exact)
	//   sim('Go Advanced',    'Go Programming') ≈ 0.13
	//   sim('Lego Stories',   'Go Programming') ≈ 0.10
	// We assert that the exact match comes first regardless of insertion
	// order and regardless of the requested ?sort= (which should be
	// overridden by relevance when search is set).
	createTestBoard(t, "Go Advanced", 5, userID, nil)
	time.Sleep(2 * time.Millisecond)
	createTestBoard(t, "Go Programming", 5, userID, nil)
	time.Sleep(2 * time.Millisecond)
	createTestBoard(t, "Lego Stories", 5, userID, nil)

	// Pick sort=oldest deliberately: under that sort, 'Go Advanced' would
	// come first. Relevance ranking must override it.
	w := doRequest(http.MethodGet, "/api/boards?search=Go+Programming&sort=oldest", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)
	if len(resp) == 0 {
		t.Fatal("expected at least one match")
	}
	assertJSONField(t, resp[0], "title", "Go Programming")
}

func TestGetBoards_SearchNoMatches(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "nomatchauthor", "nomatch@example.com")
	createTestBoard(t, "Go Programming", 5, userID, nil)
	createTestBoard(t, "Vue Basics", 5, userID, nil)

	// A query with no shared trigrams and no substring match returns
	// empty (similarity = 0, well below the 0.3 threshold).
	w := doRequest(http.MethodGet, "/api/boards?search=xyzqrs", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)
	if len(resp) != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", len(resp))
	}
}
