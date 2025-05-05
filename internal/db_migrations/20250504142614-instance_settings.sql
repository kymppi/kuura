
-- +migrate Up
CREATE TABLE instance_settings(
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- +migrate Down
DROP TABLE instance_settings;
