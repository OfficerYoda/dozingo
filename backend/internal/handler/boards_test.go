package handler

import (
	"fmt"
	"net/http"
	"testing"
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
