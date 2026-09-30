package schedjobexecserviceimpl

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// volumeFiles is the file layer, in memory.
type volumeFiles struct {
	fileservice.Service
	files   map[string]string
	removed []string
	volumes map[string]*entity.Setting
}

type bufferWriter struct {
	bytes.Buffer
	keep func(string)
}

func (w *bufferWriter) Close() error { w.keep(w.String()); return nil }
func (w *bufferWriter) Abort(error)  {}

func (v *volumeFiles) Create(_ context.Context, _ database.IDB, file *entity.File) (fileservice.FileWriter, error) {
	return &bufferWriter{keep: func(s string) { v.files[file.StorageID+"|"+file.Path] = s }}, nil
}

func (v *volumeFiles) Remove(_ context.Context, _ database.IDB, file *entity.File) error {
	v.removed = append(v.removed, file.StorageID+"|"+file.Path)
	delete(v.files, file.StorageID+"|"+file.Path)
	return nil
}

func (v *volumeFiles) ProjectVolume(_ context.Context, _ database.IDB, projectID string) (*entity.Setting, error) {
	if vol := v.volumes[projectID]; vol != nil {
		return vol, nil
	}
	return nil, hperrors.NewNotFound("Project default volume")
}

type insertedFiles struct {
	repository.FileRepo
	inserted []*entity.File
}

func (r *insertedFiles) Insert(
	_ context.Context, _ database.IDB, file *entity.File, _ ...bunex.InsertQueryOption,
) error {
	r.inserted = append(r.inserted, file)
	return nil
}

func saveToFileJob(fileName string) *entity.Setting {
	setting := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob}
	setting.MustSetData(&entity.SchedJob{Command: &entity.CommandTemplate{Command: "pg_dump app"},
		CommandOutput: &entity.SchedJobCommandOutput{Enabled: true, SaveToFile: &entity.SchedJobCommandOutputSaveToFile{
			FileName:          fileName,
			CompressionFormat: base.FileCompressionNone,
			EncryptionFormat:  base.FileEncryptionNone,
		}}})
	return setting
}

func runSaveToFile(t *testing.T, exec containerexecservice.Service) (*volumeFiles, *insertedFiles, error) {
	t.Helper()
	return runSaveToFileNamed(t, exec, "dump.sql")
}

func runSaveToFileNamed(
	t *testing.T, exec containerexecservice.Service, fileName string,
) (*volumeFiles, *insertedFiles, error) {
	t.Helper()
	files := &volumeFiles{files: map[string]string{}, volumes: map[string]*entity.Setting{"p1": {ID: "vol-p1"}}}
	repo := &insertedFiles{}
	svc := &service{containerExecService: exec, commandService: &commandServiceStub{}, fileService: files,
		fileRepo: repo}

	_, err := svc.SchedJobExec(context.Background(), database.Tx{}, &schedjobexecservice.SchedJobExecReq{
		TaskExecData:    &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		SchedJobSetting: saveToFileJob(fileName),
		DestApp: &entity.App{ID: "a1", Key: "web", ProjectID: "p1",
			ProjectEnv: &entity.ProjectEnv{Key: "prod"}},
	})
	return files, repo, err
}

// A job's output is written to its project's default volume, by env and app.
func TestAJobOutputIsWrittenToTheProjectVolume(t *testing.T) {
	files, repo, err := runSaveToFile(t, &fakeExec{output: "dump data"})

	assert.NoError(t, err)
	if assert.Len(t, repo.inserted, 1) {
		file := repo.inserted[0]
		assert.Equal(t, base.FileStorageVolume, file.StorageType)
		assert.Equal(t, "vol-p1", file.StorageID)
		assert.Equal(t, ".hivepaas/job-output/prod/web/"+file.ID+"-dump.sql", file.Path)
		assert.Equal(t, "dump data", files.files["vol-p1|"+file.Path])
		assert.Equal(t, int64(len("dump data")), file.Size)
	}
}

type failingExec struct{ fakeExec }

func (f *failingExec) ContainerExec(
	ctx context.Context, req *containerexecservice.ContainerExecReq,
) (*containerexecservice.ContainerExecResp, error) {
	_, _ = f.fakeExec.ContainerExec(ctx, req)
	return nil, errors.New("container is gone")
}

// A job that fails leaves no file behind, and no record of one.
func TestAFailedJobLeavesNoOutputFile(t *testing.T) {
	files, repo, err := runSaveToFile(t, &failingExec{fakeExec{output: "half a dump"}})

	assert.Error(t, err)
	assert.Empty(t, repo.inserted)
	assert.Empty(t, files.files)
	assert.Len(t, files.removed, 1)
}

// A file name that climbs out of the app's own directory is refused: the volume
// also holds the data of the project's other apps.
func TestAJobOutputNamedOutOfItsDirectoryIsRefused(t *testing.T) {
	files, repo, err := runSaveToFileNamed(t, &fakeExec{output: "dump data"}, "x/../../../../../prod/db/data")

	assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot)
	assert.Empty(t, repo.inserted)
	assert.Empty(t, files.files)
}

// A file name with a directory of its own stays inside the app's directory.
func TestAJobOutputNamedWithADirectoryStaysInside(t *testing.T) {
	_, repo, err := runSaveToFileNamed(t, &fakeExec{output: "dump data"}, "daily/dump.sql")

	assert.NoError(t, err)
	if assert.Len(t, repo.inserted, 1) {
		assert.Equal(t, ".hivepaas/job-output/prod/web/"+repo.inserted[0].ID+"-daily/dump.sql", repo.inserted[0].Path)
	}
}
