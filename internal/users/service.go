package users

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/models"
)

func (s *UserService) GetUser(ctx context.Context, uid string) (*models.User, error) {
	row, err := s.db.GetUser(ctx, uid)
	if err != nil {
		return nil, errs.New(errcode.UserNotFound, err)
	}

	obj := &models.User{
		Id:       row.ID,
		Username: row.Username,
	}

	if row.LastLoginAt.Valid {
		obj.LastLoginAt = &row.LastLoginAt.Time
	}

	return obj, nil
}

func (s *UserService) Logout(ctx context.Context, sessionId string, uid string) error {
	s.logger.Info("User logging out", slog.String("session_id", sessionId), slog.String("uid", uid))

	return s.db.DeleteUserSession(ctx, db_gen.DeleteUserSessionParams{
		ID:     sessionId,
		UserID: uid,
	})
}

func (s *UserService) LoginToService(ctx context.Context, uid string, serviceId uuid.UUID) (string, error) {
	code, err := generateOpaqueToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate opaque code: %w", err)
	}

	service, err := s.services.GetService(ctx, serviceId)
	if err != nil {
		return "", err
	}

	sessionId, err := s.CreateSessionForFutureUse(ctx, uid, serviceId)
	if err != nil {
		return "", err
	}

	hashedCode := hashCodeHMAC(code, s.tokenCodeHashingSecret)

	if err := s.db.InsertCodeToSessionTokenExchange(ctx, db_gen.InsertCodeToSessionTokenExchangeParams{
		SessionID: sessionId,
		ExpiresAt: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().Add(5 * time.Minute),
		},
		HashedCode: hashedCode,
	}); err != nil {
		return "", fmt.Errorf("failed to insert code to exchange: %w", err)
	}

	return fmt.Sprintf("%s?code=%s", service.LoginRedirect, code), nil
}
