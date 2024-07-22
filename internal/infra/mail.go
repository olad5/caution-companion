package infra

import (
	"context"
)

type MailOptions struct {
	To      string
	Subject string
	Body    string
}
type MailService interface {
	Send(ctx context.Context, opts MailOptions) error
}
