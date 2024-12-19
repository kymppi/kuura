-- name: CreateM2MRoleTemplate :exec
INSERT INTO m2m_session_templates (id, roles)
VALUES ($1, $2);

-- name: GetM2MRoleTemplates :many
SELECT * FROM m2m_session_templates;

-- name: CreateM2MSession :exec
INSERT INTO m2m_sessions (
    id,
    subject_id,
    refresh_token,
    roles,
    expires_at,
    service_id
)
SELECT 
    $1 AS id,
    $2 AS subject_id,
    $3 AS refresh_token, -- hashed
    t.roles AS roles,
    $4 AS expires_at,
    $5 as service_id
FROM m2m_session_templates t
WHERE t.id = $6;

-- name: GetM2MSessionAndService :one
SELECT 
    m.id,
    m.subject_id,
    m.refresh_token,
    m.roles,
    m.created_at,
    m.last_authenticated_at,
    m.expires_at,
    m.service_id,
    s.name as service_name,
    s.description as service_description,
    s.jwt_audience as service_jwt_audience,
    s.modified_at as service_modified_at,
    s.created_at as service_created_at
FROM m2m_sessions m
JOIN services s ON s.id = m.service_id
WHERE m.id = $1;

-- name: UpdateM2MSessionLastAuthenticatedAt :exec
UPDATE m2m_sessions 
SET last_authenticated_at = NOW()
WHERE id = $1;

-- name: RotateM2MSessionRefreshToken :exec
UPDATE m2m_sessions
SET refresh_token = $1
WHERE id = $2;
