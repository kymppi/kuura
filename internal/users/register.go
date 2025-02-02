package users

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/db_gen"
)

func (s *UserService) Register(ctx context.Context, username string, srp_salt string, srp_verifier string) (uid string, err error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to create new uuid: %w", err)
	}

	if err = s.db.CreateUser(ctx, db_gen.CreateUserParams{
		ID:       id.String(),
		Username: username,
		Salt:     srp_salt,
		Verifier: srp_verifier,
	}); err != nil {
		return "", fmt.Errorf("failed to create user in db: %w", err)
	}

	s.logger.Info("Created new user", slog.String("username", username))

	return id.String(), nil
}
