package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/models"
	"github.com/kymppi/kuura/internal/utils"
)

type ServiceManager struct {
	db *db_gen.Queries
}

func NewServiceManager(databaseQueries *db_gen.Queries) *ServiceManager {
	return &ServiceManager{
		db: databaseQueries,
	}
}

func (m *ServiceManager) CreateService(ctx context.Context, name string, jwtAudience string, apiDomain string) (*uuid.UUID, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return nil, handleAppServiceError("CreateService", err, nil)
	}

	err = m.db.CreateAppService(ctx, db_gen.CreateAppServiceParams{
		ID:          utils.UUIDToPgType(id),
		JwtAudience: jwtAudience,
		Name:        name,
		ApiDomain:   apiDomain,
	})

	if err != nil {
		return nil, handleAppServiceError("CreateService", err, &id)
	}

	return &id, nil
}

func (m *ServiceManager) GetService(ctx context.Context, id uuid.UUID) (*models.AppService, error) {
	data, err := m.db.GetAppService(ctx, utils.UUIDToPgType(id))
	if err != nil {
		return nil, handleAppServiceError("GetService", err, &id)
	}
	return &models.AppService{
		Id:            id,
		JWTAudience:   data.JwtAudience,
		CreatedAt:     data.CreatedAt.Time,
		ModifiedAt:    data.ModifiedAt,
		Name:          data.Name,
		Description:   data.Description.String,
		ContactName:   data.ContactName,
		ContactEmail:  data.ContactEmail,
		LoginRedirect: data.LoginRedirect,
	}, nil
}

func (m *ServiceManager) GetServices(ctx context.Context) ([]*models.AppService, error) {
	data, err := m.db.GetAppServices(ctx)
	if err != nil {
		return nil, handleAppServiceError("GetServices", err, nil)
	}

	var result []*models.AppService
	for _, row := range data {
		result = append(result, &models.AppService{
			Id:            row.ID.Bytes,
			JWTAudience:   row.JwtAudience,
			CreatedAt:     row.CreatedAt.Time,
			ModifiedAt:    row.ModifiedAt,
			Name:          row.Name,
			Description:   row.Description.String,
			ContactName:   row.ContactName,
			ContactEmail:  row.ContactEmail,
			LoginRedirect: row.LoginRedirect,
		})
	}

	return result, nil
}

func (m *ServiceManager) DeleteService(ctx context.Context, id uuid.UUID) error {
	err := m.db.DeleteAppService(ctx, utils.UUIDToPgType(id))

	return handleAppServiceError("DeleteService", err, &id)
}

func handleAppServiceError(operation string, err error, id *uuid.UUID) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		baseMsg := operation
		if id != nil {
			baseMsg += fmt.Sprintf(" (ID: %s)", id)
		}

		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%s: already exists: %w", baseMsg, err)
		case pgerrcode.NoData:
			return fmt.Errorf("%s: not found: %w", baseMsg, err)
		default:
			return fmt.Errorf("%s: database error (code %s): %w", baseMsg, pgErr.Code, err)
		}
	}

	baseMsg := operation
	if id != nil {
		baseMsg += fmt.Sprintf(" (ID: %s)", id)
	}
	return fmt.Errorf("%s: error occurred: %w", baseMsg, err)
}
