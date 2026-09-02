package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

func (uc *UC) DeleteBackupRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.DeleteBackupRepoReq,
) (*backuprepodto.DeleteBackupRepoResp, error) {
	req.Type = currentSettingType
	_, err := uc.DeleteSetting(ctx, &req.DeleteSettingReq, &settings.DeleteSettingData{
		BeforePersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.DeleteSettingData,
			pData *settings.PersistingSettingDeletionData,
		) error {
			return uc.deleteRepoSnapshots(ctx, db, req.Scope, data.Setting.ID, pData)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.DeleteBackupRepoResp{}, nil
}

// deleteRepoSnapshots soft-deletes every snapshot setting created under the repo (linked through
// RefID, see newSnapshotSettings) along with their tags. Deleting the repo without them would
// leave snapshot settings pointing at a repo that no longer exists.
//
// NOTE: this only removes the app's records of the snapshots. It does not touch the data on the
// storage backend - deleting a Kopia repository means deleting the underlying storage objects,
// which is a separate concern from cleaning up these settings.
func (uc *UC) deleteRepoSnapshots(
	ctx context.Context,
	db database.Tx,
	scope *entity.ObjectScope,
	repoSettingID string,
	pData *settings.PersistingSettingDeletionData,
) error {
	snapshots, _, err := uc.SettingRepo.List(ctx, db, scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeBackupSnapshot),
		bunex.SelectWhere("setting.ref_id = ?", repoSettingID),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(snapshots) == 0 {
		return nil
	}

	timeNow := timeutil.NowUTC()
	snapshotIDs := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		snapshot.UpdateVer++
		snapshot.UpdatedAt = timeNow
		snapshot.DeletedAt = timeNow
		snapshotIDs = append(snapshotIDs, snapshot.ID)
	}
	pData.UpsertingSettings = append(pData.UpsertingSettings, snapshots...)

	if err := uc.TagRepo.DeleteAllByObjects(ctx, db, snapshotIDs); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
