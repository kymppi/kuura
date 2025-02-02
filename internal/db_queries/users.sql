-- name: CreateUser :exec
INSERT INTO users (id, username, salt, verifier)
VALUES ($1, $2, $3, $4);
