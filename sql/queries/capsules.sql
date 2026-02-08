-- name: CreateCapsule :one
INSERT INTO capsule (id, user_id, title, created_at, s3key, unlock_at, is_unlocked)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetCapsuleForUnlock :one
SELECT id, user_id, title, s3key, is_unlocked FROM capsule
WHERE id = $1 LIMIT 1;

-- name: MarkAsUnlocked :exec
UPDATE capsule
SET is_unlocked = true
WHERE id = $1;

-- name: GetCapsulesByUserID :many
SELECT * FROM capsule WHERE user_id = $1 ORDER BY created_at DESC;
