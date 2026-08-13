package internal

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

func InitSystemSettings(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	settingRepo repository.SettingRepo,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting system settings initialization...")
			err := sysSettingsLoadAppDomain(ctx, db, settingRepo, cfg)
			if err != nil {
				return fmt.Errorf("failed to load system settings: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping system settings initialization...")
			return nil
		},
	})
}

func sysSettingsLoadAppDomain(
	ctx context.Context,
	db *database.DB,
	settingRepo repository.SettingRepo,
	cfg *config.Config,
) (err error) {
	loaderFunc := func() (string, error) {
		dbHttpSettings, _, err := settingRepo.List(ctx, db, nil, nil,
			bunex.SelectJoin("JOIN apps AS app ON app.id = setting.object_id"),
			bunex.SelectJoin("JOIN projects AS project ON project.id = app.project_id"),
			bunex.SelectWhere("project.key = ?", base.HivepaasProjectKey),
			bunex.SelectWhere("app.key = ?", base.HivepaasAppKey),
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
			bunex.SelectLimit(1),
		)
		if err != nil {
			return "", apperrors.Wrap(err)
		}
		if len(dbHttpSettings) == 0 {
			return "", nil
		}
		httpSettings := dbHttpSettings[0].MustAsAppHttpSettings()
		if len(httpSettings.Domains) > 0 {
			return httpSettings.Domains[0].Domain, nil
		}
		return "", nil
	}

	config.SetAppDomainReloadFunc(loaderFunc)
	cfg.AppDomain, err = loaderFunc()
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
