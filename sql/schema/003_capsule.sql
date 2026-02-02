-- +goose Up
CREATE TABLE capsule (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title TEXT,
  created_at TIMESTAMP NOT NULL,
  s3key TEXT NOT NULL,
  unlock_at TIMESTAMP NOT NULL,
  is_unlocked BOOLEAN DEFAULT false
);

-- +goose Down
DROP TABLE capsule;
