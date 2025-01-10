package m2m

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	tokenhasher "github.com/kymppi/kuura/internal/argon2"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/jwks"
	"github.com/kymppi/kuura/internal/models"
	"github.com/kymppi/kuura/internal/utils"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/oklog/ulid/v2"
)

type M2MService struct {
	db          *db_gen.Queries
	tokenhasher *tokenhasher.TokenHasher
	jwtIssuer   string
	jwkManager  *jwks.JWKManager
}

func NewM2MService(generatedQueries *db_gen.Queries, jwtIssuer string, jwkManager *jwks.JWKManager) *M2MService {
	return &M2MService{
		db: generatedQueries,
		tokenhasher: tokenhasher.NewTokenHasher(tokenhasher.Argon2Params{
			Memory:      64 * 1024,
			Iterations:  3,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		}),
		jwtIssuer:  jwtIssuer,
		jwkManager: jwkManager,
	}
}

func (s *M2MService) CreateRoleTemplate(ctx context.Context, serviceId uuid.UUID, name string, roles []string) error {
	err := s.db.CreateM2MRoleTemplate(ctx, db_gen.CreateM2MRoleTemplateParams{
		ID:        name,
		Roles:     roles,
		ServiceID: utils.UUIDToPgType(serviceId),
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *M2MService) GetRoleTemplates(ctx context.Context, serviceId uuid.UUID) ([]*models.M2MRoleTemplate, error) {
	data, err := s.db.GetM2MRoleTemplates(ctx, utils.UUIDToPgType(serviceId))

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

func (s *M2MService) CreateSession(ctx context.Context, serviceId uuid.UUID, subjectId string, template string) (id string, initialToken string, err error) {
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
		ID_2:      template,
		ServiceID: utils.UUIDToPgType(serviceId),
	})

	if err != nil {
		return "", "", err
	}

	return id, initialToken, nil
}

func (s *M2MService) CreateAccessToken(ctx context.Context, sessionId string, refreshToken string) (accessToken string, newRefreshToken string, err error) {
	valid, roles, service, subjectId, err := s.validateRefreshTokenAndGetRolesAndServiceAndSubjectId(ctx, sessionId, refreshToken)

	if err != nil || !valid {
		//TODO: log real error to console, could be like "role doesn't exist"
		return "", "", errors.New("invalid token")
	}

	err = s.db.UpdateM2MSessionLastAuthenticatedAt(ctx, sessionId)

	if err != nil {
		return "", "", fmt.Errorf("failed to update session last authentication date: %w", err)
	}

	// important that we try to generate the jwt BEFORE updating refresh token, if it fails then the client can't even retry
	exp := time.Now().Add(30 * time.Minute) //TODO: move to services db table

	token, err := jwt.NewBuilder().
		Audience([]string{service.JWTAudience}).
		Issuer(s.jwtIssuer).
		Subject(subjectId).
		IssuedAt(time.Now()).
		Expiration(exp).
		Claim("session_id", sessionId).
		Claim("roles", roles).
		Claim("client_type", "machine").
		Build()

	if err != nil {
		return "", "", fmt.Errorf("failed to build jwt: %w", err)
	}

	signingKey, err := s.jwkManager.GetSigningKey(ctx, service.Id)

	if err != nil {
		return "", "", fmt.Errorf("failed to get signing key: %w", err)
	}

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.ES384, signingKey))

	if err != nil {
		return "", "", fmt.Errorf("failed to sign jwt: %w", err)
	}

	accessToken = string(signedToken)

	newRefreshToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to rotate refresh token: generate opaque token: %w", err)
	}

	hashedToken, err := s.tokenhasher.HashValue(newRefreshToken)

	if err != nil {
		return "", "", fmt.Errorf("failed to hash new refresh token: %w", err)
	}

	err = s.db.RotateM2MSessionRefreshToken(ctx, db_gen.RotateM2MSessionRefreshTokenParams{
		RefreshToken: hashedToken,
		ID:           sessionId,
	})

	if err != nil {
		return "", "", fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (s *M2MService) validateRefreshTokenAndGetRolesAndServiceAndSubjectId(ctx context.Context, sessionId string, refreshToken string) (valid bool, roles []string, service *models.AppService, subjectId string, err error) {
	session, err := s.db.GetM2MSessionAndService(ctx, sessionId)

	if err != nil {
		return false, nil, nil, "", err
	}

	if time.Now().After(session.ExpiresAt.Time) {
		return false, nil, nil, "", errors.New("the session is expired")
	}

	valid, err = s.tokenhasher.CompareHashAndValue(session.RefreshToken, refreshToken)

	if err != nil {
		return false, nil, nil, "", err
	} else if !valid {
		return false, nil, nil, "", errors.New("invalid token")
	}

	serviceId, err := utils.PgTypeUUIDToUUID(session.ServiceID)

	if err != nil {
		return false, nil, nil, "", fmt.Errorf("failed to parse session's service id")
	}

	service = &models.AppService{
		Id:          serviceId,
		JWTAudience: session.ServiceJwtAudience,
		CreatedAt:   session.ServiceCreatedAt.Time,
		ModifiedAt:  session.ServiceModifiedAt,
		Name:        session.ServiceName,
		Description: session.ServiceDescription.String,
	}

	return true, session.Roles, service, session.SubjectID, nil
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
