-- name: CreateUser :exec
INSERT INTO users (id, username, salt, verifier)
VALUES ($1, $2, $3, $4);

-- name: GetUserSaltAndVerifier :one
SELECT id, salt, verifier FROM users
WHERE username = $1;

-- name: StorePremaster :exec
INSERT INTO srp_premasters (id, data, expires_at)
VALUES ($1, $2, $3);

-- name: GetAndDeletePremaster :one
DELETE FROM srp_premasters 
WHERE id = $1 AND expires_at > NOW()
RETURNING data;

-- name: GetUserIDFromUsername :one
SELECT id FROM users
WHERE username = $1;
