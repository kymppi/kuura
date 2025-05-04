
-- +migrate Up
ALTER TABLE services ADD COLUMN access_token_duration INT NOT NULL DEFAULT 3600;
ALTER TABLE services ADD COLUMN access_token_cookie TEXT NOT NULL DEFAULT 'access_token';

-- +migrate Down
ALTER TABLE services DROP COLUMN access_token_duration;
ALTER TABLE services DROP COLUMN access_token_cookie;
