package handler

import (
	"fmt"
	"net/http"
	"testing"
)

// setupForGameCells creates a user, a board, and a game (with auto-seeded
// game cells), and returns the relevant IDs. The board uses size 5, so the
// game has 25 game cells out of the box. Tests that need finer control over
// the board size should construct their own setup.
func setupForGameCells(t *testing.T) (userID, boardID, gameID string) {
	t.Helper()
	userID = createTestUser(t, "gamecelluser", "gamecelluser@example.com")
	boardID = createTestBoard(t, "GameCell Board", 5, userID, nil)
	gameID = createTestGame(t, userID, boardID)
	return userID, boardID, gameID
}

func TestGetGameCellsByGameID(t *testing.T) {
	setupTest(t)
	_, _, gameID := setupForGameCells(t)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	// createTestGame seeds size*size cells (5*5 = 25).
	if len(resp) != 25 {
		t.Errorf("expected 25 game cells, got %d", len(resp))
		return
	}

	// Should be ordered by position 0..24.
	for i, cell := range resp {
		if pos, ok := cell["position"].(float64); !ok || int(pos) != i {
			t.Errorf("expected position %d, got %v", i, cell["position"])
		}
	}
}

// TestGetGameCellsByGameID_NoSuchGame asserts that listing cells for a
// nonexistent game id returns an empty list (the GET endpoint does not 404
// on missing games — it just has nothing to return).
func TestGetGameCellsByGameID_NoSuchGame(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/games/00000000-0000-0000-0000-000000000000/cells", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 game cells for unknown game, got %d", len(resp))
	}
}

func TestGetGameCellsByGameID_InvalidGameID(t *testing.T) {
	w := doRequest(http.MethodGet, "/api/games/not-a-uuid/cells", nil)
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestCreateGame_IsMarkedDefaultsFalse(t *testing.T) {
	setupTest(t)
	_, _, gameID := setupForGameCells(t)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) == 0 {
		t.Fatalf("expected seeded game cells, got 0")
	}
	for i, c := range resp {
		marked, ok := c["is_marked"].(bool)
		if !ok || marked {
			t.Errorf("cell %d: expected is_marked = false by default, got %v", i, c["is_marked"])
		}
	}
}

func TestUpdateGameCellMark_MarkTrue(t *testing.T) {
	setupTest(t)
	_, _, gameID := setupForGameCells(t)

	gameCellID := createTestGameCell(t, gameID, "", "", 0)

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
	_, _, gameID := setupForGameCells(t)

	gameCellID := createTestGameCell(t, gameID, "", "", 0)
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
	_, _, gameID := setupForGameCells(t)

	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/games/%s/cells/00000000-0000-0000-0000-000000000000", gameID),
		map[string]any{"is_marked": true},
		cookiesForGame(t, gameID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateGameCellMark_WrongGame(t *testing.T) {
	setupTest(t)
	userID, boardID, gameID := setupForGameCells(t)

	gameCellID := createTestGameCell(t, gameID, "", "", 0)

	// Create a second game (same player so cookie lookup works for both)
	game2ID := createTestGame(t, userID, boardID)

	// Try to mark cell from game1 using game2's path
	w := doRequestWithCookies(http.MethodPut, fmt.Sprintf("/api/games/%s/cells/%s", game2ID, gameCellID),
		map[string]any{"is_marked": true},
		cookiesForGame(t, game2ID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

// TestUpdateGameCellMark_HijackOtherPlayersCell verifies that a caller
// who legitimately owns their own game cannot toggle marks on someone
// else's game by smuggling that game_cell_id through their own URL.
// The service must catch the cross-game id at the membership-assert
// step before the SQL UPDATE runs.
func TestUpdateGameCellMark_HijackOtherPlayersCell(t *testing.T) {
	setupTest(t)
	_, boardID, victimGameID := setupForGameCells(t)
	victimCellID := createTestGameCell(t, victimGameID, "", "", 0)

	// Attacker creates their own game on the same board.
	attacker := createTestUser(t, "attacker_gamecell", "attacker_gc@example.com")
	attackerGameID := createTestGame(t, attacker, boardID)

	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/games/%s/cells/%s", attackerGameID, victimCellID),
		map[string]any{"is_marked": true},
		cookiesFor(attacker),
	)
	assertStatus(t, w, http.StatusNotFound)
}

// TestGetGameCellsByGameID_CellIDNullable asserts that when a board cell is
// deleted, any game_cells referencing it have their cell_id set to NULL but
// keep their snapshotted content. We pick one of the auto-seeded cells from
// the game, look up its source board cell_id, delete that board cell, and
// re-read the game cells list to verify.
func TestGetGameCellsByGameID_CellIDNullable(t *testing.T) {
	setupTest(t)
	_, boardID, gameID := setupForGameCells(t)

	// Read game cells, pick the first one's cell_id (the board cell it
	// references) and remember its content snapshot.
	listResp := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, listResp, http.StatusOK)
	var before []map[string]any
	decodeJSON(t, listResp, &before)
	if len(before) == 0 {
		t.Fatalf("expected seeded game cells, got 0")
	}
	first := before[0]
	sourceCellID, _ := first["cell_id"].(string)
	if sourceCellID == "" {
		t.Fatalf("expected cell_id on freshly created game cell, got empty")
	}
	wantContent, _ := first["content"].(string)
	gameCellID, _ := first["game_cell_id"].(string)

	// Delete the source board cell. The schema should SET NULL on
	// game_cells.cell_id.
	delResp := doRequestWithCookies(http.MethodDelete,
		fmt.Sprintf("/api/boards/%s/cells/%s", boardID, sourceCellID),
		nil,
		cookiesForBoard(t, boardID),
	)
	if delResp.Code >= 400 {
		t.Fatalf("failed to delete source cell: status=%d body=%s", delResp.Code, delResp.Body.String())
	}

	// Re-fetch and find the same game cell by id.
	w := doRequest(http.MethodGet, fmt.Sprintf("/api/games/%s/cells", gameID), nil)
	assertStatus(t, w, http.StatusOK)
	var after []map[string]any
	decodeJSON(t, w, &after)

	var found map[string]any
	for _, c := range after {
		if id, _ := c["game_cell_id"].(string); id == gameCellID {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatalf("game cell %s missing after source cell deletion", gameCellID)
	}
	if found["cell_id"] != nil {
		t.Errorf("expected cell_id = null after source cell deletion, got %v", found["cell_id"])
	}
	// Content snapshot should be preserved.
	if got, _ := found["content"].(string); got != wantContent {
		t.Errorf("expected content %q to be preserved after cell deletion, got %q", wantContent, got)
	}
}
