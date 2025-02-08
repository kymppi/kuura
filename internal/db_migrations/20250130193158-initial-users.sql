
-- +migrate Up
CREATE TABLE users (
    id text PRIMARY KEY,
    username text UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    disabled boolean DEFAULT false,
    salt text NOT NULL, -- SRP
    verifier text NOT NULL, -- SRP
    roles text[] DEFAULT '{}'
);

CREATE TABLE srp_premasters (
    id text PRIMARY KEY REFERENCES users(id),
    data text NOT NULL,
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
DROP TABLE srp_premasters;
DROP TABLE users;
