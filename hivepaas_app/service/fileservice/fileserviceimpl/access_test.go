package fileserviceimpl

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	agentfile "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/fileservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type fakeVolumes struct {
	repository.SettingRepo
	volumes map[string]*entity.Setting
}

func (f *fakeVolumes) GetByID(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType, id string, _ bool,
	_ ...bunex.SelectQueryOption,
) (*entity.Setting, error) {
	if v := f.volumes[id]; v != nil {
		return v, nil
	}
	return nil, hperrors.NewNotFound("Volume")
}

type fakeNodes struct {
	docker.Manager
	current string
}

func (f *fakeNodes) NodeCurrentID(context.Context) (string, error) { return f.current, nil }

type fakeAgents struct {
	agentservice.Service
	labeled map[string][]string
}

func (f *fakeAgents) GetAgentAddrForNode(_ context.Context, nodeID string) (string, error) {
	return "agent@" + nodeID, nil
}

func (f *fakeAgents) NodeIDsWithLabel(_ context.Context, label string) ([]string, error) {
	return f.labeled[label], nil
}

// agentDisk is the files the agents hold, by agent address, root and path.
type agentDisk struct {
	files map[string]string
	calls []string
}

func (d *agentDisk) client(addr string) (agentfile.FileServiceClient, error) {
	return &agentDiskClient{disk: d, addr: addr}, nil
}

type agentDiskClient struct {
	agentfile.FileServiceClient
	disk *agentDisk
	addr string
}

func (c *agentDiskClient) key(root, path string) string { return c.addr + ":" + root + "|" + path }

func (c *agentDiskClient) Read(_ context.Context, root, path string) (io.ReadCloser, error) {
	c.disk.calls = append(c.disk.calls, "read "+c.key(root, path))
	content, ok := c.disk.files[c.key(root, path)]
	if !ok {
		return nil, hperrors.NewNotFound("File")
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func (c *agentDiskClient) Write(_ context.Context, root, path string, content io.Reader) (int64, error) {
	c.disk.calls = append(c.disk.calls, "write "+c.key(root, path))
	b, err := io.ReadAll(content)
	if err != nil {
		return 0, err
	}
	c.disk.files[c.key(root, path)] = string(b)
	return int64(len(b)), nil
}

func (c *agentDiskClient) Remove(_ context.Context, root, path string) error {
	c.disk.calls = append(c.disk.calls, "remove "+c.key(root, path))
	delete(c.disk.files, c.key(root, path))
	return nil
}

func (c *agentDiskClient) Stat(_ context.Context, root, path string) (int64, error) {
	content, ok := c.disk.files[c.key(root, path)]
	if !ok {
		return 0, hperrors.NewNotFound("File")
	}
	return int64(len(content)), nil
}

func (c *agentDiskClient) Close() error { return nil }

func volume(t *testing.T, id string, vol *entity.ClusterVolume) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, RefID: id, Type: base.SettingTypeClusterVolume, Name: id}
	assert.NoError(t, setting.SetData(vol))
	return setting
}

func bindOpts(dir string) map[string]string {
	return map[string]string{"type": "none", "device": dir, "o": "bind,rw"}
}

func newAccessService(t *testing.T, volumes ...*entity.Setting) (*service, *agentDisk) {
	t.Helper()
	disk := &agentDisk{files: map[string]string{}}
	byID := map[string]*entity.Setting{}
	for _, v := range volumes {
		byID[v.ID] = v
	}
	return &service{
		settingRepo:   &fakeVolumes{volumes: byID},
		dockerManager: &fakeNodes{current: "node-mgr"},
		agentService: &fakeAgents{labeled: map[string][]string{
			"disk=one": {"node-3"}, "disk=many": {"node-3", "node-4"},
		}},
		agentFiles: disk.client,
	}, disk
}

func writeFile(t *testing.T, s *service, file *entity.File, content string) error {
	t.Helper()
	w, err := s.Create(context.Background(), nil, file)
	if err != nil {
		return err
	}
	if _, err = io.Copy(w, strings.NewReader(content)); err != nil {
		return err
	}
	return w.Close()
}

func readFile(t *testing.T, s *service, file *entity.File) string {
	t.Helper()
	r, err := s.Open(context.Background(), nil, file)
	if !assert.NoError(t, err) {
		return ""
	}
	defer r.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// A file with no volume named is on the system volume: HivePaaS's data
// directory, read and written directly.
func TestAFileOnTheSystemVolumeIsUnderTheDataDirectory(t *testing.T) {
	appPath := t.TempDir()
	config.SetCurrent(&config.Config{AppPath: appPath})
	t.Cleanup(func() { config.SetCurrent(nil) })
	s, disk := newAccessService(t)
	file := &entity.File{StorageType: base.FileStorageVolume, Path: "files/01J-report.csv"}

	assert.NoError(t, writeFile(t, s, file, "a,b"))

	onDisk, _ := os.ReadFile(filepath.Join(appPath, "files/01J-report.csv"))
	assert.Equal(t, "a,b", string(onDisk))
	assert.Equal(t, "a,b", readFile(t, s, file))
	size, err := s.Stat(context.Background(), nil, file)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), size)
	assert.NoError(t, s.Remove(context.Background(), nil, file))
	assert.NoError(t, s.Remove(context.Background(), nil, file), "removing what is gone is done")
	_, err = s.Stat(context.Background(), nil, file)
	assert.ErrorIs(t, err, hperrors.ErrNotFound)
	assert.Empty(t, disk.calls, "no agent is asked")
}

// A file on a volume pinned to a node goes to that node's agent, inside the
// volume's directory.
func TestAFileOnAPinnedVolumeGoesToItsNode(t *testing.T) {
	s, disk := newAccessService(t, volume(t, "vol-p1", &entity.ClusterVolume{
		NodeID: "node-2", Driver: "local", DriverOpts: bindOpts("/srv/project_data/p1"),
	}))
	file := &entity.File{StorageType: base.FileStorageVolume, StorageID: "vol-p1",
		Path: ".hivepaas/cache/repos/r1.tar.lz4"}

	assert.NoError(t, writeFile(t, s, file, "cache"))
	assert.Equal(t, "cache", readFile(t, s, file))
	assert.NoError(t, s.Remove(context.Background(), nil, file))

	assert.Equal(t, []string{
		"write agent@node-2:/srv/project_data/p1|.hivepaas/cache/repos/r1.tar.lz4",
		"read agent@node-2:/srv/project_data/p1|.hivepaas/cache/repos/r1.tar.lz4",
		"remove agent@node-2:/srv/project_data/p1|.hivepaas/cache/repos/r1.tar.lz4",
	}, disk.calls)
}

// A shared bind volume is the same directory everywhere: reached through the
// agent of the node asking.
func TestAFileOnASharedVolumeGoesToTheCurrentNode(t *testing.T) {
	s, disk := newAccessService(t, volume(t, "vol-nfs", &entity.ClusterVolume{
		Driver: "local", DriverOpts: bindOpts("/mnt/shared/p1"),
	}))
	file := &entity.File{StorageType: base.FileStorageVolume, StorageID: "vol-nfs", Path: "a.txt"}

	assert.NoError(t, writeFile(t, s, file, "x"))

	assert.Equal(t, []string{"write agent@node-mgr:/mnt/shared/p1|a.txt"}, disk.calls)
}

// A volume pinned by a label is on the one node the label names; a label naming
// several nodes cannot say which holds the files.
func TestAVolumePinnedByALabelHoldsFilesOnlyOnOneNode(t *testing.T) {
	s, disk := newAccessService(t,
		volume(t, "vol-one", &entity.ClusterVolume{NodeLabel: "disk=one", DriverOpts: bindOpts("/data")}),
		volume(t, "vol-many", &entity.ClusterVolume{NodeLabel: "disk=many", DriverOpts: bindOpts("/data")}),
	)

	assert.NoError(t, writeFile(t, s,
		&entity.File{StorageType: base.FileStorageVolume, StorageID: "vol-one", Path: "f"}, "x"))
	assert.Equal(t, []string{"write agent@node-3:/data|f"}, disk.calls)

	err := writeFile(t, s, &entity.File{StorageType: base.FileStorageVolume, StorageID: "vol-many", Path: "f"}, "x")
	assert.ErrorIs(t, err, hperrors.ErrVolumeCannotHoldFiles)
}

// A volume every node mounts through a driver is on no host path.
func TestAnNFSVolumeCannotHoldFiles(t *testing.T) {
	s, _ := newAccessService(t, volume(t, "vol-nfs", &entity.ClusterVolume{
		Driver: "local", DriverOpts: map[string]string{"type": "nfs", "device": ":/exports", "o": "addr=10.0.0.5"},
	}))

	_, err := s.Open(context.Background(), nil,
		&entity.File{StorageType: base.FileStorageVolume, StorageID: "vol-nfs", Path: "f"})

	assert.ErrorIs(t, err, hperrors.ErrBackupVolumeSharedNotBind)
}

// A path stored for a file stays inside its volume, whatever the database says.
func TestAFilePathOutOfItsVolumeIsRefused(t *testing.T) {
	config.SetCurrent(&config.Config{AppPath: t.TempDir()})
	t.Cleanup(func() { config.SetCurrent(nil) })
	s, _ := newAccessService(t)

	for _, path := range []string{"../etc/passwd", "/etc/passwd", ""} {
		_, err := s.Open(context.Background(), nil, &entity.File{StorageType: base.FileStorageVolume, Path: path})
		assert.ErrorIs(t, err, hperrors.ErrFilePathOutsideRoot, path)
	}
}

// A write given up half way leaves nothing, on the system volume as on another.
func TestAnAbortedWriteLeavesNothing(t *testing.T) {
	appPath := t.TempDir()
	config.SetCurrent(&config.Config{AppPath: appPath})
	t.Cleanup(func() { config.SetCurrent(nil) })
	s, disk := newAccessService(t, volume(t, "vol-p1", &entity.ClusterVolume{
		NodeID: "node-2", Driver: "local", DriverOpts: bindOpts("/srv/p1"),
	}))

	for _, file := range []*entity.File{
		{StorageType: base.FileStorageVolume, Path: "files/out.sql"},
		{StorageType: base.FileStorageVolume, StorageID: "vol-p1", Path: "out.sql"},
	} {
		w, err := s.Create(context.Background(), nil, file)
		if !assert.NoError(t, err) {
			continue
		}
		_, _ = w.Write([]byte("half of a dump"))
		w.Abort(io.ErrUnexpectedEOF)
	}

	entries, _ := os.ReadDir(filepath.Join(appPath, "files"))
	assert.Empty(t, entries)
	assert.Empty(t, disk.files)
}

func (f *fakeVolumes) List(
	_ context.Context, _ database.IDB, scope *entity.ObjectScope, _ *basedto.Paging, _ ...bunex.SelectQueryOption,
) ([]*entity.Setting, *basedto.PagingMeta, error) {
	var found []*entity.Setting
	for _, v := range f.volumes {
		if v.Default && v.ObjectID == scope.ProjectID {
			found = append(found, v)
		}
	}
	return found, nil, nil
}

func projectDefaultVolume(t *testing.T, projectID, dir string) *entity.Setting {
	t.Helper()
	v := volume(t, "vol-"+projectID, &entity.ClusterVolume{NodeID: "node-2", Driver: "local", DriverOpts: bindOpts(dir)})
	v.Default, v.ObjectID, v.Scope = true, projectID, base.ObjectScopeProject
	return v
}

// A file uploaded to an app goes to its project's default volume, in
// HivePaaS's own directory, by env and app.
func TestAFileUploadedToAnAppGoesToItsProjectVolume(t *testing.T) {
	s, disk := newAccessService(t, projectDefaultVolume(t, "p1", "/srv/project_data/shop"))
	s.fileRepo = nil
	app := &entity.App{ID: "a1", Key: "web", ProjectID: "p1", ProjectEnv: &entity.ProjectEnv{Key: "prod"}}

	resp, err := s.Upload(context.Background(), nil, &fileservice.UploadReq{
		Scope:       &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "a1", ProjectID: "p1"},
		App:         app,
		FileType:    base.FileTypeDataFile,
		StorageType: base.FileStorageVolume,
		Items: []*fileservice.UploadItemReq{
			{FilePath: "dump.sql", FileSize: 4, FileData: io.NopCloser(strings.NewReader("data"))},
		},
	})

	if assert.NoError(t, err) && assert.Len(t, resp.Files, 1) {
		file := resp.Files[0]
		assert.Equal(t, "vol-p1", file.StorageID)
		assert.Equal(t, ".hivepaas/files/prod/web/"+file.ID+"-dump.sql", file.Path)
		assert.Equal(t, "data", disk.files["agent@node-2:/srv/project_data/shop|"+file.Path])
	}
}

// A file uploaded anywhere else is on the system volume, named by its id so two
// files of one name never meet.
func TestAFileUploadedOutsideAProjectIsOnTheSystemVolume(t *testing.T) {
	appPath := t.TempDir()
	config.SetCurrent(&config.Config{AppPath: appPath})
	t.Cleanup(func() { config.SetCurrent(nil) })
	s, disk := newAccessService(t)

	resp, err := s.Upload(context.Background(), nil, &fileservice.UploadReq{
		Scope:       &entity.ObjectScope{ScopeType: base.ObjectScopeUser, UserID: "u1"},
		FileType:    base.FileTypeTmp,
		StorageType: base.FileStorageVolume,
		Items: []*fileservice.UploadItemReq{
			{FilePath: "logo.png", FileSize: 3, FileData: io.NopCloser(strings.NewReader("png"))},
		},
	})

	if assert.NoError(t, err) && assert.Len(t, resp.Files, 1) {
		file := resp.Files[0]
		assert.Empty(t, file.StorageID)
		assert.Equal(t, "files/"+file.ID+"-logo.png", file.Path)
		onDisk, _ := os.ReadFile(filepath.Join(appPath, file.Path))
		assert.Equal(t, "png", string(onDisk))
	}
	assert.Empty(t, disk.calls)
}
