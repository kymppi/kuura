package users

import (
	"context"
	"log/slog"

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
	s.logger.Info("User logged out", slog.String("session_id", sessionId), slog.String("uid", uid))

	return s.db.DeleteUserSession(ctx, db_gen.DeleteUserSessionParams{
		ID:     sessionId,
		UserID: uid,
	})
}
