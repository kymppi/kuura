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

-- name: UpdateService :exec
UPDATE services
SET 
    jwt_audience = COALESCE($2, jwt_audience),
    modified_at = NOW(),
    name = COALESCE($3, name),
    description = COALESCE($4, description),
    access_token_duration = COALESCE($5, access_token_duration),
    access_token_cookie = COALESCE($6, access_token_cookie),
    login_redirect = COALESCE($7, login_redirect),
    contact_name = COALESCE($8, contact_name),
    contact_email = COALESCE($9, contact_email)
WHERE id = $1;
