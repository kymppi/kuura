package users

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/opencoff/go-srp"
)

// validates the client and returns server proof
func (s *UserService) ClientVerify(ctx context.Context, ih string, data string) (string, error) {
	uidBytes, err := hex.DecodeString(ih)
	if err != nil {
		return "", fmt.Errorf("failed to decode string: %w", err)
	}

	uid := string(uidBytes)

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

	return proof, nil
}

func (s *UserService) ClientBegin(ctx context.Context, creds string) (string, error) {
	ih, A, err := srp.ServerBegin(creds)
	if err != nil {
		return "", fmt.Errorf("failed to begin server: %w", err)
	}

	uidBytes, err := hex.DecodeString(ih)
	if err != nil {
		return "", fmt.Errorf("failed to decode string: %w", err)
	}

	uid := string(uidBytes)

	// lookup the user db using "I" as the key and
	// fetch salt, verifier etc.
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

	// Generate the credentials to send to client
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
