package handler

import (
	"net/http"
	"testing"
	"time"
)

/// ===== GET /me/boards =====

func TestListBoardsBySession_NoCookie_401(t *testing.T) {
	setupTest(t)

	w := doRequest(http.MethodGet, "/api/me/boards", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestListBoardsBySession_AnonSession_401(t *testing.T) {
	setupTest(t)

	_, cookie := mintAnonSession(t, 30*24*time.Hour)
	w := doRequestWithCookies(http.MethodGet, "/api/me/boards", nil, []*http.Cookie{cookie})
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestListBoardsBySession_ReturnsOnlyMine(t *testing.T) {
	setupTest(t)

	alice := createTestUser(t, "alice", "alice@example.com")
	bob := createTestUser(t, "bob", "bob@example.com")

	aliceBoard1 := createTestBoard(t, "alice-1", 5, alice, nil)
	aliceBoard2 := createTestBoard(t, "alice-2", 5, alice, nil)
	bobBoard := createTestBoard(t, "bob-1", 5, bob, nil)

	w := doRequestWithCookies(http.MethodGet, "/api/me/boards", nil, cookiesFor(alice))
	assertStatus(t, w, http.StatusOK)

	var boards []map[string]any
	decodeJSON(t, w, &boards)

	if len(boards) != 2 {
		t.Fatalf("expected 2 boards for alice, got %d (%v)", len(boards), boards)
	}

	gotIDs := map[string]bool{}
	for _, b := range boards {
		id, _ := b["board_id"].(string)
		gotIDs[id] = true

		// Every returned board must be authored by the caller.
		if author, _ := b["author_id"].(string); author != alice {
			t.Errorf("board %s has author_id=%s, expected %s (caller)", id, author, alice)
		}
	}
	if !gotIDs[aliceBoard1] || !gotIDs[aliceBoard2] {
		t.Errorf("expected both alice's boards in response, got %v", gotIDs)
	}
	if gotIDs[bobBoard] {
		t.Errorf("bob's board %s leaked into alice's /me/boards response", bobBoard)
	}
}

func TestListBoardsBySession_RespectsLimit(t *testing.T) {
	setupTest(t)

	user := createTestUser(t, "limituser", "lim@example.com")
	for i := 0; i < 3; i++ {
		createTestBoard(t, "lim-board", 5, user, nil)
	}

	w := doRequestWithCookies(http.MethodGet, "/api/me/boards?limit=2", nil, cookiesFor(user))
	assertStatus(t, w, http.StatusOK)

	var boards []map[string]any
	decodeJSON(t, w, &boards)

	if len(boards) != 2 {
		t.Errorf("expected limit=2 to cap response at 2 boards, got %d", len(boards))
	}
}
