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
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("cannot connect to DB:", err)
	}
	store := database.NewStore(db)
	cfg := &config.Config{
		DB:        store,
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
	app := NewAPI(cfg)
	mux := http.NewServeMux()
	createHandler := http.HandlerFunc(app.HandlerCreateCapsule)
	mux.Handle("POST /v1/capsules", auth.WithAuthMiddleware(cfg, createHandler))
	log.Fatal(http.ListenAndServe(":8081", mux))
}
