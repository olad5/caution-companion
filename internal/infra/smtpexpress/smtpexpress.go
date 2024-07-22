package smtpexpress

import (
	"context"
	"fmt"

	"github.com/olad5/caution-companion/config"
	"github.com/olad5/caution-companion/internal/infra"
	"github.com/prime-labs/smtpexpress-client-go/lib"
)

type SMTPExpress struct {
	Client *lib.APIClient
	cfg    *config.Configurations
}

func New(ctx context.Context, cfg *config.Configurations) (*SMTPExpress, error) {
	client := lib.CreateClient(cfg.SMTPExpressProjectSecret, &lib.Config{})

	return &SMTPExpress{
		Client: client,
		cfg:    cfg,
	}, nil
}

func (r *SMTPExpress) Send(ctx context.Context, opts infra.MailOptions) error {
	smtpOpts := lib.SendMailOptions{
		Message: opts.Body,
		Subject: opts.Subject,
		Sender: lib.MailSender{
			Email: r.cfg.SenderEmail,
			Name:  "caution-companion",
		},
		Recipients: []lib.MailRecipient{
			{
				Email: opts.To,
			},
		},
	}
	_, err := r.Client.Send.SendMail(ctx, smtpOpts)
	if err != nil {
		// TODO:TODO: you need to log the email has been sent successfully
		return fmt.Errorf("Error sending email: %w", err)
	}
	return nil
}
