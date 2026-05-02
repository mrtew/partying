package middleware

import (
	"context"
	"net/http"
	"partying/internal/service"
	"partying/pkg/response"
	"strings"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
)

func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				response.Unauthorized(w, "missing token")
				return
			}

			userID, username, err := authService.ValidateToken(token)
			if err != nil {
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UsernameKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken 同时支持Header和Query参数两种方式
func extractToken(r *http.Request) string {
	// 方式1：Authorization: Bearer xxx（Postman用）
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	// 方式2：?token=xxx（浏览器WebSocket用）
	return r.URL.Query().Get("token")
}

func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(UserIDKey).(string)
	return id
}

func GetUsername(r *http.Request) string {
	username, _ := r.Context().Value(UsernameKey).(string)
	return username
}
