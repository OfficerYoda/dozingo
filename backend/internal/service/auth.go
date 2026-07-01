// Package service implements the application's business logic on top of the repository layer.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/auth"
	"github.com/officeryoda/dozingo/internal/domain"
	emailpkg "github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/middleware"
	"github.com/officeryoda/dozingo/internal/pgmap"
	"github.com/officeryoda/dozingo/internal/repository"
	"github.com/officeryoda/dozingo/internal/storage"
)

const (
	emailVerificationTokenTTL = 12 * time.Hour
	passwordResetTokenTTL     = 30 * time.Minute
)

type AvatarGenerator func(seed string) (*storage.Image, error)

type Auth struct {
	users              *repository.Users
	passwords          *repository.Passwords
	sessions           *repository.Sessions
	twoFactor          *repository.TwoFactor
	verificationTokens *repository.VerificationTokens
	emailSender        emailpkg.Sender
	queries            *generated.Queries
	txRunner           repository.TxRunner
	avatarGen          AvatarGenerator
	uploader           storage.ObjectUploader
}

func NewAuth(
	repos repository.Repos,
	queries *generated.Queries,
	emailSender emailpkg.Sender,
	txRunner repository.TxRunner,
	avatarGen AvatarGenerator,
	uploader storage.ObjectUploader,
) *Auth {
	return &Auth{
		users:              repos.Users,
		passwords:          repos.Passwords,
		sessions:           repos.Sessions,
		twoFactor:          repos.TwoFactor,
		verificationTokens: repos.VerificationTokens,
		emailSender:        emailSender,
		txRunner:           txRunner,
		queries:            queries,
		avatarGen:          avatarGen,
		uploader:           uploader,
	}
}

type RegisterInput struct {
	Username string
	Password string
	Email    *string
}

type LoginInput struct {
	Identifier string // username or email address
	Password   string
}

// LoginResult is the return value of Login. When TwoFAPending is true the
// session has been marked pending and the caller must direct the user to
// POST /auth/2fa/verify (or /auth/2fa/verify-recovery) before accessing
// protected endpoints. User is zero-value in that case.
type LoginResult struct {
	User         generated.User
	TwoFAPending bool
}

type NewPasswordInput struct {
	Token       string
	NewPassword string
}

func (s *Auth) Register(ctx context.Context, in RegisterInput) (generated.User, error) {
	user, err := s.generateUser(ctx, in)
	if err != nil {
		return generated.User{}, fmt.Errorf("user creation: %w", err)
	}

	// generate and upload profile
	updated, err := s.assignGeneratedAvatar(ctx, user)
	if err != nil {
		slog.Warn("failed to assign generated avatar on register",
			"error", err, "user_id", user.ID.String(), "username", user.Username)
	} else {
		user = updated
	}

	// send email verification
	if in.Email != nil && strings.TrimSpace(*in.Email) != "" {
		err = sendEmailVerification(ctx, s.txRunner, s.emailSender, user.ID, *in.Email)
		if err != nil {
			slog.Warn("failed to send email verification", "error", err)
		}
	}

	_, err = s.attachUserToSession(ctx, user)
	if err != nil {
		return generated.User{}, fmt.Errorf("attach user to session: %w", err)
	}

	return user, nil
}

func (s *Auth) assignGeneratedAvatar(ctx context.Context, user generated.User) (generated.User, error) {
	if s.avatarGen == nil || s.uploader == nil {
		return user, fmt.Errorf("avatar generator or uploader not configured")
	}

	avatarID, err := uuid.NewRandom()
	if err != nil {
		return user, fmt.Errorf("generate avatar key uuid: %w", err)
	}

	img, err := s.avatarGen(avatarID.String())
	if err != nil {
		return user, fmt.Errorf("generate avatar: %w", err)
	}

	objectKey := fmt.Sprintf("%s%s", avatarID, img.Extension)

	err = s.uploader.Upload(ctx, objectKey, img)
	if err != nil {
		return user, fmt.Errorf("upload avatar: %w", err)
	}

	updatedUser, err := s.users.SetAvatar(ctx, user.ID, objectKey)
	if err != nil {
		return user, fmt.Errorf("set avatar key: %w", err)
	}

	return updatedUser, nil
}

func (s *Auth) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	type loginUser struct {
		id           pgtype.UUID
		username     string
		email        string
		avatarKey    string
		passwordHash string
	}

	var lu loginUser
	var fetchErr error
	if strings.Contains(in.Identifier, "@") {
		u, err := s.users.GetForPasswordLoginByEmail(ctx, in.Identifier)
		if err != nil {
			fetchErr = err
		} else {
			lu = loginUser{u.ID, u.Username, u.Email, u.AvatarKey, u.PasswordHash}
		}
	} else {
		u, err := s.users.GetForPasswordLoginByUsername(ctx, in.Identifier)
		if err != nil {
			fetchErr = err
		} else {
			lu = loginUser{u.ID, u.Username, u.Email, u.AvatarKey, u.PasswordHash}
		}
	}

	if fetchErr != nil {
		if errors.Is(fetchErr, domain.ErrNotFound) {
			auth.CheckPasswordAgainstDummy(in.Password)
			return LoginResult{}, domain.ErrUnauthorized
		}
		return LoginResult{}, fmt.Errorf("user retrieval for login: %w", fetchErr)
	}

	err := auth.CheckPassword(in.Password, lu.passwordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("password mismatch: %w", err)
	}

	vanillaUser := generated.User{
		ID:        lu.id,
		Username:  lu.username,
		Email:     lu.email,
		AvatarKey: lu.avatarKey,
	}
	sessionToken, err := s.attachUserToSession(ctx, vanillaUser)
	if err != nil {
		return LoginResult{}, fmt.Errorf("attach user to session: %w", err)
	}

	twoFA, err := s.twoFactor.GetByUserID(ctx, vanillaUser.ID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return LoginResult{}, fmt.Errorf("check 2fa status: %w", err)
	}
	if twoFA.TotpVerifiedAt.Valid {
		_, err = s.sessions.SetTwoFAPending(ctx, sessionToken, true)
		if err != nil {
			return LoginResult{}, fmt.Errorf("mark session pending 2fa: %w", err)
		}

		return LoginResult{TwoFAPending: true}, nil
	}

	err = s.emailSender.SendLoginNotification(lu.email, time.Now())
	if err != nil {
		slog.Warn("failed to send login notification", "error", err, "user", lu.email)
	}

	return LoginResult{User: vanillaUser}, nil
}

func (s *Auth) Logout(ctx context.Context) error {
	sessionUser, err := requiresVerifiedSession(ctx, s.queries)
	if err != nil {
		return err
	}

	err = s.sessions.Delete(ctx, sessionUser.Token)
	if err != nil {
		return fmt.Errorf("delete session token: %w", err)
	}

	err = middleware.ClearSessionTokenCookie(ctx)
	if err != nil {
		slog.Warn("failed to clear session cookie on logout", "error", err)
	}

	return nil
}

func (s *Auth) ChangePassword(ctx context.Context, oldPassword, newPassword string) error {
	sessionUser, err := requiresVerifiedSession(ctx, s.queries)
	if err != nil {
		return err
	}

	if oldPassword == newPassword {
		return fmt.Errorf("password match: %w", domain.ErrUnprocessableEntity)
	}

	oldPasswordHash, err := s.passwords.GetHashForUserID(ctx, sessionUser.UserID)
	if err != nil {
		return fmt.Errorf("get password hash: %w", err)
	}

	err = auth.CheckPassword(oldPassword, oldPasswordHash)
	if err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}

	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		_, err = r.Passwords.Upsert(ctx, sessionUser.UserID, newPasswordHash)
		if err != nil {
			return fmt.Errorf("update password: %w", err)
		}

		err = r.Sessions.DeleteOtherSessionsFromUser(ctx, sessionUser.UserID, sessionUser.SessionID)
		if err != nil {
			return fmt.Errorf("delete other sessions: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Auth) ForgotPassword(ctx context.Context, address string) error {
	user, err := s.users.GetByEmail(ctx, address)
	if err != nil {
		return fmt.Errorf("retrieve user: %w", err)
	}

	token, err := upsertToken(
		ctx,
		s.txRunner,
		user.ID,
		generated.TokenTypePasswordReset,
		nil,
		passwordResetTokenTTL,
	)
	if err != nil {
		return err
	}

	err = s.emailSender.SendResetPassword(address, token)
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

func (s *Auth) NewPassword(ctx context.Context, in NewPasswordInput) (generated.User, error) {
	tokenHash := auth.HashToken(in.Token)
	token, err := s.verificationTokens.GetByToken(ctx, tokenHash)
	if err != nil {
		return generated.User{}, fmt.Errorf("retrieve verification token: %w", err)
	}

	if token.Type != generated.TokenTypePasswordReset {
		return generated.User{}, fmt.Errorf("invalid token type: %w", domain.ErrBadInput)
	}

	if token.ExpiresAt.Time.Before(time.Now()) {
		return generated.User{}, fmt.Errorf("expired token: %w", domain.ErrGone)
	}

	passwordHash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		return generated.User{}, fmt.Errorf("hash password: %w", err)
	}

	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		if err = r.VerificationTokens.Delete(ctx, tokenHash); err != nil {
			return fmt.Errorf("delete verification token: %w", err)
		}

		if err = r.Sessions.DeleteByUserID(ctx, token.UserID); err != nil {
			return fmt.Errorf("delete user sessions: %w", err)
		}

		if _, err = r.Passwords.Upsert(ctx, token.UserID, passwordHash); err != nil {
			return fmt.Errorf("update password: %w", err)
		}

		return nil
	})
	if err != nil {
		return generated.User{}, err
	}

	user, err := s.users.GetByID(ctx, token.UserID)
	if err != nil {
		return generated.User{}, fmt.Errorf("retrieve user: %w", err)
	}

	return user, nil
}

func (s *Auth) SendEmailVerification(ctx context.Context) error {
	sessionUser, err := requiresVerifiedSession(ctx, s.queries)
	if err != nil {
		return err
	}

	if !sessionUser.Email.Valid {
		return fmt.Errorf("missing email: %w", domain.ErrUnauthorized)
	}

	if sessionUser.EmailVerifiedAt.Valid {
		return fmt.Errorf("email already verified: %w", domain.ErrConflict)
	}

	return sendEmailVerification(ctx, s.txRunner, s.emailSender, sessionUser.UserID, sessionUser.Email.String)
}

func (s *Auth) VerifyEmail(ctx context.Context, token string) (generated.User, error) {
	tokenHash := auth.HashToken(token)
	verificationToken, err := s.verificationTokens.GetByToken(ctx, tokenHash)
	if err != nil {
		return generated.User{}, fmt.Errorf("retrieve verification token: %w", err)
	}

	if verificationToken.Type != generated.TokenTypeEmailVerification {
		return generated.User{}, fmt.Errorf("invalid token type: %w", domain.ErrBadInput)
	}

	if verificationToken.ExpiresAt.Time.Before(time.Now()) {
		return generated.User{}, fmt.Errorf("expired token: %w", domain.ErrGone)
	}

	user, err := s.users.GetByID(ctx, verificationToken.UserID)
	if err != nil {
		return generated.User{}, fmt.Errorf("retrieve user: %w", err)
	}

	if verificationToken.Email.String != user.Email {
		return generated.User{}, fmt.Errorf("token email mismatch: %w", domain.ErrGone)
	}

	now := time.Now()
	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		err = r.VerificationTokens.Delete(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("delete verification token : %w", err)
		}

		user, err = r.Users.SetEmailVerifiedAt(ctx, verificationToken.UserID, &now)
		if err != nil {
			return fmt.Errorf("set email valid: %w", err)
		}

		return nil
	})
	if err != nil {
		return generated.User{}, err
	}

	return user, nil
}

func (s *Auth) generateUser(ctx context.Context, in RegisterInput) (generated.User, error) {
	passwordHash, err := auth.HashPassword(in.Password)
	if err != nil {
		return generated.User{}, fmt.Errorf("hash password: %w", err)
	}

	var user generated.User
	err = s.txRunner.WithTx(ctx, func(r repository.Repos) error {
		user, err = r.Users.Create(ctx, in.Username, in.Email)
		if err != nil {
			return err
		}
		_, err = r.Passwords.Upsert(ctx, user.ID, passwordHash)
		if err != nil {
			return fmt.Errorf("update password: %w", err)
		}

		return nil
	})
	if err != nil {
		return generated.User{}, err
	}

	return user, nil
}

func (s *Auth) attachUserToSession(ctx context.Context, user generated.User) (string, error) {
	sessionUser, err := middleware.RequireSession(ctx, s.queries)
	if err != nil {
		return "", fmt.Errorf("session required: %w", err)
	}
	_, err = s.sessions.AttachUser(ctx, sessionUser.Token, user.ID)
	if err != nil {
		return "", fmt.Errorf("attach user to session: %w", err)
	}

	return sessionUser.Token, nil
}

func upsertToken(
	ctx context.Context,
	txRunner repository.TxRunner,
	userID pgtype.UUID,
	tokenType generated.TokenType,
	email *string,
	tokenTTL time.Duration,
) (string, error) {
	plaintext := auth.GenerateToken()
	err := txRunner.WithTx(ctx, func(r repository.Repos) error {
		existingToken, err := r.VerificationTokens.GetValidTokenForUser(ctx, repository.GetByTokenForUserInput{
			UserID:    userID,
			TokenType: tokenType,
		})
		// "no existing valid token" is the normal first-issuance case;
		// only bail on other errors.
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return pgmap.TranslatePgErr(err)
		}

		if existingToken.UserID.Valid {
			err = r.VerificationTokens.Delete(ctx, existingToken.Token)
			if err != nil {
				return fmt.Errorf("delete existing token: %w", err)
			}
		}

		_, err = r.VerificationTokens.Create(ctx, repository.CreateVerificationTokenInput{
			UserID:    userID,
			TokenHash: auth.HashToken(plaintext),
			TokenType: tokenType,
			Email:     email,
			ExpiresAt: time.Now().Add(tokenTTL),
		})
		if err != nil {
			return fmt.Errorf("create token: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	// Return plaintext so the caller can embed it in the verification email
	// link. It is never persisted.
	return plaintext, nil
}
