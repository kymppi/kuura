-- name: CreateUser :exec
INSERT INTO users (id, username, hashed_username, encoded_verifier)
VALUES ($1, $2, $3, $4);

-- name: GetSRPVerifier :one
SELECT encoded_verifier FROM users WHERE id = $1;

-- name: GetAndDeleteSRPServer :one
DELETE FROM user_srp
WHERE uid = $1 AND expires_at > NOW()
RETURNING *;

-- name: GetUserIDFromUsername :one
SELECT id FROM users
WHERE username = $1;

-- name: GetUserIDFromUsernameHash :one
SELECT id FROM users
WHERE hashed_username = $1;

-- name: CheckSRPServerNotExpired :one
SELECT EXISTS (
    SELECT 1 FROM user_srp
    WHERE uid = $1 AND expires_at > CURRENT_TIMESTAMP
) AS record_not_expired;

-- name: UpsertSRPServer :exec
INSERT INTO user_srp (uid, encoded_server, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (uid) DO UPDATE
SET encoded_server = EXCLUDED.encoded_server, expires_at = EXCLUDED.expires_at;

-- name: CreateUserSession :exec
INSERT INTO user_sessions (id, user_id, service_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdateUserLastSignInDate :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: GetUserSession :one
SELECT * FROM user_sessions
WHERE id = $1;

-- name: UpdateUserSessionLastAuthenticatedAt :exec
UPDATE user_sessions 
SET last_authenticated_at = NOW()
WHERE id = $1;

-- name: RotateUserSessionRefreshToken :exec
UPDATE user_sessions
SET refresh_token_hash = $1
WHERE id = $2;

-- name: GetUserRoles :one
SELECT roles FROM users
WHERE id = $1;

-- name: GetUser :one
SELECT id, username, last_login_at FROM users
WHERE id = $1;

-- name: InsertCodeToSessionTokenExchange :exec
INSERT INTO user_token_code_exchange (session_id, expires_at, hashed_code)
VALUES ($1, $2, $3);

-- name: UseTokenExchangeCode :one
DELETE FROM user_token_code_exchange AS token
WHERE token.hashed_code = $1
  AND token.expires_at > NOW()
RETURNING
  token.session_id;

-- name: GetAccessTokenDurationUsingSessionId :one
SELECT svc.access_token_duration
FROM services AS svc
JOIN user_sessions AS us ON us.service_id = svc.id
WHERE us.id = $1;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions
WHERE id = $1 AND user_id = $2;
