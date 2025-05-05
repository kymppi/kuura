-- name: GetSettingsByKey :one
SELECT value FROM instance_settings
WHERE key = $1;

-- name: UpsertSetting :exec
INSERT INTO instance_settings (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value;
