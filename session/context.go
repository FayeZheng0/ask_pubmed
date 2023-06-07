package session

import (
	"context"

	"github.com/fox-one/mixin-sdk-go"
)

type contextKey struct{}

type key int

const (
	userKey key = iota
	clientIPKey
	envKey
)

func With(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, contextKey{}, s)
}

func From(ctx context.Context) *Session {
	return ctx.Value(contextKey{}).(*Session)
}

func WithUser(ctx context.Context, user *mixin.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFrom(ctx context.Context) (*mixin.User, bool) {
	u, ok := ctx.Value(userKey).(*mixin.User)
	return u, ok
}

func WithEnv(ctx context.Context, env string) context.Context {
	return context.WithValue(ctx, envKey, env)
}

func EnvFrom(ctx context.Context) (string, bool) {
	env, ok := ctx.Value(envKey).(string)
	return env, ok
}
