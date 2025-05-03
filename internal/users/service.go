package users

import (
	"context"

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
