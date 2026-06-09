package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

// issueAndSendEmailVerification mints a fresh email-verification token for the
// given user and sends a verification mail to `to`. Shared by:
//
//   - Auth.SendEmailVerification: the user requests a verification mail for
//     their existing (unverified) address.
//   - Users.UpdateUser: when the user's email has just changed, fire off a
//     fresh verification mail to the new address.
//
// The caller is responsible for any policy gating (already verified, missing
// email, session check) before calling this. This helper is purely the
// "issue token + dispatch mail" plumbing.
func issueAndSendEmailVerification(
	ctx context.Context,
	txRunner repository.TxRunner,
	sender email.Sender,
	userID pgtype.UUID,
	to string,
) error {
	token, err := upsertToken(
		ctx,
		txRunner,
		userID,
		generated.TokenTypeEmailVerification,
		emailVerificationTokenTTL,
	)
	if err != nil {
		return err
	}

	if err := sender.SendEmailVerification(to, token); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}
