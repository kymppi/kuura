package users

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/opencoff/go-srp"
)

// validates the client and returns server proof
func (s *UserService) ClientVerify(ctx context.Context, ih string, data string) (string, error) {
	uid, err := s.db.GetUserIDFromUsernameHash(ctx, ih)
	if err != nil {
		return "", fmt.Errorf("failed to get uid from identity hash: %w", err)
	}

	row, err := s.db.GetAndDeleteSRPServer(ctx, uid)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve srp server: %w", err)
	}

	srv, err := srp.UnmarshalServer(string(row.EncodedServer))
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal server: %w", err)
	}

	proof, ok := srv.ClientOk(data)
	if !ok {
		return "", fmt.Errorf("unauthorized")
	}

	s.logger.Info("User logged in", slog.String("uid", uid))

	return proof, nil
}

func (s *UserService) ClientBegin(ctx context.Context, creds string) (string, error) {
	ih, A, err := srp.ServerBegin(creds)
	if err != nil {
		return "", fmt.Errorf("failed to begin server: %w", err)
	}

	uid, err := s.db.GetUserIDFromUsernameHash(ctx, ih)
	if err != nil {
		return "", fmt.Errorf("failed to get uid from identity hash: %w", err)
	}

	s.logger.Info("User initiated login", slog.String("uid", uid))

	encodedVerifier, err := s.db.GetSRPVerifier(ctx, uid)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user verifier and salt: %w", err)
	}

	srpObj, v, err := srp.MakeSRPVerifier(encodedVerifier)
	if err != nil {
		return "", fmt.Errorf("failed to make srp verifier: %w", err)
	}

	srv, err := srpObj.NewServer(v, A)
	if err != nil {
		return "", fmt.Errorf("failed to compute the shared secret: %w", err)
	}

	creds = srv.Credentials()

	if err = s.db.SaveSRPServer(ctx, db_gen.SaveSRPServerParams{
		Uid:           uid,
		EncodedServer: []byte(srv.Marshal()),
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(5 * time.Second),
			Valid: true,
		},
	}); err != nil {
		return "", fmt.Errorf("failed to store SRP server: %w", err)
	}

	return creds, nil
}
