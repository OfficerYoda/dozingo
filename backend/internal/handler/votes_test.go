package handler

import (
	"fmt"
	"net/http"
	"testing"
)

// setupBoardForVotes creates a user and board, returning the user ID and board ID.
func setupBoardForVotes(t *testing.T) (userID, boardID string) {
	t.Helper()
	userID = createTestUser(t, "voteuser", "voteuser@example.com")
	boardID = createTestBoard(t, "Vote Board", 5, userID, nil)
	return userID, boardID
}

func TestUpsertVote_Create(t *testing.T) {
	setupTest(t)
	userID, boardID := setupBoardForVotes(t)

	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		map[string]any{"vote_value": 1},
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "user_id", userID)
	assertJSONField(t, resp, "board_id", boardID)

	if _, ok := resp["vote_id"]; !ok {
		t.Error("expected 'id' field in response")
	}
}

func TestUpsertVote_Update(t *testing.T) {
	setupTest(t)
	userID, boardID := setupBoardForVotes(t)

	// Create initial upvote
	doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		map[string]any{"vote_value": 1},
		cookiesFor(userID),
	)

	// Change to downvote
	w := doRequestWithCookies(http.MethodPut,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		map[string]any{"vote_value": -1},
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "user_id", userID)
	assertJSONField(t, resp, "board_id", boardID)
}

func TestGetVotesByBoardID(t *testing.T) {
	setupTest(t)
	userID, boardID := setupBoardForVotes(t)

	// Create an upvote
	createTestVote(t, boardID, userID, 1)

	w := doRequestWithCookies(http.MethodGet,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	score := resp["score"].(float64)
	voteCount := resp["vote_count"].(float64)
	userVote := resp["user_vote"].(float64)

	if int(score) != 1 {
		t.Errorf("expected score = 1, got %v", score)
	}
	if int(voteCount) != 1 {
		t.Errorf("expected vote_count = 1, got %v", voteCount)
	}
	if int(userVote) != 1 {
		t.Errorf("expected user_vote = 1, got %v", userVote)
	}
}

func TestGetVotesByBoardID_NoVotes(t *testing.T) {
	setupTest(t)
	userID, boardID := setupBoardForVotes(t)

	w := doRequestWithCookies(http.MethodGet,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	score := resp["score"].(float64)
	voteCount := resp["vote_count"].(float64)

	if int(score) != 0 {
		t.Errorf("expected score = 0, got %v", score)
	}
	if int(voteCount) != 0 {
		t.Errorf("expected vote_count = 0, got %v", voteCount)
	}
	if resp["user_vote"] != nil {
		t.Errorf("expected user_vote = null, got %v", resp["user_vote"])
	}
}

func TestGetVotesByBoardID_MultipleVoters(t *testing.T) {
	setupTest(t)
	user1, boardID := setupBoardForVotes(t)
	user2 := createTestUser(t, "voter2", "voter2@example.com")

	// user1 upvotes
	createTestVote(t, boardID, user1, 1)
	// user2 downvotes
	createTestVote(t, boardID, user2, -1)

	// Check aggregated results from user1's perspective
	w := doRequestWithCookies(http.MethodGet,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		nil,
		cookiesFor(user1),
	)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	score := resp["score"].(float64)
	voteCount := resp["vote_count"].(float64)
	userVote := resp["user_vote"].(float64)

	if int(score) != 0 {
		t.Errorf("expected score = 0 (1 + -1), got %v", score)
	}
	if int(voteCount) != 2 {
		t.Errorf("expected vote_count = 2, got %v", voteCount)
	}
	if int(userVote) != 1 {
		t.Errorf("expected user_vote = 1 (user1's vote), got %v", userVote)
	}
}

func TestDeleteVote(t *testing.T) {
	setupTest(t)
	userID, boardID := setupBoardForVotes(t)

	// Create a vote
	createTestVote(t, boardID, userID, 1)

	// Delete the vote
	w := doRequestWithCookies(http.MethodDelete,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusNoContent)

	// Verify vote is gone via GET
	getResp := doRequestWithCookies(http.MethodGet,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, getResp, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, getResp, &resp)

	if int(resp["vote_count"].(float64)) != 0 {
		t.Errorf("expected vote_count = 0 after delete, got %v", resp["vote_count"])
	}
	if resp["user_vote"] != nil {
		t.Errorf("expected user_vote = null after delete, got %v", resp["user_vote"])
	}
}

func TestDeleteVote_NotFound(t *testing.T) {
	setupTest(t)
	userID, boardID := setupBoardForVotes(t)

	// Try to delete a vote that doesn't exist
	w := doRequestWithCookies(http.MethodDelete,
		fmt.Sprintf("/api/boards/%s/vote", boardID),
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetVotesByBoardID_InvalidBoardID(t *testing.T) {
	setupTest(t)
	userID := createTestUser(t, "invalidvoteuser", "invalidvote@example.com")

	w := doRequestWithCookies(http.MethodGet,
		"/api/boards/not-a-uuid/vote",
		nil,
		cookiesFor(userID),
	)
	assertStatus(t, w, http.StatusUnprocessableEntity)
}
