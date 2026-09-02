package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dblock"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
	"github.com/hivepaas/hivepaas/services/backup"
)

// SyncBackupRepo adopts whatever the repository actually says, in both directions the setting can
// drift: the options it is configured with, and the snapshots it holds.
//
// A repository is shared state. Anything holding its password can change it - the engine run by
// hand, another node, another install pointed at the same bucket - and none of that reaches the
// DB. This is what repairs that, and it is deliberately a read of the repository followed by a
// write to the DB, never the other way round: the repository is the source of truth.
func (uc *UC) SyncBackupRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.SyncBackupRepoReq,
) (*backuprepodto.SyncBackupRepoResp, error) {
	req.Type = currentSettingType

	// Same lock a cleanup takes, and for the same records. A sync never writes to the repository,
	// but it does reconcile: listing while a cleanup prunes and then reconciling afterwards would
	// re-add records for the snapshots that cleanup just expired, under fresh IDs.
	//
	// Refused rather than queued, as with cleanup - a sync that waits for a prune to finish would
	// only be reading a repository the prune has already reconciled.
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

	getResp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Reading the repository happens outside any transaction: it reaches the storage backend and
	// can take a while, and holding a transaction open across it buys nothing.
	syncResp, err := uc.backupRepoService.SyncRepo(ctx, uc.DB, &backupreposervice.SyncRepoReq{
		Scope:       req.Scope,
		RepoSetting: getResp.Data,
		RefObjects:  getResp.RefObjects,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var optionsChanged bool
	var snapshotResp *backupreposervice.SyncRepoSnapshotsResp
	err = transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		optionsChanged, err = uc.syncRepoOptions(ctx, db, getResp.Data, syncResp.Config)
		if err != nil {
			return hperrors.Wrap(err)
		}

		snapshotResp, err = uc.backupRepoService.SyncRepoSnapshots(ctx, db,
			&backupreposervice.SyncRepoSnapshotsReq{
				Scope:       req.Scope,
				RepoSetting: getResp.Data,
				Remaining:   syncResp.Snapshots,
			})
		return hperrors.Wrap(err)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.SyncBackupRepoResp{
		Data: backuprepodto.TransformSyncBackupRepo(getResp.Data.MustAsBackupRepo(), optionsChanged,
			len(syncResp.Snapshots), snapshotResp.Removed, snapshotResp.Added),
	}, nil
}

// syncRepoOptions stores what the repository reported, and reports whether that was a change.
//
// Nothing is pushed back to the repository here. A sync exists precisely because the repository
// won, so writing the old setting onto it would undo the change this is meant to pick up.
func (uc *UC) syncRepoOptions(
	ctx context.Context,
	db database.Tx,
	setting *entity.Setting,
	config *backup.RepoConfig,
) (bool, error) {
	repo, err := setting.AsBackupRepo()
	if err != nil {
		return false, hperrors.Wrap(err)
	}

	// applyRepoConfig mutates in place, so the comparison needs the values from before it runs.
	// A shallow copy is enough: it replaces the Retention pointer rather than writing through it.
	before := *repo
	applyRepoConfig(repo, config)
	if !repoConfigChanged(&before, repo) {
		return false, nil
	}

	if err := setting.SetData(repo); err != nil {
		return false, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()

	err = uc.SettingRepo.UpsertMulti(ctx, db, []*entity.Setting{setting},
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, nil
}

// repoConfigChanged compares only what a repository actually holds. Everything else on the setting
// - name, description, password, which storage it lives on - is the app's, not the repository's.
func repoConfigChanged(before, after *entity.BackupRepo) bool {
	if before.PackSize != after.PackSize || before.Compression != after.Compression {
		return true
	}
	switch {
	case before.Retention == nil && after.Retention == nil:
		return false
	case before.Retention == nil || after.Retention == nil:
		return true
	default:
		return *before.Retention != *after.Retention
	}
}
