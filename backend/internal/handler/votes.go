package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/service"
	"github.com/officeryoda/dozingo/internal/types"
)

// ===== Input/Output types =====

type voteOutput struct {
	VoteID    string `json:"vote_id" format:"uuid"`
	UserID    string `json:"user_id" format:"uuid"`
	BoardID   string `json:"board_id" format:"uuid"`
	VoteValue int32  `json:"vote_value"`
}

type getVotesByBoardIDInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

type getVotesByBoardIDBody struct {
	Score     int32  `json:"score"`
	VoteCount int32  `json:"vote_count"`
	UserVote  *int32 `json:"user_vote"`
}

type getVotesByBoardIDOutput struct {
	Body getVotesByBoardIDBody
}

type upsertVoteInputBody struct {
	VoteValue int32 `json:"vote_value" required:"true" minimum:"-1" maximum:"1"`
}

type upsertVoteInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
	Body    upsertVoteInputBody
}

type upsertVoteOutput struct {
	Body voteOutput
}

type deleteVoteInput struct {
	BoardID types.UUIDParam `path:"board_id" format:"uuid"`
}

// ===== Handler =====

type VotesHandler struct {
	svc *service.Votes
}

func NewVotesHandler(svc *service.Votes) *VotesHandler {
	return &VotesHandler{svc: svc}
}

func (h *VotesHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-votes-by-board-id",
		Method:      http.MethodGet,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Get all votes for a Board",
		Tags:        []string{"Votes"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.ReadLimiter)},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "upsert-vote",
		Method:      http.MethodPut,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Upsert a vote",
		Description: "Update or Create a vote",
		Tags:        []string{"Votes"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.upsert)

	huma.Register(api, huma.Operation{
		OperationID: "delete-vote",
		Method:      http.MethodDelete,
		Path:        "/boards/{board_id}/vote",
		Summary:     "Delete a vote",
		Tags:        []string{"Votes"},
		Middlewares: huma.Middlewares{middleware.RateLimit(api, middleware.WriteLimiter)},
	}, h.delete)
}

func (h *VotesHandler) get(ctx context.Context, in *getVotesByBoardIDInput) (*getVotesByBoardIDOutput, error) {
	sessionUser, _ := middleware.SessionUserFromContext(ctx)

	votes, err := h.svc.GetAggregateByBoardID(ctx, service.GetVotesAggregateInput{
		BoardID: in.BoardID.Value,
		UserID:  sessionUser.UserID,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to get votes")
	}

	out := &getVotesByBoardIDOutput{}
	out.Body.Score = votes.Score
	out.Body.VoteCount = votes.VoteCount
	var userVotePtr *int32
	if votes.UserVote != 0 {
		userVotePtr = &votes.UserVote
	}
	out.Body.UserVote = userVotePtr

	return out, nil
}

func (h *VotesHandler) upsert(ctx context.Context, in *upsertVoteInput) (*upsertVoteOutput, error) {
	vote, err := h.svc.Upsert(ctx, service.UpsertVoteInput{
		BoardID:   in.BoardID.Value,
		VoteValue: in.Body.VoteValue,
	})
	if err != nil {
		return nil, toHumaErr(err, "", "failed to upsert vote")
	}

	return &upsertVoteOutput{Body: voteToOutput(vote)}, nil
}

func (h *VotesHandler) delete(ctx context.Context, in *deleteVoteInput) (*struct{}, error) {
	err := h.svc.Delete(ctx, in.BoardID.Value)
	if err != nil {
		return nil, toHumaErr(err, "vote not found on this board", "failed to delete vote")
	}

	return &struct{}{}, nil
}

func voteToOutput(vote generated.Vote) voteOutput {
	return voteOutput{
		VoteID:    vote.ID.String(),
		UserID:    vote.UserID.String(),
		BoardID:   vote.BoardID.String(),
		VoteValue: vote.VoteValue,
	}
}
