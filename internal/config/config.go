package config

import (
	"github.com/mnhsh/time-capsule/internal/database"
)

type Config struct {
	DB        *database.Queries
	JWTSecret string
}
