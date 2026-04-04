package internal

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

func InitSystemData(
	lc fx.Lifecycle,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	settingService settingservice.SettingService,
	userService userservice.UserService,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sysStatus, err := sysStatusRepo.Get(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to load system status: %w", err)
			}
			if sysStatus.InstallationComplete {
				logger.Info("system data is already initialized")
				return nil
			}

			logger.Info("initializing system data...")
			var userCleanupFunc func()
			err = transaction.Execute(ctx, db, func(db database.Tx) error {
				sysStatus, err = sysStatusRepo.Get(ctx, db,
					bunex.SelectFor("UPDATE"),
				)
				if err != nil {
					return fmt.Errorf("failed to load system status: %w", err)
				}
				if sysStatus.InstallationComplete {
					return nil
				}

				if userCleanupFunc, err = userService.InitAdminUser(ctx, db); err != nil {
					return fmt.Errorf("failed to initialize admin user: %w", err)
				}

				if err := settingService.InitDefaults(ctx, db); err != nil {
					return fmt.Errorf("failed to initialize default settings: %w", err)
				}

				sysStatus.InstallationComplete = true
				sysStatus.UpdateVer++
				sysStatus.UpdatedAt = timeutil.NowUTC()
				err = sysStatusRepo.Upsert(ctx, db, sysStatus,
					entity.SystemStatusUpsertingConflictCols, entity.SystemStatusUpsertingUpdateCols)
				if err != nil {
					return fmt.Errorf("failed to save system status: %w", err)
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to initialize system data: %w", err)
			}

			if userCleanupFunc != nil {
				userCleanupFunc()
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
