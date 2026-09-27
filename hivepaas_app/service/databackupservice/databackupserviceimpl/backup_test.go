package databackupserviceimpl

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// fakeRepos is the backup repository service: it reads what it is streamed, and
// records what it is asked to take, list, sync and delete.
type fakeRepos struct {
	backupreposervice.Service
	streamed  string
	stream    *backupreposervice.BackupStreamReq
	streamErr error
	// connectErr fails the stream before the repository is connected; stopErr
	// after, without reading what it is streamed.
	connectErr error
	stopErr    error
	connected  bool
	snapshot   *backupreposervice.RepoSnapshot
	directory  *backupreposervice.BackupDirectoryReq
	listed     []*backupreposervice.RepoSnapshot
	listedTags []string
	synced     []*backupreposervice.RepoSnapshot
	deleted    []string
}

func (f *fakeRepos) BackupStream(
	_ context.Context, _ database.IDB, req *backupreposervice.BackupStreamReq,
) (*backupreposervice.BackupResp, error) {
	if f.connectErr != nil {
		return nil, f.connectErr
	}
	f.connected = true
	f.stream = req
	if req.Progress != nil {
		req.Progress("Repository server ready at https://10.0.1.5:40123")
	}
	req.OnConnected()
	if f.stopErr != nil {
		return nil, f.stopErr
	}
	// A stream that ends in an error is, to kopia, an end of stream: it commits
	// what it read.
	data, _ := io.ReadAll(req.Stdin)
	f.streamed = string(data)
	if f.streamErr != nil {
		return nil, f.streamErr
	}
	return &backupreposervice.BackupResp{Snapshot: f.snapshot}, nil
}

func (f *fakeRepos) BackupDirectory(
	_ context.Context, _ database.IDB, req *backupreposervice.BackupDirectoryReq,
) (*backupreposervice.BackupResp, error) {
	f.directory = req
	if req.Progress != nil {
		req.Progress("Starting the repository server on node node-1")
	}
	return &backupreposervice.BackupResp{Snapshot: f.snapshot}, nil
}

func (f *fakeRepos) ListSnapshots(
	_ context.Context, _ database.IDB, req *backupreposervice.ListSnapshotsReq,
) (*backupreposervice.ListSnapshotsResp, error) {
	if req.Options != nil {
		f.listedTags = req.Options.Tags
	}
	return &backupreposervice.ListSnapshotsResp{Snapshots: f.listed}, nil
}

func (f *fakeRepos) SyncRepoSnapshots(
	_ context.Context, _ database.Tx, req *backupreposervice.SyncRepoSnapshotsReq,
) (*backupreposervice.SyncRepoSnapshotsResp, error) {
	f.synced = req.Remaining
	return &backupreposervice.SyncRepoSnapshotsResp{}, nil
}

func (f *fakeRepos) DeleteSnapshot(
	_ context.Context, _ database.IDB, req *backupreposervice.DeleteSnapshotReq,
) error {
	f.deleted = append(f.deleted, req.SnapshotID)
	return nil
}

// fakeJobExec "runs" the command: writes its output, then fails or not.
type fakeJobExec struct {
	schedjobexecservice.Service
	repos   *fakeRepos
	output  string
	err     error
	command *entity.CommandTemplate
	// afterConnect is whether the repository was connected when the command ran.
	afterConnect bool
}

func (f *fakeJobExec) SchedJobExec(
	_ context.Context, _ database.Tx, req *schedjobexecservice.SchedJobExecReq,
) (*schedjobexecservice.SchedJobExecResp, error) {
	f.command = req.Command
	f.afterConnect = f.repos != nil && f.repos.connected
	_, writeErr := req.StdoutWriter.Write([]byte(f.output))
	if f.err == nil && writeErr != nil {
		// A command whose output cannot be written fails as a real one does.
		return &schedjobexecservice.SchedJobExecResp{}, writeErr
	}
	return &schedjobexecservice.SchedJobExecResp{}, f.err
}

func snapshotOf(id string) *backupreposervice.RepoSnapshot {
	return &backupreposervice.RepoSnapshot{Snapshot: &entity.BackupSnapshot{ID: id, SizeBytes: 42}}
}

func backupReq(t *testing.T, dataBackup *entity.SchedJobDataBackup) *databackupservice.BackupReq {
	t.Helper()
	dataBackup.TargetRepository = entity.ObjectID{ID: "repo1"}
	job := &entity.Setting{ID: "j1", Name: "nightly", Type: base.SettingTypeSchedJob}
	job.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeDataBackup, DataBackup: dataBackup})
	repo := &entity.Setting{ID: "repo1", Type: base.SettingTypeBackupRepo, Status: base.SettingStatusActive}
	refObjects := entity.NewRefObjects()
	refObjects.RefSettings["repo1"] = repo
	return &databackupservice.BackupReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		JobSetting:   job,
		App:          &entity.App{ID: "a1"},
		RefObjects:   refObjects,
	}
}

func commandBackup() *entity.SchedJobDataBackup {
	return &entity.SchedJobDataBackup{
		Source:         base.SchedJobDataBackupSourceCommand,
		SourceCommand:  &entity.CommandTemplate{Command: "pg_dump app"},
		SourceFileName: "db.sql",
	}
}

func newTestService(repos *fakeRepos, jobExec *fakeJobExec) *service {
	jobExec.repos = repos
	return &service{
		backupRepoService:   repos,
		schedJobExecService: jobExec,
		repoScope: func(context.Context, database.IDB, *entity.Setting) (*entity.ObjectScope, error) {
			return entity.NewObjectScopeGlobal(), nil
		},
	}
}

// A command's output is streamed into the repository; the snapshot is synced
// into the repository's list and is the run's output.
func TestBackupOfACommand(t *testing.T) {
	repos := &fakeRepos{snapshot: snapshotOf("k1"), listed: []*backupreposervice.RepoSnapshot{snapshotOf("k1")}}
	jobExec := &fakeJobExec{output: "CREATE TABLE"}
	req := backupReq(t, commandBackup())

	resp, err := newTestService(repos, jobExec).Backup(context.Background(), database.Tx{}, req)

	assert.NoError(t, err)
	assert.Equal(t, "pg_dump app", jobExec.command.Command)
	// The engine is built, reading the database, before the command shares it.
	assert.True(t, jobExec.afterConnect)
	assert.Equal(t, "CREATE TABLE", repos.streamed)
	// kopia records every run of the job under one source, which its retention goes by.
	assert.Equal(t, "hivepaas@data-backup:/j1", repos.stream.Source)
	assert.Equal(t, "nightly (run t1)", repos.stream.Description)
	assert.Equal(t, &entity.SchedJobDataBackupResult{SnapshotID: "k1", SizeBytes: 42}, resp.Result)
	assert.Len(t, repos.synced, 1)
	output, _ := req.Task.OutputAsDataBackup()
	assert.Equal(t, "k1", output.SnapshotID)
	assert.Empty(t, repos.deleted)
}

// A command that fails leaves no snapshot: what kopia made of the output it read
// before the end is deleted, and the run fails.
func TestBackupOfAFailedCommandLeavesNoSnapshot(t *testing.T) {
	repos := &fakeRepos{snapshot: snapshotOf("k1")}
	jobExec := &fakeJobExec{output: "CREATE TA", err: errors.New("exit code 1")}

	_, err := newTestService(repos, jobExec).Backup(context.Background(), database.Tx{},
		backupReq(t, commandBackup()))

	assert.Error(t, err)
	assert.Equal(t, []string{"k1"}, repos.deleted)
	assert.Nil(t, repos.synced)
}

// When kopia fails too it names no snapshot: the ones tagged with the run are
// found and deleted.
func TestBackupOfAFailedCommandDeletesTheRunsSnapshots(t *testing.T) {
	repos := &fakeRepos{streamErr: errors.New("kopia: stream closed"),
		listed: []*backupreposervice.RepoSnapshot{snapshotOf("k2")}}
	jobExec := &fakeJobExec{err: errors.New("exit code 1")}

	_, err := newTestService(repos, jobExec).Backup(context.Background(), database.Tx{},
		backupReq(t, commandBackup()))

	assert.Error(t, err)
	assert.Equal(t, []string{"hivepaas.run:t1"}, repos.listedTags)
	assert.Equal(t, []string{"k2"}, repos.deleted)
}

// A repository that cannot be connected fails the run before the command runs.
func TestBackupOfACommandIntoAnUnreachableRepository(t *testing.T) {
	repos := &fakeRepos{connectErr: errors.New("kopia: repository not found")}
	jobExec := &fakeJobExec{output: "CREATE TABLE"}

	_, err := newTestService(repos, jobExec).Backup(context.Background(), database.Tx{},
		backupReq(t, commandBackup()))

	assert.ErrorContains(t, err, "repository not found")
	assert.Nil(t, jobExec.command)
}

// kopia stopping while the command runs fails the run with kopia's error, not
// the command's broken pipe.
func TestBackupOfACommandWhenKopiaStops(t *testing.T) {
	repos := &fakeRepos{stopErr: errors.New("kopia: no space left")}
	jobExec := &fakeJobExec{output: "CREATE TABLE"}

	_, err := newTestService(repos, jobExec).Backup(context.Background(), database.Tx{},
		backupReq(t, commandBackup()))

	assert.ErrorContains(t, err, "no space left")
}

// A step of a job sequence keeps the sequence's output; its snapshot is the
// step's result.
func TestBackupAsASequenceStepLeavesTheTaskOutputAlone(t *testing.T) {
	repos := &fakeRepos{snapshot: snapshotOf("k1")}
	req := backupReq(t, commandBackup())
	req.Sequence = &schedjobexecservice.SequenceStep{Step: 1, Steps: 2}

	resp, err := newTestService(repos, &fakeJobExec{}).Backup(context.Background(), database.Tx{}, req)

	assert.NoError(t, err)
	assert.Equal(t, "k1", resp.Result.SnapshotID)
	assert.Empty(t, req.Task.Output)
}

// A volume's directory is the volume's on its host, the app's part of it and
// the job's subpath; it is read on the volume's node.
func TestBackupOfAVolume(t *testing.T) {
	repos := &fakeRepos{snapshot: snapshotOf("k1")}
	svc := newTestService(repos, &fakeJobExec{})
	svc.findAppVolume = func(context.Context, database.IDB, *entity.App, string) (*databackupservice.AppVolume, error) {
		return &databackupservice.AppVolume{HostDir: "/srv/data/p1/dev/a1", NodeID: "node-2"}, nil
	}
	req := backupReq(t, &entity.SchedJobDataBackup{
		Source: base.SchedJobDataBackupSourceVolume, SourceVolume: entity.ObjectID{ID: "vol1"},
		SourceVolumeSubpath: "uploads",
	})

	_, err := svc.Backup(context.Background(), database.Tx{}, req)

	assert.NoError(t, err)
	if assert.NotNil(t, repos.directory) {
		assert.Equal(t, "/srv/data/p1/dev/a1/uploads", repos.directory.HostDir)
		assert.Equal(t, "node-2", repos.directory.NodeID)
		assert.Contains(t, repos.directory.Tags, "hivepaas.run:t1")
		assert.Equal(t, "nightly (run t1)", repos.directory.Description)
		assert.Equal(t, "hivepaas@data-backup:/j1", repos.directory.Source)
	}
}

func TestBackupFailsWithoutAnActiveRepository(t *testing.T) {
	req := backupReq(t, commandBackup())
	req.RefObjects.RefSettings["repo1"].Status = base.SettingStatusDisabled

	_, err := newTestService(&fakeRepos{}, &fakeJobExec{}).Backup(context.Background(), database.Tx{}, req)

	assert.Error(t, err)
}

// The app's part of a volume is where its mount reaches: a bind mount names the
// directory itself, a volume mount a subpath of the volume's directory.
func TestAppVolumeDir(t *testing.T) {
	bind := mount.Mount{Type: mount.TypeBind, Source: "/srv/data/p1/dev/a1"}
	assert.Equal(t, "/srv/data/p1/dev/a1", appVolumeDir(&bind, "/srv/data"))

	volume := mount.Mount{Type: mount.TypeVolume, VolumeOptions: &mount.VolumeOptions{Subpath: "p1/dev/a1"}}
	assert.Equal(t, "/var/lib/docker/volumes/v/_data/p1/dev/a1", appVolumeDir(&volume,
		"/var/lib/docker/volumes/v/_data"))

	whole := mount.Mount{Type: mount.TypeVolume}
	assert.Equal(t, "/data", appVolumeDir(&whole, "/data"))
}

// A job's subpath stays inside the app's part of the volume.
func TestJoinSubpath(t *testing.T) {
	dir, err := joinSubpath("/srv/a1", "uploads/2026")
	assert.NoError(t, err)
	assert.Equal(t, "/srv/a1/uploads/2026", dir)

	dir, err = joinSubpath("/srv/a1", "")
	assert.NoError(t, err)
	assert.Equal(t, "/srv/a1", dir)

	for _, bad := range []string{"../a2", "/etc", "uploads/../../a2"} {
		_, err = joinSubpath("/srv/a1", bad)
		assert.Error(t, err, bad)
	}
}

// The mount a volume source reads is the app's own directory or the whole
// volume, never another app's directory the app was given.
func TestPickAppVolumeMount(t *testing.T) {
	mounts := []mount.Mount{{Target: "/other"}, {Target: "/data"}, {Target: "/whole"}}
	descs := []*volumeservice.AppMountDesc{
		{VolumeID: "vol1", AppKey: "a2", Own: false},
		{VolumeID: "vol1", AppKey: "a1", Own: true},
		{VolumeID: "vol2"},
	}

	picked := pickAppVolumeMount(mounts, descs, "vol1")
	if assert.NotNil(t, picked) {
		assert.Equal(t, "/data", picked.Target)
	}
	assert.Nil(t, pickAppVolumeMount(mounts[:1], descs[:1], "vol1"))
	assert.Nil(t, pickAppVolumeMount(mounts, descs, "vol3"))
}

// The app is gone, or never loaded: the run fails, it does not panic.
func TestBackupFailsWithoutItsApp(t *testing.T) {
	req := backupReq(t, commandBackup())
	req.App = nil

	_, err := newTestService(&fakeRepos{}, &fakeJobExec{}).Backup(context.Background(), database.Tx{}, req)

	assert.Error(t, err)
}

// logText is what the run's log says.
func logText(t *testing.T, req *databackupservice.BackupReq) string {
	t.Helper()
	frames, err := req.LogStore.GetLocalData(context.Background(), 0)
	assert.NoError(t, err)
	var text string
	for _, frame := range frames {
		text += frame.Data
	}
	return text
}

// The run's log says what the backup did on the way, such as a repository server
// started for it.
func TestBackupLogsItsProgress(t *testing.T) {
	repos := &fakeRepos{snapshot: snapshotOf("k1")}
	req := backupReq(t, commandBackup())
	req.LogStore = tasklog.NewLocalStore("t1")

	_, err := newTestService(repos, &fakeJobExec{}).Backup(context.Background(), database.Tx{}, req)

	assert.NoError(t, err)
	assert.Contains(t, logText(t, req), "Repository server ready at https://10.0.1.5:40123\n")

	repos = &fakeRepos{snapshot: snapshotOf("k1")}
	svc := newTestService(repos, &fakeJobExec{})
	svc.findAppVolume = func(context.Context, database.IDB, *entity.App, string) (*databackupservice.AppVolume, error) {
		return &databackupservice.AppVolume{HostDir: "/srv/data/a1", NodeID: "node-2"}, nil
	}
	req = backupReq(t, &entity.SchedJobDataBackup{
		Source: base.SchedJobDataBackupSourceVolume, SourceVolume: entity.ObjectID{ID: "vol1"},
	})
	req.LogStore = tasklog.NewLocalStore("t1")

	_, err = svc.Backup(context.Background(), database.Tx{}, req)

	assert.NoError(t, err)
	assert.Contains(t, logText(t, req), "Starting the repository server on node node-1\n")
}
