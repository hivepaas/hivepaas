package internal

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/repository"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/service/settingservice"
	"github.com/localpaas/localpaas/localpaas_app/service/userservice"
)

func CompleteInstallation(
	lc fx.Lifecycle,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	projectRepo repository.ProjectRepo,
	userService userservice.Service,
	settingService settingservice.Service,
	projectService projectservice.Service,
	logger logging.Logger,
) {
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
	var userCleanupFunc func()
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

		if userCleanupFunc, err = userService.InitAdminUser(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize admin user: %w", err)
		}

		if err := settingService.InitDefaults(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize default settings: %w", err)
		}

		if err := projectService.InitRootProject(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize root project: %w", err)
		}

		if err := installationInitDevProjects(ctx, db, projectRepo, projectService, logger); err != nil {
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

	if userCleanupFunc != nil {
		userCleanupFunc()
	}
	return nil
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
		return apperrors.Wrap(err)
	}

	_, _, err = projectService.SyncProject(ctx, db, projectA)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
