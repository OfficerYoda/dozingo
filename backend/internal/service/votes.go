package service

import (
	"context"

	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Votes struct {
	repo *repository.Votes
}

func NewVotes(repo *repository.Votes) *Votes {
	return &Votes{repo: repo}
}

func (s *Votes) GetAggregateByBoardID(ctx context.Context, in repository.GetVotesAggregateInput) (generated.GetVotesByBoardIDRow, error) {
	return s.repo.GetAggregateByBoardID(ctx, in)
}

func (s *Votes) Upsert(ctx context.Context, in repository.UpsertVoteInput) (generated.Vote, error) {
	if in.VoteValue != -1 && in.VoteValue != 1 {
		return generated.Vote{}, domain.ErrUnprocessableEntity
	}
	// TODO(authz): once handlers pass the session, verify in.UserID matches
	// the authenticated user.
	return s.repo.Upsert(ctx, in)
}

func (s *Votes) Delete(ctx context.Context, in repository.DeleteVoteInput) error {
	// TODO(authz): once handlers pass the session, verify in.UserID matches
	// the authenticated user.
	_, err := s.repo.Delete(ctx, in)
	return err
}
