-- name: CreateM2MRoleTemplate :exec
INSERT INTO m2m_session_templates (id, roles)
VALUES ($1, $2);

-- name: GetM2MRoleTemplates :many
SELECT * FROM m2m_session_templates;
