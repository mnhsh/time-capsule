package config

import (
	"github.com/mnhsh/time-capsule/internal/database"
	storage "github.com/mnhsh/time-capsule/internal/storage"
)

type Config struct {
	DB        database.Store
	JWTSecret string
	Storage   storage.S3Storage
}
