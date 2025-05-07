
-- +migrate Up
CREATE TABLE user_token_code_exchange(
    session_id TEXT NOT NULL PRIMARY KEY REFERENCES user_sessions(id),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    hashed_code TEXT NOT NULL
);

-- +migrate Down
DROP TABLE user_token_code_exchange;
