package settings

import (
	"log/slog"

	"github.com/kymppi/kuura/internal/db_gen"
)

type SettingsService struct {
	logger *slog.Logger
	db     *db_gen.Queries
}

func NewSettingsService(logger *slog.Logger, db *db_gen.Queries) *SettingsService {
	return &SettingsService{
		logger: logger,
		db:     db,
	}
}
