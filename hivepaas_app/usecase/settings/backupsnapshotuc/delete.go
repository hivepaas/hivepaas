package backupsnapshotuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
	"github.com/hivepaas/hivepaas/services/backup"
)

// DeleteBackupSnapshot deletes a snapshot from its repository, then its record.
// A snapshot the repository no longer holds counts as deleted.
func (uc *UC) DeleteBackupSnapshot(
	ctx context.Context,
	auth *basedto.Auth,
	req *backupsnapshotdto.DeleteBackupSnapshotReq,
) (*backupsnapshotdto.DeleteBackupSnapshotResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		if err := uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
			return hperrors.Wrap(err)
		}
		record, repos, err := uc.findInReach(ctx, db, auth, req.Scope, req.ID, base.ActionTypeDelete)
		if err != nil {
			return hperrors.Wrap(err)
		}
		var repo *entity.Setting
		for _, r := range repos {
			if r.ID == record.RefID {
				repo = r
			}
		}
		if repo == nil || !repo.IsActive() {
			return hperrors.NewArgumentInvalid("repo").
				WithExtraDetail("the snapshot's repository must be active to delete from it")
		}
		snapshot, err := record.AsBackupSnapshot()
		if err != nil {
			return hperrors.Wrap(err)
		}
		repoScope, err := uc.ScopeService.LoadObjectScope(ctx, db, repo.Scope, repo.ObjectID, true)
		if err != nil {
			return hperrors.Wrap(err)
		}
		err = snapshotGoneIsDeleted(uc.backupRepoService.DeleteSnapshot(ctx, db, &backupreposervice.DeleteSnapshotReq{
			RepoTarget: backupreposervice.RepoTarget{Scope: repoScope, RepoSetting: repo},
			SnapshotID: snapshot.ID,
		}))
		if err != nil {
			return hperrors.Wrap(err)
		}

		timeNow := timeutil.NowUTC()
		record.UpdateVer++
		record.UpdatedAt = timeNow
		record.DeletedAt = timeNow
		err = uc.SettingRepo.UpsertMulti(ctx, db, []*entity.Setting{record},
			entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return hperrors.Wrap(uc.TagRepo.DeleteAllByObjects(ctx, db, []string{record.ID}))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupsnapshotdto.DeleteBackupSnapshotResp{}, nil
}

// snapshotGoneIsDeleted counts a snapshot the repository no longer holds -
// deleted with kopia directly, or by an earlier try - as deleted.
func snapshotGoneIsDeleted(err error) error {
	if errors.Is(err, backup.ErrSnapshotNotFound) {
		return nil
	}
	return err
}
