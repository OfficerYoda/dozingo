package email

import (
	"fmt"

	"github.com/officeryoda/dozingo/internal/config"
	"github.com/resend/resend-go/v3"
)

type MailSender struct {
	client        *resend.Client
	senderAddress string
}

func New(cfg *config.Config) MailSender {
	return MailSender{
		client:        resend.NewClient(cfg.ResendAPIKey),
		senderAddress: cfg.MailSenderAddress,
	}
}

func (m *MailSender) SendResetPassword(receiverAddress, resetToken string) {
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Dozingo <%s>", m.senderAddress),
		To:      []string{receiverAddress},
		Html:    "<strong>hello world</strong> the reset token is: " + resetToken,
		Subject: "Hello from Dozingo",
	}

	sent, err := m.client.Emails.Send(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(sent.Id)
}
