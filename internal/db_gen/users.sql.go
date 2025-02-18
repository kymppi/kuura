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
INSERT INTO users (id, username, hashed_username, encoded_verifier)
VALUES ($1, $2, $3, $4)
`

type CreateUserParams struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	HashedUsername  string `json:"hashed_username"`
	EncodedVerifier string `json:"encoded_verifier"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.Exec(ctx, createUser,
		arg.ID,
		arg.Username,
		arg.HashedUsername,
		arg.EncodedVerifier,
	)
	return err
}

const createUserSession = `-- name: CreateUserSession :exec
INSERT INTO user_sessions (id, user_id, service_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5)
`

type CreateUserSessionParams struct {
	ID               string             `json:"id"`
	UserID           string             `json:"user_id"`
	ServiceID        pgtype.UUID        `json:"service_id"`
	RefreshTokenHash string             `json:"refresh_token_hash"`
	ExpiresAt        pgtype.Timestamptz `json:"expires_at"`
}

func (q *Queries) CreateUserSession(ctx context.Context, arg CreateUserSessionParams) error {
	_, err := q.db.Exec(ctx, createUserSession,
		arg.ID,
		arg.UserID,
		arg.ServiceID,
		arg.RefreshTokenHash,
		arg.ExpiresAt,
	)
	return err
}

const getAndDeleteSRPServer = `-- name: GetAndDeleteSRPServer :one
DELETE FROM user_srp
WHERE uid = $1 AND expires_at > NOW()
RETURNING uid, encoded_server, expires_at
`

func (q *Queries) GetAndDeleteSRPServer(ctx context.Context, uid string) (UserSrp, error) {
	row := q.db.QueryRow(ctx, getAndDeleteSRPServer, uid)
	var i UserSrp
	err := row.Scan(&i.Uid, &i.EncodedServer, &i.ExpiresAt)
	return i, err
}

const getSRPVerifier = `-- name: GetSRPVerifier :one
SELECT encoded_verifier FROM users WHERE id = $1
`

func (q *Queries) GetSRPVerifier(ctx context.Context, id string) (string, error) {
	row := q.db.QueryRow(ctx, getSRPVerifier, id)
	var encoded_verifier string
	err := row.Scan(&encoded_verifier)
	return encoded_verifier, err
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

const getUserIDFromUsernameHash = `-- name: GetUserIDFromUsernameHash :one
SELECT id FROM users
WHERE hashed_username = $1
`

func (q *Queries) GetUserIDFromUsernameHash(ctx context.Context, hashedUsername string) (string, error) {
	row := q.db.QueryRow(ctx, getUserIDFromUsernameHash, hashedUsername)
	var id string
	err := row.Scan(&id)
	return id, err
}

const getUserRoles = `-- name: GetUserRoles :one
SELECT roles FROM users
WHERE id = $1
`

func (q *Queries) GetUserRoles(ctx context.Context, id string) ([]string, error) {
	row := q.db.QueryRow(ctx, getUserRoles, id)
	var roles []string
	err := row.Scan(&roles)
	return roles, err
}

const getUserSession = `-- name: GetUserSession :one
SELECT id, user_id, service_id, refresh_token_hash, expires_at, created_at, last_authenticated_at FROM user_sessions
WHERE id = $1
`

func (q *Queries) GetUserSession(ctx context.Context, id string) (UserSession, error) {
	row := q.db.QueryRow(ctx, getUserSession, id)
	var i UserSession
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ServiceID,
		&i.RefreshTokenHash,
		&i.ExpiresAt,
		&i.CreatedAt,
		&i.LastAuthenticatedAt,
	)
	return i, err
}

const rotateUserSessionRefreshToken = `-- name: RotateUserSessionRefreshToken :exec
UPDATE user_sessions
SET refresh_token_hash = $1
WHERE id = $2
`

type RotateUserSessionRefreshTokenParams struct {
	RefreshTokenHash string `json:"refresh_token_hash"`
	ID               string `json:"id"`
}

func (q *Queries) RotateUserSessionRefreshToken(ctx context.Context, arg RotateUserSessionRefreshTokenParams) error {
	_, err := q.db.Exec(ctx, rotateUserSessionRefreshToken, arg.RefreshTokenHash, arg.ID)
	return err
}

const saveSRPServer = `-- name: SaveSRPServer :exec
INSERT INTO user_srp (uid, encoded_server, expires_at)
VALUES ($1, $2, $3)
`

type SaveSRPServerParams struct {
	Uid           string             `json:"uid"`
	EncodedServer []byte             `json:"encoded_server"`
	ExpiresAt     pgtype.Timestamptz `json:"expires_at"`
}

func (q *Queries) SaveSRPServer(ctx context.Context, arg SaveSRPServerParams) error {
	_, err := q.db.Exec(ctx, saveSRPServer, arg.Uid, arg.EncodedServer, arg.ExpiresAt)
	return err
}

const updateUserLastSignInDate = `-- name: UpdateUserLastSignInDate :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1
`

func (q *Queries) UpdateUserLastSignInDate(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, updateUserLastSignInDate, id)
	return err
}

const updateUserSessionLastAuthenticatedAt = `-- name: UpdateUserSessionLastAuthenticatedAt :exec
UPDATE user_sessions 
SET last_authenticated_at = NOW()
WHERE id = $1
`

func (q *Queries) UpdateUserSessionLastAuthenticatedAt(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, updateUserSessionLastAuthenticatedAt, id)
	return err
}
