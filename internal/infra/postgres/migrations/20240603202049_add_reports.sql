-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE reports(
    id UUID PRIMARY KEY,
    owner_id TEXT NOT NULL,
    incident_type TEXT NOT NULL,
    longitude TEXT NOT NULL,
    latitude TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE reports;
-- +goose StatementEnd
