package backupreposerviceimpl

import (
	"context"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/backup"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func (s *service) BackupStream(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.BackupStreamReq,
) (*backupreposervice.BackupResp, error) {
	engine, err := s.buildTargetEngine(ctx, db, &req.RepoTarget)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = engine.ConnectRepo(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	result, err := engine.BackupStream(ctx, req.Stdin, req.FileName, &backup.BackupOptions{Tags: req.Tags})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupreposervice.BackupResp{Snapshot: toRepoSnapshot(result.Item)}, nil
}

func (s *service) BackupDirectory(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.BackupDirectoryReq,
) (*backupreposervice.BackupResp, error) {
	repo, err := req.RepoSetting.AsBackupRepo()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	storage, _, err := s.buildStorage(ctx, db, req.Scope, repo, req.RepoSetting.ID, req.RefObjects, "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = checkRepoReachableFrom(storage, req.NodeID, req.NodeLabel); err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The directory is read where it is: the engine runs on its node, whatever the
	// repository's storage would have chosen.
	engine, err := backup.NewEngine(repo.Engine, storage,
		onNodeExecutor(s.buildCommandExecutor(true), req.NodeID, req.NodeLabel))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = engine.ConnectRepo(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The agent mounts the host root: the directory is expressed from inside it.
	dir := filepath.Join(volumeservice.HostPathPrefix, req.HostDir)
	result, err := engine.BackupDirectory(ctx, dir, &backup.BackupOptions{Tags: req.Tags})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupreposervice.BackupResp{Snapshot: toRepoSnapshot(result.Item)}, nil
}

func (s *service) DeleteSnapshot(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.DeleteSnapshotReq,
) error {
	engine, err := s.buildTargetEngine(ctx, db, &req.RepoTarget)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = engine.ConnectRepo(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	_, err = engine.DeleteSnapshot(ctx, req.SnapshotID)
	return hperrors.Wrap(err)
}

func (s *service) VolumeHostDir(
	ctx context.Context,
	volume *entity.Setting,
) (*backupreposervice.VolumeHostDir, error) {
	clusterVolume, err := volume.AsClusterVolume()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dir, err := s.resolveVolumeHostPath(ctx, volume, clusterVolume)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupreposervice.VolumeHostDir{
		Dir: dir, NodeID: clusterVolume.NodeID, NodeLabel: clusterVolume.NodeLabel,
	}, nil
}

func (s *service) buildTargetEngine(
	ctx context.Context,
	db database.IDB,
	target *backupreposervice.RepoTarget,
) (backup.Engine, error) {
	repo, err := target.RepoSetting.AsBackupRepo()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	engine, err := s.buildEngine(ctx, db, target.Scope, repo, target.RepoSetting.ID, target.RefObjects)
	return engine, hperrors.Wrap(err)
}

// onNodeExecutor runs a command the engine gives no node on the node asked for.
func onNodeExecutor(base backupmodel.CommandExecutor, nodeID, nodeLabel string) backupmodel.CommandExecutor {
	return func(ctx context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		if req.NodeID == "" && req.NodeLabel == "" {
			req.NodeID, req.NodeLabel = nodeID, nodeLabel
		}
		return base(ctx, req)
	}
}

// checkRepoReachableFrom refuses a repository on a volume of another node than the
// one a directory is read on: it is reachable there only, until the repository
// server makes it reachable from anywhere.
func checkRepoReachableFrom(storage *backup.Storage, nodeID, nodeLabel string) error {
	local := storage.StorageLocal
	if local == nil {
		return nil
	}
	if (local.NodeID != "" && local.NodeID == nodeID) || (local.NodeLabel != "" && local.NodeLabel == nodeLabel) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrNotImplemented).WithExtraDetail(
		"the repository's volume is on another node than the data: backing it up there is not supported yet")
}

// toRepoSnapshot is a snapshot the engine made, as the repository's snapshots read.
func toRepoSnapshot(item *backupmodel.Snapshot) *backupreposervice.RepoSnapshot {
	if item == nil {
		return nil
	}
	return &backupreposervice.RepoSnapshot{
		Snapshot: &entity.BackupSnapshot{
			ID:        item.ID,
			ShortID:   item.ShortID,
			Time:      item.Time,
			Paths:     item.Paths,
			Hostname:  item.Hostname,
			SizeBytes: item.SizeBytes,
		},
		Tags: item.Tags,
	}
}
