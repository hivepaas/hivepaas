package backupsnapshotuc

import (
	"context"
	"io"
	"path"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
	"github.com/hivepaas/hivepaas/services/backup"
)

// DownloadBackupSnapshotFile gives a file of a snapshot the view reaches, to a
// person who may write on the snapshot's owner: the file is the data itself.
// The download is recorded before anything is read.
func (uc *UC) DownloadBackupSnapshotFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *backupsnapshotdto.DownloadBackupSnapshotFileReq,
) (*backupsnapshotdto.DownloadBackupSnapshotFileResp, error) {
	db := uc.DB
	if err := uc.ScopeService.LoadObjectScopeData(ctx, db, req.Scope); err != nil {
		return nil, hperrors.Wrap(err)
	}
	record, repos, err := uc.findInReach(ctx, db, auth, req.Scope, req.ID, base.ActionTypeWrite)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	target, snapshot, err := uc.snapshotTarget(ctx, db, record, repos)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dir, name := path.Split(req.Path)
	entries, err := uc.backupRepoService.ListEntries(ctx, db, &backupreposervice.ListEntriesReq{
		RepoTarget: *target, SnapshotID: snapshot.ID, Path: dir,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry, err := downloadableEntry(entries, name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	tags, err := uc.tagsOf(ctx, db, []*entity.Setting{record})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	scope, objectID := downloadAuditScope(entity.ParseDataBackupSnapshotTags(tags[record.ID]).AppID, target.RepoSetting)
	err = auditservice.RecordAllowed(ctx, uc.AuditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeBackupDownload,
		Scope:    scope,
		ObjectID: objectID,
		Source:   base.AuditLogSourceAPIAction,
		Auth:     auth,
		ResType:  base.ResourceTypeSetting,
		ResID:    record.ID,
		ResName:  "snapshot " + snapshot.ShortID,
		Detail: auditdetail.New().Set("repository", target.RepoSetting.Name).Set("path", req.Path).
			Set("sizeBytes", entry.SizeBytes).String(),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backupsnapshotdto.DownloadBackupSnapshotFileResp{
		FileName:  name,
		SizeBytes: entry.SizeBytes,
		Write: func(ctx context.Context, w io.Writer) error {
			err := uc.backupRepoService.RestoreStream(ctx, uc.DB, &backupreposervice.RestoreStreamReq{
				RepoTarget: *target, SnapshotID: snapshot.ID, FileName: req.Path, Stdout: w,
			})
			return hperrors.Wrap(err)
		},
	}, nil
}

// downloadableEntry is the file name of a directory's entries; a directory is
// restored, not downloaded.
func downloadableEntry(entries []backup.SnapshotEntry, name string) (*backup.SnapshotEntry, error) {
	for i := range entries {
		if entries[i].Name != name {
			continue
		}
		if entries[i].Dir {
			return nil, hperrors.NewArgumentInvalid("path").
				WithExtraDetail("%s is a directory: a directory is restored, not downloaded", name)
		}
		return &entries[i], nil
	}
	return nil, hperrors.NewNotFound("File in the snapshot")
}

// downloadAuditScope is where a download is recorded: against the snapshot's
// app, or its repository's scope when no app owns it.
func downloadAuditScope(appID string, repo *entity.Setting) (base.ObjectScopeType, string) {
	if appID != "" {
		return base.ObjectScopeApp, appID
	}
	return repo.Scope, repo.ObjectID
}
