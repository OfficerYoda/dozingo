package handler

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// setupForGames creates a user and board, returning both IDs.
func setupForGames(t *testing.T) (userID, boardID string) {
	t.Helper()
	userID = createTestUser(t, "gameuser", "gameuser@example.com")
	boardID = createTestBoard(t, "Game Board", 5, userID, nil)
	return userID, boardID
}

func TestCreateGame(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)

	body := seedGameCellsBody(t, boardID)
	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), body, cookiesFor(userID))
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "player_id", userID)
	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "status", "active")

	if _, ok := resp["game_id"]; !ok {
		t.Error("expected 'id' field in response")
	}
}

func TestCreateGame_Anonymous(t *testing.T) {
	setupTest(t)
	_, boardID := setupForGames(t)

	// No cookie -> server mints a fresh anonymous session and creates the
	// game with player_id = NULL, session_id set.
	body := seedGameCellsBody(t, boardID)
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/games", boardID), body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	if val, ok := resp["player_id"]; !ok || val != nil {
		t.Errorf("expected player_id to be null for anonymous game, got %v", val)
	}
	sessionID, ok := resp["session_id"].(string)
	if !ok || sessionID == "" {
		t.Errorf("expected session_id to be a non-empty string, got %v", resp["session_id"])
	}
	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "status", "active")

	// The response must carry a Set-Cookie for the freshly minted session.
	if cookie := extractSessionCookie(w); cookie == nil || cookie.Value == "" {
		t.Error("expected a session_token cookie to be set on anonymous game creation")
	}
}

func TestCreateGame_AnonymousReusesExistingSession(t *testing.T) {
	setupTest(t)
	_, boardID := setupForGames(t)

	// Pre-mint an anonymous session and post with that cookie.
	token, cookie := mintAnonSession(t, 30*24*time.Hour)
	row, ok := loadSessionByToken(t, token)
	if !ok {
		t.Fatalf("failed to load freshly minted anon session by token")
	}
	expectedSessionID := row.SessionID.String()

	gameID := createAnonGame(t, cookie, boardID)

	// Read it back and assert the server reused our pre-minted session.
	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s", gameID), nil)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "session_id", expectedSessionID)
	if val, ok := resp["player_id"]; !ok || val != nil {
		t.Errorf("expected player_id to be null, got %v", val)
	}
}

func TestGetGameByID(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "game_id", gameID)
	assertJSONField(t, resp, "player_id", userID)
	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "status", "active")
}

func TestGetGameByID_NotFound(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/games/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestListGamesByPlayer(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)

	createTestGame(t, userID, boardID)
	createTestGame(t, userID, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s/games", userID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 games, got %d", len(resp))
	}
}

func TestListGamesByPlayer_Empty(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "nogameuser", "nogame@example.com")

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s/games", userID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 games, got %d", len(resp))
	}
}

func TestListGamesByBoard(t *testing.T) {
	setupTest(t)
	user1, boardID := setupForGames(t)
	user2 := createTestUser(t, "gameuser2", "gameuser2@example.com")

	createTestGame(t, user1, boardID)
	createTestGame(t, user2, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 games, got %d", len(resp))
	}
}

func TestListGamesByBoard_Empty(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "emptyboarduser", "emptyboard@example.com")
	boardID := createTestBoard(t, "Empty Board", 5, userID, nil)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 games, got %d", len(resp))
	}
}

func TestUpdateGameStatus(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/games/%s/status", gameID),
		map[string]any{"status": "completed"},
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "game_id", gameID)
	assertJSONField(t, resp, "status", "completed")
}

func TestUpdateGameStatus_Abandoned(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/games/%s/status", gameID),
		map[string]any{"status": "abandoned"},
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "status", "abandoned")
}

// The status field is constrained to a closed enum at the Huma input
// layer. Off-enum values must be rejected at validation time so they
// never reach the repository (which would happily forward whatever the
// underlying Postgres column accepts). This test guards against the
// enum tag silently disappearing.
func TestUpdateGameStatus_InvalidValue_Rejected(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	for _, bad := range []string{"garbage", "", "ACTIVE", "deleted", "pending"} {
		t.Run("status="+bad, func(t *testing.T) {
			w := doRequestWithCookies(http.MethodPut,
				fmt.Sprintf("/api/games/%s/status", gameID),
				map[string]any{"status": bad},
				cookiesFor(userID),
			)
			if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
				t.Errorf("expected 422/400 for status %q, got %d (body: %s)", bad, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateGameStatus_NotFound(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "ghostplayer", "ghost@example.com")

	w := doRequestWithCookies(http.MethodPut,
		"/api/games/00000000-0000-0000-0000-000000000000/status",
		map[string]any{"status": "completed"},
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateGameStatus_WrongPlayer(t *testing.T) {
	setupTest(t)
	owner, boardID := setupForGames(t)
	gameID := createTestGame(t, owner, boardID)

	stranger := createTestUser(t, "stranger", "stranger@example.com")

	// stranger tries to update owner's game; service returns
	// domain.ErrUnauthorized -> 401.
	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/games/%s/status", gameID),
		map[string]any{"status": "completed"},
		cookiesFor(stranger),
	)
	assertStatus(t, w, http.StatusForbidden)
}

func TestDeleteGame(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	// Delete the game
	w := doRequestWithCookies(http.MethodDelete, fmt.Sprintf("/api/games/%s", gameID), nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	// Verify game is gone
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s", gameID), nil)
	// Should return error (game not found)
	if getResp.Code == http.StatusOK {
		t.Error("expected game to be deleted, but it still exists")
	}
}

func TestDeleteGame_NotFound(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "delnogame", "delnogame@example.com")

	w := doRequestWithCookies(http.MethodDelete,
		"/api/games/00000000-0000-0000-0000-000000000000",
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestDeleteGame_NotFound_Anonymous(t *testing.T) {
	setupTest(t)

	// Anonymous caller -> RequireSession mints a session, then the service
	// loads the (nonexistent) game and returns ErrNotFound -> 404.
	// Previously this returned 401; the relaxed authentication policy means
	// the absence of a cookie is no longer an error in itself.
	w := doRequest(http.MethodDelete, "/api/games/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestCreateGame_MultipleOnSameBoard(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)

	// A player can have multiple games on the same board
	game1 := createTestGame(t, userID, boardID)
	game2 := createTestGame(t, userID, boardID)

	if game1 == game2 {
		t.Error("expected different game IDs for multiple games on same board")
	}
}

func TestDeleteGame_CascadesGameCells(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	// createTestGame seeds size*size game cells (board size 5 -> 25 cells).
	// Sanity-check they exist before deletion so the post-delete assertion
	// is meaningful.
	pre := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, pre, http.StatusOK)
	var preCells []map[string]any
	decodeJSON(t, pre, &preCells)
	if len(preCells) == 0 {
		t.Fatalf("expected game cells to be seeded by createTestGame, got 0")
	}

	// Delete the game (should cascade to game_cells)
	w := doRequestWithCookies(http.MethodDelete, fmt.Sprintf("/api/games/%s", gameID), nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	// Verify game cells are gone too. After the game is deleted the cells
	// endpoint returns 404 (the service checks game existence before querying
	// cells), which also proves the cascade removed the game row.
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, getResp, http.StatusNotFound)
}

// ===== Anonymous + /me/games coverage =====

func TestGetGameByID_Anonymous(t *testing.T) {
	setupTest(t)
	_, boardID := setupForGames(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)
	gameID := createAnonGame(t, cookie, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "game_id", gameID)
	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "status", "active")

	if val, ok := resp["player_id"]; !ok || val != nil {
		t.Errorf("expected player_id to be null for anonymous game, got %v", val)
	}
	sessionID, ok := resp["session_id"].(string)
	if !ok || sessionID == "" {
		t.Errorf("expected session_id to be a non-empty string, got %v", resp["session_id"])
	}
}

func TestListByCurrentSession_Authenticated(t *testing.T) {
	setupTest(t)
	user1, boardID := setupForGames(t)
	user2 := createTestUser(t, "othergameuser", "othergameuser@example.com")

	createTestGame(t, user1, boardID)
	createTestGame(t, user1, boardID)
	createTestGame(t, user2, boardID)

	w := doRequestWithCookies(http.MethodGet, "/api/me/games", nil, cookiesFor(user1))
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 games for user1's session, got %d", len(resp))
	}
	for _, g := range resp {
		if pid, ok := g["player_id"].(string); !ok || pid != user1 {
			t.Errorf("expected each game to belong to user1 (%s), got %v", user1, g["player_id"])
		}
	}
}

func TestListByCurrentSession_Anonymous(t *testing.T) {
	setupTest(t)
	_, boardID := setupForGames(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)
	createAnonGame(t, cookie, boardID)
	createAnonGame(t, cookie, boardID)

	w := doRequestWithCookies(http.MethodGet, "/api/me/games", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 anonymous games for the session, got %d", len(resp))
	}
	for _, g := range resp {
		if val, ok := g["player_id"]; !ok || val != nil {
			t.Errorf("expected each anonymous game to have player_id == null, got %v", val)
		}
	}
}

func TestListByCurrentSession_Empty(t *testing.T) {
	setupTest(t)

	// Pre-minted session has no games on it.
	_, cookie := mintAnonSession(t, 30*24*time.Hour)

	w := doRequestWithCookies(http.MethodGet, "/api/me/games", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 games on a fresh session, got %d", len(resp))
	}
}

func TestListByCurrentSession_AfterLogin(t *testing.T) {
	setupTest(t)
	_, boardID := setupForGames(t)

	// Anonymous session creates a game...
	_, cookie := mintAnonSession(t, 30*24*time.Hour)
	anonGameID := createAnonGame(t, cookie, boardID)

	// ...then registers a user reusing the same cookie. Register attaches
	// the user to the existing session row, so the session_id stays the
	// same and the previously-anonymous game is still listed under it.
	body := map[string]any{
		"username": "promoteduser",
		"password": "testpassword123",
		"email":    "promoteduser@example.com",
	}
	regResp := doRequestWithCookies(http.MethodPost, "/api/auth/register", body, []*http.Cookie{cookie})
	assertStatus(t, regResp, http.StatusOK)

	w := doRequestWithCookies(http.MethodGet, "/api/me/games", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Fatalf("expected 1 game on the promoted session, got %d", len(resp))
	}
	if id, ok := resp[0]["game_id"].(string); !ok || id != anonGameID {
		t.Errorf("expected listed game to be %s, got %v", anonGameID, resp[0]["game_id"])
	}
	// The anon game's player_id is still NULL even though the session now
	// has a user_id; only games created *after* login get the player_id.
	if val, ok := resp[0]["player_id"]; !ok || val != nil {
		t.Errorf("expected player_id to remain null on the pre-login game, got %v", val)
	}
}

// ===== Anonymous game authorization =====

// TestUpdateGameStatus_Anonymous verifies that an anonymous player can update
// the status of their own game. Authorization is by session_id: the SQL WHERE
// clause uses (player_id IS NULL AND session_id = $4) for anon callers, so the
// update matches and returns the updated game.
func TestUpdateGameStatus_Anonymous(t *testing.T) {
	setupTest(t)
	_, boardID := setupForGames(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)
	gameID := createAnonGame(t, cookie, boardID)

	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/games/%s/status", gameID),
		map[string]any{"status": "completed"},
		[]*http.Cookie{cookie},
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)
	assertJSONField(t, resp, "status", "completed")
}

// TestDeleteGame_AnonymousNonOwner verifies that a different anonymous caller
// cannot delete another user's anonymous game. The service-layer ownership
// check branches on game.PlayerID.Valid: when player_id is NULL (anon game)
// it compares sessionUser.SessionID == game.SessionID, so a stranger with a
// different session correctly receives 403 Forbidden.
func TestDeleteGame_AnonymousNonOwner(t *testing.T) {
	setupTest(t)
	_, boardID := setupForGames(t)

	_, ownerCookie := mintAnonSession(t, 30*24*time.Hour)
	gameID := createAnonGame(t, ownerCookie, boardID)

	_, strangerCookie := mintAnonSession(t, 30*24*time.Hour)
	w := doRequestWithCookies(http.MethodDelete,
		fmt.Sprintf("/api/games/%s", gameID),
		nil,
		[]*http.Cookie{strangerCookie},
	)
	assertStatus(t, w, http.StatusForbidden)
}

// ===== Game-create input validation =====

// setupForGamesSize creates a user and a board with the given size, returning
// both IDs. Helper for tests that want to pin down the size to keep payloads
// small.
func setupForGamesSize(t *testing.T, size int) (userID, boardID string) {
	t.Helper()
	userID = createTestUser(t, fmt.Sprintf("gameuser_size%d", size), fmt.Sprintf("gameuser_size%d@example.com", size))
	boardID = createTestBoard(t, fmt.Sprintf("Board %d", size), size, userID, nil)
	return userID, boardID
}

// TestCreateGame_RejectsMismatchedCellCount documents the size^2 contract:
// the body must carry exactly board.size*board.size items. One short is a
// bad-input error from the service.
func TestCreateGame_RejectsMismatchedCellCount(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGamesSize(t, 4)

	// Seed full body, then truncate by one to force the mismatch.
	body := seedGameCellsBody(t, boardID)
	body = body[:len(body)-1]

	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), body, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

// TestCreateGame_RejectsEmptyBody asserts an empty cell list is rejected
// for any board with size > 0.
func TestCreateGame_RejectsEmptyBody(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGamesSize(t, 4)

	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), []map[string]any{}, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

// TestCreateGame_RejectsTooManyCells exercises the maxItems:64 cap on the
// Huma input schema. 65 items must be rejected at validation time, before
// the handler runs.
func TestCreateGame_RejectsTooManyCells(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGamesSize(t, 4)

	// Build 65 syntactically-valid items. The cell IDs need not exist;
	// validation must reject the request before the service is called.
	items := make([]map[string]any, 65)
	for i := range items {
		items[i] = map[string]any{
			"cell_id":  "00000000-0000-0000-0000-000000000000",
			"position": i,
		}
	}

	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), items, cookiesFor(userID))
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

// TestCreateGame_RejectsCellsFromDifferentBoard pins down the security fix
// from the review: a client cannot supply cell IDs that belong to a
// different board. The board-scoped cell lookup in the service must reject
// these as bad input.
func TestCreateGame_RejectsCellsFromDifferentBoard(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "cross_board_user", "cross_board@example.com")

	// Two same-size boards, each fully seeded.
	targetBoardID := createTestBoard(t, "Target Board", 4, userID, nil)
	otherBoardID := createTestBoard(t, "Other Board", 4, userID, nil)

	// Build a body whose cell_ids belong to the *other* board, but post to
	// the target board. The size matches, so only the board-scope check
	// can reject the request.
	body := seedGameCellsBody(t, otherBoardID)

	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", targetBoardID), body, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

// TestCreateGame_RejectsUnknownCellID asserts a non-existent cell id is
// rejected even if the count and board match up.
func TestCreateGame_RejectsUnknownCellID(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGamesSize(t, 4)

	// 4 items where one is a zero UUID.
	body := seedGameCellsBody(t, boardID)
	body[0]["cell_id"] = "00000000-0000-0000-0000-000000000000"

	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), body, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

// TestCreateGame_RejectsBoardNotFound asserts a missing board surfaces as
// bad input (per the service: "board does not exist: ErrBadInput"). The
// body shape is irrelevant; the board check runs first.
func TestCreateGame_RejectsBoardNotFound(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "no_board_user", "noboard@example.com")

	body := []map[string]any{
		{"cell_id": "00000000-0000-0000-0000-000000000000", "position": 0},
	}
	w := doRequestWithCookies(http.MethodPost,
		"/api/boards/00000000-0000-0000-0000-000000000000/games", body, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)
}

// TestCreateGame_AtomicityOnCellFailure pins down the transaction wrapping:
// when the cell-resolution step fails, the outer game row must not be
// persisted. Without the transaction, the failed second insert would leave
// an orphan game discoverable via the by-board listing.
func TestCreateGame_AtomicityOnCellFailure(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGamesSize(t, 4)

	// Force a cell-resolution failure by including a zero UUID after the
	// board check has passed.
	body := seedGameCellsBody(t, boardID)
	body[0]["cell_id"] = "00000000-0000-0000-0000-000000000000"

	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), body, cookiesFor(userID))
	assertStatus(t, w, http.StatusBadRequest)

	// Verify no orphan game row was committed for this board.
	listResp := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/games", boardID), nil)
	assertStatus(t, listResp, http.StatusOK)
	var games []map[string]any
	decodeJSON(t, listResp, &games)
	if len(games) != 0 {
		t.Errorf("expected 0 games on the board after a failed create, got %d (transaction rollback regressed)", len(games))
	}
}

// TestCreateGame_SnapshotsBoardCellContent verifies the new behavior: the
// content for each game cell is sourced from the board's cell, not from
// anything the client supplied. The handler input has no content field.
func TestCreateGame_SnapshotsBoardCellContent(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGamesSize(t, 4)

	// Seed cells with deterministic contents so we can match them in the
	// game cells output. Board size is 4, so we need 16 cells.
	wantContent := map[string]string{}
	body := make([]map[string]any, 0, 16)
	for i := range 16 {
		content := fmt.Sprintf("Board Content %d", i)
		cellID := createTestCell(t, boardID, content)
		wantContent[cellID] = content
		body = append(body, map[string]any{
			"cell_id":  cellID,
			"position": i,
		})
	}

	createResp := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), body, cookiesFor(userID))
	assertStatus(t, createResp, http.StatusOK)
	var game map[string]any
	decodeJSON(t, createResp, &game)
	gameID := game["game_id"].(string)

	// Read game cells back and check each one's content matches the
	// board cell it was created from.
	cellsResp := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, cellsResp, http.StatusOK)
	var gameCells []map[string]any
	decodeJSON(t, cellsResp, &gameCells)
	if len(gameCells) != 16 {
		t.Fatalf("expected 16 game cells, got %d", len(gameCells))
	}
	for _, gc := range gameCells {
		cellID, _ := gc["cell_id"].(string)
		gotContent, _ := gc["content"].(string)
		want, ok := wantContent[cellID]
		if !ok {
			t.Errorf("game cell references unknown cell_id %q", cellID)
			continue
		}
		if gotContent != want {
			t.Errorf("game cell content for cell %s: got %q, want %q", cellID, gotContent, want)
		}
	}
}

// TestCreateGame_SeedsAllCells happy path: a successful create returns a
// game whose cells exactly cover positions 0..(size^2 - 1).
func TestCreateGame_SeedsAllCells(t *testing.T) {
	setupTest(t)
	userID, boardID := setupForGamesSize(t, 4)

	gameID := createTestGame(t, userID, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, w, http.StatusOK)
	var cells []map[string]any
	decodeJSON(t, w, &cells)

	if len(cells) != 16 {
		t.Fatalf("expected 16 game cells (size 4 -> 16), got %d", len(cells))
	}
	seen := make(map[int]bool, 16)
	for _, c := range cells {
		pos, ok := c["position"].(float64)
		if !ok {
			t.Errorf("expected position to be a number, got %T", c["position"])
			continue
		}
		seen[int(pos)] = true
	}
	for i := range 16 {
		if !seen[i] {
			t.Errorf("expected position %d to be present", i)
		}
	}
}
