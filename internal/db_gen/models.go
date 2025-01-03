// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db_gen

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type JwkPrivate struct {
	ID               string             `json:"id"`
	ServiceID        pgtype.UUID        `json:"service_id"`
	EncryptedKeyData []byte             `json:"encrypted_key_data"`
	Nonce            []byte             `json:"nonce"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
}

type JwkPublicKey struct {
	ID        string             `json:"id"`
	ServiceID pgtype.UUID        `json:"service_id"`
	KeyData   []byte             `json:"key_data"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type M2mSession struct {
	ID                  string             `json:"id"`
	SubjectID           string             `json:"subject_id"`
	RefreshToken        string             `json:"refresh_token"`
	Roles               []string           `json:"roles"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	LastAuthenticatedAt pgtype.Timestamptz `json:"last_authenticated_at"`
	ExpiresAt           pgtype.Timestamptz `json:"expires_at"`
	ServiceID           pgtype.UUID        `json:"service_id"`
}

type M2mSessionTemplate struct {
	ID        string      `json:"id"`
	Roles     []string    `json:"roles"`
	ServiceID pgtype.UUID `json:"service_id"`
}

type Service struct {
	ID          pgtype.UUID        `json:"id"`
	JwtAudience string             `json:"jwt_audience"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	ModifiedAt  time.Time          `json:"modified_at"`
	Name        string             `json:"name"`
	Description pgtype.Text        `json:"description"`
}

type ServiceKeyState struct {
	ServiceID    pgtype.UUID `json:"service_id"`
	JwkPrivateID string      `json:"jwk_private_id"`
	Status       string      `json:"status"`
}
