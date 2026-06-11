package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
)

type Votes struct {
	votes   *repository.Votes
	queries *generated.Queries
}

func NewVotes(repos *repository.Repos, queries *generated.Queries) *Votes {
	return &Votes{votes: repos.Votes, queries: queries}
}

type GetVotesAggregateInput struct {
	BoardID pgtype.UUID
	UserID  pgtype.UUID
}

type UpsertVoteInput struct {
	BoardID   pgtype.UUID
	VoteValue int32
}

func (s *Votes) GetAggregateByBoardID(ctx context.Context, boardID pgtype.UUID) (generated.GetVotesByBoardIDRow, error) {
	sessionUser, _ := middleware.SessionUserFromContext(ctx)

	return s.votes.GetAggregateByBoardID(ctx, repository.GetVotesAggregateInput{
		BoardID: boardID,
		UserID:  sessionUser.UserID,
	})
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

func (s *Votes) ListVotesFromUser(ctx context.Context, userID pgtype.UUID) ([]generated.ListVotesFromUserRow, error) {
	votes, err := s.votes.ListVotesFromUser(ctx, userID)
	if err != nil {
		return []generated.ListVotesFromUserRow{}, fmt.Errorf("list user votes: %w", err)
	}

	return votes, nil
}

func (s *Votes) ListVotesFromMe(ctx context.Context) ([]generated.ListVotesFromUserRow, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return []generated.ListVotesFromUserRow{}, err
	}

	return s.ListVotesFromUser(ctx, sessionUser.UserID)
}
