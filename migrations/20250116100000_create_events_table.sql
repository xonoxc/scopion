-- +goose Up
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    level TEXT NOT NULL,
    service TEXT NOT NULL,
    name TEXT NOT NULL,
    trace_id TEXT NOT NULL
);

-- +goose Down
DROP TABLE events;