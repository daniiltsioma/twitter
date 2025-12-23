package auth

import "context"

type contextKey string

const userIdKey contextKey = "userId"

func WithUserID(ctx context.Context, userId int64) context.Context {
	return context.WithValue(ctx, userIdKey, userId)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userId, ok := ctx.Value(userIdKey).(int64)
	return userId, ok
}