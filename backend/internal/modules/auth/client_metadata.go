package auth

import (
	"context"
	"strings"
)

type clientIPContextKey struct{}
type userAgentContextKey struct{}

func WithClientMetadata(ctx context.Context, clientIP, userAgent string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, clientIPContextKey{}, strings.TrimSpace(clientIP))
	ctx = context.WithValue(ctx, userAgentContextKey{}, strings.TrimSpace(userAgent))
	return ctx
}

func ClientIPFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(clientIPContextKey{}).(string)
	return strings.TrimSpace(value)
}

func UserAgentFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(userAgentContextKey{}).(string)
	return strings.TrimSpace(value)
}
