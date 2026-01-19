-- +goose Up
ALTER TABLE events ADD COLUMN data TEXT;

-- +goose Down
ALTER TABLE events DROP COLUMN data;