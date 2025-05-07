
-- +migrate Up
ALTER TABLE user_sessions
ALTER COLUMN refresh_token_hash DROP NOT NULL;

-- +migrate Down
ALTER TABLE user_sessions
ALTER COLUMN refresh_token_hash SET NOT NULL;
