package jwks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kymppi/kuura/internal/db_gen"
	jwk_storage "github.com/kymppi/kuura/internal/jwks/storage"
	"github.com/kymppi/kuura/internal/utils"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type FullJWK struct {
	id        string
	public    jwk.ECDSAPublicKey
	private   jwk.ECDSAPrivateKey
	createdAt time.Time // ignored when used as a param for StoreKey
}

type PublicJWK struct {
	id        string
	public    jwk.ECDSAPublicKey
	createdAt time.Time
}

type KeyStorage interface {
	StoreKey(ctx context.Context, serviceId uuid.UUID, key FullJWK) error
	GetPublic(ctx context.Context, serviceId uuid.UUID, id string) (PublicJWK, error)
	GetPublicKeys(ctx context.Context, serviceId uuid.UUID) ([]PublicJWK, error)
	GetPrivate(ctx context.Context, serviceId uuid.UUID, id string) (FullJWK, error)
	GetCurrentPrivateKey(ctx context.Context, serviceId uuid.UUID) (FullJWK, error)
	DeleteKey(ctx context.Context, serviceId uuid.UUID, id string) error
	SetCurrentKey(ctx context.Context, serviceId uuid.UUID, nextKey string) error
	GetUpcomingKey(ctx context.Context, serviceId uuid.UUID) (id string, err error)
	GetOldestRetired(ctx context.Context, serviceId uuid.UUID) (id string, err error)
	GetKeyStates(ctx context.Context, serviceId uuid.UUID) (map[string]string, error)
}

type PostgresQLKeyStorage struct {
	db            *db_gen.Queries
	encryptor     *jwk_storage.SymmetricKeyEncryptor
	encryptionKey []byte
}

func NewPostgresQLKeyStorage(databaseQueries *db_gen.Queries, encryptionKey []byte) *PostgresQLKeyStorage {
	return &PostgresQLKeyStorage{
		db:            databaseQueries,
		encryptor:     jwk_storage.NewSymmetricKeyEncryptor(),
		encryptionKey: encryptionKey,
	}
}

func (ks *PostgresQLKeyStorage) StoreKey(ctx context.Context, serviceId uuid.UUID, key FullJWK) error {
	privateKeyJSON, err := json.Marshal(key.private)
	if err != nil {
		return fmt.Errorf("failed to serialize private key: %w", err)
	}

	encryptedPrivateKey, nonce, err := ks.encryptor.Encrypt(privateKeyJSON, ks.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	publicKeyJSON, err := json.Marshal(key.public)
	if err != nil {
		return fmt.Errorf("failed to serialize public key: %w", err)
	}

	err = ks.db.InsertJWKTransaction(ctx, db_gen.InsertJWKTransactionParams{
		ID:               key.id,
		EncryptedKeyData: encryptedPrivateKey, // encrypted private key serialized w/ json
		Nonce:            nonce,               // IV
		KeyData:          publicKeyJSON,       // public key as json
		ServiceID:        utils.UUIDToPgType(serviceId),
	})

	if err != nil {
		return handlePgError("StoreKey", err, key.id)
	}

	err = ks.db.CreateServiceKey(ctx, db_gen.CreateServiceKeyParams{
		ServiceID:    utils.UUIDToPgType(serviceId),
		JwkPrivateID: key.id,
	})

	if err != nil {
		return handlePgError("StoreKeyStep2", err, key.id)
	}

	return nil
}

func (ks *PostgresQLKeyStorage) GetPublic(ctx context.Context, serviceId uuid.UUID, id string) (PublicJWK, error) {
	row, err := ks.db.GetJWKPublic(ctx, db_gen.GetJWKPublicParams{
		ID:        id,
		ServiceID: utils.UUIDToPgType(serviceId),
	})
	if err != nil {
		return PublicJWK{}, handlePgError("GetPublic", err, id)
	}

	publicKey, err := parsePublicKey(row.KeyData, id)
	if err != nil {
		return PublicJWK{}, fmt.Errorf("failed to deserialize public key: %w", err)
	}

	return PublicJWK{
		id:        row.ID,
		public:    publicKey,
		createdAt: row.CreatedAt.Time,
	}, nil
}

func (ks *PostgresQLKeyStorage) GetPublicKeys(ctx context.Context, serviceId uuid.UUID) ([]PublicJWK, error) {
	rows, err := ks.db.GetPublicJWKs(ctx, utils.UUIDToPgType(serviceId))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public keys: %w", err)
	}

	var publicKeys []PublicJWK
	for _, row := range rows {
		publicKey, err := parsePublicKey(row.KeyData, row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize public key with ID %s: %w", row.ID, err)
		}

		publicKeys = append(publicKeys, PublicJWK{
			id:        row.ID,
			public:    publicKey,
			createdAt: row.CreatedAt.Time,
		})
	}

	return publicKeys, nil
}

func (ks *PostgresQLKeyStorage) GetPrivate(ctx context.Context, serviceId uuid.UUID, id string) (FullJWK, error) {
	row, err := ks.db.GetJWKPrivate(ctx, db_gen.GetJWKPrivateParams{
		ID:        id,
		ServiceID: utils.UUIDToPgType(serviceId),
	})

	if err != nil {
		return FullJWK{}, handlePgError("GetPrivate", err, id)
	}

	decryptedPrivateKey, err := ks.encryptor.Decrypt(row.EncryptedKeyData, ks.encryptionKey, row.Nonce)
	if err != nil {
		return FullJWK{}, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	privateKey, err := parsePrivateKey(decryptedPrivateKey)
	if err != nil {
		return FullJWK{}, fmt.Errorf("failed to deserialize private key: %w", err)
	}

	publicKey, err := parsePublicKey(row.PublicKeyData, id)
	if err != nil {
		return FullJWK{}, fmt.Errorf("failed to deserialize public key: %w", err)
	}

	return FullJWK{
		id:        id,
		private:   privateKey,
		public:    publicKey,
		createdAt: row.CreatedAt.Time,
	}, nil
}

func (ks *PostgresQLKeyStorage) GetCurrentPrivateKey(ctx context.Context, serviceId uuid.UUID) (FullJWK, error) {
	row, err := ks.db.GetCurrentJWKPrivate(ctx, utils.UUIDToPgType(serviceId))

	if err != nil {
		return FullJWK{}, handlePgError("GetCurrentPrivate", err, serviceId.String())
	}

	decryptedPrivateKey, err := ks.encryptor.Decrypt(row.EncryptedKeyData, ks.encryptionKey, row.Nonce)
	if err != nil {
		return FullJWK{}, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	privateKey, err := parsePrivateKey(decryptedPrivateKey)
	if err != nil {
		return FullJWK{}, fmt.Errorf("failed to deserialize private key: %w", err)
	}

	publicKey, err := parsePublicKey(row.PublicKeyData, row.ID)
	if err != nil {
		return FullJWK{}, fmt.Errorf("failed to deserialize public key: %w", err)
	}

	return FullJWK{
		id:        row.ID,
		private:   privateKey,
		public:    publicKey,
		createdAt: row.CreatedAt.Time,
	}, nil
}

func (ks *PostgresQLKeyStorage) DeleteKey(ctx context.Context, serviceId uuid.UUID, id string) error {
	err := ks.db.DeleteJWKPublic(ctx, db_gen.DeleteJWKPublicParams{
		ID:        id,
		ServiceID: utils.UUIDToPgType(serviceId),
	})
	if err != nil {
		return handlePgError("DeleteJWKPublic", err, id)
	}

	err = ks.db.DeleteJWKPrivate(ctx, db_gen.DeleteJWKPrivateParams{
		ID:        id,
		ServiceID: utils.UUIDToPgType(serviceId),
	})
	if err != nil {
		return handlePgError("DeleteJWKPrivate", err, id)
	}

	return nil
}

func (ks *PostgresQLKeyStorage) SetCurrentKey(ctx context.Context, serviceId uuid.UUID, nextKey string) error {
	var currentKeyID string

	currentKey, err := ks.GetCurrentPrivateKey(ctx, serviceId)
	if err != nil {
		currentKeyID = ""
	} else {
		currentKeyID = currentKey.id
	}

	err = ks.db.SetJWKStatusToCurrent(ctx, db_gen.SetJWKStatusToCurrentParams{
		ServiceID:    utils.UUIDToPgType(serviceId),
		JwkPrivateID: nextKey,
	})
	if err != nil {
		return handlePgError("SetCurrentKey", err, nextKey)
	}

	if currentKeyID != "" {
		err = ks.db.SetJWKStatusToRetired(ctx, db_gen.SetJWKStatusToRetiredParams{
			ServiceID:    utils.UUIDToPgType(serviceId),
			JwkPrivateID: currentKeyID,
		})
		if err != nil {
			return handlePgError("SetCurrentKeyRetired", err, currentKeyID)
		}
	}

	return nil
}

func (ks *PostgresQLKeyStorage) GetUpcomingKey(ctx context.Context, serviceId uuid.UUID) (id string, err error) {
	id, err = ks.db.GetUpcomingKey(ctx, utils.UUIDToPgType(serviceId))

	if err != nil {
		return "", handlePgError("GetUpcomingKey", err, id)
	}

	return id, nil
}

func (ks *PostgresQLKeyStorage) GetOldestRetired(ctx context.Context, serviceId uuid.UUID) (id string, err error) {
	id, err = ks.db.GetOldestRetiredKey(ctx, utils.UUIDToPgType(serviceId))

	if err != nil {
		return "", handlePgError("GetOldestRetired", err, id)
	}

	return id, nil
}

func (ks *PostgresQLKeyStorage) GetKeyStates(ctx context.Context, serviceId uuid.UUID) (map[string]string, error) {
	data, err := ks.db.GetKeyStatus(ctx, utils.UUIDToPgType(serviceId))

	if err != nil {
		return nil, handlePgError("GetKeyStates", err, "")
	}

	statusMap := make(map[string]string, len(data))
	for _, row := range data {
		statusMap[row.JwkPrivateID] = row.Status
	}

	return statusMap, nil
}

func handlePgError(operation string, err error, id string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%s: key with ID %s already exists: %w", operation, id, err)
		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%s: key with ID %s cannot be deleted due to a foreign key violation: %w", operation, id, err)
		case pgerrcode.NoData:
			return fmt.Errorf("%s: key with ID %s not found: %w", operation, id, err)
		default:
			return fmt.Errorf("%s: database error with ID %s, code %s: %w", operation, id, pgErr.Code, err)
		}
	}

	return fmt.Errorf("%s: error with key ID %s: %w", operation, id, err)
}

func parsePublicKey(keyData []byte, id string) (jwk.ECDSAPublicKey, error) {
	publicKey, err := jwk.ParseKey(keyData)

	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecPublicKey, ok := publicKey.(jwk.ECDSAPublicKey)

	if !ok {
		return nil, errors.New("public key is not ECDSA")
	}

	ecPublicKey.Set("kid", id)

	return ecPublicKey, nil
}

func parsePrivateKey(keyData []byte) (jwk.ECDSAPrivateKey, error) {
	privateKey, err := jwk.ParseKey(keyData)

	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecPublicKey, ok := privateKey.(jwk.ECDSAPrivateKey)

	if !ok {
		return nil, errors.New("private key is not ECDSA")
	}

	return ecPublicKey, nil
}
