package email

import "context"

type Client interface {
	Send(ctx context.Context, to string, subject string, body string) error
}
