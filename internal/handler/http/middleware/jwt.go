package middleware

import (
	"context"
	"net/http"
	"pos-backend/pkg/jwtutil"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user_claims"

func RequireAuth(secretKey string) func(http.HandlerFunc) http.HandlerFunc {
	return func (next http.HandlerFunc) http.HandlerFunc  {
		return func (w http.ResponseWriter, r *http.Request)  {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "Unauthorized: Token not found", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Unauthorized: Format is must Bearer <token>", http.StatusUnauthorized)
				return
			}

			claims, err := jwtutil.ValidateToken(parts[1], secretKey)
			if err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}