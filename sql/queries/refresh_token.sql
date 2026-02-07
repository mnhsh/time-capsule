-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  token,
  created_at,
  updated_at,
  user_id,
  expires_at
) VALUES (
  $1,
  now(),
  now(),
  $2,
  $3
)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
    revoked_at = now(),
    updated_at = now()
WHERE token = $1
  AND revoked_at IS NULL;
