package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kymppi/kuura/internal/constants"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/enums/instance_setting"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/models"
	"github.com/kymppi/kuura/internal/utils"
)

const KUURA_AUDIENCE = "kuura"

func serviceToModel(service db_gen.Service) (*models.AppService, error) {
	id, err := utils.PgTypeUUIDToUUID(service.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uuid from db: %w", err)
	}

	return &models.AppService{
		Id:                      id,
		JWTAudience:             service.JwtAudience,
		CreatedAt:               service.CreatedAt.Time,
		ModifiedAt:              service.ModifiedAt,
		Name:                    service.Name,
		Description:             service.Description.String,
		ContactName:             service.ContactName,
		ContactEmail:            service.ContactEmail,
		LoginRedirect:           service.LoginRedirect,
		AccessTokenDuration:     time.Duration(service.AccessTokenDuration) * time.Second,
		AccessTokenCookieDomain: service.ApiDomain,
		AccessTokenCookie:       service.AccessTokenCookie,
	}, nil
}

func (m *ServiceManager) CreateService(
	ctx context.Context,
	name string,
	jwtAudience string,
	apiDomain string,
	loginRedirect string,
) (*uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", err)
	}

	err = m.db.CreateAppService(ctx, db_gen.CreateAppServiceParams{
		ID:            utils.UUIDToPgType(id),
		JwtAudience:   jwtAudience,
		Name:          name,
		ApiDomain:     apiDomain,
		LoginRedirect: loginRedirect,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return &id, nil
}

func (m *ServiceManager) GetService(ctx context.Context, id uuid.UUID) (*models.AppService, error) {
	data, err := m.db.GetAppService(ctx, utils.UUIDToPgType(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return serviceToModel(data)
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
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	return utils.MapSliceE(data, serviceToModel)
}

func (m *ServiceManager) DeleteService(ctx context.Context, id uuid.UUID) error {
	err := m.db.DeleteAppService(ctx, utils.UUIDToPgType(id))

	return fmt.Errorf("failed to delete service: %w", err)
}

func (m *ServiceManager) UpdateService(ctx context.Context, service *models.AppService) error {
	if service == nil {
		return errors.New("provided service is nil")
	}

	return m.db.UpdateService(ctx, db_gen.UpdateServiceParams{
		ID:          utils.UUIDToPgType(service.Id),
		JwtAudience: service.JWTAudience,
		Name:        service.Name,
		Description: pgtype.Text{
			String: service.Description,
			Valid:  service.Description != "",
		},
		AccessTokenDuration: int32(service.AccessTokenDuration.Seconds()),
		AccessTokenCookie:   service.AccessTokenCookie,
		LoginRedirect:       service.LoginRedirect,
		ContactName:         service.ContactName,
		ContactEmail:        service.ContactEmail,
	})
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
			return m.VerifyServiceSettingsOrUpdate(ctx, existingServiceUUID, publicKuuraDomain)
		}
	}

	newServiceUUID, err := m.CreateService(ctx, "Kuura", KUURA_AUDIENCE, publicKuuraDomain, fmt.Sprintf("https://%s/home", publicKuuraDomain))
	if err != nil {
		return err
	}

	if err := m.settings.SaveValue(ctx, instance_setting.InternalServiceId, newServiceUUID.String()); err != nil {
		return err
	}

	return m.VerifyServiceSettingsOrUpdate(ctx, *newServiceUUID, publicKuuraDomain)
}

func (m *ServiceManager) VerifyServiceSettingsOrUpdate(ctx context.Context, existingServiceUUID uuid.UUID, publicKuuraDomain string) error {
	service, err := m.GetService(ctx, existingServiceUUID)
	if err != nil {
		return err
	}

	needsUpdate := false
	updatedService := *service

	if service.JWTAudience != KUURA_AUDIENCE {
		updatedService.JWTAudience = KUURA_AUDIENCE
		needsUpdate = true
	}

	if service.Name != "Kuura" {
		updatedService.Name = "Kuura"
		needsUpdate = true
	}

	expectedCookie := constants.INTERNAL_ACCESS_TOKEN_COOKIE
	if service.AccessTokenCookie != expectedCookie {
		updatedService.AccessTokenCookie = expectedCookie
		needsUpdate = true
	}

	expectedRedirect := fmt.Sprintf("https://%s/home", publicKuuraDomain)
	if service.LoginRedirect != expectedRedirect {
		updatedService.LoginRedirect = expectedRedirect
		needsUpdate = true
	}

	if needsUpdate {
		return m.UpdateService(ctx, &updatedService)
	}

	return nil
}
