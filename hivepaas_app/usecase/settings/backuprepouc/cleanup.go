package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dblock"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

// cleanupLockPrefix namespaces the advisory lock so it cannot collide with locks taken elsewhere.
const cleanupLockPrefix = "backup-repo:cleanup:"

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
	// the first one just finished.
	lock, acquired, err := dblock.TryAcquire(ctx, uc.DB, cleanupLockPrefix+req.ID)
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

	var syncResult repoSnapshotSyncResult
	err = transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		syncResult, err = uc.syncRepoSnapshots(ctx, db, req.Scope, resp.Data, cleanupResp.Remaining)
		return err
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.CleanupBackupRepoResp{
		Data: backuprepodto.TransformCleanupBackupRepo(len(cleanupResp.Remaining),
			syncResult.Removed, syncResult.Added),
	}, nil
}

// repoSnapshotSyncResult reports what the DB reconciliation changed. Removed carries the snapshots
// themselves, not just a count, so the caller can report which ones are gone.
type repoSnapshotSyncResult struct {
	Removed []*entity.BackupSnapshot
	Added   int
}

// syncRepoSnapshots makes the stored snapshot records match what the repository actually holds.
//
// It does not simply delete what the prune expired: a snapshot can also vanish or appear behind
// the app's back (someone running the engine directly, a backup taken by another node). Diffing
// against the full remaining list repairs that drift in the same pass.
//
// It runs on the caller's transaction rather than opening its own, so everything it writes is
// committed or discarded as one.
func (uc *UC) syncRepoSnapshots(
	ctx context.Context,
	db database.Tx,
	scope *entity.ObjectScope,
	repoSetting *entity.Setting,
	remaining []*backupreposervice.RepoSnapshot,
) (result repoSnapshotSyncResult, err error) {
	stored, _, err := uc.SettingRepo.List(ctx, db, scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeBackupSnapshot),
		bunex.SelectWhere("setting.ref_id = ?", repoSetting.ID),
	)
	if err != nil {
		return result, hperrors.Wrap(err)
	}

	// Records are keyed by the engine's snapshot ID, which is what survives across runs.
	storedBySnapshotID := make(map[string]*entity.Setting, len(stored))
	for _, setting := range stored {
		storedBySnapshotID[setting.MustAsBackupSnapshot().ID] = setting
	}

	var missing []*backupreposervice.RepoSnapshot
	for _, item := range remaining {
		if _, ok := storedBySnapshotID[item.Snapshot.ID]; ok {
			delete(storedBySnapshotID, item.Snapshot.ID)
			continue
		}
		missing = append(missing, item)
	}

	// Whatever is left in the map no longer exists in the repository.
	timeNow := timeutil.NowUTC()
	upsertingSettings := make([]*entity.Setting, 0, len(storedBySnapshotID)+len(missing))
	goneIDs := make([]string, 0, len(storedBySnapshotID))
	for _, setting := range storedBySnapshotID {
		// Read the snapshot before flipping DeletedAt: the parsed data is what the report needs.
		result.Removed = append(result.Removed, setting.MustAsBackupSnapshot())
		setting.UpdateVer++
		setting.UpdatedAt = timeNow
		setting.DeletedAt = timeNow
		upsertingSettings = append(upsertingSettings, setting)
		goneIDs = append(goneIDs, setting.ID)
	}

	addedSettings, addedTags := newSnapshotSettings(repoSetting, missing)
	upsertingSettings = append(upsertingSettings, addedSettings...)

	err = uc.SettingRepo.UpsertMulti(ctx, db, upsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return result, hperrors.Wrap(err)
	}

	if err := uc.TagRepo.DeleteAllByObjects(ctx, db, goneIDs); err != nil {
		return result, hperrors.Wrap(err)
	}

	err = uc.TagRepo.UpsertMulti(ctx, db, addedTags,
		entity.TagUpsertingConflictCols, entity.TagUpsertingUpdateCols)
	if err != nil {
		return result, hperrors.Wrap(err)
	}

	result.Added = len(addedSettings)
	return result, nil
}
