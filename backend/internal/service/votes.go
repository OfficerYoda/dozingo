package service

import (
	"context"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Votes interface {
	GetAggregateByBoardID(ctx context.Context, in repository.GetVotesAggregateInput) (generated.GetVotesByBoardIDRow, error)
	Upsert(ctx context.Context, in repository.UpsertVoteInput) (generated.Vote, error)
	Delete(ctx context.Context, in repository.DeleteVoteInput) error
}

type votes struct {
	repo *repository.Votes
}

func NewVotes(repo *repository.Votes) Votes {
	return &votes{repo: repo}
}

func (s *votes) GetAggregateByBoardID(ctx context.Context, in repository.GetVotesAggregateInput) (generated.GetVotesByBoardIDRow, error) {
	return s.repo.GetAggregateByBoardID(ctx, in)
}

func (s *votes) Upsert(ctx context.Context, in repository.UpsertVoteInput) (generated.Vote, error) {
	// TODO(authz): once handlers pass the session, verify in.UserID matches
	// the authenticated user.
	return s.repo.Upsert(ctx, in)
}

func (s *votes) Delete(ctx context.Context, in repository.DeleteVoteInput) error {
	// TODO(authz): once handlers pass the session, verify in.UserID matches
	// the authenticated user.
	_, err := s.repo.Delete(ctx, in)
	return err
}
