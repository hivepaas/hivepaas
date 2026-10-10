package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
	"github.com/hivepaas/hivepaas/services/backup"
)

// updateBackupRepoData carries what the post-commit step needs out of the transaction.
type updateBackupRepoData struct {
	Repo      *entity.BackupRepo
	SettingID string
	// Options and Retention are set only when the repository itself has to be told about the change.
	Options   *backup.RepoOptions
	Retention *entity.BackupRetentionPolicy
}

func (uc *UC) UpdateBackupRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.UpdateBackupRepoReq,
) (*backuprepodto.UpdateBackupRepoResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	data := &updateBackupRepoData{}
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName: req.Name,
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			settingData *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			backupRepo := pData.Setting.MustAsBackupRepo()
			// Custom validator that can only be called here as it requires an extra arg
			if err := req.Validate(backupRepo.Engine); err != nil {
				return hperrors.Wrap(err)
			}

			// Read what the repository is running with now, before Apply overwrites it.
			currentOptions := backup.NewRepoOptions(
				int(backupRepo.PackSize.MBytes()), backupRepo.Compression)
			newOptions := req.RepoOptions(backupRepo)
			currentRetention := backupRepo.Retention

			req.Apply(backupRepo)
			if err := pData.Setting.SetData(backupRepo); err != nil {
				return hperrors.Wrap(err)
			}

			if newOptions != currentOptions {
				data.Options = &newOptions
			}
			data.Retention = retentionToApply(currentRetention, backupRepo.Retention)
			if data.Options != nil || data.Retention != nil {
				data.Repo = backupRepo
				data.SettingID = pData.Setting.ID
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := uc.applyRepoOptions(ctx, req, data); err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.UpdateBackupRepoResp{}, nil
}

// retentionToApply is the retention the repository has to be told about: the new one when it
// differs from what the repository keeps by now. An update without one leaves the repository's.
func retentionToApply(current, updated *entity.BackupRetentionPolicy) *entity.BackupRetentionPolicy {
	if updated == nil || (current != nil && *current == *updated) {
		return nil
	}
	return updated
}

// applyRepoOptions pushes compression, pack size and retention to the repository, which is where
// the engine reads them from: stored in the DB alone they would have no effect on any later backup,
// and a sync would read the repository's back over them.
//
// This runs after the commit on purpose. Failing inside the transaction would throw away the rest
// of the update - a rename included - over a setting that the next successful update pushes again.
func (uc *UC) applyRepoOptions(
	ctx context.Context,
	req *backuprepodto.UpdateBackupRepoReq,
	data *updateBackupRepoData,
) error {
	if data.Options == nil && data.Retention == nil {
		return nil
	}

	err := uc.backupRepoService.ApplyRepoOptions(ctx, uc.DB, &backupreposervice.ApplyRepoOptionsReq{
		Scope:     req.Scope,
		Repo:      data.Repo,
		RepoID:    data.SettingID,
		Options:   data.Options,
		Retention: data.Retention,
	})
	if err != nil {
		// The settings are already saved, so reporting a plain failure would suggest nothing
		// changed. What actually happened is that the DB and the repository now disagree.
		return hperrors.Wrap(hperrors.ErrBackupRepoOptionsNotApplied).WithCause(err)
	}
	return nil
}
