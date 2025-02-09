package users

import (
	"log/slog"

	"github.com/kymppi/kuura/internal/db_gen"
)

type UserService struct {
	logger *slog.Logger
	db     *db_gen.Queries
}

func NewUserService(logger *slog.Logger, db *db_gen.Queries) *UserService {
	return &UserService{
		logger: logger,
		db:     db,
	}
}
