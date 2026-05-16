package auth

import (
	"context"
	"encoding/json"
	"net/http"
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
		claims, err := service.AuthenticateRequest(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			writeUnauthorized(w)
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
		if claims.SessionID != "" {
			ctx = context.WithValue(ctx, sessionIDContextKey, claims.SessionID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": http.StatusUnauthorized,
		"msg":  ErrUnauthorized.Error(),
		"data": nil,
	})
}
