package handler

import (
	"fmt"
	"net/http"
	"testing"
)

// setupForGames creates a user and board, returning both IDs.
func setupForGames(t *testing.T) (userID, boardID string) {
	t.Helper()
	userID = createTestUser(t, "gameuser", "gameuser@example.com")
	boardID = createTestBoard(t, "Game Board", 5, userID)
	return userID, boardID
}

func TestCreateGame(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID, boardID := setupForGames(t)

	w := doRequest(http.MethodPost,
		fmt.Sprintf("/api/boards/%s/games?player_id=%s", boardID, userID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "player_id", userID)
	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "status", "active")

	if _, ok := resp["id"]; !ok {
		t.Error("expected 'id' field in response")
	}
}

func TestGetGameByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "id", gameID)
	assertJSONField(t, resp, "player_id", userID)
	assertJSONField(t, resp, "board_id", boardID)
	assertJSONField(t, resp, "status", "active")
}

func TestGetGameByID_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodGet, "/api/games/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestListGamesByPlayer(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
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
	t.Cleanup(func() { cleanupTables(t) })
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
	t.Cleanup(func() { cleanupTables(t) })
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
	t.Cleanup(func() { cleanupTables(t) })
	userID := createTestUser(t, "emptyboarduser", "emptyboard@example.com")
	boardID := createTestBoard(t, "Empty Board", 5, userID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/games", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 games, got %d", len(resp))
	}
}

func TestUpdateGameStatus(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	w := doRequest(http.MethodPut,
		fmt.Sprintf("/api/games/%s/status?player_id=%s", gameID, userID),
		map[string]any{"status": "completed"},
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "id", gameID)
	assertJSONField(t, resp, "status", "completed")
}

func TestUpdateGameStatus_Abandoned(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	w := doRequest(http.MethodPut,
		fmt.Sprintf("/api/games/%s/status?player_id=%s", gameID, userID),
		map[string]any{"status": "abandoned"},
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "status", "abandoned")
}

func TestUpdateGameStatus_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID := createTestUser(t, "ghostplayer", "ghost@example.com")

	w := doRequest(http.MethodPut,
		fmt.Sprintf("/api/games/00000000-0000-0000-0000-000000000000/status?player_id=%s", userID),
		map[string]any{"status": "completed"},
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateGameStatus_WrongPlayer(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	user1, boardID := setupForGames(t)
	user2 := createTestUser(t, "otherplayer", "other@example.com")
	gameID := createTestGame(t, user1, boardID)

	// user2 tries to update user1's game
	w := doRequest(http.MethodPut,
		fmt.Sprintf("/api/games/%s/status?player_id=%s", gameID, user2),
		map[string]any{"status": "completed"},
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestDeleteGame(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID, boardID := setupForGames(t)
	gameID := createTestGame(t, userID, boardID)

	// Delete the game
	w := doRequest(http.MethodDelete, fmt.Sprintf("/api/games/%s", gameID), nil)
	assertStatus(t, w, http.StatusNoContent)

	// Verify game is gone
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s", gameID), nil)
	// Should return error (game not found)
	if getResp.Code == http.StatusOK {
		t.Error("expected game to be deleted, but it still exists")
	}
}

func TestDeleteGame_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodDelete, "/api/games/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestCreateGame_MultipleOnSameBoard(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID, boardID := setupForGames(t)

	// A player can have multiple games on the same board
	game1 := createTestGame(t, userID, boardID)
	game2 := createTestGame(t, userID, boardID)

	if game1 == game2 {
		t.Error("expected different game IDs for multiple games on same board")
	}
}

func TestDeleteGame_CascadesGameCells(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })
	userID, boardID := setupForGames(t)
	cellID := createTestCell(t, boardID, "Test Cell")
	gameID := createTestGame(t, userID, boardID)

	// Create game cells
	doRequest(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), []map[string]any{
		{"cell_id": cellID, "content": "Test Cell", "position": 0},
	})

	// Delete the game (should cascade to game_cells)
	w := doRequest(http.MethodDelete, fmt.Sprintf("/api/games/%s", gameID), nil)
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
