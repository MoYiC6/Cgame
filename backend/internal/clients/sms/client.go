package sms

import "context"

type Client interface {
	Send(ctx context.Context, to string, message string) error
}
