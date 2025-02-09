-- name: CreateUser :exec
INSERT INTO users (id, username, encoded_verifier)
VALUES ($1, $2, $3);

-- name: GetSRPVerifier :one
SELECT encoded_verifier FROM users WHERE id = $1;

-- name: GetAndDeleteSRPServer :one
DELETE FROM user_srp
WHERE uid = $1 AND expires_at > NOW()
RETURNING *;

-- name: GetUserIDFromUsername :one
SELECT id FROM users
WHERE username = $1;

-- name: SaveSRPServer :exec
INSERT INTO user_srp (uid, encoded_server, expires_at)
VALUES ($1, $2, $3);
