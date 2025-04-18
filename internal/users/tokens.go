package users

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/utils"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/oklog/ulid/v2"
)

func (s *UserService) CreateSession(ctx context.Context, uid string, serviceId uuid.UUID) (id string, refreshToken string, err error) {
	id = ulid.Make().String()

	refreshToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate opaque token: %w", err)
	}

	hashedToken, err := s.tokenhasher.HashValue(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	if err = s.db.CreateUserSession(ctx, db_gen.CreateUserSessionParams{
		ID:     id,
		UserID: uid,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24 * 7),
			Valid: true,
		},
		RefreshTokenHash: hashedToken,
		ServiceID:        utils.UUIDToPgType(serviceId),
	}); err != nil {
		return "", "", err
	}

	if err = s.db.UpdateUserLastSignInDate(ctx, uid); err != nil {
		return "", "", err
	}

	return id, refreshToken, nil
}

func (s *UserService) CreateAccessToken(ctx context.Context, sessionId string, refreshToken string) (accessToken string, newRefreshToken string, serviceDomain string, err error) {
	session, err := s.GetSession(ctx, sessionId)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get session: %w", err)
	}

	roles, err := s.db.GetUserRoles(ctx, session.UserID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get user roles: %w", err)
	}

	tokenValid, err := s.validateRefreshToken(session, refreshToken)
	if err != nil || !tokenValid {
		logFields := []any{slog.String("session", sessionId)}
		if err != nil {
			logFields = append(logFields, slog.String("error", err.Error()))
		}

		s.logger.Error("Failed to validate refresh token", logFields...)
		return "", "", "", errors.New("invalid refresh token")
	}

	service, err := s.db.GetAppService(ctx, session.ServiceID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get service: %w", err)
	}

	err = s.db.UpdateUserSessionLastAuthenticatedAt(ctx, sessionId)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to update session last authentication date: %w", err)
	}

	// important that we try to generate the jwt BEFORE updating refresh token, if it fails then the client can't even retry
	exp := time.Now().Add(15 * time.Minute) //TODO: move to services db table

	token, err := jwt.NewBuilder().
		Audience([]string{service.JwtAudience}).
		Issuer(s.jwtIssuer).
		Subject(session.UserID).
		IssuedAt(time.Now()).
		Expiration(exp).
		Claim("session_id", sessionId).
		Claim("roles", roles).
		Claim("client_type", "user").
		Build()

	if err != nil {
		return "", "", "", fmt.Errorf("failed to build jwt: %w", err)
	}

	serviceId, err := utils.PgTypeUUIDToUUID(session.ServiceID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse session service id: %w", err)
	}

	signingKey, err := s.jwkManager.GetSigningKey(ctx, serviceId)

	if err != nil {
		return "", "", "", fmt.Errorf("failed to get signing key: %w", err)
	}

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.ES384, signingKey))

	if err != nil {
		return "", "", "", fmt.Errorf("failed to sign jwt: %w", err)
	}

	accessToken = string(signedToken)

	newRefreshToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to rotate refresh token: generate opaque token: %w", err)
	}

	hashedToken, err := s.tokenhasher.HashValue(newRefreshToken)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to hash new refresh token: %w", err)
	}

	if err = s.db.RotateUserSessionRefreshToken(ctx, db_gen.RotateUserSessionRefreshTokenParams{
		RefreshTokenHash: hashedToken,
		ID:               sessionId,
	}); err != nil {
		return "", "", "", fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, service.ApiDomain, nil
}

func (s *UserService) GetSession(ctx context.Context, sessionId string) (*db_gen.UserSession, error) {
	session, err := s.db.GetUserSession(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt.Time) {
		return nil, errors.New("the session is expired")
	}

	return &session, nil
}

func (s *UserService) validateRefreshToken(session *db_gen.UserSession, token string) (bool, error) {
	valid, err := s.tokenhasher.CompareHashAndValue(session.RefreshTokenHash, token)
	if err != nil {
		return false, err
	}

	if !valid {
		return false, errors.New("invalid token")
	}

	return true, nil
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
