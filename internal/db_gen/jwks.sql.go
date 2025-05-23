// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: jwks.sql

package db_gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createServiceKey = `-- name: CreateServiceKey :exec
INSERT INTO service_key_states (service_id, jwk_private_id, status)
VALUES ($1, $2, 'future')
`

type CreateServiceKeyParams struct {
	ServiceID    pgtype.UUID `json:"service_id"`
	JwkPrivateID string      `json:"jwk_private_id"`
}

func (q *Queries) CreateServiceKey(ctx context.Context, arg CreateServiceKeyParams) error {
	_, err := q.db.Exec(ctx, createServiceKey, arg.ServiceID, arg.JwkPrivateID)
	return err
}

const deleteJWKPrivate = `-- name: DeleteJWKPrivate :exec
DELETE FROM jwk_private
WHERE id = $1
AND EXISTS (
    SELECT 1 
    FROM service_key_states
    WHERE service_key_states.service_id = $2
    AND service_key_states.jwk_private_id = $1
)
`

type DeleteJWKPrivateParams struct {
	ID        string      `json:"id"`
	ServiceID pgtype.UUID `json:"service_id"`
}

func (q *Queries) DeleteJWKPrivate(ctx context.Context, arg DeleteJWKPrivateParams) error {
	_, err := q.db.Exec(ctx, deleteJWKPrivate, arg.ID, arg.ServiceID)
	return err
}

const deleteJWKPublic = `-- name: DeleteJWKPublic :exec
DELETE FROM jwk_public_keys
WHERE id = $1
AND EXISTS (
    SELECT 1 
    FROM service_key_states
    WHERE service_key_states.service_id = $2
    AND service_key_states.jwk_private_id = $1
)
`

type DeleteJWKPublicParams struct {
	ID        string      `json:"id"`
	ServiceID pgtype.UUID `json:"service_id"`
}

func (q *Queries) DeleteJWKPublic(ctx context.Context, arg DeleteJWKPublicParams) error {
	_, err := q.db.Exec(ctx, deleteJWKPublic, arg.ID, arg.ServiceID)
	return err
}

const getCurrentJWKPrivate = `-- name: GetCurrentJWKPrivate :one
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
LIMIT 1
`

type GetCurrentJWKPrivateRow struct {
	ID               string             `json:"id"`
	ServiceID        pgtype.UUID        `json:"service_id"`
	EncryptedKeyData []byte             `json:"encrypted_key_data"`
	Nonce            []byte             `json:"nonce"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	PublicKeyData    []byte             `json:"public_key_data"`
}

func (q *Queries) GetCurrentJWKPrivate(ctx context.Context, serviceID pgtype.UUID) (GetCurrentJWKPrivateRow, error) {
	row := q.db.QueryRow(ctx, getCurrentJWKPrivate, serviceID)
	var i GetCurrentJWKPrivateRow
	err := row.Scan(
		&i.ID,
		&i.ServiceID,
		&i.EncryptedKeyData,
		&i.Nonce,
		&i.CreatedAt,
		&i.PublicKeyData,
	)
	return i, err
}

const getJWKPrivate = `-- name: GetJWKPrivate :one
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
    )
`

type GetJWKPrivateParams struct {
	ID        string      `json:"id"`
	ServiceID pgtype.UUID `json:"service_id"`
}

type GetJWKPrivateRow struct {
	PrivateID        string             `json:"private_id"`
	ServiceID        pgtype.UUID        `json:"service_id"`
	EncryptedKeyData []byte             `json:"encrypted_key_data"`
	Nonce            []byte             `json:"nonce"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	PublicKeyData    []byte             `json:"public_key_data"`
}

func (q *Queries) GetJWKPrivate(ctx context.Context, arg GetJWKPrivateParams) (GetJWKPrivateRow, error) {
	row := q.db.QueryRow(ctx, getJWKPrivate, arg.ID, arg.ServiceID)
	var i GetJWKPrivateRow
	err := row.Scan(
		&i.PrivateID,
		&i.ServiceID,
		&i.EncryptedKeyData,
		&i.Nonce,
		&i.CreatedAt,
		&i.PublicKeyData,
	)
	return i, err
}

const getJWKPublic = `-- name: GetJWKPublic :one
SELECT id, service_id, key_data, created_at 
FROM jwk_public_keys
WHERE id = $1
AND EXISTS (
    SELECT 1 
    FROM service_key_states
    WHERE service_key_states.service_id = $2
    AND service_key_states.jwk_private_id = $1
)
`

type GetJWKPublicParams struct {
	ID        string      `json:"id"`
	ServiceID pgtype.UUID `json:"service_id"`
}

func (q *Queries) GetJWKPublic(ctx context.Context, arg GetJWKPublicParams) (JwkPublicKey, error) {
	row := q.db.QueryRow(ctx, getJWKPublic, arg.ID, arg.ServiceID)
	var i JwkPublicKey
	err := row.Scan(
		&i.ID,
		&i.ServiceID,
		&i.KeyData,
		&i.CreatedAt,
	)
	return i, err
}

const getKeyStatus = `-- name: GetKeyStatus :many
SELECT status, jwk_private_id FROM service_key_states
WHERE service_id = $1
`

type GetKeyStatusRow struct {
	Status       string `json:"status"`
	JwkPrivateID string `json:"jwk_private_id"`
}

func (q *Queries) GetKeyStatus(ctx context.Context, serviceID pgtype.UUID) ([]GetKeyStatusRow, error) {
	rows, err := q.db.Query(ctx, getKeyStatus, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetKeyStatusRow{}
	for rows.Next() {
		var i GetKeyStatusRow
		if err := rows.Scan(&i.Status, &i.JwkPrivateID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOldestRetiredKey = `-- name: GetOldestRetiredKey :one
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
LIMIT 1
`

func (q *Queries) GetOldestRetiredKey(ctx context.Context, serviceID pgtype.UUID) (string, error) {
	row := q.db.QueryRow(ctx, getOldestRetiredKey, serviceID)
	var id string
	err := row.Scan(&id)
	return id, err
}

const getPublicJWKs = `-- name: GetPublicJWKs :many
SELECT id, service_id, key_data, created_at 
FROM jwk_public_keys
WHERE id IN (
    SELECT jwk_private_id
    FROM service_key_states
    WHERE service_key_states.service_id = $1
)
`

func (q *Queries) GetPublicJWKs(ctx context.Context, serviceID pgtype.UUID) ([]JwkPublicKey, error) {
	rows, err := q.db.Query(ctx, getPublicJWKs, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []JwkPublicKey{}
	for rows.Next() {
		var i JwkPublicKey
		if err := rows.Scan(
			&i.ID,
			&i.ServiceID,
			&i.KeyData,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUpcomingKey = `-- name: GetUpcomingKey :one
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
LIMIT 1
`

func (q *Queries) GetUpcomingKey(ctx context.Context, serviceID pgtype.UUID) (string, error) {
	row := q.db.QueryRow(ctx, getUpcomingKey, serviceID)
	var id string
	err := row.Scan(&id)
	return id, err
}

const insertJWKTransaction = `-- name: InsertJWKTransaction :exec
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
SELECT inserted_private_key.id, inserted_public_key.id FROM inserted_private_key, inserted_public_key
`

type InsertJWKTransactionParams struct {
	ID               string      `json:"id"`
	ServiceID        pgtype.UUID `json:"service_id"`
	EncryptedKeyData []byte      `json:"encrypted_key_data"`
	Nonce            []byte      `json:"nonce"`
	KeyData          []byte      `json:"key_data"`
}

func (q *Queries) InsertJWKTransaction(ctx context.Context, arg InsertJWKTransactionParams) error {
	_, err := q.db.Exec(ctx, insertJWKTransaction,
		arg.ID,
		arg.ServiceID,
		arg.EncryptedKeyData,
		arg.Nonce,
		arg.KeyData,
	)
	return err
}

const setJWKStatusToCurrent = `-- name: SetJWKStatusToCurrent :exec
UPDATE service_key_states
SET status = 'current'
WHERE service_id = $1
  AND jwk_private_id = $2
`

type SetJWKStatusToCurrentParams struct {
	ServiceID    pgtype.UUID `json:"service_id"`
	JwkPrivateID string      `json:"jwk_private_id"`
}

func (q *Queries) SetJWKStatusToCurrent(ctx context.Context, arg SetJWKStatusToCurrentParams) error {
	_, err := q.db.Exec(ctx, setJWKStatusToCurrent, arg.ServiceID, arg.JwkPrivateID)
	return err
}

const setJWKStatusToRetired = `-- name: SetJWKStatusToRetired :exec
UPDATE service_key_states
SET status = 'retired'
WHERE service_id = $1
  AND jwk_private_id = $2
`

type SetJWKStatusToRetiredParams struct {
	ServiceID    pgtype.UUID `json:"service_id"`
	JwkPrivateID string      `json:"jwk_private_id"`
}

func (q *Queries) SetJWKStatusToRetired(ctx context.Context, arg SetJWKStatusToRetiredParams) error {
	_, err := q.db.Exec(ctx, setJWKStatusToRetired, arg.ServiceID, arg.JwkPrivateID)
	return err
}
