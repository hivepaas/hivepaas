package fileserviceimpl

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	agentfile "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/cloudstorageservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// dirModeOwn is for the directories a file's path needs on the system volume.
const dirModeOwn = 0o700

// Open reads a file.
func (s *service) Open(ctx context.Context, db database.IDB, file *entity.File) (io.ReadCloser, error) {
	switch file.StorageType {
	case base.FileStorageVolume:
		if file.StorageID == "" {
			path, err := systemPath(file.Path)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			f, err := os.Open(path) //nolint:gosec // held inside the data directory above
			return f, notFoundOr(err)
		}
		agent, root, err := s.volumeAgent(ctx, db, file.StorageID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		reader, err := agent.Read(ctx, root, file.Path)
		if err != nil {
			_ = agent.Close()
			return nil, hperrors.Wrap(err)
		}
		return &closingBoth{ReadCloser: reader, agent: agent}, nil
	case base.FileStorageCloud:
		return s.openCloud(ctx, db, file)
	default:
		return nil, hperrors.Wrap(hperrors.ErrStorageTypeUnsupported).WithParam("Type", file.StorageType)
	}
}

// Create writes a file. What is written reaches its place once the writer is
// closed without error: a failed write leaves no file, and replaces none.
func (s *service) Create(ctx context.Context, db database.IDB, file *entity.File) (fileservice.FileWriter, error) {
	if file.StorageType != base.FileStorageVolume {
		return nil, hperrors.Wrap(hperrors.ErrStorageTypeUnsupported).WithParam("Type", file.StorageType)
	}
	if file.StorageID == "" {
		return createSystemFile(file.Path)
	}
	agent, root, err := s.volumeAgent(ctx, db, file.StorageID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	pr, pw := io.Pipe()
	w := &agentWriter{PipeWriter: pw, done: make(chan error, 1)}
	go func() {
		var writeErr error
		defer func() {
			_ = agent.Close()
			_ = pr.CloseWithError(writeErr)
			w.done <- writeErr
		}()
		defer safego.RecoverTo(&writeErr)
		_, writeErr = agent.Write(ctx, root, file.Path, pr)
	}()
	return w, nil
}

// Remove removes a file. A file already gone is removed.
func (s *service) Remove(ctx context.Context, db database.IDB, file *entity.File) error {
	switch file.StorageType {
	case base.FileStorageVolume:
		if file.StorageID == "" {
			path, err := systemPath(file.Path)
			if err != nil {
				return hperrors.Wrap(err)
			}
			if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return hperrors.Wrap(err)
			}
			return nil
		}
		agent, root, err := s.volumeAgent(ctx, db, file.StorageID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		defer agent.Close()
		return hperrors.Wrap(agent.Remove(ctx, root, file.Path))
	case base.FileStorageCloud:
		return s.removeCloud(ctx, db, file)
	default:
		return hperrors.Wrap(hperrors.ErrStorageTypeUnsupported).WithParam("Type", file.StorageType)
	}
}

// Stat is a file's size, as it is stored now.
func (s *service) Stat(ctx context.Context, db database.IDB, file *entity.File) (int64, error) {
	switch file.StorageType {
	case base.FileStorageVolume:
		if file.StorageID == "" {
			path, err := systemPath(file.Path)
			if err != nil {
				return 0, hperrors.Wrap(err)
			}
			info, err := os.Stat(path)
			if err != nil {
				return 0, notFoundOr(err)
			}
			return info.Size(), nil
		}
		agent, root, err := s.volumeAgent(ctx, db, file.StorageID)
		if err != nil {
			return 0, hperrors.Wrap(err)
		}
		defer agent.Close()
		size, err := agent.Stat(ctx, root, file.Path)
		return size, hperrors.Wrap(err)
	case base.FileStorageCloud:
		return file.Size, nil
	default:
		return 0, hperrors.Wrap(hperrors.ErrStorageTypeUnsupported).WithParam("Type", file.StorageType)
	}
}

// CountOnVolume counts the files a volume holds. Deleting a volume is refused
// while it is not zero; deleting the files with the volume would remove each of
// them through Remove.
func (s *service) CountOnVolume(ctx context.Context, db database.IDB, volumeID string) (int, error) {
	files, _, err := s.fileRepo.List(ctx, db, nil,
		bunex.SelectColumns("file.id"),
		bunex.SelectWhere("file.storage_type = ?", base.FileStorageVolume),
		bunex.SelectWhere("file.storage_id = ?", volumeID),
		bunex.SelectWhere("file.deleted IS NOT TRUE"),
	)
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return len(files), nil
}

// volumeAgent is the agent that reaches a volume's files, and the volume's
// directory on its host: the node the volume is pinned to, or, for a volume
// every node shares, the node asking.
func (s *service) volumeAgent(
	ctx context.Context,
	db database.IDB,
	volumeID string,
) (agentfile.FileServiceClient, string, error) {
	setting, err := s.settingRepo.GetByID(ctx, db, nil, base.SettingTypeClusterVolume, volumeID, false)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	dir, err := volumeservice.ResolveHostDir(ctx, s.dockerManager, setting)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}

	nodeID := dir.NodeID
	if nodeID == "" {
		nodeIDs, err := s.agentService.NodeIDsWithLabel(ctx, dir.NodeLabel)
		if err != nil {
			return nil, "", hperrors.Wrap(err)
		}
		if len(nodeIDs) != 1 {
			return nil, "", hperrors.Wrap(hperrors.ErrVolumeCannotHoldFiles).
				WithParam("Name", setting.Name).WithParam("Label", dir.NodeLabel).WithParam("Nodes", len(nodeIDs))
		}
		nodeID = nodeIDs[0]
	}

	addr, err := s.agentService.GetAgentAddrForNode(ctx, nodeID)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	agent, err := s.agentFiles(addr)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	return agent, dir.Dir, nil
}

// systemPath is a path on the system volume, HivePaaS's data directory, which
// must stay inside it.
func systemPath(path string) (string, error) {
	clean := filepath.Clean(path)
	if path == "" || filepath.IsAbs(path) || clean == "." || clean == ".." ||
		strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", hperrors.Wrap(hperrors.ErrFilePathOutsideRoot).
			WithParam("Path", path).WithParam("Root", config.Current().AppPath)
	}
	return filepath.Join(config.Current().AppPath, clean), nil
}

// createSystemFile writes a file on the system volume through a temporary file,
// renamed into place on Close.
func createSystemFile(path string) (fileservice.FileWriter, error) {
	target, err := systemPath(path)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = os.MkdirAll(filepath.Dir(target), dirModeOwn); err != nil {
		return nil, hperrors.Wrap(err)
	}
	temp, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &renamingWriter{File: temp, target: target}, nil
}

// renamingWriter is a temporary file that becomes target when closed. Abort
// discards it instead.
type renamingWriter struct {
	*os.File
	target string
	failed bool
}

func (w *renamingWriter) Write(p []byte) (int, error) {
	n, err := w.File.Write(p)
	if err != nil {
		w.failed = true
	}
	return n, err //nolint:wrapcheck
}

func (w *renamingWriter) Close() error {
	if err := w.File.Close(); err != nil || w.failed {
		_ = os.Remove(w.Name())
		return hperrors.Wrap(errors.Join(err, errWriteFailed))
	}
	if err := os.Rename(w.Name(), w.target); err != nil {
		_ = os.Remove(w.Name())
		return hperrors.Wrap(err)
	}
	return nil
}

func (w *renamingWriter) Abort(error) {
	_ = w.File.Close()
	_ = os.Remove(w.Name())
}

var errWriteFailed = errors.New("the file was not written whole")

// agentWriter feeds an agent's write; Close waits for the agent's answer.
type agentWriter struct {
	*io.PipeWriter
	done chan error
}

func (w *agentWriter) Close() error {
	_ = w.PipeWriter.Close()
	return hperrors.Wrap(<-w.done)
}

// Abort ends the agent's write with err, which makes the agent drop it.
func (w *agentWriter) Abort(err error) {
	_ = w.CloseWithError(err)
	<-w.done
}

// closingBoth is a reader whose agent connection closes with it.
type closingBoth struct {
	io.ReadCloser
	agent agentfile.FileServiceClient
}

func (c *closingBoth) Close() error {
	err := c.ReadCloser.Close()
	_ = c.agent.Close()
	return hperrors.Wrap(err)
}

func (s *service) openCloud(ctx context.Context, db database.IDB, file *entity.File) (io.ReadCloser, error) {
	if file.Storage == nil {
		return nil, hperrors.NewInactive("Storage setting")
	}
	switch base.CloudStorageKind(file.Storage.Kind) {
	case base.CloudStorageKindS3:
		s3Client, err := cloudstorageservice.NewS3Client(ctx, db, file.Storage, nil)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		object, err := s3Client.GetObject(ctx, file.Bucket, file.Path)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return object.Body, nil
	default:
		return nil, hperrors.NewUnsupported("Storage type")
	}
}

func notFoundOr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return hperrors.NewNotFound("File")
	}
	return hperrors.Wrap(err)
}
