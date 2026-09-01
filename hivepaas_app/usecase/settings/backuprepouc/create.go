package backuprepouc

import (
	"context"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
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
				snapshotSettings, snapshotTags := newSnapshotSettings(pData.Setting, initResp.Snapshots)
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

// newSnapshotSettings turns the snapshots read back from an imported repository into settings
// linked to that repository through RefID, plus the rows for their tags. Tags go to the tags
// table instead of into the snapshot data so they can be indexed and searched.
func newSnapshotSettings(
	repoSetting *entity.Setting,
	snapshots []*backupreposervice.RepoSnapshot,
) (settings []*entity.Setting, tags []*entity.Tag) {
	if len(snapshots) == 0 {
		return nil, nil
	}

	timeNow := time.Now()
	settings = make([]*entity.Setting, 0, len(snapshots))
	for _, item := range snapshots {
		snapshot := item.Snapshot
		setting := &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     repoSetting.Scope,
			ObjectID:  repoSetting.ObjectID,
			RefID:     repoSetting.ID,
			Type:      base.SettingTypeBackupSnapshot,
			Kind:      repoSetting.Kind,
			Status:    base.SettingStatusActive,
			Name:      snapshot.ShortID,
			Size:      snapshot.SizeBytes,
			Version:   entity.CurrentBackupSnapshotVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		setting.MustSetData(snapshot)
		settings = append(settings, setting)

		// A repository can hold the same tag on many snapshots, but the tags table is keyed by
		// (object_id, tag), so duplicates only matter within one snapshot.
		seen := make(map[string]struct{}, len(item.Tags))
		index := 0
		for _, tag := range item.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			tags = append(tags, &entity.Tag{
				ObjectID: setting.ID,
				Tag:      tag,
				Index:    index,
			})
			index++
		}
	}
	return settings, tags
}
