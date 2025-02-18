
-- +migrate Up
ALTER TABLE services ADD COLUMN api_domain text NOT NULL DEFAULT 'api.example.com';

CREATE TABLE users (
    id text PRIMARY KEY,
    username text UNIQUE NOT NULL,
    hashed_username text UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    disabled boolean DEFAULT false,
    encoded_verifier text NOT NULL, -- SRP
    roles text[] DEFAULT '{}'
);

CREATE TABLE user_srp (
    uid TEXT PRIMARY KEY REFERENCES users(id),
    encoded_server bytea,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE user_sessions (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id),
    service_id uuid NOT NULL REFERENCES services(id),
    refresh_token_hash text NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_authenticated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);

-- +migrate Down
DROP INDEX idx_sessions_user_id;
DROP TABLE sessions;
DROP TABLE user_srp;
DROP TABLE users;
ALTER TABLE services DROP COLUMN api_domain;
