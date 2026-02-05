package auth

import (
	"context"
	"net/http"

	"github.com/mnhsh/time-capsule/internal/config"
	response "github.com/mnhsh/time-capsule/internal/response"
)

type contextKey string

const UserIDKey contextKey = "userID"

func WithAuthMiddleware(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetBearerToken(r.Header)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
			return
		}

		userID, err := ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid Token", err)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
