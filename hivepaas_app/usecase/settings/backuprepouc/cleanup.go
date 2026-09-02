package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dblock"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

func (uc *UC) CleanupBackupRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.CleanupBackupRepoReq,
) (*backuprepodto.CleanupBackupRepoResp, error) {
	req.Type = currentSettingType

	// Pruning rewrites the repository and can run for minutes, far too long to keep a transaction
	// - and therefore a row lock - open for. An advisory lock belongs to the session instead, so
	// it holds across the commits below while no transaction stays open during the prune itself.
	//
	// Taking it without waiting is deliberate: a queued second cleanup would only redo the work
	// the first one just finished. The scheduled job takes the same lock, so a manual run and a
	// scheduled one cannot collide either.
	lock, acquired, err := dblock.TryAcquire(ctx, uc.DB, backupreposervice.RepoLockName(req.ID))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !acquired {
		return nil, hperrors.Wrap(hperrors.ErrBackupRepoCleanupInProgress).WithNTParam("Name", req.ID)
	}
	defer func() {
		_ = lock.Release(ctx)
	}()

	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// NOTE: this prunes the repository before the DB is touched, and outside any transaction. The
	// repository is the source of truth for what exists, so a failure while reconciling below
	// leaves records that the next cleanup reconciles again rather than data that cannot be
	// recovered.
	cleanupResp, err := uc.backupRepoService.CleanupRepo(ctx, uc.DB, &backupreposervice.CleanupRepoReq{
		Scope:       req.Scope,
		RepoSetting: resp.Data,
		RefObjects:  resp.RefObjects,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var syncResp *backupreposervice.SyncRepoSnapshotsResp
	err = transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		syncResp, err = uc.backupRepoService.SyncRepoSnapshots(ctx, db,
			&backupreposervice.SyncRepoSnapshotsReq{
				Scope:       req.Scope,
				RepoSetting: resp.Data,
				Remaining:   cleanupResp.Remaining,
			})
		return hperrors.Wrap(err)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.CleanupBackupRepoResp{
		Data: backuprepodto.TransformCleanupBackupRepo(len(cleanupResp.Remaining),
			syncResp.Removed, syncResp.Added),
	}, nil
}
