package service

import (
	"context"
	"fmt"

	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Votes struct {
	votes *repository.Votes
}

func NewVotes(votes *repository.Votes) *Votes {
	return &Votes{votes: votes}
}

func (s *Votes) GetAggregateByBoardID(ctx context.Context, in repository.GetVotesAggregateInput) (generated.GetVotesByBoardIDRow, error) {
	return s.votes.GetAggregateByBoardID(ctx, in)
}

func (s *Votes) Upsert(ctx context.Context, in repository.UpsertVoteInput) (generated.Vote, error) {
	if !in.UserID.Valid {
		return generated.Vote{}, fmt.Errorf("anonymous users can't vote: %w", domain.ErrUnauthorized)
	}

	if in.VoteValue != -1 && in.VoteValue != 1 {
		return generated.Vote{}, fmt.Errorf("invalid vote_value: %w", domain.ErrUnprocessableEntity)
	}
	// TODO(authz): once handlers pass the session, verify in.UserID matches
	// the authenticated user.
	return s.votes.Upsert(ctx, in)
}

func (s *Votes) Delete(ctx context.Context, in repository.DeleteVoteInput) error {
	if !in.UserID.Valid {
		return fmt.Errorf("anonymous users can't vote: %w", domain.ErrUnauthorized)
	}

	// TODO(authz): once handlers pass the session, verify in.UserID matches
	// the authenticated user.
	_, err := s.votes.Delete(ctx, in)
	return err
}
