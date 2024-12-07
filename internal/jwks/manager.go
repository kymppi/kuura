package jwks

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/oklog/ulid/v2"
)

type JWKManager struct {
	storage   KeyStorage
	serviceId uuid.UUID
}

func NewJWKManager(storage KeyStorage, serviceId uuid.UUID) *JWKManager {
	return &JWKManager{
		storage:   storage,
		serviceId: serviceId,
	}
}

func (m *JWKManager) CreateKey(ctx context.Context) (keyId string, err error) {
	keyId = ulid.Make().String()

	raw, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	if err != nil {
		return "", fmt.Errorf("failed to generate new ECDSA private key: %w", err)
	}

	privateKey, err := jwk.FromRaw(raw)
	if err != nil {
		return "", fmt.Errorf("failed to create symmetric key: %w", err)
	}

	if _, ok := privateKey.(jwk.ECDSAPrivateKey); !ok {
		return "", fmt.Errorf("expected jwk.ECDSAPrivateKey, got %T", privateKey)
	}

	publicKey, err := privateKey.PublicKey()

	if err != nil {
		return "", fmt.Errorf("failed to get public key from private key: %w", err)
	}

	if _, ok := publicKey.(jwk.ECDSAPublicKey); !ok {
		return "", fmt.Errorf("expected jwk.ECDSAPublicKey, got %T", publicKey)
	}

	fullJWK := FullJWK{
		id:      keyId,
		private: privateKey.(jwk.ECDSAPrivateKey),
		public:  publicKey.(jwk.ECDSAPublicKey),
	}

	err = m.storage.StoreKey(ctx, m.serviceId, fullJWK)
	if err != nil {
		return "", fmt.Errorf("failed to store key: %w", err)
	}

	return keyId, nil
}

func (m *JWKManager) GetJWKS(ctx context.Context) (jwk.Set, error) {
	publicKeys, err := m.storage.GetPublicKeys(ctx, m.serviceId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve public keys: %w", err)
	}

	set := jwk.NewSet()
	for _, publicKey := range publicKeys {
		set.AddKey(publicKey.public)
	}

	return set, nil
}

func (m *JWKManager) Rotate(ctx context.Context) error {
	_, err := m.CreateKey(ctx)

	if err != nil {
		return fmt.Errorf("failed to create a new key: %w", err)
	}

	upcomingKeyId, err := m.storage.GetUpcomingKey(ctx, m.serviceId)

	if err != nil {
		return fmt.Errorf("failed to get upcoming key id: %w", err)
	}

	err = m.storage.SetCurrentKey(ctx, m.serviceId, upcomingKeyId)

	if err != nil {
		return fmt.Errorf("failed to promote key: %w", err)
	}

	oldestRetiredKeyId, err := m.storage.GetOldestRetired(ctx, m.serviceId)

	if err != nil {
		return fmt.Errorf("failed to get oldest retired key: %w", err)
	}

	err = m.storage.DeleteKey(ctx, m.serviceId, oldestRetiredKeyId)

	if err != nil {
		return fmt.Errorf("failed to remove oldest retired key: %w", err)
	}

	return nil
}

func (m *JWKManager) Remove(ctx context.Context, id string) error {
	currentKey, err := m.storage.GetCurrentPrivateKey(ctx, m.serviceId)
	if err != nil {
		return fmt.Errorf("failed to retrieve current key: %w", err)
	}

	if currentKey.id == id {
		return errors.New("currently used key must not be removed")
	}

	if err := m.storage.DeleteKey(ctx, m.serviceId, id); err != nil {
		return fmt.Errorf("failed to remove key: %w", err)
	}

	return nil
}

func (m *JWKManager) Export(ctx context.Context, id string) (jwk.Key, error) {
	fullKey, err := m.storage.GetPrivate(ctx, m.serviceId, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key: %w", err)
	}
	return fullKey.private, nil
}