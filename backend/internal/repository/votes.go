package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/pgmap"
)

type Votes struct {
	queries *generated.Queries
}

type GetVotesAggregateInput struct {
	BoardID pgtype.UUID
	UserID  pgtype.UUID
}

type UpsertVoteInput struct {
	UserID    pgtype.UUID
	BoardID   pgtype.UUID
	VoteValue int32
}

type DeleteVoteInput struct {
	UserID  pgtype.UUID
	BoardID pgtype.UUID
}

func (r *Votes) GetAggregateByBoardID(ctx context.Context, in GetVotesAggregateInput) (generated.GetVotesByBoardIDRow, error) {
	row, err := r.queries.GetVotesByBoardID(ctx, generated.GetVotesByBoardIDParams{
		BoardID: in.BoardID,
		UserID:  in.UserID,
	})
	if err != nil {
		return generated.GetVotesByBoardIDRow{}, pgmap.TranslatePgErr(err)
	}

	return row, nil
}

func (r *Votes) Upsert(ctx context.Context, in UpsertVoteInput) (generated.Vote, error) {
	vote, err := r.queries.UpsertVote(ctx, generated.UpsertVoteParams{
		UserID:    in.UserID,
		BoardID:   in.BoardID,
		VoteValue: in.VoteValue,
	})
	if err != nil {
		return generated.Vote{}, pgmap.TranslatePgErr(err)
	}

	return vote, nil
}

func (r *Votes) Delete(ctx context.Context, in DeleteVoteInput) (generated.Vote, error) {
	vote, err := r.queries.DeleteVote(ctx, generated.DeleteVoteParams{
		UserID:  in.UserID,
		BoardID: in.BoardID,
	})
	if err != nil {
		return generated.Vote{}, pgmap.TranslatePgErr(err)
	}

	return vote, nil
}

func (r *Votes) ListVotesFromUser(ctx context.Context, userID pgtype.UUID) ([]generated.ListVotesFromUserRow, error) {
	votes, err := r.queries.ListVotesFromUser(ctx, userID)
	if err != nil {
		return []generated.ListVotesFromUserRow{}, pgmap.TranslatePgErr(err)
	}

	return votes, nil
}
