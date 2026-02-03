package capsule

import (
	"context"
	"time"
)

type Capsule struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	S3Key      string    `json:"s3_key"`
	UnlockAt   time.Time `json:"unlock_at"`
	IsUnlocked bool      `json:"is_unlocked"`
	CreatedAt  time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, c *Capsule) error
	GetByID(ctx context.Context, id string) (*Capsule, error)
	MarkAsUnlocked(ctx context.Context, id string) error
}
