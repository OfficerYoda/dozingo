// Package email provides the transactional email sender and templated message helpers.
package email

import (
	"fmt"
	"net/url"
	"time"

	"github.com/resend/resend-go/v3"

	"github.com/officeryoda/dozingo/internal/config"
)

// frontendURL is the base URL of the Dozingo frontend used to build the
// action links (password reset, email verification) included in transactional
// emails.
const frontendURL = "https://dozingo.de"

// Sender is the abstract interface used by services that need to send mail.
// Production code wires up *ResendSender; tests substitute a fake.
type Sender interface {
	SendResetPassword(receiverAddress, token string) error
	SendEmailVerification(receiverAddress, token string) error
	SendLoginNotification(receiverAddress string, loginTime time.Time) error
	Send2FAActivated(receiverAddress string, activatedAt time.Time) error
}

type ResendSender struct {
	client        *resend.Client
	senderAddress string
}

func New(cfg *config.Config) Sender {
	return &ResendSender{
		client:        resend.NewClient(cfg.ResendAPIKey),
		senderAddress: cfg.MailSenderAddress,
	}
}

func (s *ResendSender) SendResetPassword(receiverAddress, token string) error {
	actionURL := fmt.Sprintf(
		"%s/reset-password?token=%s",
		frontendURL,
		url.QueryEscape(token),
	)

	html, err := render("password_reset.html", templateData{ActionURL: actionURL})
	if err != nil {
		return err
	}

	text := "Hi there,\n\n" +
		"We got a request to reset your Dozingo password. " +
		"Open the link below to choose a new one:\n\n" +
		actionURL + "\n\n" +
		"If you didn't request this, you can safely ignore this email — " +
		"your password won't change.\n\n" +
		"— Dozingo"

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Dozingo <%s>", s.senderAddress),
		To:      []string{receiverAddress},
		Html:    html,
		Text:    text,
		Subject: "[Dozingo] Password Reset",
	}

	_, err = s.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}

func (s *ResendSender) SendEmailVerification(receiverAddress, token string) error {
	actionURL := fmt.Sprintf(
		"%s/verify-email?token=%s",
		frontendURL,
		url.QueryEscape(token),
	)

	html, err := render("verify_email.html", templateData{ActionURL: actionURL})
	if err != nil {
		return err
	}

	text := "Welcome to Dozingo!\n\n" +
		"Please confirm your email address to finish setting up your account " +
		"by opening the link below:\n\n" +
		actionURL + "\n\n" +
		"If you didn't sign up for Dozingo, feel free to ignore this email.\n\n" +
		"— Dozingo"

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Dozingo <%s>", s.senderAddress),
		To:      []string{receiverAddress},
		Html:    html,
		Text:    text,
		Subject: "[Dozingo] Email verification",
	}

	_, err = s.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}

func (s *ResendSender) SendLoginNotification(receiverAddress string, loginTime time.Time) error {
	ts := loginTime.UTC().Format("Jan 02, 2006 at 15:04 UTC")

	html, err := render("login_notification.html", templateData{Timestamp: ts})
	if err != nil {
		return err
	}

	text := "New login detected on your Dozingo account.\n\n" +
		"Time: " + ts + "\n\n" +
		"If this was you, no action is needed. " +
		"If you don't recognize this activity, please reset your password immediately.\n\n" +
		"— Dozingo"

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Dozingo <%s>", s.senderAddress),
		To:      []string{receiverAddress},
		Html:    html,
		Text:    text,
		Subject: "[Dozingo] New login to your account",
	}

	_, err = s.client.Emails.Send(params)

	return err
}

func (s *ResendSender) Send2FAActivated(receiverAddress string, activatedAt time.Time) error {
	ts := activatedAt.UTC().Format("Jan 02, 2006 at 15:04 UTC")

	html, err := render("two_fa_activated.html", templateData{Timestamp: ts})
	if err != nil {
		return err
	}

	text := "Two-factor authentication has been activated on your Dozingo account.\n\n" +
		"Activated at: " + ts + "\n\n" +
		"From now on, you'll need your authenticator app when signing in. " +
		"If you didn't make this change, please reset your password immediately.\n\n" +
		"— Dozingo"

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Dozingo <%s>", s.senderAddress),
		To:      []string{receiverAddress},
		Html:    html,
		Text:    text,
		Subject: "[Dozingo] Two-factor authentication activated",
	}

	_, err = s.client.Emails.Send(params)

	return err
}
