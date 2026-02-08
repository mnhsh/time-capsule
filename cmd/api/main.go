package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/mnhsh/time-capsule/internal/auth"
	"github.com/mnhsh/time-capsule/internal/config"
	"github.com/mnhsh/time-capsule/internal/database"
	"github.com/mnhsh/time-capsule/internal/storage"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("couldn't open database: %v", err)
	}
	defer db.Close()

	store := database.NewStore(db)

	s3Storage, err := storage.NewS3Storage(context.Background(), "capsule-bucket", "us-east-1", true)
	if err != nil {
		log.Fatalf("couldn't create S3 client: %v", err)
	}

	cfg := &config.Config{
		DB:        store,
		JWTSecret: os.Getenv("JWT_SECRET"),
		Storage:   *s3Storage,
	}

	app := newAPI(cfg)
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /v1/users", app.handlerUsers)
	mux.HandleFunc("POST /v1/login", app.handlerLogin)
	mux.HandleFunc("POST /v1/refresh", app.handlerRefreshToken)
	mux.HandleFunc("POST /v1/revoke", app.handlerRevoke)

	// Protected routes
	mux.Handle("POST /v1/capsules", auth.WithAuthMiddleware(cfg, http.HandlerFunc(app.handlerCreateCapsule)))
	mux.Handle("GET /v1/capsules", auth.WithAuthMiddleware(cfg, http.HandlerFunc(app.handlerGetCapsule)))

	// Wrap with CORS middleware
	handler := auth.CORSMiddleware(mux)

	log.Println("Server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", handler))
}
