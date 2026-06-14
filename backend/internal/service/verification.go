package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/officeryoda/dozingo/internal/email"
	"github.com/officeryoda/dozingo/internal/generated"
	"github.com/officeryoda/dozingo/internal/repository"
)

// sendEmailVerification mints a fresh email-verification token for the
// given user and sends a verification mail to `to`.
func sendEmailVerification(
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
