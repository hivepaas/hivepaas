package internal

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settinginitservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

func SystemInstallation(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	projectRepo repository.ProjectRepo,
	userService userservice.Service,
	settingInitService settinginitservice.Service,
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
			config.SetInstallationStep(sysStatus.NextStep)

			if sysStatus.NextStep == base.InstallationStepInitData {
				err = sysInstallationInitData(ctx, db, sysStatusRepo, projectRepo, userService,
					settingInitService, projectService, logger)
				if err != nil {
					return fmt.Errorf("failed to initialize system data: %w", err)
				}
				firstBootRan.Store(true)
			}
			// The data exists, whoever made it: what the first boot needed is
			// not kept on disk a moment longer.
			forgetFirstBootEnv(logger)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

// firstBootRan says this process created the installation's data, for what
// has to wait for services started after SystemInstallation.
var firstBootRan atomic.Bool

// forgetFirstBootEnv deletes the settings the installer left for the first boot.
// A failure is logged, not fatal: the boot has what it needs, and the next boot
// tries again.
func forgetFirstBootEnv(logger logging.Logger) {
	if err := config.RemoveFirstBootEnv(); err != nil {
		logger.Errorf("failed to remove the first-boot settings: %v", err)
	}
}

// DashboardCertOnFirstBoot asks for the dashboard's certificate after a boot
// that created the installation's data. It runs once the task queue has
// started, so the task is scheduled at once rather than found by the queue's
// next scan.
func DashboardCertOnFirstBoot(
	lc fx.Lifecycle,
	db *database.DB,
	getStartedService getstartedservice.Service,
	taskQueue queue.TaskQueue,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if firstBootRan.Load() {
				requestDashboardCert(ctx, db, getStartedService, taskQueue, logger)
			}
			return nil
		},
	})
}

func sysInstallationInitData(
	ctx context.Context,
	db *database.DB,
	sysStatusRepo repository.SystemStatusRepo,
	projectRepo repository.ProjectRepo,
	userService userservice.Service,
	settingInitService settinginitservice.Service,
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
		config.SetInstallationStep(sysStatus.NextStep)
		if sysStatus.NextStep == "" {
			return nil
		}

		if err = userService.InitAdminUser(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize admin user: %w", err)
		}

		if err = settingInitService.InitDefaultsWithTx(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize default settings: %w", err)
		}

		if err = settingInitService.InitSelfSignedCert(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize the self-signed certificate: %w", err)
		}

		if postInitFunc, err = projectService.InitRootProject(ctx, db); err != nil {
			return fmt.Errorf("failed to initialize root project: %w", err)
		}

		if err = sysInstallationInitDevProjects(ctx, db, projectRepo, projectService, logger); err != nil {
			return fmt.Errorf("failed to initialize dev projects: %w", err)
		}

		sysStatus.NextStep = base.InstallationStepGetStarted
		sysStatus.UpdateVer++
		sysStatus.UpdatedAt = timeutil.NowUTC()
		err = sysStatusRepo.Upsert(ctx, db, sysStatus,
			entity.SystemStatusUpsertingConflictCols, entity.SystemStatusUpsertingUpdateCols)
		if err != nil {
			return fmt.Errorf("failed to save system status: %w", err)
		}
		config.SetInstallationStep(sysStatus.NextStep)

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

func sysInstallationInitDevProjects(
	ctx context.Context,
	db database.IDB,
	projectRepo repository.ProjectRepo,
	projectService projectservice.Service,
	logger logging.Logger,
) error {
	if !config.Current().IsDevEnv() {
		return nil
	}

	logger.Info("initializing development projects...")

	projectA, err := projectRepo.GetByKey(ctx, db, "project_a")
	if err != nil {
		return hperrors.Wrap(err)
	}

	_, _, _, err = projectService.SyncProject(ctx, db, projectA) //nolint:dogsled
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

// requestDashboardCert asks for the dashboard's certificate once the first
// boot has created the data: where the domain already points here and port 80
// is open, it is there before anyone opens the dashboard. A failure is only
// logged - the Get started card asks again - and never stops the boot.
func requestDashboardCert(
	ctx context.Context,
	db *database.DB,
	getStartedService getstartedservice.Service,
	taskQueue queue.TaskQueue,
	logger logging.Logger,
) {
	certRequest, err := getStartedService.RequestDashboardCert(ctx, db, false)
	if err != nil {
		logger.Errorf("failed to request the dashboard's certificate: %v", err)
		return
	}
	if certRequest.NotAsked != "" {
		logger.Warnf("the dashboard's certificate was not requested: %s", certRequest.NotAsked)
	}
	if len(certRequest.Tasks) == 0 {
		return
	}
	if err = taskQueue.ScheduleTask(ctx, certRequest.Tasks...); err != nil {
		logger.Errorf("failed to schedule obtaining the dashboard's certificate: %v", err)
	}
}
