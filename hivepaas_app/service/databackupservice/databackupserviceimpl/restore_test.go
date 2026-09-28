package databackupserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// restoreWorld is everything a restore touches, faked: the repository, the
// app's command, the app itself and the node's host. It writes what happens,
// in order, into events.
type restoreWorld struct {
	events []string

	// the repository
	dump       string
	streamErr  error
	restoreErr error
	// the app's command
	commandErr error
	loaded     string
	// the app
	running bool
	// the host: which paths exist
	exists map[string]bool
}

func (w *restoreWorld) log(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

type restoreRepos struct {
	backupreposervice.Service
	w *restoreWorld
}

func (r restoreRepos) RestoreStream(_ context.Context, _ database.IDB, req *backupreposervice.RestoreStreamReq) error {
	r.w.log("stream %s/%s", req.SnapshotID, req.FileName)
	req.OnConnected()
	_, _ = io.WriteString(req.Stdout, r.w.dump)
	return r.w.streamErr
}

func (r restoreRepos) RestoreDirectory(
	_ context.Context, _ database.IDB, req *backupreposervice.RestoreDirectoryReq,
) error {
	r.w.log("restore %s[%s] into %s on %s", req.SnapshotID, req.Path, req.HostDir, req.NodeID)
	return r.w.restoreErr
}

type restoreExec struct {
	schedjobexecservice.Service
	w *restoreWorld
}

func (e restoreExec) RunCommand(
	_ context.Context, _ database.Tx, req *schedjobexecservice.RunCommandReq,
) (*schedjobexecservice.RunCommandResp, error) {
	e.w.log("run %q in %s", req.Command.Command, req.App.ID)
	in, _ := io.ReadAll(req.Stdin)
	e.w.loaded = string(in)
	return &schedjobexecservice.RunCommandResp{}, e.w.commandErr
}

func (w *restoreWorld) service() *service {
	return &service{
		backupRepoService:   restoreRepos{w: w},
		schedJobExecService: restoreExec{w: w},
		findAppVolume: func(context.Context, database.IDB, *entity.App, string) (*databackupservice.AppVolume, error) {
			return &databackupservice.AppVolume{HostDir: "/srv/p1/app1", NodeID: "node-2"}, nil
		},
		apps: fakeAppControl{w: w},
		hostRun: func(_ context.Context, nodeID, _ string, cmd ...string) (int, error) {
			w.log("%s: %s", nodeID, strings.Join(cmd, " "))
			if cmd[0] == "test" {
				if w.exists[cmd[2]] {
					return 0, nil
				}
				return 1, nil
			}
			return 0, nil
		},
		now: func() time.Time { return time.Date(2026, 9, 28, 15, 30, 0, 0, time.UTC) },
	}
}

type fakeAppControl struct {
	w *restoreWorld
}

func (a fakeAppControl) Running(context.Context, *entity.App) (bool, error) {
	return a.w.running, nil
}

func (a fakeAppControl) Stop(_ context.Context, app *entity.App) error {
	a.w.log("stop %s", app.ID)
	a.w.running = false
	return nil
}

func (a fakeAppControl) Start(_ context.Context, app *entity.App) error {
	a.w.log("start %s", app.ID)
	a.w.running = true
	return nil
}

func restoreReq(mod func(*databackupservice.RestoreReq)) *databackupservice.RestoreReq {
	req := &databackupservice.RestoreReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		SnapshotID:   "k1",
		App:          &entity.App{ID: "app1", ServiceID: "svc1"},
	}
	mod(req)
	return req
}

func volumeRestore(mode base.BackupRestoreMode, stop bool) func(*databackupservice.RestoreReq) {
	return func(req *databackupservice.RestoreReq) {
		req.Volume = &databackupservice.RestoreVolume{
			VolumeID: "vol1", Subpath: "data", SnapshotPath: "uploads", StopApp: stop, Mode: mode,
		}
	}
}

// Replace: the app stops, the directory is moved aside, the snapshot is
// restored into an empty one, and the app starts again.
func TestRestoreVolumeReplacesTheDirectory(t *testing.T) {
	w := &restoreWorld{running: true, exists: map[string]bool{"/host/srv/p1/app1/data/uploads": true}}

	err := w.service().Restore(context.Background(), database.Tx{},
		restoreReq(volumeRestore(base.BackupRestoreModeReplace, true)))

	assert.NoError(t, err)
	assert.Equal(t, []string{
		"stop app1",
		"node-2: test -e /host/srv/p1/app1/data/uploads",
		"node-2: mv -- /host/srv/p1/app1/data/uploads /host/srv/p1/app1/data/uploads.before-restore-20260928-153000",
		"node-2: mkdir -p -- /host/srv/p1/app1/data/uploads",
		"restore k1[uploads] into /srv/p1/app1/data/uploads on node-2",
		"start app1",
	}, w.events)
}

// A directory that is not there has nothing to move aside.
func TestRestoreVolumeReplacesADirectoryThatIsNotThere(t *testing.T) {
	w := &restoreWorld{running: true, exists: map[string]bool{}}

	err := w.service().Restore(context.Background(), database.Tx{},
		restoreReq(volumeRestore(base.BackupRestoreModeReplace, true)))

	assert.NoError(t, err)
	assert.Equal(t, []string{
		"stop app1",
		"node-2: test -e /host/srv/p1/app1/data/uploads",
		"node-2: mkdir -p -- /host/srv/p1/app1/data/uploads",
		"restore k1[uploads] into /srv/p1/app1/data/uploads on node-2",
		"start app1",
	}, w.events)
}

// A restore that fails puts the directory back as it was, and starts the app.
func TestRestoreVolumeThatFailsPutsTheDirectoryBack(t *testing.T) {
	w := &restoreWorld{running: true, exists: map[string]bool{"/host/srv/p1/app1/data/uploads": true},
		restoreErr: errors.New("kopia: repository gone")}

	err := w.service().Restore(context.Background(), database.Tx{},
		restoreReq(volumeRestore(base.BackupRestoreModeReplace, true)))

	assert.ErrorContains(t, err, "repository gone")
	assert.Equal(t, []string{
		"stop app1",
		"node-2: test -e /host/srv/p1/app1/data/uploads",
		"node-2: mv -- /host/srv/p1/app1/data/uploads /host/srv/p1/app1/data/uploads.before-restore-20260928-153000",
		"node-2: mkdir -p -- /host/srv/p1/app1/data/uploads",
		"restore k1[uploads] into /srv/p1/app1/data/uploads on node-2",
		"node-2: rm -rf -- /host/srv/p1/app1/data/uploads",
		"node-2: mv -- /host/srv/p1/app1/data/uploads.before-restore-20260928-153000 /host/srv/p1/app1/data/uploads",
		"start app1",
	}, w.events)
}

// A canceled restore is a failed one: the same steps back, even though the
// run's context is done.
func TestRestoreVolumeCanceledPutsTheDirectoryBack(t *testing.T) {
	w := &restoreWorld{running: true, exists: map[string]bool{"/host/srv/p1/app1/data/uploads": true},
		restoreErr: context.Canceled}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := w.service().Restore(ctx, database.Tx{}, restoreReq(volumeRestore(base.BackupRestoreModeReplace, true)))

	assert.ErrorIs(t, err, context.Canceled)
	assert.Contains(t, w.events,
		"node-2: mv -- /host/srv/p1/app1/data/uploads.before-restore-20260928-153000 /host/srv/p1/app1/data/uploads")
	assert.Equal(t, "start app1", w.events[len(w.events)-1])
}

// Overwrite, the app left running: nothing is stopped, nothing moved.
func TestRestoreVolumeOverwritesWithoutStopping(t *testing.T) {
	w := &restoreWorld{running: true}

	err := w.service().Restore(context.Background(), database.Tx{},
		restoreReq(volumeRestore(base.BackupRestoreModeOverwrite, false)))

	assert.NoError(t, err)
	assert.Equal(t, []string{
		"node-2: mkdir -p -- /host/srv/p1/app1/data/uploads",
		"restore k1[uploads] into /srv/p1/app1/data/uploads on node-2",
	}, w.events)
}

// An app that was not running stays stopped.
func TestRestoreVolumeLeavesAStoppedAppStopped(t *testing.T) {
	w := &restoreWorld{running: false}

	err := w.service().Restore(context.Background(), database.Tx{},
		restoreReq(volumeRestore(base.BackupRestoreModeOverwrite, true)))

	assert.NoError(t, err)
	assert.NotContains(t, w.events, "start app1")
	assert.NotContains(t, w.events, "stop app1")
}

// Replace never leaves the app running: its containers would go on writing to
// the directory moved aside.
func TestRestoreVolumeReplaceRefusesARunningApp(t *testing.T) {
	w := &restoreWorld{running: true}

	err := w.service().Restore(context.Background(), database.Tx{},
		restoreReq(volumeRestore(base.BackupRestoreModeReplace, false)))

	assert.Error(t, err)
	assert.Empty(t, w.events)
}

func commandRestore(req *databackupservice.RestoreReq) {
	req.Command = &databackupservice.RestoreCommand{
		Command: &entity.CommandTemplate{Command: "psql app"}, FileName: "db.sql",
	}
}

// A command snapshot is streamed into the command, run in the target app.
func TestRestoreCommandStreamsTheFileIntoTheCommand(t *testing.T) {
	w := &restoreWorld{dump: "CREATE TABLE t();\n", running: true}

	err := w.service().Restore(context.Background(), database.Tx{}, restoreReq(commandRestore))

	assert.NoError(t, err)
	assert.Equal(t, "CREATE TABLE t();\n", w.loaded)
	assert.Equal(t, []string{`stream k1/db.sql`, `run "psql app" in app1`}, w.events)
}

// A stream that breaks fails the restore, however the command ended: it read a
// part of the file.
func TestRestoreCommandFailsWhenTheStreamBreaks(t *testing.T) {
	w := &restoreWorld{dump: "CREATE TA", streamErr: errors.New("kopia: connection reset")}

	err := w.service().Restore(context.Background(), database.Tx{}, restoreReq(commandRestore))

	assert.ErrorContains(t, err, "connection reset")
}

func TestRestoreCommandFailsWhenTheCommandFails(t *testing.T) {
	w := &restoreWorld{dump: "CREATE TABLE t();\n", commandErr: errors.New("exit code 3")}

	err := w.service().Restore(context.Background(), database.Tx{}, restoreReq(commandRestore))

	assert.ErrorContains(t, err, "exit code 3")
}
