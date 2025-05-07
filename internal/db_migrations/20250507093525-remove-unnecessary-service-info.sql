
-- +migrate Up
ALTER TABLE services DROP COLUMN access_token_cookie;
ALTER TABLE services DROP COLUMN api_domain;

-- +migrate Down
ALTER TABLE services ADD COLUMN access_token_cookie TEXT NOT NULL DEFAULT 'access_token';
ALTER TABLE services ADD COLUMN api_domain text NOT NULL DEFAULT 'api.example.com';
