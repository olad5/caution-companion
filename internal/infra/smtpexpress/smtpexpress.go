package smtpexpress

import (
	"context"
	"fmt"

	"github.com/olad5/caution-companion/config"
	"github.com/olad5/caution-companion/internal/infra"
	"github.com/prime-labs/smtpexpress-client-go/lib"
	"go.uber.org/zap"
)

type SMTPExpress struct {
	logger *zap.Logger
	Client *lib.APIClient
	cfg    *config.Configurations
}

func New(ctx context.Context, logger *zap.Logger, cfg *config.Configurations) (*SMTPExpress, error) {
	client := lib.CreateClient(cfg.SMTPExpressProjectSecret, &lib.Config{})

	return &SMTPExpress{
		logger: logger,
		Client: client,
		cfg:    cfg,
	}, nil
}

func (s *SMTPExpress) Send(ctx context.Context, opts infra.MailOptions) error {
	smtpOpts := lib.SendMailOptions{
		Message: opts.Body,
		Subject: opts.Subject,
		Sender: lib.MailSender{
			Email: s.cfg.SenderEmail,
			Name:  "caution-companion",
		},
		Recipients: []lib.MailRecipient{
			{
				Email: opts.To,
			},
		},
	}
	_, err := s.Client.Send.SendMail(ctx, smtpOpts)
	if err != nil {
		s.logger.Error("Error sending mail: ", zap.Error(err))
		return fmt.Errorf("Error sending email: %w", err)
	}
	s.logger.Info("Mail has been sent successfully: ")
	return nil
}
