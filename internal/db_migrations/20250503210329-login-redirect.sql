
-- +migrate Up
ALTER TABLE services ADD COLUMN login_redirect TEXT NOT NULL DEFAULT 'https://midka.dev';

-- +migrate Down
ALTER TABLE services DROP COLUMN login_redirect;
