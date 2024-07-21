-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255) DEFAULT 'https://res.cloudinary.com/deda4nfxl/image/upload/v1721583338/caution-companion/caution-companion/avatars/4608bc1b98c84a06838fafb5e38fb552.jpg';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE users DROP COLUMN avatar_url ;
-- +goose StatementEnd
