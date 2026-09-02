package backupreposerviceimpl

import (
	"context"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
)

// SyncRepoSnapshots makes the stored snapshot records match what the repository actually holds.
//
// It does not just drop what a prune expired: a snapshot can also vanish or appear behind the
// app's back - somebody running the engine directly, a backup taken from another node - so
// diffing against the full remaining list repairs that drift in the same pass.
//
// It runs on the caller's transaction rather than opening its own, so everything it writes is
// committed or discarded as one.
func (s *service) SyncRepoSnapshots(
	ctx context.Context,
	db database.Tx,
	req *backupreposervice.SyncRepoSnapshotsReq,
) (resp *backupreposervice.SyncRepoSnapshotsResp, err error) {
	resp = &backupreposervice.SyncRepoSnapshotsResp{}

	stored, _, err := s.settingRepo.List(ctx, db, req.Scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeBackupSnapshot),
		bunex.SelectWhere("setting.ref_id = ?", req.RepoSetting.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Records are keyed by the engine's snapshot ID, which is what survives across runs.
	storedBySnapshotID := make(map[string]*entity.Setting, len(stored))
	for _, setting := range stored {
		storedBySnapshotID[setting.MustAsBackupSnapshot().ID] = setting
	}

	var missing []*backupreposervice.RepoSnapshot
	for _, item := range req.Remaining {
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
		resp.Removed = append(resp.Removed, setting.MustAsBackupSnapshot())
		setting.UpdateVer++
		setting.UpdatedAt = timeNow
		setting.DeletedAt = timeNow
		upsertingSettings = append(upsertingSettings, setting)
		goneIDs = append(goneIDs, setting.ID)
	}

	addedSettings, addedTags := NewSnapshotSettings(req.RepoSetting, missing)
	upsertingSettings = append(upsertingSettings, addedSettings...)

	err = s.settingRepo.UpsertMulti(ctx, db, upsertingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := s.tagRepo.DeleteAllByObjects(ctx, db, goneIDs); err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.tagRepo.UpsertMulti(ctx, db, addedTags,
		entity.TagUpsertingConflictCols, entity.TagUpsertingUpdateCols)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.Added = len(addedSettings)
	return resp, nil
}

// NewSnapshotSettings turns snapshots read from a repository into settings linked to it through
// RefID, plus the rows for their tags. Tags go to the tags table instead of into the snapshot data
// so they can be indexed and searched.
func NewSnapshotSettings(
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
