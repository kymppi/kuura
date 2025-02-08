package users

import (
	"log/slog"

	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/srp"
)

type UserService struct {
	logger *slog.Logger
	db     *db_gen.Queries
	srp    *srp.SRPOptions
}

func NewUserService(logger *slog.Logger, db *db_gen.Queries, srp *srp.SRPOptions) *UserService {
	return &UserService{
		logger: logger,
		db:     db,
		srp:    srp,
	}
}
