package backupreposerviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
)

// SyncRepo reads a repository back so the caller can adopt what it finds there.
//
// This is the read half of what CleanupRepo does, without the prune: a repository can be changed
// by anything holding its password - the engine run by hand, another node, another install - and
// then the setting no longer describes what backups will actually do.
func (s *service) SyncRepo(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.SyncRepoReq,
) (resp *backupreposervice.SyncRepoResp, err error) {
	repo, err := req.RepoSetting.AsBackupRepo()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	engine, err := s.buildEngine(ctx, db, req.Scope, repo, req.RepoSetting.ID, req.RefObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The engine refuses to work on a repository it is not connected to, and a freshly started
	// container has never written the config file.
	if err = engine.ConnectRepo(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// A long-lived process stays connected, and a connected client answers from what it cached
	// about the repository. Without this the sync would report the options it already had and
	// silently find no drift - which is the one thing it exists to find.
	if err = engine.RefreshCache(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}

	config, err := engine.ReadRepoConfig(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	snapshots, err := readSnapshots(ctx, engine, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backupreposervice.SyncRepoResp{
		Config:    &config,
		Snapshots: snapshots,
	}, nil
}
