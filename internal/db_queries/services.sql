-- name: GetAppService :one
SELECT * FROM services
WHERE id = $1;

-- name: GetAppServices :many
SELECT * FROM services;

-- name: CreateAppService :exec
INSERT INTO services (id, jwt_audience, name, api_domain, login_redirect)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteAppService :exec
DELETE FROM services
WHERE id = $1;
