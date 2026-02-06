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

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
