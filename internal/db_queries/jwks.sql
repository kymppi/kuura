
-- name: InsertJWKTransaction :exec
WITH inserted_private_key AS (
    INSERT INTO jwk_private (id, service_id, encrypted_key_data, nonce)
    VALUES ($1, $2, $3, $4)
    RETURNING id
),
inserted_public_key AS (
    INSERT INTO jwk_public_keys (id, service_id, key_data)
    VALUES ($1, $2, $5)
    RETURNING id
)
SELECT * FROM inserted_private_key, inserted_public_key;

-- name: DeleteJWKPublic :exec
DELETE FROM jwk_public_keys
WHERE id = $1
AND EXISTS (
    SELECT 1 
    FROM service_key_states
    WHERE service_key_states.service_id = $2
    AND service_key_states.jwk_private_id = $1
);

-- name: DeleteJWKPrivate :exec
DELETE FROM jwk_private
WHERE id = $1
AND EXISTS (
    SELECT 1 
    FROM service_key_states
    WHERE service_key_states.service_id = $2
    AND service_key_states.jwk_private_id = $1
);

-- name: GetJWKPrivate :one
SELECT 
    p.id AS private_id,
    p.service_id,
    p.encrypted_key_data,
    p.nonce,
    p.created_at,
    k.key_data AS public_key_data
FROM 
    jwk_private p
INNER JOIN 
    jwk_public_keys k ON p.id = k.id
WHERE 
    p.id = $1
    AND EXISTS (
        SELECT 1
        FROM service_key_states
        WHERE service_key_states.service_id = $2
        AND service_key_states.jwk_private_id = p.id
    );

-- name: GetCurrentJWKPrivate :one
SELECT 
    p.id,
    p.service_id,
    p.encrypted_key_data,
    p.nonce,
    p.created_at,
    k.key_data AS public_key_data
FROM 
    jwk_private p
INNER JOIN 
    service_key_states sks ON p.id = sks.jwk_private_id
INNER JOIN 
    jwk_public_keys k ON p.id = k.id
WHERE 
    sks.service_id = $1
    AND sks.status = 'current'
ORDER BY p.created_at DESC -- multiple keys can be 'current' during rotation
LIMIT 1;

-- name: GetJWKPublic :one
SELECT * 
FROM jwk_public_keys
WHERE id = $1
AND EXISTS (
    SELECT 1 
    FROM service_key_states
    WHERE service_key_states.service_id = $2
    AND service_key_states.jwk_private_id = $1
);

-- name: GetPublicJWKs :many
SELECT * 
FROM jwk_public_keys
WHERE id IN (
    SELECT jwk_private_id
    FROM service_key_states
    WHERE service_key_states.service_id = $1
);

-- name: CreateServiceKey :exec
INSERT INTO service_key_states (service_id, jwk_private_id, status)
VALUES ($1, $2, 'future');

-- name: SetJWKStatusToCurrent :exec
UPDATE service_key_states
SET status = 'current'
WHERE service_id = $1
  AND jwk_private_id = $2;

-- name: SetJWKStatusToRetired :exec
UPDATE service_key_states
SET status = 'retired'
WHERE service_id = $1
  AND jwk_private_id = $2;

-- name: GetUpcomingKey :one
SELECT 
    p.id
FROM 
    jwk_private p
INNER JOIN 
    service_key_states sks ON p.id = sks.jwk_private_id
WHERE 
    sks.service_id = $1
    AND sks.status = 'future'
ORDER BY p.created_at ASC
LIMIT 1;

-- name: GetOldestRetiredKey :one
SELECT 
    p.id
FROM 
    jwk_private p
INNER JOIN 
    service_key_states sks ON p.id = sks.jwk_private_id
WHERE 
    sks.service_id = $1
    AND sks.status = 'retired'
ORDER BY p.created_at ASC
LIMIT 1;
