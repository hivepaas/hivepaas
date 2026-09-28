package backupsnapshotuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

// ListBackupSnapshotEntries lists a directory of a snapshot the view reaches.
func (uc *UC) ListBackupSnapshotEntries(
	ctx context.Context,
	auth *basedto.Auth,
	req *backupsnapshotdto.ListBackupSnapshotEntriesReq,
) (*backupsnapshotdto.ListBackupSnapshotEntriesResp, error) {
	db := uc.DB
	if err := uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return nil, hperrors.Wrap(err)
	}
	record, repos, err := uc.findInReach(ctx, db, auth, req.Scope, req.ID, base.ActionTypeRead)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	target, snapshot, err := uc.snapshotTarget(ctx, db, record, repos)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entries, err := uc.backupRepoService.ListEntries(ctx, db, &backupreposervice.ListEntriesReq{
		RepoTarget: *target, SnapshotID: snapshot.ID, Path: req.Path,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &backupsnapshotdto.ListBackupSnapshotEntriesResp{
		Data: make([]*backupsnapshotdto.BackupSnapshotEntryResp, 0, len(entries)),
	}
	for _, entry := range entries {
		resp.Data = append(resp.Data, &backupsnapshotdto.BackupSnapshotEntryResp{
			Name: entry.Name, Dir: entry.Dir, SizeBytes: entry.SizeBytes,
		})
	}
	return resp, nil
}

// snapshotTarget is the repository a snapshot of the view is in, which must be
// active to be read, and the snapshot.
func (uc *UC) snapshotTarget(
	ctx context.Context,
	db database.IDB,
	record *entity.Setting,
	repos []*entity.Setting,
) (*backupreposervice.RepoTarget, *entity.BackupSnapshot, error) {
	var repo *entity.Setting
	for _, r := range repos {
		if r.ID == record.RefID {
			repo = r
		}
	}
	if repo == nil || !repo.IsActive() {
		return nil, nil, hperrors.NewArgumentInvalid("repo").
			WithExtraDetail("the snapshot's repository must be active to read it")
	}
	snapshot, err := record.AsBackupSnapshot()
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	repoScope, err := uc.ScopeService.LoadObjectScope(ctx, db, repo.Scope, repo.ObjectID, true)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return &backupreposervice.RepoTarget{Scope: repoScope, RepoSetting: repo}, snapshot, nil
}
