package handler

import (
	"context"
	"net/http"
	"testing"
)

// fetchRecentStats GETs /api/stats/recent with the given query string (e.g.
// "?duration=1h"), asserts a 200 OK, and returns the decoded body. Pass an
// empty string to omit the query.
func fetchRecentStats(t *testing.T, query string) map[string]any {
	t.Helper()
	w := doRequest(http.MethodGet, "/api/stats/recent"+query, nil)
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp
}

func TestGetRecentStats_NoData(t *testing.T) {
	setupTest(t)

	resp := fetchRecentStats(t, "?duration=1h")
	assertJSONInt(t, resp, "bingos", 0)
	assertJSONInt(t, resp, "games", 0)
	assertJSONInt(t, resp, "boards", 0)
	assertJSONInt(t, resp, "cells", 0)
}

// TestGetRecentStats_DefaultDuration verifies that omitting the duration query
// param results in a zero time.Duration on the server. The SQL window then
// becomes (now - 0s, now], so freshly-created rows fall outside it and every
// counter reads zero.
func TestGetRecentStats_DefaultDuration(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "statsuser_default", "statsuser_default@example.com")
	boardID := createTestBoard(t, "Default Board", 4, userID, nil)
	createTestGame(t, userID, boardID)

	resp := fetchRecentStats(t, "")
	assertJSONInt(t, resp, "bingos", 0)
	assertJSONInt(t, resp, "games", 0)
	assertJSONInt(t, resp, "boards", 0)
	assertJSONInt(t, resp, "cells", 0)
}

// TestGetRecentStats_InvalidDuration documents the current behavior: the
// DurationParam.OnParamSet hook swallows time.ParseDuration errors and leaves
// Value as the zero duration. The handler accepts the request and returns
// zero counts (the SQL window collapses to (now, now]). This may not be ideal
// long-term but it is the contract today; pin it down so any change is
// deliberate.
func TestGetRecentStats_InvalidDuration(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "statsuser_invalid", "statsuser_invalid@example.com")
	createTestBoard(t, "Invalid Board", 4, userID, nil)

	resp := fetchRecentStats(t, "?duration=notvalid")
	assertJSONInt(t, resp, "bingos", 0)
	assertJSONInt(t, resp, "games", 0)
	assertJSONInt(t, resp, "boards", 0)
	assertJSONInt(t, resp, "cells", 0)
}

func TestGetRecentStats_CountsBoards(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "statsuser_boards", "statsuser_boards@example.com")
	createTestBoard(t, "Board A", 4, userID, nil)
	createTestBoard(t, "Board B", 4, userID, nil)

	resp := fetchRecentStats(t, "?duration=1h")
	assertJSONInt(t, resp, "boards", 2)
}

func TestGetRecentStats_CountsCells(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "statsuser_cells", "statsuser_cells@example.com")
	boardID := createTestBoard(t, "Cells Board", 4, userID, nil)
	// Four explicit cells via the API; createTestGame would seed more.
	createTestCell(t, boardID, "c1")
	createTestCell(t, boardID, "c2")
	createTestCell(t, boardID, "c3")
	createTestCell(t, boardID, "c4")

	resp := fetchRecentStats(t, "?duration=1h")
	assertJSONInt(t, resp, "cells", 4)
}

func TestGetRecentStats_CountsGames(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "statsuser_games", "statsuser_games@example.com")
	boardID := createTestBoard(t, "Games Board", 4, userID, nil)
	createTestGame(t, userID, boardID)
	createTestGame(t, userID, boardID)

	resp := fetchRecentStats(t, "?duration=1h")
	assertJSONInt(t, resp, "games", 2)
}

func TestGetRecentStats_CountsBingos(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "statsuser_bingos", "statsuser_bingos@example.com")
	boardID := createTestBoard(t, "Bingos Board", 4, userID, nil)
	gameID := createTestGame(t, userID, boardID)

	// Mark the game as completed so the bingos filter picks it up. The
	// API doesn't expose a "complete the game" endpoint that bypasses the
	// bingo-detection logic, so we flip the row directly.
	_, err := testPool.Exec(context.Background(),
		`UPDATE games SET status = 'completed', updated_at = now() WHERE id = $1`,
		gameID)
	if err != nil {
		t.Fatalf("failed to mark game completed: %v", err)
	}

	resp := fetchRecentStats(t, "?duration=1h")
	assertJSONInt(t, resp, "bingos", 1)
	assertJSONInt(t, resp, "games", 1)
}

// TestGetRecentStats_ExcludesOldData backdates every row's timestamps so they
// fall outside a 1h window, then asserts each counter reads zero. This is the
// closest we can get to validating the SQL window without time-traveling the
// clock.
//
// The set_updated_at trigger on boards/cells/games would otherwise reset
// updated_at to now() on every UPDATE we issue, defeating the test. We
// disable it for the duration of the backdate (replica role skips
// non-replica triggers).
func TestGetRecentStats_ExcludesOldData(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "statsuser_old", "statsuser_old@example.com")
	boardID := createTestBoard(t, "Old Board", 4, userID, nil)
	gameID := createTestGame(t, userID, boardID)

	ctx := context.Background()

	// Mark the game completed BEFORE backdating; the trigger will bump
	// updated_at to now() but the next stage backdates it again.
	if _, err := testPool.Exec(ctx,
		`UPDATE games SET status = 'completed' WHERE id = $1`, gameID); err != nil {
		t.Fatalf("failed to mark game completed: %v", err)
	}

	// Suppress the set_updated_at triggers so the backdate sticks.
	if _, err := testPool.Exec(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatalf("failed to enter replica role: %v", err)
	}
	defer func() {
		_, _ = testPool.Exec(ctx, `SET session_replication_role = origin`)
	}()

	stmts := []string{
		`UPDATE boards SET created_at = now() - interval '2 hours', updated_at = now() - interval '2 hours'`,
		`UPDATE cells  SET created_at = now() - interval '2 hours', updated_at = now() - interval '2 hours'`,
		`UPDATE games  SET created_at = now() - interval '2 hours', updated_at = now() - interval '2 hours'`,
	}
	for _, s := range stmts {
		if _, err := testPool.Exec(ctx, s); err != nil {
			t.Fatalf("failed to backdate rows (%q): %v", s, err)
		}
	}

	resp := fetchRecentStats(t, "?duration=1h")
	assertJSONInt(t, resp, "bingos", 0)
	assertJSONInt(t, resp, "games", 0)
	assertJSONInt(t, resp, "boards", 0)
	assertJSONInt(t, resp, "cells", 0)

	// Sanity check: a wide window should pick them all up again, proving
	// the rows still exist and only the duration filter excluded them.
	wide := fetchRecentStats(t, "?duration=24h")
	assertJSONInt(t, wide, "bingos", 1)
	assertJSONInt(t, wide, "games", 1)
	assertJSONInt(t, wide, "boards", 1)
	// 4x4 board => 16 cells were seeded for game creation.
	if cells, ok := wide["cells"].(float64); !ok || int64(cells) <= 0 {
		t.Errorf("expected cells > 0 with a 24h window, got %v", wide["cells"])
	}
}
