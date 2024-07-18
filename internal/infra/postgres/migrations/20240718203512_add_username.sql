-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE users ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE users ADD COLUMN user_name VARCHAR(12) DEFAULT '';
ALTER TABLE users ADD COLUMN location VARCHAR(255) DEFAULT '';
ALTER TABLE users ADD COLUMN phone VARCHAR(11) DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE users DROP COLUMN user_name ;
ALTER TABLE users DROP COLUMN location ;
ALTER TABLE users DROP COLUMN phone ;
ALTER TABLE users DROP COLUMN created_at ;
ALTER TABLE users DROP COLUMN updated_at ;
-- +goose StatementEnd
