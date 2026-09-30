package backupreposerviceimpl

import (
	"context"
	"path/filepath"
	"strings"

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
	repo, storage, err := s.targetStorage(ctx, db, &req.RepoTarget)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	opts := &backup.BackupOptions{Tags: req.Tags, Description: req.Description, Source: req.Source}
	take := func(engine backup.Engine) (backupmodel.BackupResult, error) {
		if req.OnConnected != nil {
			req.OnConnected()
		}
		result, err := engine.BackupStream(ctx, req.Stdin, req.FileName, opts)
		return result, hperrors.Wrap(err)
	}

	var result backupmodel.BackupResult
	if !repoServerNeeded(storage, true, "", "") {
		result, err = s.backupWith(ctx, repo, storage, s.buildCommandExecutor(storage.StorageLocal != nil), take)
	} else {
		// The agent's commands carry no stdin: kopia runs here, through the
		// repository's server.
		result, err = s.backupThroughServer(ctx, req.RepoSetting.ID, repo, storage, req.Source, req.Progress,
			s.buildCommandExecutor(false), take)
	}
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
	repo, storage, err := s.targetStorage(ctx, db, &req.RepoTarget)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// The agent mounts the host root: the directory is expressed from inside it.
	dir := filepath.Join(volumeservice.HostPathPrefix, req.HostDir)
	opts := &backup.BackupOptions{Tags: req.Tags, Description: req.Description, Source: req.Source}
	take := func(engine backup.Engine) (backupmodel.BackupResult, error) {
		result, err := engine.BackupDirectory(ctx, dir, opts)
		return result, hperrors.Wrap(err)
	}
	// The directory is read where it is: kopia runs on its node, whatever the
	// repository's storage would have chosen.
	onDataNode := onNodeExecutor(s.buildCommandExecutor(true), req.NodeID, req.NodeLabel)

	var result backupmodel.BackupResult
	if !repoServerNeeded(storage, false, req.NodeID, req.NodeLabel) {
		result, err = s.backupWith(ctx, repo, storageOnDataNode(storage, req.NodeID, req.NodeLabel), onDataNode, take)
	} else {
		result, err = s.backupThroughServer(ctx, req.RepoSetting.ID, repo, storage, req.Source, req.Progress,
			onDataNode, take)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupreposervice.BackupResp{Snapshot: toRepoSnapshot(result.Item)}, nil
}

func (s *service) BackupLocalDirectory(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.BackupLocalDirectoryReq,
) (*backupreposervice.BackupResp, error) {
	repo, storage, err := s.targetStorage(ctx, db, &req.RepoTarget)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	opts := &backup.BackupOptions{Tags: req.Tags, Description: req.Description, Source: req.Source}
	take := func(engine backup.Engine) (backupmodel.BackupResult, error) {
		result, err := engine.BackupDirectory(ctx, req.Dir, opts)
		return result, hperrors.Wrap(err)
	}
	// The directory is this process's: kopia runs here, and reaches a repository
	// on a volume through its server.
	var result backupmodel.BackupResult
	if storage.StorageLocal == nil {
		result, err = s.backupWith(ctx, repo, storage, s.buildCommandExecutor(false), take)
	} else {
		result, err = s.backupThroughServer(ctx, req.RepoSetting.ID, repo, storage, req.Source, req.Progress,
			s.buildCommandExecutor(false), take)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupreposervice.BackupResp{Snapshot: toRepoSnapshot(result.Item)}, nil
}

// targetStorage is the repository and where it is.
func (s *service) targetStorage(
	ctx context.Context,
	db database.IDB,
	target *backupreposervice.RepoTarget,
) (*entity.BackupRepo, *backup.Storage, error) {
	repo, err := target.RepoSetting.AsBackupRepo()
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	storage, _, err := s.buildStorage(ctx, db, target.Scope, repo, target.RepoSetting.ID, target.RefObjects, "")
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return repo, storage, nil
}

// backupWith connects an engine on the storage, its commands run by exec, and
// takes the backup.
func (s *service) backupWith(
	ctx context.Context,
	repo *entity.BackupRepo,
	storage *backup.Storage,
	exec backupmodel.CommandExecutor,
	take func(backup.Engine) (backupmodel.BackupResult, error),
) (result backupmodel.BackupResult, err error) {
	err = s.runWith(ctx, repo, storage, exec, func(engine backup.Engine) error {
		result, err = take(engine)
		return err
	})
	return result, hperrors.Wrap(err)
}

// runWith connects an engine on the storage, its commands run by exec, and
// runs fn with it.
func (s *service) runWith(
	ctx context.Context,
	repo *entity.BackupRepo,
	storage *backup.Storage,
	exec backupmodel.CommandExecutor,
	fn func(backup.Engine) error,
) error {
	engine, err := backup.NewEngine(repo.Engine, storage, exec)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = engine.ConnectRepo(ctx); err != nil {
		return hperrors.Wrap(err)
	}
	return fn(engine)
}

// backupThroughServer takes the backup through a repository server run for it on
// the repository's node, as the identity of the snapshot's source, user@host.
func (s *service) backupThroughServer(
	ctx context.Context,
	repoID string,
	repo *entity.BackupRepo,
	storage *backup.Storage,
	source string,
	progress func(string),
	exec backupmodel.CommandExecutor,
	take func(backup.Engine) (backupmodel.BackupResult, error),
) (result backupmodel.BackupResult, err error) {
	username, _, _ := strings.Cut(source, ":")
	err = s.throughServer(ctx, repoID, repo, storage, username, progress, exec, func(engine backup.Engine) error {
		result, err = take(engine)
		return err
	})
	return result, hperrors.Wrap(err)
}

// throughServer runs fn with an engine reaching the repository through a server
// run for it on the repository's node, as username, user@host. The engine's
// connection goes with the session.
func (s *service) throughServer(
	ctx context.Context,
	repoID string,
	repo *entity.BackupRepo,
	storage *backup.Storage,
	username string,
	progress func(string),
	exec backupmodel.CommandExecutor,
	fn func(backup.Engine) error,
) error {
	say := func(msg string) {
		if progress != nil {
			progress(msg)
		}
	}
	node := storage.StorageLocal.NodeID
	if node == "" {
		node = storage.StorageLocal.NodeLabel
	}
	say("Starting the repository server on node " + node)
	err := s.withRepoServer(ctx, storage, username, repoID, func(client *backup.Storage) error {
		say("Repository server ready at " + client.StorageServer.URL)
		engine, err := backup.NewEngine(repo.Engine, client, exec)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = engine.ConnectRepo(ctx); err != nil {
			return hperrors.Wrap(err)
		}
		defer func() { _ = engine.DisconnectRepo(context.WithoutCancel(ctx)) }()
		return hperrors.Wrap(fn(engine))
	})
	say("Repository server stopped")
	return hperrors.Wrap(err)
}

// repoServerNeeded says whether a backup needs the repository's server: a
// repository on a volume is reached directly on its own node only - or on any
// node, when it is on storage they all share - and there by commands that carry
// no stdin.
func repoServerNeeded(storage *backup.Storage, stream bool, nodeID, nodeLabel string) bool {
	local := storage.StorageLocal
	if local == nil {
		return false
	}
	if stream {
		return true
	}
	if local.Shared {
		return false
	}
	onRepoNode := (local.NodeID != "" && local.NodeID == nodeID) ||
		(local.NodeLabel != "" && local.NodeLabel == nodeLabel)
	return !onRepoNode
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
) (*volumeservice.HostDir, error) {
	dir, err := volumeservice.ResolveHostDir(ctx, s.dockerManager, volume)
	return dir, hperrors.Wrap(err)
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

// storageOnDataNode is the repository's storage as reached from the data's node:
// a repository every node reaches is reached from there, so its engine's commands
// go there. Any other storage is itself.
func storageOnDataNode(storage *backup.Storage, nodeID, nodeLabel string) *backup.Storage {
	if storage.StorageLocal == nil || !storage.StorageLocal.Shared {
		return storage
	}
	moved := *storage
	local := *storage.StorageLocal
	local.NodeID, local.NodeLabel = nodeID, nodeLabel
	moved.StorageLocal = &local
	return &moved
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
