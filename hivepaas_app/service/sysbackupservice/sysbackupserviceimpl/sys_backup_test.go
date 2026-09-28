package sysbackupserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type fakeRepos struct {
	backupreposervice.Service
	// taken is what the directory held when it was taken, by name.
	taken   []string
	req     *backupreposervice.BackupLocalDirectoryReq
	takeErr error
	synced  bool
	deleted []string
}

func (f *fakeRepos) BackupLocalDirectory(
	_ context.Context, _ database.IDB, req *backupreposervice.BackupLocalDirectoryReq,
) (*backupreposervice.BackupResp, error) {
	f.req = req
	entries, _ := os.ReadDir(req.Dir)
	for _, entry := range entries {
		f.taken = append(f.taken, entry.Name())
	}
	sort.Strings(f.taken)
	if f.takeErr != nil {
		return nil, f.takeErr
	}
	return &backupreposervice.BackupResp{Snapshot: &backupreposervice.RepoSnapshot{
		Snapshot: &entity.BackupSnapshot{ID: "k1full", SizeBytes: 42},
	}}, nil
}

func (f *fakeRepos) ListSnapshots(
	_ context.Context, _ database.IDB, req *backupreposervice.ListSnapshotsReq,
) (*backupreposervice.ListSnapshotsResp, error) {
	if req.Options != nil && len(req.Options.Tags) > 0 {
		return &backupreposervice.ListSnapshotsResp{Snapshots: []*backupreposervice.RepoSnapshot{
			{Snapshot: &entity.BackupSnapshot{ID: "k-partial"}},
		}}, nil
	}
	return &backupreposervice.ListSnapshotsResp{}, nil
}

func (f *fakeRepos) SyncRepoSnapshots(
	context.Context, database.Tx, *backupreposervice.SyncRepoSnapshotsReq,
) (*backupreposervice.SyncRepoSnapshotsResp, error) {
	f.synced = true
	return &backupreposervice.SyncRepoSnapshotsResp{}, nil
}

func (f *fakeRepos) DeleteSnapshot(_ context.Context, _ database.IDB, req *backupreposervice.DeleteSnapshotReq) error {
	f.deleted = append(f.deleted, req.SnapshotID)
	return nil
}

type fakeSpecs struct {
	specservice.Service
	req *specservice.ExportReq
	err error
}

func (f *fakeSpecs) Export(
	_ context.Context, _ database.IDB, req *specservice.ExportReq,
) (*specservice.ExportResp, error) {
	f.req = req
	if f.err != nil {
		return nil, f.err
	}
	name := "hivepaas-spec-20260928.tar.gz"
	if req.SecretsMode == specmodel.SecretsModeEncrypted {
		name += ".age"
	}
	path := filepath.Join(req.WorkDir, name)
	if err := os.WriteFile(path, []byte("bundle"), 0o600); err != nil {
		return nil, err
	}
	return &specservice.ExportResp{Path: path, Filename: name, Size: 6}, nil
}

type repoByID struct {
	repository.SettingRepo
}

func (repoByID) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	repo := &entity.Setting{ID: id, Type: base.SettingTypeBackupRepo, Scope: base.ObjectScopeGlobal}
	repo.MustSetData(&entity.BackupRepo{Engine: "kopia"})
	return repo, nil
}

type world struct {
	repos   *fakeRepos
	specs   *fakeSpecs
	dumpErr error
	dumped  bool
}

func (w *world) service(t *testing.T) *service {
	return &service{
		settingRepo:       repoByID{},
		backupRepoService: w.repos,
		specService:       w.specs,
		repoScope: func(context.Context, database.IDB, *entity.Setting) (*entity.ObjectScope, error) {
			return &entity.ObjectScope{ScopeType: base.ObjectScopeGlobal}, nil
		},
		dumpDB: func(_ context.Context, path string, _ *tasklog.Store) error {
			w.dumped = true
			if w.dumpErr != nil {
				return w.dumpErr
			}
			return os.WriteFile(path, []byte("PGDMP"), 0o600)
		},
		workDir: func() (string, error) { return os.MkdirTemp(t.TempDir(), "sys-backup-*") },
	}
}

func newWorld() *world {
	return &world{repos: &fakeRepos{}, specs: &fakeSpecs{}}
}

func backupReq(settings *entity.SystemBackup) *sysbackupservice.SysBackupReq {
	settings.TargetRepository = entity.ObjectID{ID: "repo1"}
	return &sysbackupservice.SysBackupReq{
		TaskExecData:      &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		SysBackupSettings: settings,
	}
}

// The database and the spec go into one snapshot of a directory, recorded under
// the system backup's source with its tags; the repository's records follow.
func TestSystemBackupTakesTheDatabaseAndTheSpecInOneSnapshot(t *testing.T) {
	w := newWorld()
	settings := &entity.SystemBackup{IncludeDB: true, IncludeSpec: true,
		SpecSecrets: string(specmodel.SecretsModeEncrypted)}
	settings.SpecPassphrase.Set("correct horse battery staple")
	req := backupReq(settings)

	_, err := w.service(t).Backup(context.Background(), database.Tx{}, req)

	assert.NoError(t, err)
	assert.Equal(t, []string{"db.pg_dump", "spec.tar.gz.age"}, w.repos.taken)
	if assert.NotNil(t, w.specs.req) {
		assert.Equal(t, base.ObjectScopeGlobal, w.specs.req.Scope.ScopeType)
		assert.Equal(t, specmodel.SecretsModeEncrypted, w.specs.req.SecretsMode)
		assert.Equal(t, "correct horse battery staple", w.specs.req.Passphrase)
	}
	assert.Equal(t, "hivepaas@system-backup:/system", w.repos.req.Source)
	assert.Equal(t, []string{"hivepaas.source:system-backup", "hivepaas.run:t1"}, w.repos.req.Tags)
	assert.Equal(t, "System backup (run t1)", w.repos.req.Description)
	assert.Equal(t, "repo1", w.repos.req.RepoSetting.ID)
	assert.True(t, w.repos.synced)

	out, err := req.Task.OutputAsSystemBackup()
	assert.NoError(t, err)
	assert.Equal(t, &entity.TaskSystemBackupOutput{SnapshotID: "k1full", SizeBytes: 42,
		Includes: []string{"database", "spec"}}, out)
}

func TestSystemBackupOfTheDatabaseOnly(t *testing.T) {
	w := newWorld()

	_, err := w.service(t).Backup(context.Background(), database.Tx{}, backupReq(&entity.SystemBackup{IncludeDB: true}))

	assert.NoError(t, err)
	assert.Equal(t, []string{"db.pg_dump"}, w.repos.taken)
	assert.Nil(t, w.specs.req, "no spec was exported")
}

func TestSystemBackupOfTheSpecOnly(t *testing.T) {
	w := newWorld()

	_, err := w.service(t).Backup(context.Background(), database.Tx{}, backupReq(&entity.SystemBackup{
		IncludeSpec: true, SpecSecrets: string(specmodel.SecretsModeOmit),
	}))

	assert.NoError(t, err)
	assert.Equal(t, []string{"spec.tar.gz"}, w.repos.taken)
	assert.False(t, w.dumped, "the database was not dumped")
}

// A failed dump or export fails the run before anything is taken.
func TestSystemBackupThatCannotDumpTakesNothing(t *testing.T) {
	w := newWorld()
	w.dumpErr = errors.New("pg_dump: connection refused")

	_, err := w.service(t).Backup(context.Background(), database.Tx{},
		backupReq(&entity.SystemBackup{IncludeDB: true, IncludeSpec: true, SpecSecrets: "omit"}))

	assert.ErrorContains(t, err, "connection refused")
	assert.Nil(t, w.repos.req)

	w = newWorld()
	w.specs.err = errors.New("spec: gather failed")
	_, err = w.service(t).Backup(context.Background(), database.Tx{},
		backupReq(&entity.SystemBackup{IncludeDB: true, IncludeSpec: true, SpecSecrets: "omit"}))
	assert.ErrorContains(t, err, "gather failed")
	assert.Nil(t, w.repos.req)
}

// A snapshot failing half way is deleted by the run's tag.
func TestSystemBackupDeletesAHalfMadeSnapshot(t *testing.T) {
	w := newWorld()
	w.repos.takeErr = errors.New("kopia: repository gone")

	_, err := w.service(t).Backup(context.Background(), database.Tx{}, backupReq(&entity.SystemBackup{IncludeDB: true}))

	assert.ErrorContains(t, err, "repository gone")
	assert.Equal(t, []string{"k-partial"}, w.repos.deleted)
}

// A backup that takes nothing is refused: the configuration asks for one.
func TestSystemBackupOfNothingIsRefused(t *testing.T) {
	w := newWorld()

	_, err := w.service(t).Backup(context.Background(), database.Tx{}, backupReq(&entity.SystemBackup{}))

	assert.Error(t, err)
	assert.Nil(t, w.repos.req)
}
