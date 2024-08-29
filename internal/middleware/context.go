package middleware

import (
	"context"
)

type userCtxKeyType string

const userCtxKey userCtxKeyType = "user"

func WithUser(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, userCtxKey, subject)
}

func GetUser(ctx context.Context) string {
	subject, ok := ctx.Value(userCtxKey).(string)
	if !ok {
		return ""
	}
	return subject
}
