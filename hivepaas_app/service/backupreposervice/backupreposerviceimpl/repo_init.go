package backupreposerviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/services/backup"
)

func (s *service) InitRepo(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.InitRepoReq,
) (resp *backupreposervice.InitRepoResp, err error) {
	engine, err := s.buildEngine(ctx, db, req.Scope, req.Repo, req.RepoID, req.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if !req.ImportExisting {
		err = engine.InitRepo(ctx, &backup.InitRepoOptions{
			Description: req.Repo.Description,
			PackSizeMB:  int(req.Repo.PackSize.MBytes()),
			Compression: req.Repo.Compression,
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return &backupreposervice.InitRepoResp{}, nil
	}

	// Importing: the repository is already out there, only attach to it. Its own format settings
	// win, so the compression / pack size on the request are deliberately not applied.
	if err = engine.ConnectRepo(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp = &backupreposervice.InitRepoResp{}
	if !req.SyncData {
		return resp, nil
	}

	resp.Snapshots, err = readSnapshots(ctx, engine, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

// readSnapshots reads the snapshots held by a repository and maps them onto the app entities.
func readSnapshots(
	ctx context.Context,
	engine backup.Engine,
	opts *backup.ListSnapshotsOptions,
) ([]*backupreposervice.RepoSnapshot, error) {
	listResp, err := engine.ListSnapshots(ctx, opts)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	snapshots := make([]*backupreposervice.RepoSnapshot, 0, len(listResp.Items))
	for _, item := range listResp.Items {
		if item == nil {
			continue
		}
		snapshots = append(snapshots, &backupreposervice.RepoSnapshot{
			Snapshot: &entity.BackupSnapshot{
				ID:        item.ID,
				ShortID:   item.ShortID,
				Time:      item.Time,
				Paths:     item.Paths,
				Hostname:  item.Hostname,
				SizeBytes: item.SizeBytes,
			},
			Tags: item.Tags,
		})
	}
	return snapshots, nil
}
