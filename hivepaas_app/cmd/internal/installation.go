package internal

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
)

func CompleteInstallation(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	projectRepo repository.ProjectRepo,
	userService userservice.Service,
	settingService settingservice.Service,
	projectService projectservice.Service,
	logger logging.Logger,
) {
	stepEnabled := cfg.RunMode != config.RunModeUpdater
	if !stepEnabled {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sysStatus, err := sysStatusRepo.Get(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to load system status: %w", err)
			}
			config.Current.SystemInfo.NextStep = sysStatus.NextStep

			if sysStatus.NextStep == base.InstallationStepInitData {
				err = installationInitData(ctx, db, sysStatusRepo, projectRepo, userService,
					settingService, projectService, logger)
				if err != nil {
					return fmt.Errorf("failed to initialize system data: %w", err)
				}
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

func installationInitData(
	ctx context.Context,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	projectRepo repository.ProjectRepo,
	userService userservice.Service,
	settingService settingservice.Service,
	projectService projectservice.Service,
	logger logging.Logger,
) error {
	logger.Info("initializing system data...")
	var postInitFunc func() error
	err := transaction.Execute(ctx, db, func(db database.Tx) error {
		sysStatus, err := sysStatusRepo.Get(ctx, db,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return fmt.Errorf("failed to load system status: %w", err)
		}
		config.Current.SystemInfo.NextStep = sysStatus.NextStep
		if sysStatus.NextStep == "" {
			return nil
		}

		if err = userService.InitAdminUser(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize admin user: %w", err)
		}

		if err = settingService.InitDefaults(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize default settings: %w", err)
		}

		if postInitFunc, err = projectService.InitRootProject(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize root project: %w", err)
		}

		if err = installationInitDevProjects(ctx, db, projectRepo, projectService, logger); err != nil {
			return fmt.Errorf("failed to initialize dev projects: %w", err)
		}

		sysStatus.NextStep = base.InstallationStepObtainAppSSL
		sysStatus.UpdateVer++
		sysStatus.UpdatedAt = timeutil.NowUTC()
		err = sysStatusRepo.Upsert(ctx, db, sysStatus,
			entity.SystemStatusUpsertingConflictCols, entity.SystemStatusUpsertingUpdateCols)
		if err != nil {
			return fmt.Errorf("failed to save system status: %w", err)
		}
		config.Current.SystemInfo.NextStep = sysStatus.NextStep

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to initialize system data: %w", err)
	}

	if postInitFunc != nil {
		e := postInitFunc()
		if e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func installationInitDevProjects(
	ctx context.Context,
	db database.IDB,
	projectRepo repository.ProjectRepo,
	projectService projectservice.Service,
	logger logging.Logger,
) error {
	if !config.Current.IsDevEnv() {
		return nil
	}

	logger.Info("initializing development projects...")

	projectA, err := projectRepo.GetByKey(ctx, db, "project_a")
	if err != nil {
		return apperrors.New(err)
	}

	_, _, _, err = projectService.SyncProject(ctx, db, projectA) //nolint:dogsled
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
