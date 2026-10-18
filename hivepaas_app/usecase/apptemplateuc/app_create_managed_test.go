package apptemplateuc

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// fakeSettings answers GetByID from the settings it holds, by id and type.
type fakeSettings struct {
	repository.SettingRepo
	byID map[string]*entity.Setting
}

func (f *fakeSettings) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, typ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if setting := f.byID[id]; setting != nil && setting.Type == typ {
		return setting, nil
	}
	return nil, hperrors.NewNotFound("Setting")
}

func (f *fakeSettings) add(id string, typ base.SettingType, kind string, data entity.SettingData) {
	setting := &entity.Setting{ID: id, Type: typ, Kind: kind, Status: base.SettingStatusActive}
	setting.MustSetData(data)
	f.byID[id] = setting
}

type fakeEngine struct {
	prepared   *managedmodel.PrepareReq
	configured *managedmodel.ConfigureReq
	err        error
}

func (e *fakeEngine) Name() templatemodel.ManagedEngine { return templatemodel.ManagedEnginePostgres }
func (e *fakeEngine) APIRange() (int, int)              { return 1, 1 }

func (e *fakeEngine) PrepareDoc(req *managedmodel.PrepareReq) error {
	e.prepared = req
	return e.err
}

func (e *fakeEngine) Configure(req *managedmodel.ConfigureReq) (*managedmodel.ConfigureResp, error) {
	e.configured = req
	managed := &entity.Setting{ID: "managed-1", Type: base.SettingTypeAppManaged, ObjectID: req.App.ID}
	return &managedmodel.ConfigureResp{Settings: []*entity.Setting{managed}}, nil
}

func (e *fakeEngine) Status(*managedmodel.StatusReq) (*managedmodel.Status, error) {
	return &managedmodel.Status{}, nil
}

func (e *fakeEngine) Reconfigure(*managedmodel.ReconfigureReq) (*managedmodel.ReconfigureResp, error) {
	return &managedmodel.ReconfigureResp{}, nil
}

type fakeManaged struct{ engine *fakeEngine }

func (f *fakeManaged) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	if name != templatemodel.ManagedEnginePostgres {
		return nil, false
	}
	return f.engine, true
}

type fakeArtifacts struct {
	artifactservice.Service
	specs  []*swarm.ServiceSpec
	mounts []mount.Mount
}

func (f *fakeArtifacts) MountsFor(_ context.Context, _ *entity.AppArtifactsSettings,
	spec *swarm.ServiceSpec) ([]mount.Mount, error) {
	f.specs = append(f.specs, spec)
	return f.mounts, nil
}

type managedFakes struct {
	*createFakes
	settings  *fakeSettings
	engine    *fakeEngine
	artifacts *fakeArtifacts
}

// newManagedTest is the create test's template made managed, its data on vol-1,
// which is pinned to node n1, with an S3 repository, a volume repository and a
// key to choose from.
func newManagedTest(t *testing.T) (*UC, *managedFakes) {
	t.Helper()
	uc, create := newCreateTest(t)
	resp := create.templates.resp
	resp.Template.Metadata.Requires = templatemodel.Requires{VersionCode: "v000002",
		Features: []string{"managed-databases"}}
	resp.Template.Managed = &templatemodel.Managed{Engine: templatemodel.ManagedEnginePostgres, EngineAPI: 1,
		Postgres: &templatemodel.ManagedPostgresDefaults{RetainFull: 9, BaseBackupSchedule: "0 3 * * *"}}
	resp.Template.Artifacts = []*templatemodel.Artifact{{Name: "wal-g", Version: "3.0.8",
		MountPath: "/opt/hivepaas/bin/wal-g", Files: map[base.CPUArch]*templatemodel.ArtifactFile{
			base.CPUArchAMD64: {URL: "https://example.com/wal-g", SHA256: "aa", Size: 1}}}}
	resp.Result.Doc.Deployment.Storage = &specmodel.Storage{Mounts: map[string]specmodel.Mount{
		"/var/lib/postgresql": {Type: mount.TypeVolume, Source: "vol-1"}}}

	settings := &fakeSettings{byID: map[string]*entity.Setting{}}
	settings.add("vol-1", base.SettingTypeClusterVolume, "", &entity.ClusterVolume{NodeID: "n1"})
	settings.add("ka-1", base.SettingTypeKeyAuth, "", &entity.KeyAuth{KeyID: "AKIA-REPO",
		SecretKey: entity.NewEncryptedField("repo-secret")})
	settings.add("ka-narrow", base.SettingTypeKeyAuth, "", &entity.KeyAuth{KeyID: "AKIA-NARROW",
		SecretKey: entity.NewEncryptedField("narrow-secret")})
	settings.add("cs-1", base.SettingTypeCloudStorage, string(base.CloudStorageKindS3), &entity.CloudStorage{
		S3: &entity.CloudStorageS3{KeyAuth: entity.ObjectID{ID: "ka-1"}, Region: "auto", Bucket: "backups",
			Endpoint: "https://s3.example.com"}})
	settings.add("repo-s3", base.SettingTypeBackupRepo, "", &entity.BackupRepo{
		CloudStorage: entity.ObjectID{ID: "cs-1"}, StoragePrefix: "kopia"})
	settings.add("vol-same", base.SettingTypeClusterVolume, "", &entity.ClusterVolume{NodeID: "n1"})
	settings.add("vol-other", base.SettingTypeClusterVolume, "", &entity.ClusterVolume{NodeID: "n2"})
	settings.add("vol-shared", base.SettingTypeClusterVolume, "", &entity.ClusterVolume{})
	for _, volume := range []string{"vol-same", "vol-other", "vol-shared"} {
		settings.add("repo-"+volume, base.SettingTypeBackupRepo, "", &entity.BackupRepo{
			Volume: entity.ObjectID{ID: volume}, StoragePrefix: "repo"})
	}

	fakes := &managedFakes{createFakes: create, settings: settings, engine: &fakeEngine{},
		artifacts: &fakeArtifacts{mounts: []mount.Mount{
			artifactservice.Mount("/var/lib/hivepaas/cache/artifacts/sha256/aa", "/opt/hivepaas/bin/wal-g")}}}
	uc.settingRepo = settings
	uc.managedService = &fakeManaged{engine: fakes.engine}
	uc.artifactService = fakes.artifacts
	return uc, fakes
}

func managedReq(storage apptemplatedto.ManagedStorageReq) *apptemplatedto.CreateAppFromTemplateReq {
	req := testCreateReq()
	req.Managed = &apptemplatedto.ManagedCreateReq{Storage: storage}
	return req
}

func plan(t *testing.T, uc *UC, fakes *managedFakes, req *apptemplatedto.CreateAppFromTemplateReq) (
	*managedPlan, error) {
	t.Helper()
	return uc.planManaged(context.Background(), req, planApps(req, fakes.templates.resp))
}

func TestAnOrdinaryTemplateHasNoManagedPlan(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := testCreateReq()

	planned, err := uc.planManaged(context.Background(), req, planApps(req, fakes.templates.resp))

	assert.NoError(t, err)
	assert.Nil(t, planned)
}

func TestAManagedTemplateNeedsAnArchive(t *testing.T) {
	uc, fakes := newManagedTest(t)

	_, err := plan(t, uc, fakes, testCreateReq())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
}

// Its artifacts and its archive are decided by the node its data is on, and a
// volume any node may mount gives no node to decide by.
func TestAManagedAppsDataMustBePinned(t *testing.T) {
	uc, fakes := newManagedTest(t)
	fakes.settings.add("vol-1", base.SettingTypeClusterVolume, "", &entity.ClusterVolume{})

	_, err := plan(t, uc, fakes, managedReq(apptemplatedto.ManagedStorageReq{
		BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}}))

	assert.ErrorIs(t, err, hperrors.ErrAppManagedVolumeNotPinned)
	assert.Nil(t, fakes.engine.prepared, "nothing is prepared")
}

func TestAManagedAppArchivesToAnS3Repository(t *testing.T) {
	uc, fakes := newManagedTest(t)

	planned, err := plan(t, uc, fakes, managedReq(apptemplatedto.ManagedStorageReq{
		BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}}))

	assert.NoError(t, err)
	assert.Equal(t, managedPlacementOffSite, planned.placement)
	prepared := fakes.engine.prepared
	if assert.NotNil(t, prepared) && assert.NotNil(t, prepared.Archive.S3) {
		s3 := prepared.Archive.S3
		assert.Equal(t, "backups", s3.Bucket)
		assert.Equal(t, "kopia/walg/main-db", s3.Prefix, "beside the repository's own, never in it")
		assert.Equal(t, "AKIA-REPO", s3.AccessKey)
		assert.Equal(t, "repo-secret", s3.SecretKey)
		assert.True(t, s3.ForcePathStyle, "an S3-compatible endpoint")
		assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{64}$`), prepared.EncryptionKey)
		assert.Same(t, fakes.templates.resp.Result.Doc, prepared.Doc)
		assert.Equal(t, 9, prepared.Options.RetainFull, "the template's default")
		assert.Equal(t, "0 3 * * *", prepared.Options.BaseBackupSchedule)
	}
	assert.Equal(t, entity.ManagedStorage{BackupRepo: entity.ObjectID{ID: "repo-s3"}, Prefix: "kopia/walg/main-db",
		Placement: managedPlacementOffSite}, planned.storage)
	if assert.Len(t, fakes.artifacts.specs, 1) {
		assert.Equal(t, []string{"node.id==n1"}, fakes.artifacts.specs[0].TaskTemplate.Placement.Constraints,
			"the artifacts go where the data is")
	}
	assert.Equal(t, "wal-g", planned.artifacts.Artifacts[0].Name)
}

func TestTheRequestOverridesTheTemplatesDefaults(t *testing.T) {
	uc, fakes := newManagedTest(t)
	req := managedReq(apptemplatedto.ManagedStorageReq{CloudStorage: &basedto.ObjectIDReq{ID: "cs-1"},
		KeyAuth: &basedto.ObjectIDReq{ID: "ka-narrow"}})
	req.Managed.RetainFull = 3
	req.Managed.ArchiveTimeout = timeutil.Duration(2 * time.Minute)

	planned, err := plan(t, uc, fakes, req)

	assert.NoError(t, err)
	prepared := fakes.engine.prepared
	assert.Equal(t, 3, prepared.Options.RetainFull)
	assert.Equal(t, 2*time.Minute, prepared.Options.ArchiveTimeout)
	assert.Equal(t, "walg/main-db", prepared.Archive.S3.Prefix)
	assert.Equal(t, "AKIA-NARROW", prepared.Archive.S3.AccessKey, "the narrower key, not the storage's")
	assert.Equal(t, "narrow-secret", prepared.Archive.S3.SecretKey)
	assert.Equal(t, entity.ManagedStorage{CloudStorage: entity.ObjectID{ID: "cs-1"},
		KeyAuth: entity.ObjectID{ID: "ka-narrow"}, Prefix: "walg/main-db", Placement: managedPlacementOffSite},
		planned.storage)
}

func TestAVolumeRepositoryArchivesWhereTheDatabaseCanMountIt(t *testing.T) {
	for repo, want := range map[string]string{
		"repo-vol-same":   managedPlacementSameNode,
		"repo-vol-shared": managedPlacementShared,
	} {
		uc, fakes := newManagedTest(t)

		planned, err := plan(t, uc, fakes, managedReq(apptemplatedto.ManagedStorageReq{
			BackupRepo: &basedto.ObjectIDReq{ID: repo}}))

		assert.NoError(t, err, repo)
		assert.Equal(t, want, planned.placement, repo)
		assert.Equal(t, want, planned.storage.Placement, "%s: kept for Status", repo)
		if volume := fakes.engine.prepared.Archive.Volume; assert.NotNil(t, volume, repo) {
			assert.Equal(t, "repo/walg/main-db", volume.Subpath, repo)
		}
	}
}

func TestAVolumeRepositoryOnAnotherNodeIsRefused(t *testing.T) {
	uc, fakes := newManagedTest(t)

	_, err := plan(t, uc, fakes, managedReq(apptemplatedto.ManagedStorageReq{
		BackupRepo: &basedto.ObjectIDReq{ID: "repo-vol-other"}}))

	assert.ErrorIs(t, err, hperrors.ErrAppManagedArchiveUnreachable)
}

// The engine finishes the settings the document was built into, and the app is
// given its artifacts: their setting, and their mounts on the service.
func TestConfigureAddsTheEnginesSettingsAndTheArtifacts(t *testing.T) {
	uc, fakes := newManagedTest(t)
	req := managedReq(apptemplatedto.ManagedStorageReq{BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}})
	apps := planApps(req, fakes.templates.resp)
	planned, err := uc.planManaged(context.Background(), req, apps)
	assert.NoError(t, err)
	planned.attach(apps)

	provisioned, err := uc.provisionAll(context.Background(), database.Tx{}, testAuth(), req, apps)

	assert.NoError(t, err)
	app := provisioned.Apps[0].App
	types := map[base.SettingType]bool{}
	for _, setting := range app.Settings {
		types[setting.Type] = true
	}
	assert.True(t, types[base.SettingTypeAppManaged], "the engine's")
	assert.True(t, types[base.SettingTypeAppArtifacts])
	assert.True(t, types[base.SettingTypeAppTemplate], "still bound to its template")
	if configured := fakes.engine.configured; assert.NotNil(t, configured) {
		assert.Equal(t, app.ID, configured.App.ID)
		assert.Equal(t, planned.storage, configured.Storage)
	}
	assert.Contains(t, fakes.provision.specs[0].TaskTemplate.ContainerSpec.Mounts, fakes.artifacts.mounts[0])
}

func TestABadScheduleIsRefusedBeforeAnythingIsMade(t *testing.T) {
	uc, fakes := newManagedTest(t)
	req := managedReq(apptemplatedto.ManagedStorageReq{BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}})
	req.Managed.RestoreTestSchedule = "every sunday"

	_, err := plan(t, uc, fakes, req)

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
	assert.Nil(t, fakes.engine.prepared)
}

func TestAManagedRequestForAnOrdinaryTemplateIsRefused(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := managedReq(apptemplatedto.ManagedStorageReq{BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}})

	_, err := uc.planManaged(context.Background(), req, planApps(req, fakes.templates.resp))

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
}
