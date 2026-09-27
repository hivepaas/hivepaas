package databackupserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

type service struct {
	backupRepoService   backupreposervice.Service
	schedJobExecService schedjobexecservice.Service

	// repoScope is the scope a repository's storage is resolved in: its own.
	repoScope func(ctx context.Context, db database.IDB, repo *entity.Setting) (*entity.ObjectScope, error)
	// findAppVolume is FindAppVolume, replaced in tests.
	findAppVolume func(ctx context.Context, db database.IDB, app *entity.App,
		volumeID string) (*databackupservice.AppVolume, error)

	clusterService clusterservice.Service
	settingService settingservice.Service
	volumeService  volumeservice.Service
}

func New(
	backupRepoService backupreposervice.Service,
	clusterService clusterservice.Service,
	schedJobExecService schedjobexecservice.Service,
	scopeService scopeservice.Service,
	settingService settingservice.Service,
	volumeService volumeservice.Service,
) databackupservice.Service {
	svc := &service{
		backupRepoService:   backupRepoService,
		schedJobExecService: schedJobExecService,
		clusterService:      clusterService,
		settingService:      settingService,
		volumeService:       volumeService,
	}
	svc.repoScope = func(ctx context.Context, db database.IDB, repo *entity.Setting) (*entity.ObjectScope, error) {
		scope, err := scopeService.LoadObjectScope(ctx, db, repo.Scope, repo.ObjectID, true)
		return scope, hperrors.Wrap(err)
	}
	svc.findAppVolume = svc.FindAppVolume
	return svc
}
