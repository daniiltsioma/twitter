package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwt"
)

type contextKey string

const userIdKey contextKey = "userId"

func WithUserID(ctx context.Context, userId int64) context.Context {
	return context.WithValue(ctx, userIdKey, userId)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userId, ok := ctx.Value(userIdKey).(int64)
	return userId, ok
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userId := int64(claims["user_id"].(float64))

		ctx := WithUserID(r.Context(), userId)

		// Token is authenticated, pass user ID through
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}