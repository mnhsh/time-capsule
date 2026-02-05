-- name: CreateOutboxEvent :one
INSERT INTO outbox (id, payload, status, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPendingOutboxEvents :many
SELECT * FROM outbox
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1;

-- name: UpdateOutboxStatus :exec
UPDATE outbox
SET status = $2
WHERE id = $1;
