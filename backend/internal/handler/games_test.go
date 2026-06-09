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

	w := doRequestWithCookies(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games", boardID), nil, cookiesFor(userID))
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
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/games", boardID), nil)
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
	cellID := createTestCell(t, boardID, "Test Cell")
	gameID := createTestGame(t, userID, boardID)

	// Create game cells
	doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), []map[string]any{
		{"cell_id": cellID, "content": "Test Cell", "position": 0},
	}, cookiesFor(userID))

	// Delete the game (should cascade to game_cells)
	w := doRequestWithCookies(http.MethodDelete, fmt.Sprintf("/api/games/%s", gameID), nil, cookiesFor(userID))
	assertStatus(t, w, http.StatusNoContent)

	// Verify game cells are gone too
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	// The game is gone so game cells should be empty or the game_id is invalid
	var cells []map[string]any
	if getResp.Code == http.StatusOK {
		decodeJSON(t, getResp, &cells)
		if len(cells) != 0 {
			t.Errorf("expected 0 game cells after game deletion, got %d", len(cells))
		}
	}
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

// ===== Latent bugs (these tests are expected to fail until the bugs are
// fixed in a follow-up commit). =====

// TestUpdateGameStatus_Anonymous documents that anonymous players cannot
// update their own anonymous game's status today. The repository's
// UpdateStatus method only forwards PlayerID to the SQL query, so the
// session-id branch in the WHERE clause never matches for anon callers and
// the UPDATE affects 0 rows -> ErrNoRows -> 404. Should be 200.
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

// TestDeleteGame_AnonymousNonOwner documents that a *different* anonymous
// caller can currently delete someone else's anonymous game. The
// service-layer ownership check compares sessionUser.UserID == game.PlayerID;
// for anon-vs-anon both are zero pgtype.UUID, so the equality holds and
// authorization passes incorrectly. Authorization for anon games must be by
// session_id, not by player_id.
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
