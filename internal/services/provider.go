package services

import (
	"log/slog"

	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/settings"
)

type ServiceManager struct {
	logger   *slog.Logger
	db       *db_gen.Queries
	settings *settings.SettingsService
}

func NewServiceManager(logger *slog.Logger, databaseQueries *db_gen.Queries, settings *settings.SettingsService) *ServiceManager {
	return &ServiceManager{
		logger:   logger,
		db:       databaseQueries,
		settings: settings,
	}
}
