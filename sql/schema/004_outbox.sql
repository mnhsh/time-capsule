-- +goose Up
CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    payload JSONB NOT NULL,    -- The Capsule ID and metadata
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE outbox;
