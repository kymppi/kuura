package m2m

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	tokenhasher "github.com/kymppi/kuura/internal/argon2"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/models"
	"github.com/oklog/ulid/v2"
)

type M2MService struct {
	db          *db_gen.Queries
	tokenhasher *tokenhasher.TokenHasher
}

func NewM2MService(generatedQueries *db_gen.Queries) *M2MService {
	return &M2MService{
		db: generatedQueries,
		tokenhasher: tokenhasher.NewTokenHasher(tokenhasher.Argon2Params{

			Memory:      64 * 1024,
			Iterations:  3,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		}),
	}
}

func (s *M2MService) CreateRoleTemplate(ctx context.Context, name string, roles []string) error {
	err := s.db.CreateM2MRoleTemplate(ctx, db_gen.CreateM2MRoleTemplateParams{
		ID:    name,
		Roles: roles,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *M2MService) GetRoleTemplates(ctx context.Context) ([]*models.M2MRoleTemplate, error) {
	data, err := s.db.GetM2MRoleTemplates(ctx)

	if err != nil {
		return nil, err
	}

	var result []*models.M2MRoleTemplate
	for _, row := range data {
		result = append(result, &models.M2MRoleTemplate{
			Id:    row.ID,
			Roles: row.Roles,
		})
	}

	return result, nil
}

func (s *M2MService) CreateSession(ctx context.Context, subjectId string, template string) (id string, initialToken string, err error) {
	id = ulid.Make().String()

	initialToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate opaque token: %w", err)
	}

	hashedToken, err := s.tokenhasher.HashValue(initialToken)

	if err != nil {
		return "", "", fmt.Errorf("failed to hash initial token: %w", err)
	}

	err = s.db.CreateM2MSession(ctx, db_gen.CreateM2MSessionParams{
		ID:           id,
		SubjectID:    subjectId,
		RefreshToken: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 1),
			Valid: true,
		},
		ID_2: template,
	})

	if err != nil {
		return "", "", err
	}

	return id, initialToken, nil
}

func generateOpaqueToken(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length <= 0 {
		return "", nil
	}

	result := make([]byte, length)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[index.Int64()]
	}

	return string(result), nil
}
