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
    expires_at
)
SELECT 
    $1 AS id,
    $2 AS subject_id,
    $3 AS refresh_token, -- hashed
    t.roles AS roles,
    $4 AS expires_at
FROM m2m_session_templates t
WHERE t.id = $5;
