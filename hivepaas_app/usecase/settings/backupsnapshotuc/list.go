package backupsnapshotuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

func (uc *UC) ListBackupSnapshot(
	ctx context.Context,
	auth *basedto.Auth,
	req *backupsnapshotdto.ListBackupSnapshotReq,
) (*backupsnapshotdto.ListBackupSnapshotResp, error) {
	db := uc.DB
	if err := uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return nil, hperrors.Wrap(err)
	}
	reach, repos, err := uc.reachOf(ctx, db, auth, req.Scope, base.ActionTypeRead)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	filter := &snapshotFilter{RepoIDs: req.RepoIDs, AppIDs: req.AppIDs, Tags: req.Tags, Search: req.Search}
	if !req.FromDate.IsZero() {
		filter.From = req.FromDate.ToTime()
	}
	if !req.ToDate.IsZero() {
		filter.Before = req.ToDate.AddDate(0, 0, 1).ToTime()
	}
	records, paging, err := uc.SettingRepo.List(ctx, db, nil, &req.Paging, snapshotQueryOpts(reach, filter)...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	tags, err := uc.tagsOf(ctx, db, records)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	refs, err := uc.snapshotRefs(ctx, db, repos, tags)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &backupsnapshotdto.ListBackupSnapshotResp{
		Meta:  &basedto.ListMeta{Page: paging},
		Data:  make([]*backupsnapshotdto.BackupSnapshotResp, 0, len(records)),
		Repos: make([]*settings.BaseSettingResp, 0, len(repos)),
	}
	for _, record := range records {
		resp.Data = append(resp.Data, backupsnapshotdto.TransformBackupSnapshot(record, tags[record.ID], refs))
	}
	for _, repo := range repos {
		repoResp, err := settings.TransformSettingBase(repo)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp.Repos = append(resp.Repos, repoResp)
	}
	return resp, nil
}
