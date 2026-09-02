package backupreposerviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/services/backup"
)

func (s *service) CleanupRepo(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.CleanupRepoReq,
) (resp *backupreposervice.CleanupRepoResp, err error) {
	repo, err := req.RepoSetting.AsBackupRepo()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	engine, err := s.buildEngine(ctx, db, req.Scope, repo, req.RepoSetting.ID, req.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The engine refuses to work on a repository it is not connected to, and a freshly started
	// container has never written the config file.
	if err = engine.ConnectRepo(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if _, err = engine.Prune(ctx, toRetentionPolicy(repo.Retention)); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// One read, after the fact. The engine does not report which snapshots it expired, so the
	// caller works out what changed by comparing this against its own records.
	remaining, err := readSnapshots(ctx, engine, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backupreposervice.CleanupRepoResp{Remaining: remaining}, nil
}

func toRetentionPolicy(retention *entity.BackupRetentionPolicy) *backup.RetentionPolicy {
	if retention == nil {
		return nil
	}
	return &backup.RetentionPolicy{
		KeepLast:    retention.KeepLast,
		KeepDaily:   retention.KeepDaily,
		KeepWeekly:  retention.KeepWeekly,
		KeepMonthly: retention.KeepMonthly,
	}
}
