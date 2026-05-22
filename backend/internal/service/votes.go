package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Votes struct {
	votes   *repository.Votes
	queries *generated.Queries
}

func NewVotes(votes *repository.Votes, queries *generated.Queries) *Votes {
	return &Votes{votes: votes, queries: queries}
}

type GetVotesAggregateInput struct {
	BoardID pgtype.UUID
	UserID  pgtype.UUID
}

type UpsertVoteInput struct {
	BoardID   pgtype.UUID
	VoteValue int32
}

func (s *Votes) GetAggregateByBoardID(ctx context.Context, in GetVotesAggregateInput) (generated.GetVotesByBoardIDRow, error) {
	return s.votes.GetAggregateByBoardID(ctx, repository.GetVotesAggregateInput(in))
}

func (s *Votes) Upsert(ctx context.Context, in UpsertVoteInput) (generated.Vote, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return generated.Vote{}, err
	}

	if in.VoteValue != -1 && in.VoteValue != 1 {
		return generated.Vote{}, fmt.Errorf("invalid vote_value: %w", domain.ErrUnprocessableEntity)
	}

	return s.votes.Upsert(ctx, repository.UpsertVoteInput{
		BoardID:   in.BoardID,
		VoteValue: in.VoteValue,
		UserID:    sessionUser.UserID,
	})
}

func (s *Votes) Delete(ctx context.Context, boardID pgtype.UUID) error {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return err
	}

	_, err = s.votes.Delete(ctx, repository.DeleteVoteInput{
		BoardID: boardID,
		UserID:  sessionUser.UserID,
	})
	return err
}
