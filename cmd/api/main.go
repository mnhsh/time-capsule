package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/mnhsh/time-capsule/internal/auth"
	"github.com/mnhsh/time-capsule/internal/config"
	"github.com/mnhsh/time-capsule/internal/database"
)

func main() {
	db, _ := sql.Open("postgres", os.Getenv("DB_URL"))
	queries := database.New(db)
	cfg := &config.Config{
		DB:        queries,
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
	mux := http.NewServeMux()
	mux.Handle("POST /v1/capsules", auth.WithAuthMiddleware(cfg, http.HandlerFunc(cfg.HandlerCreateCapsule)))

	log.Fatal(http.ListenAndServe(":8081", mux))
}
