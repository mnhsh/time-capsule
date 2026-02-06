package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier // This is the interface sqlc generated for you
	CreateCapsuleWithOutbox(ctx context.Context, arg CreateCapsuleParams) error
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	db       *sql.DB // The underlying connection pool
	*Queries         // All generated sqlc methods
}

// NewStore creates a new store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db), // Initialize sqlc queries
	}
}

func (s *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil) // Start the transaction
	if err != nil {
		return err
	}

	q := New(tx) // Create queries scoped specifically to this transaction
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit() // Finalize everything
}

func (s *SQLStore) CreateCapsuleWithOutbox(ctx context.Context, capParams CreateCapsuleParams) error {
	// We use the execTx helper we just built
	return s.execTx(ctx, func(q *Queries) error {
		// 1. Create the Capsule metadata
		_, err := q.CreateCapsule(ctx, capParams)
		if err != nil {
			return fmt.Errorf("failed to create capsule: %w", err)
		}

		// 2. Create the Outbox Event (The "To-Do" note for RabbitMQ)
		_, err = q.CreateOutboxEvent(ctx, CreateOutboxEventParams{
			ID:        uuid.New(),
			Payload:   []byte(capParams.ID.String()), // We only need the ID to unlock it
			Status:    sql.NullString{String: "pending", Valid: true},
			CreatedAt: time.Now().UTC(),
		})
		if err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})
}
