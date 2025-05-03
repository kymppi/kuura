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

-- name: SaveSRPServer :exec
INSERT INTO user_srp (uid, encoded_server, expires_at)
VALUES ($1, $2, $3);

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
