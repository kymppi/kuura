
-- +migrate Up
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

CREATE TABLE sessions (
    id text PRIMARY KEY,
    user_id text REFERENCES users(id),
    refresh_token_hash text NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- +migrate Down
DROP INDEX idx_sessions_user_id;
DROP TABLE sessions;
DROP TABLE user_srp;
DROP TABLE users;
