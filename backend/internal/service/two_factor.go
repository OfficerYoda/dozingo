package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TwoFactor struct {
	sessions  *repository.Sessions
	twoFactor *repository.TwoFactor
	queries   *generated.Queries
	txRunner  repository.TxRunner
}

func NewTwoFactor(
	repos *repository.Repos,
	queries *generated.Queries,
	txRunner repository.TxRunner,
) *TwoFactor {
	return &TwoFactor{
		sessions:  repos.Sessions,
		twoFactor: repos.TwoFactor,
		queries:   queries,
		txRunner:  txRunner,
	}
}

func (s *TwoFactor) Setup(ctx context.Context) (*otp.Key, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return nil, err
	}

	// check if already set up
	existing, err := s.twoFactor.GetByUserID(ctx, sessionUser.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("checking existing 2fa: %w", err)
	}
	if existing.TotpVerifiedAt.Valid {
		return nil, fmt.Errorf("2fa already active: %w", domain.ErrConflict)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Dozingo",
		AccountName: sessionUser.Email.String,
	})
	if err != nil {
		return nil, fmt.Errorf("generating otp key: %w", err)
	}

	_, err = s.twoFactor.Upsert(ctx, sessionUser.UserID, key.Secret())
	if err != nil {
		return nil, fmt.Errorf("store otp key: %w", err)
	}

	_, err = s.sessions.SetTwoFAPending(ctx, sessionUser.Token, true)
	if err != nil {
		return nil, fmt.Errorf("mark session pending 2fa: %w", err)
	}

	return key, nil
}

func (s *TwoFactor) Confirm(ctx context.Context, passcode string) error {
	pendingSession, err := requiresPendingSession(ctx, s.queries)
	if err != nil {
		return err
	}

	user2fa, err := s.validateTOTP(ctx, pendingSession.UserID, pendingSession.Token, passcode)
	if err != nil {
		return err
	}

	if user2fa.TotpVerifiedAt.Valid {
		return fmt.Errorf("2fa already active: %w", domain.ErrConflict)
	}

	_, err = s.twoFactor.MarkVerified(ctx, user2fa.UserID)
	if err != nil {
		return fmt.Errorf("mark user 2fa verified: %w", err)
	}

	// TODO: return recovery codes
	return nil
}

func (s *TwoFactor) Verify(ctx context.Context, passcode string) error {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return err
	}

	user2fa, err := s.validateTOTP(ctx, sessionUser.UserID, sessionUser.Token, passcode)
	if err != nil {
		return err
	}

	if !user2fa.TotpVerifiedAt.Valid {
		return fmt.Errorf("2fa not verified: %w", domain.ErrForbidden)
	}

	return nil
}

func (s *TwoFactor) validateTOTP(ctx context.Context, userID pgtype.UUID, sessionToken string, passcode string) (*generated.UserTwoFactor, error) {
	user2fa, err := s.twoFactor.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("retrieve user two factor: %w", err)
	}

	valid := totp.Validate(passcode, user2fa.TotpSecret)
	if !valid {
		return nil, fmt.Errorf("invalid code: %w", domain.ErrBadInput)
	}

	_, err = s.sessions.SetTwoFAPending(ctx, sessionToken, false)
	if err != nil {
		return nil, fmt.Errorf("clear 2fa pending status: %w", err)
	}

	return &user2fa, nil
}

func requiresPendingSession(ctx context.Context, queries *generated.Queries) (generated.GetSessionUserByTokenRow, error) {
	sessionUser, err := middleware.RequireSession(ctx, queries)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("session required: %w", err)
	}

	if !sessionUser.UserID.Valid {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("requires authenticated user: %w", domain.ErrUnauthorized)
	}

	if !sessionUser.TwoFaPending {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("no 2fa is pending: %w", domain.ErrUnauthorized)
	}

	return sessionUser, nil
}
