package settings

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/enums/instance_setting"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
)

func (s *SettingsService) GetValue(ctx context.Context, key instance_setting.InstanceSetting) (string, error) {
	row, err := s.db.GetSettingsByKey(ctx, key.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.New(errcode.SettingNotFound, err)
		}

		return "", errs.New(errcode.InternalServerError, err)
	}

	return row, nil
}

func (s *SettingsService) SaveValue(ctx context.Context, key instance_setting.InstanceSetting, value string) error {
	err := s.db.UpsertSetting(ctx, db_gen.UpsertSettingParams{
		Key:   key.String(),
		Value: value,
	})

	if err != nil {
		return err
	}

	return nil
}
