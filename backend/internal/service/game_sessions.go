package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

type GameSessions struct {
	gameSessions *repository.GameSessions
	games        *repository.Games
	queries      *generated.Queries
}

func NewGameSessions(repos *repository.Repos, queries *generated.Queries) *GameSessions {
	return &GameSessions{
		gameSessions: repos.GameSessions,
		games:        repos.Games,
		queries:      queries,
	}
}

func (s *GameSessions) Heartbeat(ctx context.Context, gameID pgtype.UUID) (generated.GameSession, error) {
	_, err := checkIfCallerOwnsGame(ctx, s.games, s.queries, gameID)
	if err != nil {
		return generated.GameSession{}, err
	}

	session, found, err := s.gameSessions.UpdateHeartbeat(ctx, gameID)
	if err != nil {
		return generated.GameSession{}, err
	}
	if found {
		return session, nil
	}

	return s.gameSessions.Create(ctx, gameID)
}

// EndSessions closes any open session for the given game. Does NOT enforce
// ownership; intended to be called from internal lifecycle hooks (e.g.
// game status change). Use the heartbeat endpoint for player-driven flows.
func (s *GameSessions) EndSessions(ctx context.Context, gameID pgtype.UUID) error {
	_, err := s.gameSessions.EndSessions(ctx, gameID)

	return err
}

func (s *GameSessions) CloseStaleSessions(ctx context.Context, timeout time.Duration) (int64, error) {
	return s.gameSessions.CloseStaleSessions(ctx, timeout)
}

func (s *GameSessions) GetPlaytimeByGame(ctx context.Context, gameID pgtype.UUID) (int64, error) {
	_, err := s.games.Get(ctx, gameID)
	if err != nil {
		return 0, err
	}

	return s.gameSessions.GetPlaytimeByGame(ctx, gameID)
}

func (s *GameSessions) GetPlaytimeByBoard(ctx context.Context, boardID pgtype.UUID) (int64, error) {
	return s.gameSessions.GetPlaytimeByBoard(ctx, boardID)
}

func (s *GameSessions) GetPlaytimeByPlayer(ctx context.Context, playerID pgtype.UUID) (int64, error) {
	return s.gameSessions.GetPlaytimeByPlayer(ctx, playerID)
}

func (s *GameSessions) GetPlaytimeForCurrentUser(ctx context.Context) (int64, error) {
	sessionUser, err := requiresVerifiedSession(ctx, s.queries)
	if err != nil {
		return 0, err
	}

	return s.gameSessions.GetPlaytimeByPlayer(ctx, sessionUser.UserID)
}

func (s *GameSessions) GetTotalPlaytime(ctx context.Context, period time.Duration) (int64, error) {
	return s.gameSessions.GetTotalPlaytime(ctx, period)
}
