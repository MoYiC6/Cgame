package security

import (
	"context"
	"strings"
)

type principalContextKey struct{}
type sessionIDContextKey struct{}

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, principalContextKey{}, p)
}

func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	if ctx == nil {
		return nil, false
	}
	p, ok := ctx.Value(principalContextKey{}).(*Principal)
	if !ok || p == nil {
		return nil, false
	}
	return p, true
}

func MustPrincipal(ctx context.Context) *Principal {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		panic("principal missing from context")
	}
	return p
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, sessionIDContextKey{}, sessionID)
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	sessionID, ok := ctx.Value(sessionIDContextKey{}).(string)
	if !ok || strings.TrimSpace(sessionID) == "" {
		return "", false
	}
	return sessionID, true
}
