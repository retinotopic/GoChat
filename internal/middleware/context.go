package middleware

import (
	"context"
	"log"
)

type userCtxKeyType string

const userCtxKey userCtxKeyType = "user"

func WithUser(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, userCtxKey, subject)
}

func GetUser(ctx context.Context) string {
	subject, ok := ctx.Value(userCtxKey).(string)
	if !ok {
		log.Println()
		return ""
	}
	return subject
}
