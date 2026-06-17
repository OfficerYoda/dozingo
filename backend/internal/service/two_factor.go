package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
)

type TwoFactor struct {
	passwords     *repository.Passwords
	recoveryCodes *repository.RecoveryCodes
	sessions      *repository.Sessions
	twoFactor     *repository.TwoFactor
	queries       *generated.Queries
	txRunner      repository.TxRunner
}

func NewTwoFactor(
	repos *repository.Repos,
	queries *generated.Queries,
	txRunner repository.TxRunner,
) *TwoFactor {
	return &TwoFactor{
		passwords:     repos.Passwords,
		recoveryCodes: repos.RecoveryCodes,
		sessions:      repos.Sessions,
		twoFactor:     repos.TwoFactor,
		queries:       queries,
		txRunner:      txRunner,
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

func (s *TwoFactor) Confirm(ctx context.Context, passcode string) ([]string, error) {
	pendingSession, err := requiresPendingSession(ctx, s.queries)
	if err != nil {
		return []string{}, err
	}

	user2fa, err := s.validateTOTP(ctx, pendingSession.UserID, pendingSession.Token, passcode)
	if err != nil {
		return []string{}, err
	}

	if user2fa.TotpVerifiedAt.Valid {
		return []string{}, fmt.Errorf("2fa already active: %w", domain.ErrConflict)
	}

	recoveryCodes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		return []string{}, fmt.Errorf("generate recovery codes: %w", err)
	}

	hashedCodes := make([]string, len(recoveryCodes))
	for i := range recoveryCodes {
		hashedCodes[i] = auth.HashToken(recoveryCodes[i])
	}

	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		_, err = r.TwoFactor.MarkVerified(ctx, user2fa.UserID)
		if err != nil {
			return fmt.Errorf("mark user 2fa verified: %w", err)
		}

		_, err = r.RecoveryCodes.Create(ctx, user2fa.UserID, hashedCodes)
		if err != nil {
			return fmt.Errorf("store recovery codes: %w", err)
		}

		return nil
	})
	if err != nil {
		return []string{}, err
	}

	return recoveryCodes, nil
}

func (s *TwoFactor) Verify(ctx context.Context, passcode string) error {
	sessionUser, err := requiresPendingSession(ctx, s.queries)
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

func (s *TwoFactor) VerifyRecoveryCode(ctx context.Context, recoveryCode string) error {
	sessionUser, err := requiresPendingSession(ctx, s.queries)
	if err != nil {
		return err
	}

	user2fa, err := s.twoFactor.GetByUserID(ctx, sessionUser.UserID)
	if err != nil {
		return fmt.Errorf("retrieve user two factor: %w", err)
	}

	if !user2fa.TotpVerifiedAt.Valid {
		return fmt.Errorf("2fa not verified: %w", domain.ErrForbidden)
	}

	err = s.consumeRecoveryCode(ctx, sessionUser.UserID, recoveryCode)
	if err != nil {
		return err
	}

	_, err = s.sessions.SetTwoFAPending(ctx, sessionUser.Token, false)
	if err != nil {
		return fmt.Errorf("clear 2fa pending status: %w", err)
	}

	return nil
}

func (s *TwoFactor) RegenerateCodes(ctx context.Context, password string, totpCode, recoveryCode *string) ([]string, error) {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return nil, err
	}

	user2fa, err := s.twoFactor.GetByUserID(ctx, sessionUser.UserID)
	if err != nil {
		return nil, fmt.Errorf("retrieve user two factor: %w", err)
	}
	if !user2fa.TotpVerifiedAt.Valid {
		return nil, fmt.Errorf("2fa not active: %w", domain.ErrForbidden)
	}

	err = s.verifyPasswordAndAuth(ctx, sessionUser.UserID, user2fa.TotpSecret, password, totpCode, recoveryCode)
	if err != nil {
		return nil, err
	}

	newCodes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		return nil, fmt.Errorf("generate recovery codes: %w", err)
	}

	hashedCodes := make([]string, len(newCodes))
	for i := range newCodes {
		hashedCodes[i] = auth.HashToken(newCodes[i])
	}

	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		err = r.RecoveryCodes.DeleteByUserID(ctx, sessionUser.UserID)
		if err != nil {
			return fmt.Errorf("delete old recovery codes: %w", err)
		}

		_, err = r.RecoveryCodes.Create(ctx, sessionUser.UserID, hashedCodes)
		if err != nil {
			return fmt.Errorf("store new recovery codes: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newCodes, nil
}

func (s *TwoFactor) Disable(ctx context.Context, password string, totpCode, recoveryCode *string) error {
	sessionUser, err := requiresSessionUser(ctx, s.queries)
	if err != nil {
		return err
	}

	user2fa, err := s.twoFactor.GetByUserID(ctx, sessionUser.UserID)
	if err != nil {
		return fmt.Errorf("retrieve user two factor: %w", err)
	}
	if !user2fa.TotpVerifiedAt.Valid {
		return fmt.Errorf("2fa not active: %w", domain.ErrForbidden)
	}

	if err := s.verifyPasswordAndAuth(ctx, sessionUser.UserID, user2fa.TotpSecret, password, totpCode, recoveryCode); err != nil {
		return err
	}

	return s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		if err := r.RecoveryCodes.DeleteByUserID(ctx, sessionUser.UserID); err != nil {
			return fmt.Errorf("delete recovery codes: %w", err)
		}

		if err := r.TwoFactor.Delete(ctx, sessionUser.UserID); err != nil {
			return fmt.Errorf("delete two factor: %w", err)
		}

		return nil
	})
}

// verifyPasswordAndAuth requires exactly one of totpCode and recoveryCode to be nil
func (s *TwoFactor) verifyPasswordAndAuth(ctx context.Context, userID pgtype.UUID, totpSecret, password string, totpCode, recoveryCode *string) error {
	if totpCode == nil && recoveryCode == nil {
		return fmt.Errorf("totp code or recovery code required: %w", domain.ErrBadInput)
	}
	if totpCode != nil && recoveryCode != nil {
		return fmt.Errorf("provide only one of totp code or recovery code: %w", domain.ErrBadInput)
	}

	passwordHash, err := s.passwords.GetHashForUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("retrieve password: %w", err)
	}
	if err := auth.CheckPassword(password, passwordHash); err != nil {
		return fmt.Errorf("invalid password: %w", domain.ErrUnauthorized)
	}

	if totpCode != nil {
		if !totp.Validate(*totpCode, totpSecret) {
			return fmt.Errorf("invalid totp code: %w", domain.ErrBadInput)
		}

		return nil
	}

	return s.consumeRecoveryCode(ctx, userID, *recoveryCode)
}

func (s *TwoFactor) consumeRecoveryCode(ctx context.Context, userID pgtype.UUID, code string) error {
	unused, err := s.recoveryCodes.GetUnusedByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("retrieve recovery codes: %w", err)
	}

	inputHash := []byte(auth.HashToken(code))
	for i := range unused {
		if subtle.ConstantTimeCompare(inputHash, []byte(unused[i].CodeHash)) == 1 {
			if _, err := s.recoveryCodes.MarkUsed(ctx, unused[i].ID); err != nil {
				return fmt.Errorf("mark recovery code used: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("invalid recovery code: %w", domain.ErrBadInput)
}

func (s *TwoFactor) validateTOTP(ctx context.Context, userID pgtype.UUID, sessionToken, passcode string) (*generated.UserTwoFactor, error) {
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
