package traefiksettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	db     *database.DB
	logger logging.Logger

	settingRepo repository.SettingRepo
	taskRepo    repository.TaskRepo

	auditService auditservice.Service
	hpAppService hpappservice.Service
	// probationService is confirm-or-revert, shared with the HivePaaS settings.
	probationService settingsprobationservice.Service
	traefikService   traefikservice.Service

	dockerManager docker.Manager
}

func New(
	db *database.DB,
	logger logging.Logger,

	settingRepo repository.SettingRepo,
	taskRepo repository.TaskRepo,

	auditService auditservice.Service,
	hpAppService hpappservice.Service,
	probationService settingsprobationservice.Service,
	traefikService traefikservice.Service,

	dockerManager docker.Manager,
) *UC {
	return &UC{
		db:     db,
		logger: logger,

		settingRepo: settingRepo,
		taskRepo:    taskRepo,

		auditService:     auditService,
		hpAppService:     hpAppService,
		probationService: probationService,
		traefikService:   traefikService,

		dockerManager: dockerManager,
	}
}

// traefikAppID is the object traefik's trials are filed under.
//
// Traefik is a real app in the HivePaaS project, so its probations get their own
// object id rather than borrowing the main app's. That keeps FindPending's
// (object, setting type) pair unambiguous even if traefik ever grows a second
// kind of change that can be put on trial.
func (uc *UC) traefikAppID(
	ctx context.Context,
	db database.IDB,
	extraOpts ...bunex.SelectQueryOption,
) (string, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	}
	opts = append(opts, extraOpts...)

	app, err := uc.hpAppService.LoadAppByKey(ctx, db, base.HivepaasTraefikKey, opts...)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return app.ID, nil
}
