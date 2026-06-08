package handler

import (
	"fmt"
	"net/http"
	"testing"
)

// setupForGameCells creates a user, board, cell, and game, returning relevant IDs.
func setupForGameCells(t *testing.T) (userID, boardID, cellID, gameID string) {
	t.Helper()
	userID = createTestUser(t, "gamecelluser", "gamecelluser@example.com")
	boardID = createTestBoard(t, "GameCell Board", 5, userID, nil)
	cellID = createTestCell(t, boardID, "Source Cell")
	gameID = createTestGame(t, userID, boardID)
	return
}

func TestCreateGameCells(t *testing.T) {
	setupTest(t)
	_, boardID, cellID, gameID := setupForGameCells(t)

	// Create a second cell
	cellID2 := createTestCell(t, boardID, "Source Cell 2")

	body := []map[string]any{
		{"cell_id": cellID, "content": "Cell Content 1", "position": 0},
		{"cell_id": cellID2, "content": "Cell Content 2", "position": 1},
	}

	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), body, cookiesForGame(t, gameID))
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 game cells, got %d", len(resp))
		return
	}

	// Verify cells are ordered by position
	if pos, ok := resp[0]["position"].(float64); !ok || int(pos) != 0 {
		t.Errorf("expected first cell position = 0, got %v", resp[0]["position"])
	}
	if pos, ok := resp[1]["position"].(float64); !ok || int(pos) != 1 {
		t.Errorf("expected second cell position = 1, got %v", resp[1]["position"])
	}

	assertJSONField(t, resp[0], "content", "Cell Content 1")
	assertJSONField(t, resp[1], "content", "Cell Content 2")
	assertJSONField(t, resp[0], "game_id", gameID)
}

func TestCreateGameCells_AnonymousNonOwner(t *testing.T) {
	setupTest(t)
	_, _, cellID, gameID := setupForGameCells(t)

	body := []map[string]any{
		{"cell_id": cellID, "content": "Anon Cell", "position": 0},
	}

	// Anonymous caller (no cookie) is automatically minted a fresh session by
	// RequireSession middleware, so the request reaches the ownership check.
	// That check fails because the game belongs to someone else, returning
	// 403 Forbidden.
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), body)
	assertStatus(t, w, http.StatusForbidden)
}

func TestCreateGameCells_FullBoard(t *testing.T) {
	setupTest(t)
	_, boardID, _, gameID := setupForGameCells(t)

	// Create 9 cells for a 3x3 board worth of game cells
	cells := make([]map[string]any, 9)
	for i := 0; i < 9; i++ {
		cellID := createTestCell(t, boardID, fmt.Sprintf("Cell %d", i))
		cells[i] = map[string]any{
			"cell_id":  cellID,
			"content":  fmt.Sprintf("Game Cell %d", i),
			"position": i,
		}
	}

	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), cells, cookiesForGame(t, gameID))
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 9 {
		t.Errorf("expected 9 game cells, got %d", len(resp))
	}
}

// The bulk endpoint accepts up to 64 items per call. Anything larger
// must be rejected at validation time (before the request reaches the
// service or the database) so a malicious caller can't drive arbitrary-
// size payloads through this endpoint. 64 sits comfortably above the
// 6x6=36 cells of the largest currently-supported board, leaving room
// for plausible future board sizes (e.g. 8x8).
func TestCreateGameCells_TooMany_Rejected(t *testing.T) {
	setupTest(t)
	_, _, cellID, gameID := setupForGameCells(t)

	// 65 items - just over the cap. The cell_id intentionally repeats
	// the same UUID; validation must reject the array length before any
	// DB lookup runs, so the references don't matter.
	cells := make([]map[string]any, 65)
	for i := 0; i < 65; i++ {
		cells[i] = map[string]any{
			"cell_id":  cellID,
			"content":  "x",
			"position": i,
		}
	}

	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), cells, cookiesForGame(t, gameID))
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusBadRequest {
		t.Errorf("expected 422/400 for >64 items, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestGetGameCellsByGameID(t *testing.T) {
	setupTest(t)
	_, _, cellID, gameID := setupForGameCells(t)

	// Create game cells
	body := []map[string]any{
		{"cell_id": cellID, "content": "Cell A", "position": 0},
		{"cell_id": cellID, "content": "Cell B", "position": 1},
		{"cell_id": cellID, "content": "Cell C", "position": 2},
	}
	doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), body, cookiesForGame(t, gameID))

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 3 {
		t.Errorf("expected 3 game cells, got %d", len(resp))
		return
	}

	// Should be ordered by position
	for i, cell := range resp {
		if pos, ok := cell["position"].(float64); !ok || int(pos) != i {
			t.Errorf("expected position %d, got %v", i, cell["position"])
		}
	}
}

func TestGetGameCellsByGameID_Empty(t *testing.T) {
	setupTest(t)
	_, _, _, gameID := setupForGameCells(t)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 game cells, got %d", len(resp))
	}
}

func TestUpdateGameCellMark_MarkTrue(t *testing.T) {
	setupTest(t)
	_, _, cellID, gameID := setupForGameCells(t)

	gameCellID := createTestGameCell(t, gameID, cellID, "Markable Cell", 0)

	// Mark the cell
	w := doRequestWithCookies(http.MethodPut, fmt.Sprintf("/api/games/%s/cells/%s", gameID, gameCellID),
		map[string]any{"is_marked": true},
		cookiesForGame(t, gameID),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "game_cell_id", gameCellID)
	if marked, ok := resp["is_marked"].(bool); !ok || !marked {
		t.Errorf("expected is_marked = true, got %v", resp["is_marked"])
	}
}

func TestUpdateGameCellMark_MarkFalse(t *testing.T) {
	setupTest(t)
	_, _, cellID, gameID := setupForGameCells(t)

	gameCellID := createTestGameCell(t, gameID, cellID, "Toggle Cell", 0)
	cookies := cookiesForGame(t, gameID)

	// Mark it true first
	doRequestWithCookies(http.MethodPut, fmt.Sprintf("/api/games/%s/cells/%s", gameID, gameCellID),
		map[string]any{"is_marked": true},
		cookies,
	)

	// Now unmark it
	w := doRequestWithCookies(http.MethodPut, fmt.Sprintf("/api/games/%s/cells/%s", gameID, gameCellID),
		map[string]any{"is_marked": false},
		cookies,
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	if marked, ok := resp["is_marked"].(bool); !ok || marked {
		t.Errorf("expected is_marked = false, got %v", resp["is_marked"])
	}
}

func TestUpdateGameCellMark_NotFound(t *testing.T) {
	setupTest(t)
	_, _, _, gameID := setupForGameCells(t)

	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/games/%s/cells/00000000-0000-0000-0000-000000000000", gameID),
		map[string]any{"is_marked": true},
		cookiesForGame(t, gameID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateGameCellMark_WrongGame(t *testing.T) {
	setupTest(t)
	userID, boardID, cellID, gameID := setupForGameCells(t)

	gameCellID := createTestGameCell(t, gameID, cellID, "Cell", 0)

	// Create a second game (same player so cookie lookup works for both)
	game2ID := createTestGame(t, userID, boardID)

	// Try to mark cell from game1 using game2's path
	w := doRequestWithCookies(http.MethodPut, fmt.Sprintf("/api/games/%s/cells/%s", game2ID, gameCellID),
		map[string]any{"is_marked": true},
		cookiesForGame(t, game2ID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetGameCellsByGameID_InvalidGameID(t *testing.T) {
	w := doRequest(http.MethodGet, "/api/games/not-a-uuid/cells", nil)
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestCreateGameCells_IsMarkedDefaultsFalse(t *testing.T) {
	setupTest(t)
	_, _, cellID, gameID := setupForGameCells(t)

	body := []map[string]any{
		{"cell_id": cellID, "content": "New Cell", "position": 0},
	}

	w := doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), body, cookiesForGame(t, gameID))
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Fatalf("expected 1 game cell, got %d", len(resp))
	}

	if marked, ok := resp[0]["is_marked"].(bool); !ok || marked {
		t.Errorf("expected is_marked = false by default, got %v", resp[0]["is_marked"])
	}
}

func TestGetGameCellsByGameID_CellIDNullable(t *testing.T) {
	setupTest(t)
	_, boardID, cellID, gameID := setupForGameCells(t)

	// Create game cells with a cell reference
	body := []map[string]any{
		{"cell_id": cellID, "content": "Linked Cell", "position": 0},
	}
	doRequestWithCookies(http.MethodPost, fmt.Sprintf("/api/games/%s/cells", gameID), body, cookiesForGame(t, gameID))

	// Delete the source cell (should SET NULL on game_cells.cell_id)
	doRequestWithCookies(http.MethodDelete, fmt.Sprintf("/api/boards/%s/cells/%s", boardID, cellID), nil, cookiesForBoard(t, boardID))

	// Fetch game cells — cell_id should now be null
	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Fatalf("expected 1 game cell, got %d", len(resp))
	}

	if resp[0]["cell_id"] != nil {
		t.Errorf("expected cell_id = null after source cell deletion, got %v", resp[0]["cell_id"])
	}

	// Content should still be preserved (snapshot)
	assertJSONField(t, resp[0], "content", "Linked Cell")
}
