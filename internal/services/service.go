package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/enums/instance_setting"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/models"
	"github.com/kymppi/kuura/internal/utils"
)

const KUURA_AUDIENCE = "kuura"

func (m *ServiceManager) CreateService(
	ctx context.Context,
	name string,
	jwtAudience string,
	apiDomain string,
	loginRedirect string,
) (*uuid.UUID, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return nil, handleAppServiceError("CreateService", err, nil)
	}

	err = m.db.CreateAppService(ctx, db_gen.CreateAppServiceParams{
		ID:            utils.UUIDToPgType(id),
		JwtAudience:   jwtAudience,
		Name:          name,
		ApiDomain:     apiDomain,
		LoginRedirect: loginRedirect,
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
		Id:                      id,
		JWTAudience:             data.JwtAudience,
		CreatedAt:               data.CreatedAt.Time,
		ModifiedAt:              data.ModifiedAt,
		Name:                    data.Name,
		Description:             data.Description.String,
		ContactName:             data.ContactName,
		ContactEmail:            data.ContactEmail,
		LoginRedirect:           data.LoginRedirect,
		AccessTokenDuration:     time.Duration(data.AccessTokenDuration) * time.Second,
		AccessTokenCookieDomain: data.ApiDomain,
		AccessTokenCookie:       data.AccessTokenCookie,
	}, nil
}

func (m *ServiceManager) GetInternalKuuraService(ctx context.Context) (*models.AppService, error) {
	internalServiceId, err := m.settings.GetValue(ctx, instance_setting.InternalServiceId)
	if err != nil {
		return nil, err
	}

	internalServiceUUID, err := uuid.Parse(internalServiceId)
	if err != nil {
		return nil, err
	}

	data, err := m.GetService(ctx, internalServiceUUID)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *ServiceManager) GetServices(ctx context.Context) ([]*models.AppService, error) {
	data, err := m.db.GetAppServices(ctx)
	if err != nil {
		return nil, handleAppServiceError("GetServices", err, nil)
	}

	var result []*models.AppService
	for _, row := range data {
		result = append(result, &models.AppService{
			Id:                      row.ID.Bytes,
			JWTAudience:             row.JwtAudience,
			CreatedAt:               row.CreatedAt.Time,
			ModifiedAt:              row.ModifiedAt,
			Name:                    row.Name,
			Description:             row.Description.String,
			ContactName:             row.ContactName,
			ContactEmail:            row.ContactEmail,
			LoginRedirect:           row.LoginRedirect,
			AccessTokenDuration:     time.Duration(row.AccessTokenDuration) * time.Second,
			AccessTokenCookieDomain: row.ApiDomain,
			AccessTokenCookie:       row.AccessTokenCookie,
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

func (m *ServiceManager) CreateInternalServiceIfNotExists(ctx context.Context, publicKuuraDomain string) error {
	existingServiceId, err := m.settings.GetValue(ctx, instance_setting.InternalServiceId)
	if err != nil && !errs.IsErrorCode(err, errcode.SettingNotFound) {
		return err
	}

	if existingServiceId != "" {
		existingServiceUUID, err := uuid.Parse(existingServiceId)
		if err != nil {
			return err
		}

		possibleService, err := m.GetService(ctx, existingServiceUUID)
		if err != nil && !errs.IsErrorCode(err, errcode.ServiceNotFound) {
			return err
		}

		if possibleService != nil {
			// service exists
			//TODO: check if domain is the same, if not -> update
			return nil
		}
	}

	newServiceUUID, err := m.CreateService(ctx, "Kuura", KUURA_AUDIENCE, publicKuuraDomain, fmt.Sprintf("https://%s/home", publicKuuraDomain))
	if err != nil {
		return err
	}

	if err := m.settings.SaveValue(ctx, instance_setting.InternalServiceId, newServiceUUID.String()); err != nil {
		return err
	}

	return nil
}
