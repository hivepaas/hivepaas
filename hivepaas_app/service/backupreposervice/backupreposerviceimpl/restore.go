package backupreposerviceimpl

import (
	"context"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/backup"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// restoreServerUser is who a restore reads a repository server as. It writes
// nothing, so it needs no source of its own.
const restoreServerUser = "hivepaas@restore"

func (s *service) RestoreStream(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.RestoreStreamReq,
) error {
	repo, storage, err := s.targetStorage(ctx, db, &req.RepoTarget)
	if err != nil {
		return hperrors.Wrap(err)
	}
	read := func(snapshot string) func(engine backup.Engine) error {
		return func(engine backup.Engine) error {
			if req.OnConnected != nil {
				req.OnConnected()
			}
			_, err := engine.RestoreStream(ctx, snapshot, req.FileName, req.Stdout, nil)
			return hperrors.Wrap(err)
		}
	}
	if !repoServerNeeded(storage, true, "", "") {
		return hperrors.Wrap(s.runWith(ctx, repo, storage, s.buildCommandExecutor(false), read(req.SnapshotID)))
	}
	// The agent's commands bring back no stream: kopia runs here, through the
	// repository's server.
	root, err := s.snapshotRoot(ctx, db, &req.RepoTarget, req.SnapshotID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(s.throughServer(ctx, req.RepoSetting.ID, repo, storage, restoreServerUser, req.Progress,
		s.buildCommandExecutor(false), read(root)))
}

func (s *service) RestoreDirectory(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.RestoreDirectoryReq,
) error {
	repo, storage, err := s.targetStorage(ctx, db, &req.RepoTarget)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// The agent mounts the host root: the directory is expressed from inside it.
	dir := filepath.Join(volumeservice.HostPathPrefix, req.HostDir)
	restore := func(snapshot string) func(engine backup.Engine) error {
		return func(engine backup.Engine) error {
			_, err := engine.RestoreDirectory(ctx, snapshot, dir, &backup.RestoreOptions{Path: req.Path})
			return hperrors.Wrap(err)
		}
	}
	// The directory is written where it is: kopia runs on its node, whatever the
	// repository's storage would have chosen.
	onDataNode := onNodeExecutor(s.buildCommandExecutor(true), req.NodeID, req.NodeLabel)
	if !repoServerNeeded(storage, false, req.NodeID, req.NodeLabel) {
		return hperrors.Wrap(s.runWith(ctx, repo, storageOnDataNode(storage, req.NodeID, req.NodeLabel),
			onDataNode, restore(req.SnapshotID)))
	}
	root, err := s.snapshotRoot(ctx, db, &req.RepoTarget, req.SnapshotID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(s.throughServer(ctx, req.RepoSetting.ID, repo, storage, restoreServerUser, req.Progress,
		onDataNode, restore(root)))
}

// snapshotRoot is the snapshot's root object, read on the repository's own
// node. Through a repository server a snapshot is listed and restored by it:
// kopia shows a user the manifests of its own snapshots only, and the restore
// user took none - but it reads what any snapshot holds.
func (s *service) snapshotRoot(
	ctx context.Context,
	db database.IDB,
	target *backupreposervice.RepoTarget,
	snapshotID string,
) (string, error) {
	engine, err := s.buildTargetEngine(ctx, db, target)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if err = engine.ConnectRepo(ctx); err != nil {
		return "", hperrors.Wrap(err)
	}
	got, err := engine.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if got.Item == nil || got.Item.RootObjectID == "" {
		return "", hperrors.Wrap(backupmodel.ErrSnapshotNotFound).WithParam("Name", snapshotID)
	}
	return got.Item.RootObjectID, nil
}

func (s *service) ListEntries(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.ListEntriesReq,
) ([]backup.SnapshotEntry, error) {
	engine, err := s.buildTargetEngine(ctx, db, &req.RepoTarget)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = engine.ConnectRepo(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	entries, err := engine.ListEntries(ctx, req.SnapshotID, req.Path)
	return entries, hperrors.Wrap(err)
}
