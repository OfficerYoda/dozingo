package handler

import (
	"fmt"
	"net/http"
	"testing"
)

// setupBoardForCells creates a user and board, returning the board ID.
func setupBoardForCells(t *testing.T) string {
	t.Helper()
	userID := createTestUser(t, "cellauthor", "cellauthor@example.com")
	return createTestBoard(t, "Cell Board", 5, userID, nil)
}

func TestCreateCell(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	w := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "Free Space",
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "content", "Free Space")
	assertJSONField(t, resp, "board_id", boardID)

	if _, ok := resp["cell_id"]; !ok {
		t.Error("expected 'id' field in response")
	}
}

func TestCreateCell_DefaultValue(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	// Create cell without specifying value — should default to 1
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "Default Value Cell",
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "content", "Default Value Cell")
	// Value should be 1 (default)
	if val, ok := resp["value"].(float64); !ok || int(val) != 1 {
		t.Errorf("expected value = 1, got %v", resp["value"])
	}
}

func TestCreateCell_ExplicitValue(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	// Create cell with explicit value
	w := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "High Value Cell",
		"value":   3,
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "content", "High Value Cell")
	if val, ok := resp["value"].(float64); !ok || int(val) != 3 {
		t.Errorf("expected value = 3, got %v", resp["value"])
	}
}

func TestGetCellsByBoardID(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	// Create two cells
	doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "Cell A",
	})
	doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "Cell B",
	})

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/cells", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 cells, got %d", len(resp))
	}
}

func TestGetCellsByBoardID_Empty(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/cells", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 cells, got %d", len(resp))
	}
}

func TestGetCellsByBoardID_InvalidBoardID(t *testing.T) {
	w := doRequest(http.MethodGet, "/api/boards/not-a-uuid/cells", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateCell(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	// Create a cell
	createResp := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "Original",
	})
	assertStatus(t, createResp, http.StatusOK)

	var created map[string]any
	decodeJSON(t, createResp, &created)
	cellID := created["cell_id"].(string)

	// Update the cell
	w := doRequest(http.MethodPut, fmt.Sprintf("/api/boards/%s/cells/%s", boardID, cellID), map[string]any{
		"content": "Updated",
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "cell_id", cellID)
	assertJSONField(t, resp, "content", "Updated")
	assertJSONField(t, resp, "board_id", boardID)
}

func TestUpdateCell_Value(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	// Create a cell with value 1
	createResp := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "Cell",
		"value":   1,
	})
	assertStatus(t, createResp, http.StatusOK)

	var created map[string]any
	decodeJSON(t, createResp, &created)
	cellID := created["cell_id"].(string)

	// Update the cell value
	w := doRequest(http.MethodPut, fmt.Sprintf("/api/boards/%s/cells/%s", boardID, cellID), map[string]any{
		"value": 5,
	})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	if val, ok := resp["value"].(float64); !ok || int(val) != 5 {
		t.Errorf("expected value = 5, got %v", resp["value"])
	}
}

func TestUpdateCell_NotFound(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	w := doRequest(http.MethodPut, fmt.Sprintf("/api/boards/%s/cells/00000000-0000-0000-0000-000000000000", boardID), map[string]any{
		"content": "Nope",
	})
	assertStatus(t, w, http.StatusNotFound)
}

func TestDeleteCell(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	// Create a cell
	createResp := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", boardID), map[string]any{
		"content": "Delete Me",
	})
	assertStatus(t, createResp, http.StatusOK)

	var created map[string]any
	decodeJSON(t, createResp, &created)
	cellID := created["cell_id"].(string)

	// Delete the cell
	w := doRequest(http.MethodDelete, fmt.Sprintf("/api/boards/%s/cells/%s", boardID, cellID), nil)
	assertStatus(t, w, http.StatusNoContent)

	// Verify cell is gone - board should have 0 cells
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s/cells", boardID), nil)
	assertStatus(t, getResp, http.StatusOK)

	var cells []map[string]any
	decodeJSON(t, getResp, &cells)

	if len(cells) != 0 {
		t.Errorf("expected 0 cells after delete, got %d", len(cells))
	}
}

func TestDeleteCell_NotFound(t *testing.T) {
	setupTest(t)
	boardID := setupBoardForCells(t)

	w := doRequest(http.MethodDelete, fmt.Sprintf("/api/boards/%s/cells/00000000-0000-0000-0000-000000000000", boardID), nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateCell_WrongBoard(t *testing.T) {
	setupTest(t)

	// Create two boards with cells
	userID := createTestUser(t, "wrongboardauthor", "wrongboard@example.com")
	board1 := createTestBoard(t, "Board 1", 5, userID, nil)
	board2 := createTestBoard(t, "Board 2", 5, userID, nil)

	// Create cell on board1
	createResp := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", board1), map[string]any{
		"content": "Board1 Cell",
	})
	assertStatus(t, createResp, http.StatusOK)
	var created map[string]any
	decodeJSON(t, createResp, &created)
	cellID := created["cell_id"].(string)

	// Try to update cell using board2's path -- should fail
	w := doRequest(http.MethodPut, fmt.Sprintf("/api/boards/%s/cells/%s", board2, cellID), map[string]any{
		"content": "Moved",
	})
	assertStatus(t, w, http.StatusNotFound)
}

func TestDeleteCell_WrongBoard(t *testing.T) {
	setupTest(t)

	userID := createTestUser(t, "wrongdelauthor", "wrongdel@example.com")
	board1 := createTestBoard(t, "Board 1", 5, userID, nil)
	board2 := createTestBoard(t, "Board 2", 5, userID, nil)

	// Create cell on board1
	createResp := doRequest(http.MethodPost, fmt.Sprintf("/api/boards/%s/cells", board1), map[string]any{
		"content": "Board1 Cell",
	})
	assertStatus(t, createResp, http.StatusOK)
	var created map[string]any
	decodeJSON(t, createResp, &created)
	cellID := created["cell_id"].(string)

	// Try to delete cell using board2's path -- should fail
	w := doRequest(http.MethodDelete, fmt.Sprintf("/api/boards/%s/cells/%s", board2, cellID), nil)
	assertStatus(t, w, http.StatusNotFound)
}
