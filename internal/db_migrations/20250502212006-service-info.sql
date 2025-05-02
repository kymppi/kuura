
-- +migrate Up
ALTER TABLE services ADD COLUMN contact_name TEXT NOT NULL DEFAULT 'Admin';
ALTER TABLE services ADD COLUMN contact_email TEXT NOT NULL DEFAULT 'kuura@midka.dev';

-- +migrate Down
ALTER TABLE services DROP COLUMN contact_email;
ALTER TABLE services DROP COLUMN contact_name;