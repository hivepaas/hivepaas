package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice/backupreposerviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
	"github.com/hivepaas/hivepaas/services/backup"
)

func (uc *UC) CreateBackupRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.CreateBackupRepoReq,
) (*backuprepodto.CreateBackupRepoResp, error) {
	req.Type = currentSettingType
	backupRepo := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: backupRepo.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.Engine)

			// NOTE: this reaches out to the storage backend from inside the transaction. A failure
			// after this point rolls back the DB rows but leaves the repository on the storage;
			// re-running the request with `importExisting` then adopts it.
			initResp, err := uc.backupRepoService.InitRepo(ctx, db, &backupreposervice.InitRepoReq{
				Scope:          req.Scope,
				Repo:           backupRepo,
				RepoID:         pData.Setting.ID,
				RepoName:       req.Name,
				ImportExisting: req.ImportExisting,
				SyncData:       req.ImportExisting,
			})
			if err != nil {
				return hperrors.Wrap(err)
			}

			if initResp != nil {
				// The repository is what decides these in the end: an imported one keeps its own
				// settings, and a new one gets engine defaults for whatever the request left out.
				// Either way the setting has to say what the repository will really do.
				applyRepoConfig(backupRepo, initResp.Config)

				snapshotSettings, snapshotTags := backupreposerviceimpl.NewSnapshotSettings(pData.Setting, initResp.Snapshots)
				pData.UpsertingSettings = append(pData.UpsertingSettings, snapshotSettings...)
				pData.UpsertingTags = append(pData.UpsertingTags, snapshotTags...)
			}

			err = pData.Setting.SetData(backupRepo)
			if err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepodto.CreateBackupRepoResp{
		Data: resp.Data,
	}, nil
}

// applyRepoConfig adopts the settings read back from an imported repository.
//
// NOTE: the description is deliberately left alone. The engine keeps it as a per-client option
// rather than in the repository, so what an import reads back is a generated placeholder, not
// what whoever created the repository wrote.
func applyRepoConfig(repo *entity.BackupRepo, config *backup.RepoConfig) {
	if config == nil {
		return
	}

	repo.Compression = config.Compression
	repo.PackSize = unit.DataSize(config.PackSizeMB) * unit.MB
	if config.Retention != nil {
		repo.Retention = &entity.BackupRetentionPolicy{
			KeepLast:    config.Retention.KeepLast,
			KeepDaily:   config.Retention.KeepDaily,
			KeepWeekly:  config.Retention.KeepWeekly,
			KeepMonthly: config.Retention.KeepMonthly,
		}
	}
}
