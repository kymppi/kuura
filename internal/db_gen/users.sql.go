// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package db_gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :exec
INSERT INTO users (id, username, salt, verifier)
VALUES ($1, $2, $3, $4)
`

type CreateUserParams struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Salt     string `json:"salt"`
	Verifier string `json:"verifier"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.Exec(ctx, createUser,
		arg.ID,
		arg.Username,
		arg.Salt,
		arg.Verifier,
	)
	return err
}

const getAndDeletePremaster = `-- name: GetAndDeletePremaster :one
DELETE FROM srp_premasters 
WHERE id = $1 AND expires_at > NOW()
RETURNING data
`

func (q *Queries) GetAndDeletePremaster(ctx context.Context, id string) (string, error) {
	row := q.db.QueryRow(ctx, getAndDeletePremaster, id)
	var data string
	err := row.Scan(&data)
	return data, err
}

const getUserIDFromUsername = `-- name: GetUserIDFromUsername :one
SELECT id FROM users
WHERE username = $1
`

func (q *Queries) GetUserIDFromUsername(ctx context.Context, username string) (string, error) {
	row := q.db.QueryRow(ctx, getUserIDFromUsername, username)
	var id string
	err := row.Scan(&id)
	return id, err
}

const getUserSaltAndVerifier = `-- name: GetUserSaltAndVerifier :one
SELECT id, salt, verifier FROM users
WHERE username = $1
`

type GetUserSaltAndVerifierRow struct {
	ID       string `json:"id"`
	Salt     string `json:"salt"`
	Verifier string `json:"verifier"`
}

func (q *Queries) GetUserSaltAndVerifier(ctx context.Context, username string) (GetUserSaltAndVerifierRow, error) {
	row := q.db.QueryRow(ctx, getUserSaltAndVerifier, username)
	var i GetUserSaltAndVerifierRow
	err := row.Scan(&i.ID, &i.Salt, &i.Verifier)
	return i, err
}

const storePremaster = `-- name: StorePremaster :exec
INSERT INTO srp_premasters (id, data, expires_at)
VALUES ($1, $2, $3)
`

type StorePremasterParams struct {
	ID        string             `json:"id"`
	Data      string             `json:"data"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
}

func (q *Queries) StorePremaster(ctx context.Context, arg StorePremasterParams) error {
	_, err := q.db.Exec(ctx, storePremaster, arg.ID, arg.Data, arg.ExpiresAt)
	return err
}
