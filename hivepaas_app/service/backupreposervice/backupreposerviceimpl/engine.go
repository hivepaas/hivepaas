package backupreposerviceimpl

import (
	"context"
	"path"
	"path/filepath"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/backup"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

const (
	// engineConfigDir holds one engine config file per repository. The engines keep their
	// repository connection state in that file, so sharing a single file across repositories
	// would make every operation act on whichever repository connected last.
	engineConfigDir      = "/tmp/hivepaas/backup-repos"
	engineConfigFileName = "repository.config"
)

// buildEngine resolves everything the backup engine needs to talk to a repository: where the
// repository lives, the credentials to reach it, and where its commands have to run.
func (s *service) buildEngine(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	repo *entity.BackupRepo,
	repoID string,
	refObjects *entity.RefObjects,
) (backup.Engine, error) {
	storage, runOnNode, err := s.buildStorage(ctx, db, scope, repo, repoID, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	engine, err := backup.NewEngine(repo.Engine, storage, s.buildCommandExecutor(runOnNode))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return engine, nil
}

// buildCommandExecutor picks where engine commands run. Repositories backed by cloud storage are
// reachable from anywhere, so they run in-process. A repository sitting on a node-local volume can
// only be reached from that node, so its commands go through the agent running there.
func (s *service) buildCommandExecutor(runOnNode bool) backupmodel.CommandExecutor {
	if !runOnNode {
		return backupmodel.DefaultCommandExecutor
	}

	return func(ctx context.Context, req *backupmodel.CommandExecReq) (*backupmodel.CommandExecResp, error) {
		if req.Stdin != nil {
			// The agent command protocol has no stdin channel, so streaming backups cannot be
			// served this way. Failing loudly beats silently backing up an empty stream.
			return nil, hperrors.Wrap(hperrors.ErrNotImplemented).
				WithExtraDetail("streaming to a volume-backed backup repository is not supported")
		}

		resp, err := s.nodeExecService.ExecCommand(ctx, &nodeexecservice.CommandExecReq{
			NodeID:    req.NodeID,
			NodeLabel: req.NodeLabel,
			CommandExecOpts: &nodeexecservice.CommandExecOpts{
				Command:    req.Command,
				Env:        req.Env,
				WorkingDir: req.WorkingDir,
				Stdout:     req.Stdout,
				Stderr:     req.Stderr,
			},
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return &backupmodel.CommandExecResp{ExitCode: resp.ExitCode}, nil
	}
}

// buildStorage turns the repo's referenced cloud storage or volume into an engine storage config.
// runOnNode reports whether the resulting location is only reachable from a specific node.
func (s *service) buildStorage(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	repo *entity.BackupRepo,
	repoID string,
	refObjects *entity.RefObjects,
) (storage *backup.Storage, runOnNode bool, err error) {
	hasCloudStorage := repo.CloudStorage.ID != ""
	hasVolume := repo.Volume.ID != ""
	switch {
	case hasCloudStorage && hasVolume:
		return nil, false, hperrors.Wrap(hperrors.ErrBackupRepoStorageAmbiguous)
	case !hasCloudStorage && !hasVolume:
		return nil, false, hperrors.Wrap(hperrors.ErrBackupRepoStorageRequired)
	}

	password, err := repo.Password.GetPlain()
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	if password == "" {
		return nil, false, hperrors.Wrap(hperrors.ErrBackupRepoPasswordRequired)
	}

	// LoadRefObjectsByIDs allocates the container when it is nil and skips the IDs already in it,
	// so whatever the caller passed in can go straight through.
	err = s.settingService.LoadRefObjectsByIDs(ctx, db, &refObjects, scope, true, repo.GetRefObjectIDs())
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}

	storage = &backup.Storage{
		RepositoryPassword: password,
		ConfigFile:         engineConfigFilePath(repoID),
	}

	if hasCloudStorage {
		storage.StorageS3, err = s.buildS3Storage(repo, refObjects)
		if err != nil {
			return nil, false, hperrors.Wrap(err)
		}
		return storage, false, nil
	}

	storage.StorageLocal, err = s.buildLocalStorage(ctx, repo, refObjects)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	return storage, true, nil
}

func (s *service) buildS3Storage(
	repo *entity.BackupRepo,
	refObjects *entity.RefObjects,
) (*backup.StorageS3, error) {
	setting := refObjects.RefSettings[repo.CloudStorage.ID]
	if setting == nil {
		return nil, hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("ID", repo.CloudStorage.ID)
	}

	cloudStorage, err := setting.AsCloudStorage()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if cloudStorage.S3 == nil || cloudStorage.S3.CloudProviderAWS == nil {
		return nil, hperrors.Wrap(hperrors.ErrStorageTypeUnsupported)
	}
	if err := cloudStorage.Decrypt(); err != nil {
		return nil, hperrors.Wrap(err)
	}

	secretKey, err := cloudStorage.S3.SecretKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backup.StorageS3{
		Endpoint:  cloudStorage.S3.Endpoint,
		Region:    gofn.Coalesce(cloudStorage.S3.Region, cloudStorage.S3.CloudProviderAWS.Region),
		Bucket:    cloudStorage.S3.Bucket,
		Prefix:    normalizeStoragePrefix(repo.StoragePrefix),
		AccessKey: cloudStorage.S3.AccessKeyID,
		SecretKey: secretKey,
	}, nil
}

func (s *service) buildLocalStorage(
	ctx context.Context,
	repo *entity.BackupRepo,
	refObjects *entity.RefObjects,
) (*backup.StorageLocal, error) {
	setting := refObjects.RefSettings[repo.Volume.ID]
	if setting == nil {
		return nil, hperrors.Wrap(hperrors.ErrSettingNotFound).WithParam("ID", repo.Volume.ID)
	}

	clusterVolume, err := setting.AsClusterVolume()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if clusterVolume.NodeID == "" && clusterVolume.NodeLabel == "" {
		return nil, hperrors.Wrap(hperrors.ErrBackupRepoVolumeNodeRequired).WithParam("Name", setting.Name)
	}

	hostPath, err := s.resolveVolumeHostPath(ctx, setting.Name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The agent mounts the host root, so the repository path has to be expressed from inside it.
	repoPath := filepath.Join(volumeservice.HostPathPrefix, hostPath)
	if prefix := normalizeStoragePrefix(repo.StoragePrefix); prefix != "" {
		repoPath = filepath.Join(repoPath, prefix)
	}

	return &backup.StorageLocal{
		Path:      repoPath,
		NodeID:    clusterVolume.NodeID,
		NodeLabel: clusterVolume.NodeLabel,
	}, nil
}

// resolveVolumeHostPath maps a docker volume onto its location on the host filesystem.
func (s *service) resolveVolumeHostPath(ctx context.Context, volumeName string) (string, error) {
	inspectResp, err := s.dockerManager.VolumeInspect(ctx, volumeName)
	if err != nil {
		return "", hperrors.Wrap(hperrors.ErrBackupRepoVolumePathUnresolved).WithParam("Name", volumeName)
	}

	// Volumes created with `--opt device=/path` are bind mounts: the device is the real location,
	// the docker-managed mountpoint is only a symlink target.
	if devicePath := inspectResp.Volume.Options["device"]; devicePath != "" {
		return devicePath, nil
	}
	if inspectResp.Volume.Mountpoint != "" {
		return inspectResp.Volume.Mountpoint, nil
	}
	return "", hperrors.Wrap(hperrors.ErrBackupRepoVolumePathUnresolved).WithParam("Name", volumeName)
}

// engineConfigFilePath gives each repository its own engine config file.
func engineConfigFilePath(repoID string) string {
	if repoID == "" {
		return ""
	}
	return path.Join(engineConfigDir, repoID, engineConfigFileName)
}

func normalizeStoragePrefix(prefix string) string {
	return strings.Trim(strings.TrimSpace(prefix), "/")
}
