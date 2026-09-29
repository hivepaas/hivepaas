package databackupserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// beforeRestoreSuffix and the time are what a directory moved aside by a
// restore is renamed to, beside where it was.
const beforeRestoreSuffix = ".before-restore-"

func (s *service) Restore(
	ctx context.Context,
	db database.Tx,
	req *databackupservice.RestoreReq,
) error {
	switch {
	case req.Command != nil:
		return hperrors.Wrap(s.restoreCommand(ctx, db, req))
	case req.Volume != nil:
		return hperrors.Wrap(s.restoreVolume(ctx, db, req))
	}
	return hperrors.Wrap(hperrors.ErrBadRequest).WithMsgLog("a restore has neither a command nor a volume")
}

// restoreCommand streams the snapshot's file into the command. The stream and
// the command share the task's transaction, which is not safe to use from two
// goroutines: the command starts once the stream is connected, and the stream
// uses the database no more.
func (s *service) restoreCommand(ctx context.Context, db database.Tx, req *databackupservice.RestoreReq) error {
	reader, writer := io.Pipe()
	streamCtx, cancelStream := context.WithCancel(ctx)
	defer cancelStream()

	connected := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		var err error
		// Whatever happens, a panic's included: the command reads to the end of the
		// stream, an error ends it too and is the command's to see, and the caller
		// waits on done.
		defer func() {
			_ = writer.CloseWithError(err)
			done <- err
		}()
		defer safego.RecoverTo(&err)
		err = s.backupRepoService.RestoreStream(streamCtx, db, &backupreposervice.RestoreStreamReq{
			RepoTarget: req.Target, SnapshotID: req.SnapshotID, FileName: req.Command.FileName,
			Stdout: writer, Progress: logRestoreProgress(ctx, req), OnConnected: func() { close(connected) },
		})
	}()

	select {
	case <-connected:
	case err := <-done:
		if err == nil {
			return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("the stream ended before it began")
		}
		return hperrors.Wrap(err)
	}

	logRestore(ctx, req, "Loading %s into %s", req.Command.FileName, req.App.Name)
	_, execErr := s.schedJobExecService.RunCommand(ctx, db, &schedjobexecservice.RunCommandReq{
		TaskExecData: req.TaskExecData, App: req.App, Command: req.Command.Command, Stdin: reader,
	})
	if execErr != nil {
		cancelStream()
	}
	// Whatever the command did not read is not waited for.
	_ = reader.CloseWithError(io.ErrClosedPipe)
	streamErr := <-done

	if execErr != nil {
		return hperrors.Wrap(execErr)
	}
	// The command read a part of the file: whatever it says, the restore failed.
	return hperrors.Wrap(streamErr)
}

// restoreVolume restores the snapshot into the app's directory of the volume,
// on the volume's node.
func (s *service) restoreVolume(ctx context.Context, db database.Tx, req *databackupservice.RestoreReq) (err error) {
	vol := req.Volume
	replace := vol.Mode == base.BackupRestoreModeReplace
	if replace && !vol.StopApp {
		// Its containers would go on writing to the directory moved aside.
		return hperrors.NewArgumentInvalid("stopApp").WithExtraDetail("replace stops the app")
	}
	appVolume, err := s.findAppVolume(ctx, db, req.App, vol.VolumeID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	dir, err := joinSubpath(appVolume.HostDir, vol.Subpath)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if dir, err = joinSubpath(dir, vol.SnapshotPath); err != nil {
		return hperrors.Wrap(err)
	}
	host := &nodeHost{run: s.hostRun, nodeID: appVolume.NodeID, nodeLabel: appVolume.NodeLabel}

	if vol.StopApp {
		var restart func() error
		if restart, err = s.stopForRestore(ctx, req); err != nil {
			return hperrors.Wrap(err)
		}
		// The restore's own error, joined with the app's failing to start: it is
		// the returned err, not one of this block.
		defer func() {
			err = errors.Join(err, restart())
		}()
	}

	aside := ""
	if replace {
		if aside, err = s.moveAside(ctx, req, host, dir); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if err = host.do(ctx, "mkdir", "-p", "--", hostPath(dir)); err != nil {
		return hperrors.Wrap(err)
	}

	logRestore(ctx, req, "Restoring %s into %s", snapshotPart(req.SnapshotID, vol.SnapshotPath), dir)
	err = s.backupRepoService.RestoreDirectory(ctx, db, &backupreposervice.RestoreDirectoryReq{
		RepoTarget: req.Target, SnapshotID: req.SnapshotID, Path: vol.SnapshotPath,
		HostDir: dir, NodeID: appVolume.NodeID, NodeLabel: appVolume.NodeLabel,
		Progress: logRestoreProgress(ctx, req),
	})
	if err != nil && aside != "" {
		err = errors.Join(err, s.putBack(context.WithoutCancel(ctx), req, host, dir, aside))
	}
	return hperrors.Wrap(err)
}

// stopForRestore stops the app, if it runs, until none of its containers do,
// and gives what starts it again. An app that was stopped stays so.
func (s *service) stopForRestore(ctx context.Context, req *databackupservice.RestoreReq) (func() error, error) {
	running, err := s.apps.Running(ctx, req.App)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !running {
		return func() error { return nil }, nil
	}
	logRestore(ctx, req, "Stopping %s", req.App.Name)
	restart := func() error {
		logRestore(ctx, req, "Starting %s", req.App.Name)
		return hperrors.Wrap(s.apps.Start(context.WithoutCancel(ctx), req.App))
	}
	if err = s.apps.Stop(ctx, req.App); err != nil {
		// Part of it may have stopped: it is started again.
		return nil, hperrors.Wrap(errors.Join(err, restart()))
	}
	return restart, nil
}

// moveAside moves the directory beside itself, when it is there; the path it
// moved it to, or "" for none.
func (s *service) moveAside(
	ctx context.Context,
	req *databackupservice.RestoreReq,
	host *nodeHost,
	dir string,
) (string, error) {
	exists, err := host.exists(ctx, hostPath(dir))
	if err != nil || !exists {
		return "", hperrors.Wrap(err)
	}
	aside := dir + beforeRestoreSuffix + s.now().UTC().Format("20060102-150405")
	if err = host.do(ctx, "mv", "--", hostPath(dir), hostPath(aside)); err != nil {
		return "", hperrors.Wrap(err)
	}
	logRestore(ctx, req, "Moved %s to %s", dir, aside)
	return aside, nil
}

// putBack undoes moveAside after a restore failed.
func (s *service) putBack(
	ctx context.Context,
	req *databackupservice.RestoreReq,
	host *nodeHost,
	dir string,
	aside string,
) error {
	if err := host.do(ctx, "rm", "-rf", "--", hostPath(dir)); err != nil {
		return hperrors.Wrap(fmt.Errorf("the directory is left at %s: %w", aside, err))
	}
	if err := host.do(ctx, "mv", "--", hostPath(aside), hostPath(dir)); err != nil {
		return hperrors.Wrap(fmt.Errorf("the directory is left at %s: %w", aside, err))
	}
	logRestore(ctx, req, "Put %s back", dir)
	return nil
}

// nodeHost runs commands on a node's host through its agent.
type nodeHost struct {
	run       func(ctx context.Context, nodeID, nodeLabel string, cmd ...string) (int, error)
	nodeID    string
	nodeLabel string
}

func (h *nodeHost) do(ctx context.Context, cmd ...string) error {
	code, err := h.run(ctx, h.nodeID, h.nodeLabel, cmd...)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if code != 0 {
		return hperrors.Wrap(hperrors.ErrActionFailed).WithExtraDetail("%s exited with status %d", cmd[0], code)
	}
	return nil
}

func (h *nodeHost) exists(ctx context.Context, path string) (bool, error) {
	code, err := h.run(ctx, h.nodeID, h.nodeLabel, "test", "-e", path)
	return code == 0 && err == nil, hperrors.Wrap(err)
}

// hostPath is a directory of the host as the agent sees it: the host root is
// mounted at volumeservice.HostPathPrefix.
func hostPath(dir string) string {
	return filepath.Join(volumeservice.HostPathPrefix, dir)
}

func snapshotPart(snapshotID, path string) string {
	if path == "" {
		return snapshotID
	}
	return snapshotID + "/" + path
}

func logRestore(ctx context.Context, req *databackupservice.RestoreReq, format string, args ...any) {
	_ = req.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf(format, args...)+"\n", tasklog.TsNow))
}

func logRestoreProgress(ctx context.Context, req *databackupservice.RestoreReq) func(string) {
	return func(msg string) {
		logRestore(ctx, req, "%s", msg)
	}
}

// appControl stops and starts an app for a restore.
type appControl interface {
	// Running is whether any of the app's containers runs.
	Running(ctx context.Context, app *entity.App) (bool, error)
	// Stop sets the app not running, and waits until none of its containers runs.
	Stop(ctx context.Context, app *entity.App) error
	Start(ctx context.Context, app *entity.App) error
}
