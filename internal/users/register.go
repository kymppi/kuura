package users

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/db_gen"
)

func (s *UserService) Register(ctx context.Context, username string, verifier string) (uid string, err error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to create new uuid: %w", err)
	}

	hash := sha256.New()
	hash.Write([]byte(username))
	hashInBytes := hash.Sum(nil)
	hashedUsername := hex.EncodeToString(hashInBytes)

	if err = s.db.CreateUser(ctx, db_gen.CreateUserParams{
		ID:              id.String(),
		Username:        username,
		EncodedVerifier: verifier,
		HashedUsername:  hashedUsername,
	}); err != nil {
		return "", fmt.Errorf("failed to create user in db: %w", err)
	}

	s.logger.Info("Created new user", slog.String("username", username))

	return id.String(), nil
}
