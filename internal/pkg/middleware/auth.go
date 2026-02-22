package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/RuLap/sportmates-api/internal/pkg/jwthelper"
	"github.com/darahayes/go-boom"
)

func AuthMiddleware(jwtHelper *jwthelper.JWTHelper) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				boom.Unathorized(w, "Authorization header required")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				boom.Unathorized(w, "Invalid authorization format")
				return
			}

			claims, err := jwtHelper.ParseJWT(tokenString)
			if err != nil {
				boom.Unathorized(w, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_email", claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
