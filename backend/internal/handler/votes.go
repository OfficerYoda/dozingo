package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/types"
)

/// ===== Input/Output types =====

type VoteOutput struct {
	VoteID    string `json:"vote_id" format:"uuid"`
	UserID    string `json:"user_id" format:"uuid"`
	BoardID   string `json:"board_id" format:"uuid"`
	VoteValue int32  `json:"vote_value" format:"integer"`
}

type GetVotesByBoardIDInput struct {
	UserID  types.UUIDParam `query:"user_id" format:"uuid"` // TODO eventually replace this when user auth is working
	BoardID types.UUIDParam `path:"board_id"`
}

type GetVotesByBoardIDOutput struct {
	Body struct {
		Score     int32  `json:"score"`
		VoteCount int32  `json:"vote_count"`
		UserVote  *int32 `json:"user_vote"`
	}
}

type UpsertVoteInput struct {
	UserID  types.UUIDParam `query:"user_id" format:"uuid" required:"true"` // TODO eventually replace this when user auth is working
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	Body    struct {
		VoteValue int32 `json:"vote_value" format:"integer" required:"true" minimum:"-1" maximum:"1"`
	}
}

type UpsertVoteOutput struct {
	Body VoteOutput
}

type DeleteVoteInput struct {
	UserID  types.UUIDParam `query:"user_id" format:"uuid" required:"true"` // TODO eventually replace this when user auth is working
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

/// ===== Register =====

func RegisterVotes(api huma.API, pool *pgxpool.Pool) {
	queries := generated.New(pool)

	huma.Register(api, huma.Operation{
		OperationID: "get-votes-by-board-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Get all votes for a Board",
		Tags:        []string{"Votes"},
	}, func(ctx context.Context, input *GetVotesByBoardIDInput) (*GetVotesByBoardIDOutput, error) {
		return getVotesByBoardID(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "upsert-vote",
		Method:      http.MethodPut,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Upsert a vote",
		Description: "Update or Create a vote",
		Tags:        []string{"Votes"},
	}, func(ctx context.Context, input *UpsertVoteInput) (*UpsertVoteOutput, error) {
		return upsertVote(ctx, queries, *input)
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-vote",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Delete a vote",
		Tags:        []string{"Votes"},
	}, func(ctx context.Context, input *DeleteVoteInput) (*struct{}, error) {
		return deleteVote(ctx, queries, *input)
	})
}

/// ===== Handlers =====

func getVotesByBoardID(ctx context.Context, queries *generated.Queries, input GetVotesByBoardIDInput) (*GetVotesByBoardIDOutput, error) {
	votes, err := queries.GetVotesByBoardID(ctx, generated.GetVotesByBoardIDParams{
		UserID:  input.UserID.Value,
		BoardID: input.BoardID.Value,
	})
	if err != nil {
		return nil, internalError(err, "failed to get votes")
	}

	out := &GetVotesByBoardIDOutput{}
	out.Body.Score = votes.Score
	out.Body.VoteCount = votes.VoteCount
	// only return a user vote value when the user actually voted
	var userVotePtr *int32
	if votes.UserVote != 0 {
		userVotePtr = &votes.UserVote
	}
	out.Body.UserVote = userVotePtr

	return out, nil
}

func upsertVote(ctx context.Context, queries *generated.Queries, input UpsertVoteInput) (*UpsertVoteOutput, error) {
	board, err := queries.UpsertVote(ctx, generated.UpsertVoteParams{
		UserID:    input.UserID.Value,
		BoardID:   input.BoardID.Value,
		VoteValue: input.Body.VoteValue,
	})
	if err != nil {
		return nil, internalError(err, "failed to upsert vote")
	}

	return &UpsertVoteOutput{Body: voteToOutput(board)}, nil
}

func deleteVote(ctx context.Context, queries *generated.Queries, input DeleteVoteInput) (*struct{}, error) {
	_, err := queries.DeleteVote(ctx, generated.DeleteVoteParams{
		UserID:  input.UserID.Value,
		BoardID: input.BoardID.Value,
	})
	if err != nil {
		return nil, notFoundOr500(err, "vote not found on this board", "failed to delete vote")
	}

	return &struct{}{}, nil
}

/// ===== Helper =====

func voteToOutput(vote generated.Vote) VoteOutput {
	return VoteOutput{
		VoteID:    vote.ID.String(),
		UserID:    vote.UserID.String(),
		BoardID:   vote.BoardID.String(),
		VoteValue: vote.VoteValue,
	}
}
