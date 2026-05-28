package email

import (
	"fmt"

	"github.com/officeryoda/dozingo/internal/config"
	"github.com/resend/resend-go/v3"
)

type Sender struct {
	client        *resend.Client
	senderAddress string
}

func New(cfg *config.Config) *Sender {
	return &Sender{
		client:        resend.NewClient(cfg.ResendAPIKey),
		senderAddress: cfg.MailSenderAddress,
	}
}

func (s *Sender) SendResetPassword(receiverAddress, token string) error {
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Dozingo <%s>", s.senderAddress),
		To:      []string{receiverAddress},
		Html:    "Well well well... look who forgot their password again.. the reset token is: " + token,
		Subject: "[Dozingo] Password Reset",
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sender) SendEmailVerification(receiverAddress, token string) error {
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Dozingo <%s>", s.senderAddress),
		To:      []string{receiverAddress},
		Html:    "So you want to verify you email hmmm...? the reset token is: " + token,
		Subject: "[Dozingo] Email verification",
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}
