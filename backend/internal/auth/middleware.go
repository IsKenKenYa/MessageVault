package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "commory.user_id"
const sessionIDContextKey contextKey = "commory.session_id"

func UserIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(userIDContextKey).(string); ok {
		return value
	}
	return ""
}

func SessionIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(sessionIDContextKey).(string); ok {
		return value
	}
	return ""
}

func Middleware(service *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		claims, err := service.ParseAccessTokenClaims(authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
		if claims.SessionID != "" {
			ctx = context.WithValue(ctx, sessionIDContextKey, claims.SessionID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
