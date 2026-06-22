package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/domain"
	"github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/repository"
)

type TwoFactor struct {
	passwords     *repository.Passwords
	recoveryCodes *repository.RecoveryCodes
	sessions      *repository.Sessions
	users         *repository.Users
	twoFactor     *repository.TwoFactor
	emailSender   email.Sender
	cipher        *auth.TOTPCipher
	queries       *generated.Queries
	txRunner      repository.TxRunner
}

func NewTwoFactor(
	repos *repository.Repos,
	queries *generated.Queries,
	emailSender email.Sender,
	txRunner repository.TxRunner,
	cipher *auth.TOTPCipher,
) *TwoFactor {
	return &TwoFactor{
		passwords:     repos.Passwords,
		recoveryCodes: repos.RecoveryCodes,
		sessions:      repos.Sessions,
		users:         repos.Users,
		twoFactor:     repos.TwoFactor,
		emailSender:   emailSender,
		cipher:        cipher,
		queries:       queries,
		txRunner:      txRunner,
	}
}

func (s *TwoFactor) Setup(ctx context.Context) (*otp.Key, error) {
	sessionUser, err := requiresAuthenticatedSession(ctx, s.queries)
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
		return nil, fmt.Errorf("generate otp key: %w", err)
	}

	encryptedSecret, err := s.cipher.Seal(sessionUser.UserID.Bytes, key.Secret())
	if err != nil {
		return nil, fmt.Errorf("encrypt totp secret: %w", err)
	}

	_, err = s.twoFactor.Upsert(ctx, sessionUser.UserID, encryptedSecret)
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

	user2fa, err := s.validateTOTP(ctx, pendingSession.UserID, passcode)
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

	hashedCodes := hashCodes(recoveryCodes)

	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		txErr := r.TwoFactor.SetLastUsedCode(ctx, user2fa.UserID, passcode)
		if txErr != nil {
			return fmt.Errorf("store last used code: %w", err)
		}

		_, txErr = r.TwoFactor.MarkVerified(ctx, user2fa.UserID)
		if txErr != nil {
			return fmt.Errorf("mark user 2fa verified: %w", err)
		}

		_, txErr = r.RecoveryCodes.Create(ctx, user2fa.UserID, hashedCodes)
		if txErr != nil {
			return fmt.Errorf("store recovery codes: %w", err)
		}

		_, txErr = r.Sessions.SetTwoFAPending(ctx, pendingSession.Token, false)
		if txErr != nil {
			return fmt.Errorf("clear 2fa pending status: %w", err)
		}

		txErr = r.Sessions.DeleteOtherSessionsFromUser(ctx, pendingSession.UserID, pendingSession.SessionID)
		if txErr != nil {
			return fmt.Errorf("invalidate other sessions: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return []string{}, err
	}

	user, err := s.users.GetByID(ctx, user2fa.UserID)
	if err != nil {
		return []string{}, fmt.Errorf("get user: %w", err)
	}

	err = s.emailSender.Send2FAActivated(user.Email, time.Now())
	if err != nil {
		return []string{}, fmt.Errorf("send email: %w", err)
	}

	return recoveryCodes, nil
}

func (s *TwoFactor) Verify(ctx context.Context, passcode string) error {
	sessionUser, err := requiresPendingSession(ctx, s.queries)
	if err != nil {
		return err
	}

	user2fa, err := s.validateTOTP(ctx, sessionUser.UserID, passcode)
	if err != nil {
		return err
	}

	if !user2fa.TotpVerifiedAt.Valid {
		return fmt.Errorf("2fa not verified: %w", domain.ErrForbidden)
	}

	err = s.twoFactor.SetLastUsedCode(ctx, sessionUser.UserID, passcode)
	if err != nil {
		return fmt.Errorf("store last used code: %w", err)
	}

	_, err = s.sessions.SetTwoFAPending(ctx, sessionUser.Token, false)
	if err != nil {
		return fmt.Errorf("clear 2fa pending status: %w", err)
	}

	err = s.emailSender.SendLoginNotification(sessionUser.Email.String, time.Now())
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
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
	sessionUser, err := requiresVerifiedSession(ctx, s.queries)
	if err != nil {
		return nil, err
	}

	user2fa, err := s.twoFactor.GetByUserID(ctx, sessionUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("2fa not active: %w", domain.ErrForbidden)
		}

		return nil, fmt.Errorf("retrieve user two factor: %w", err)
	}
	if !user2fa.TotpVerifiedAt.Valid {
		return nil, fmt.Errorf("2fa not active: %w", domain.ErrForbidden)
	}

	err = s.verifyPasswordAndAuth(ctx, sessionUser.UserID, user2fa, password, totpCode, recoveryCode)
	if err != nil {
		return nil, err
	}

	newCodes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		return nil, fmt.Errorf("generate recovery codes: %w", err)
	}

	hashedCodes := hashCodes(newCodes)

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
	sessionUser, err := requiresVerifiedSession(ctx, s.queries)
	if err != nil {
		return err
	}

	user2fa, err := s.twoFactor.GetByUserID(ctx, sessionUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("2fa not active: %w", domain.ErrForbidden)
		}

		return fmt.Errorf("retrieve user two factor: %w", err)
	}
	if !user2fa.TotpVerifiedAt.Valid {
		return fmt.Errorf("2fa not active: %w", domain.ErrForbidden)
	}

	if err := s.verifyPasswordAndAuth(ctx, sessionUser.UserID, user2fa, password, totpCode, recoveryCode); err != nil {
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

// verifyPasswordAndAuth requires exactly one of totpCode and recoveryCode to be provided.
func (s *TwoFactor) verifyPasswordAndAuth(ctx context.Context, userID pgtype.UUID, user2fa generated.UserTwoFactor, password string, totpCode, recoveryCode *string) error {
	switch {
	case totpCode == nil && recoveryCode == nil:
		return fmt.Errorf("totp code or recovery code required: %w", domain.ErrBadInput)
	case totpCode != nil && recoveryCode != nil:
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
		if user2fa.LastUsedCode.Valid &&
			subtle.ConstantTimeCompare([]byte(user2fa.LastUsedCode.String), []byte(*totpCode)) == 1 {
			return fmt.Errorf("code already used: %w", domain.ErrBadInput)
		}

		secret, err := s.cipher.Open(userID.Bytes, user2fa.TotpSecretEncrypted)
		if err != nil {
			return fmt.Errorf("decrypt totp secret: %w", err)
		}

		if !totp.Validate(*totpCode, secret) {
			return fmt.Errorf("invalid totp code: %w", domain.ErrBadInput)
		}
		if err := s.twoFactor.SetLastUsedCode(ctx, userID, *totpCode); err != nil {
			return fmt.Errorf("store last used code: %w", err)
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

func (s *TwoFactor) validateTOTP(ctx context.Context, userID pgtype.UUID, passcode string) (*generated.UserTwoFactor, error) {
	user2fa, err := s.twoFactor.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("retrieve user two factor: %w", err)
	}

	if user2fa.LastUsedCode.Valid &&
		subtle.ConstantTimeCompare([]byte(user2fa.LastUsedCode.String), []byte(passcode)) == 1 {
		return nil, fmt.Errorf("code already used: %w", domain.ErrBadInput)
	}

	secret, err := s.cipher.Open(userID.Bytes, user2fa.TotpSecretEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt totp secret: %w", err)
	}

	if !totp.Validate(passcode, secret) {
		return nil, fmt.Errorf("invalid code: %w", domain.ErrBadInput)
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

// requiresAuthenticatedSession returns the session user when the session is
// bound to a user, regardless of TwoFaPending state. Use this in flows that
// need to identify the caller but must work whether 2FA is pending or not
// (e.g. /auth/2fa/setup, which can be retried while the session is pending).
func requiresAuthenticatedSession(ctx context.Context, queries *generated.Queries) (generated.GetSessionUserByTokenRow, error) {
	sessionUser, err := middleware.RequireSession(ctx, queries)
	if err != nil {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("session required: %w", err)
	}

	if !sessionUser.UserID.Valid {
		return generated.GetSessionUserByTokenRow{}, fmt.Errorf("requires authenticated user: %w", domain.ErrUnauthorized)
	}

	return sessionUser, nil
}

func hashCodes(recoveryCodes []string) []string {
	hashedCodes := make([]string, len(recoveryCodes))
	for i := range recoveryCodes {
		hashedCodes[i] = auth.HashToken(recoveryCodes[i])
	}

	return hashedCodes
}
