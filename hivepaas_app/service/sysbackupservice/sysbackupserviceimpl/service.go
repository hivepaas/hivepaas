package sysbackupserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/scopeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice"
)

type service struct {
	settingRepo repository.SettingRepo

	backupRepoService backupreposervice.Service
	specService       specservice.Service

	// repoScope, dumpDB and workDir resolve a repository's scope, dump the
	// database into a file, and make a directory for a run; tests replace them.
	repoScope func(ctx context.Context, db database.IDB, repo *entity.Setting) (*entity.ObjectScope, error)
	dumpDB    func(ctx context.Context, path string, logStore *tasklog.Store) error
	workDir   func() (string, error)
}

func New(
	settingRepo repository.SettingRepo,
	backupRepoService backupreposervice.Service,
	scopeService scopeservice.Service,
	specService specservice.Service,
) sysbackupservice.Service {
	return &service{
		settingRepo:       settingRepo,
		backupRepoService: backupRepoService,
		specService:       specService,
		repoScope: func(ctx context.Context, db database.IDB, repo *entity.Setting) (*entity.ObjectScope, error) {
			scope, err := scopeService.LoadObjectScope(ctx, db, repo.Scope, repo.ObjectID, true)
			return scope, hperrors.Wrap(err)
		},
		dumpDB: dumpDB,
		workDir: func() (string, error) {
			dir, err := fileutil.CreateTempDirInAppPath("", "sys-backup-*", 0)
			return dir, hperrors.Wrap(err)
		},
	}
}
