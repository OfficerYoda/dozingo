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

	w := doRequestWithCookies(http.MethodDelete,
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
		Token:  auth.GenerateSessionToken(),
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
