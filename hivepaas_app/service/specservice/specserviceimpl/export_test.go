package specserviceimpl

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// The fakes embed the real interfaces, so a method the exporter starts calling
// without a stub panics loudly instead of silently returning a zero value.

// fakeSettingRepo serves only loadNetworkNames, which asks for every
// cluster-network row. Scope queries go through the settingLoader seam instead,
// because bunex options are closures a double cannot read.
type fakeSettingRepo struct {
	repository.SettingRepo
	networks []*entity.Setting
}

func (f *fakeSettingRepo) List(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption,
) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.networks, nil, nil
}

type fakeProjectRepo struct {
	repository.ProjectRepo
	projects []*entity.Project
}

func (f *fakeProjectRepo) List(
	_ context.Context, _ database.IDB, _ *basedto.Paging, _ ...bunex.SelectQueryOption,
) ([]*entity.Project, *basedto.PagingMeta, error) {
	return f.projects, nil, nil
}

func (f *fakeProjectRepo) GetByID(
	_ context.Context, _ database.IDB, id string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	return f.find(func(p *entity.Project) bool { return p.ID == id })
}

func (f *fakeProjectRepo) GetByKey(
	_ context.Context, _ database.IDB, key string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	return f.find(func(p *entity.Project) bool { return p.Key == key })
}

func (f *fakeProjectRepo) GetByName(
	_ context.Context, _ database.IDB, name string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	return f.find(func(p *entity.Project) bool { return p.Name == name })
}

func (f *fakeProjectRepo) find(match func(*entity.Project) bool) (*entity.Project, error) {
	for _, project := range f.projects {
		if match(project) {
			return project, nil
		}
	}
	return nil, hperrors.Wrap(hperrors.ErrProjectNotFound)
}

// fakeUserRepo knows the fixture's users: the project's owner, a disabled
// user, and another active one.
type fakeUserRepo struct {
	repository.UserRepo
	users []*entity.User
}

func (f *fakeUserRepo) GetByID(
	_ context.Context, _ database.IDB, id string, _ ...bunex.SelectQueryOption,
) (*entity.User, error) {
	return f.find(func(u *entity.User) bool { return u.ID == id })
}

func (f *fakeUserRepo) GetByEmail(
	_ context.Context, _ database.IDB, email string, _ ...bunex.SelectQueryOption,
) (*entity.User, error) {
	return f.find(func(u *entity.User) bool { return strings.EqualFold(u.Email, email) })
}

func (f *fakeUserRepo) find(match func(*entity.User) bool) (*entity.User, error) {
	for _, user := range f.users {
		if match(user) {
			return user, nil
		}
	}
	return nil, hperrors.Wrap(hperrors.ErrUserNotFound)
}

type fakeProjectEnvRepo struct {
	repository.ProjectEnvRepo
	envs []*entity.ProjectEnv
}

func (f *fakeProjectEnvRepo) List(
	_ context.Context, _ database.IDB, projectID string, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption,
) ([]*entity.ProjectEnv, *basedto.PagingMeta, error) {
	var out []*entity.ProjectEnv
	for _, env := range f.envs {
		if env.ProjectID == projectID {
			out = append(out, env)
		}
	}
	return out, nil, nil
}

func (f *fakeProjectEnvRepo) GetByKey(
	_ context.Context, _ database.IDB, projectID, key string, _ ...bunex.SelectQueryOption,
) (*entity.ProjectEnv, error) {
	for _, env := range f.envs {
		if env.ProjectID == projectID && env.Key == key {
			return env, nil
		}
	}
	return nil, hperrors.Wrap(hperrors.ErrProjectEnvNotFound)
}

type fakeAppRepo struct {
	repository.AppRepo
	apps []*entity.App
}

func (f *fakeAppRepo) List(
	_ context.Context, _ database.IDB, projectID string, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption,
) ([]*entity.App, *basedto.PagingMeta, error) {
	var out []*entity.App
	for _, app := range f.apps {
		if app.ProjectID == projectID {
			out = append(out, app)
		}
	}
	return out, nil, nil
}

func (f *fakeAppRepo) GetByID(
	_ context.Context, _ database.IDB, projectID, id string, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	for _, app := range f.apps {
		if app.ID == id && (projectID == "" || app.ProjectID == projectID) {
			return app, nil
		}
	}
	return nil, hperrors.Wrap(hperrors.ErrAppNotFound)
}

type fakeClusterService struct {
	clusterservice.Service
	services map[string]*swarm.Service
	// ports are the published ports other services hold, by service id.
	ports map[clusterservice.PortRef]string
	// updated are the specs each service was updated to; failUpdate makes one
	// service's update fail.
	updated    map[string][]*swarm.ServiceSpec
	failUpdate map[string]error
}

func (f *fakeClusterService) VerifyPortsAvailable(
	_ context.Context, ports []clusterservice.PortRef, ignoreServiceIDs []string,
) error {
	for _, port := range ports {
		if by, taken := f.ports[port]; taken && !slices.Contains(ignoreServiceIDs, by) {
			return hperrors.Wrap(hperrors.ErrPortInUse)
		}
	}
	return nil
}

// fakeDomainService holds domains for apps, by app id.
type fakeDomainService struct {
	domainservice.Service
	held map[string]string
}

func (f *fakeDomainService) VerifyDomainsAvailable(
	_ context.Context, _ database.IDB, domains []string, ignoreAppIDs []string,
) error {
	for _, domain := range domains {
		if by, taken := f.held[domain]; taken && !slices.Contains(ignoreAppIDs, by) {
			return hperrors.Wrap(hperrors.ErrDomainInUse)
		}
	}
	return nil
}

func (f *fakeClusterService) ServiceInspect(
	_ context.Context, serviceID string, _ bool,
) (*swarm.Service, error) {
	if svc, ok := f.services[serviceID]; ok {
		return svc, nil
	}
	return nil, notFoundError{}
}

// fakeExportVolumeService answers DescribeAppMounts from a table keyed by mount
// target, which is all export asks of volumeservice.
type fakeExportVolumeService struct {
	volumeservice.Service
	descs map[string]*volumeservice.AppMountDesc
	// storage answers InspectAppStorage for a volume id; a volume it does not
	// name is checked and empty.
	storage map[string]*volumeservice.AppStorageState
}

func (f *fakeExportVolumeService) InspectAppStorage(
	_ context.Context, _ database.IDB, req *volumeservice.InspectAppStorageReq,
) (*volumeservice.InspectAppStorageResp, error) {
	resp := &volumeservice.InspectAppStorageResp{}
	for _, query := range req.Queries {
		state := &volumeservice.AppStorageState{Checked: true, Exists: true, Empty: true}
		if known := f.storage[query.VolumeID]; known != nil {
			copied := *known
			state = &copied
		}
		state.AppKey, state.VolumeID = query.AppKey, query.VolumeID
		resp.States = append(resp.States, state)
	}
	return resp, nil
}

func (f *fakeExportVolumeService) DescribeAppMounts(
	_ context.Context, _ database.IDB, _ *entity.App, mounts []mount.Mount,
) ([]*volumeservice.AppMountDesc, error) {
	out := make([]*volumeservice.AppMountDesc, len(mounts))
	for i := range mounts {
		if out[i] = f.descs[mounts[i].Target]; out[i] == nil {
			out[i] = &volumeservice.AppMountDesc{}
		}
	}
	return out, nil
}

type notFoundError struct{}

func (notFoundError) Error() string { return "service not found" }

// exportFixture builds a small but realistic installation: one exportable
// project with one env and two apps, one of which has never been deployed;
// plus the hivepaas project, which must not appear.
func exportFixture(t *testing.T) specservice.Service {
	t.Helper()
	useDataKey(t)

	cert := &entity.Setting{
		ID: "cert_1", Type: base.SettingTypeSSLCert, Scope: base.ObjectScopeGlobal,
		Name: "localhost", Kind: "self-signed", Status: base.SettingStatusActive, Version: 1,
	}
	assert.NoError(t, cert.SetData(&entity.SSLCert{Domain: "localhost"}))

	// An app setting referencing a global one. Resolving it proves the index is
	// complete before any document is written.
	routing := &entity.Setting{
		ID: "routing_1", Type: base.SettingTypeAppRouting, Scope: base.ObjectScopeApp,
		ObjectID: "app_1", Status: base.SettingStatusActive, Version: 1,
	}
	assert.NoError(t, routing.SetData(&entity.AppRoutingSettings{
		Port: 8080,
		// Enabled, because a disabled domain's certificate is not a reference
		// GetRefObjectIDs reports - it holds no reservation - and export learns
		// what an exported setting reaches outside the export from that.
		Domains: []*entity.AppDomain{
			{Domain: "api.example.com", Enabled: true, SSLCert: entity.ObjectID{ID: "cert_1"}},
		},
	}))

	secret := &entity.Setting{
		ID: "secret_1", Type: base.SettingTypeSecret, Scope: base.ObjectScopeApp,
		ObjectID: "app_1", Name: "db-password", Status: base.SettingStatusActive,
	}
	assert.NoError(t, secret.SetData(&entity.Secret{
		Key: "DB_PASSWORD", Value: entity.NewEncryptedField("hunter2"),
	}))

	// The backend mounts the certificate the routing uses.
	mountEntry := &entity.Setting{
		ID: "mount_1", Type: base.SettingTypeAppSettingMount, Scope: base.ObjectScopeApp,
		ObjectID: "app_1", Name: "cert", Status: base.SettingStatusActive, Version: entity.CurrentAppSettingMountVersion,
	}
	assert.NoError(t, mountEntry.SetData(&entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"},
		Files:  []*entity.AppSettingMountFile{{Part: "certificate", Path: "/etc/app/tls/cert.pem"}},
	}))

	// The backend is a database, whose credential is a secret HivePaaS owns.
	kind := &entity.Setting{
		ID: "kind_1", Type: base.SettingTypeAppKind, Scope: base.ObjectScopeApp,
		ObjectID: "app_1", Status: base.SettingStatusActive, Version: entity.CurrentAppKindSettingsVersion,
	}
	assert.NoError(t, kind.SetData(&entity.AppKindSettings{
		Category: base.AppCategoryDatabase, Engine: "postgres",
		Database: &entity.AppKindDatabase{DbName: "app", Username: "app", Password: entity.NewEncryptedField("s3cret")},
	}))

	apiKey := &entity.Setting{
		ID: "key_1", Type: base.SettingTypeAPIKey, Scope: base.ObjectScopeGlobal,
		Name: "ci", Status: base.SettingStatusActive,
	}
	assert.NoError(t, apiKey.SetData(&entity.APIKey{KeyID: "abc"}))

	// The project's own volume, which the export holds, and a volume sync
	// discovered at global scope, which it never does.
	projectVolume := &entity.Setting{
		ID: "vol_setting_1", Type: base.SettingTypeClusterVolume, Scope: base.ObjectScopeProject,
		ObjectID: "p1", Name: "default", Status: base.SettingStatusActive,
	}
	assert.NoError(t, projectVolume.SetData(&entity.ClusterVolume{Managed: true}))
	sharedVolume := &entity.Setting{
		ID: "gvol_1", Type: base.SettingTypeClusterVolume, Scope: base.ObjectScopeGlobal,
		Name: "shared", Status: base.SettingStatusActive,
	}
	assert.NoError(t, sharedVolume.SetData(&entity.ClusterVolume{}))

	// Config files a project and an env hold for their apps.
	projectConfig := &entity.Setting{
		ID: "cfg_p1", Type: base.SettingTypeConfigFile, Scope: base.ObjectScopeProject,
		ObjectID: "p1", Name: "shared.yaml", Status: base.SettingStatusActive, Inheritable: true,
	}
	assert.NoError(t, projectConfig.SetData(&entity.ConfigFile{Name: "shared.yaml", Content: "level: project\n"}))
	envConfig := &entity.Setting{
		ID: "cfg_dev", Type: base.SettingTypeConfigFile, Scope: base.ObjectScopeProjectEnv,
		ObjectID: "p1:dev", Name: "dev-only.yaml", Status: base.SettingStatusActive,
	}
	assert.NoError(t, envConfig.SetData(&entity.ConfigFile{Name: "dev-only.yaml", Content: "level: env\n"}))

	// The env's own network, as cluster sync records it.
	settingRepo := &fakeSettingRepo{networks: []*entity.Setting{{
		ID: "net_1", Type: base.SettingTypeClusterNetwork, Scope: base.ObjectScopeProjectEnv, ObjectID: "p1:dev",
		RefID: "8vo4p3pwm1aksdu2ilryn8mpf", Name: "project_a_dev_net", Status: base.SettingStatusActive,
	}}}

	proj := &entity.Project{
		ID: "p1", Key: "project_a", Name: "Project A",
		OwnerID: "u1", Owner: &entity.User{ID: "u1", Email: "owner@example.com"},
	}
	hive := &entity.Project{ID: "p2", Key: base.HivepaasProjectKey, Name: "HivePaaS"}
	env := &entity.ProjectEnv{ID: "p1:dev", ProjectID: "p1", Key: "dev", Name: "development"}

	deployed := &entity.App{
		ID: "app_1", Key: "backend", Name: "Backend", ProjectID: "p1",
		ProjectEnvID: "p1:dev", ServiceID: "svc_1", Status: base.AppStatusActive,
	}
	undeployed := &entity.App{
		ID: "app_2", Key: "frontend", Name: "Frontend", ProjectID: "p1",
		ProjectEnvID: "p1:dev", Status: base.AppStatusActive,
	}
	preview := &entity.App{
		ID: "app_3", Key: "backend-pr-42", Name: "PR 42", ProjectID: "p1",
		ProjectEnvID: "p1:dev", ParentID: "app_1",
	}

	all := []*entity.Setting{cert, apiKey, routing, secret, mountEntry, kind, projectVolume, sharedVolume,
		projectConfig, envConfig}

	svc := New(
		&fakeAppRepo{apps: []*entity.App{deployed, undeployed, preview}},
		&fakeProjectEnvRepo{envs: []*entity.ProjectEnv{env}},
		&fakeProjectRepo{projects: []*entity.Project{proj, hive}},
		settingRepo,
		&fakeUserRepo{users: []*entity.User{
			{ID: "u1", Email: "owner@example.com", Status: base.UserStatusActive},
			{ID: "u2", Email: "gone@example.com", Status: base.UserStatusDisabled},
			{ID: "u3", Email: "other@example.com", Status: base.UserStatusActive},
		}},
		&fakeProjectService{},
		&fakeAppService{},
		&fakeDeploymentService{},
		&fakeProvisionService{},
		&fakeRoutingService{},
		&fakeClusterService{services: map[string]*swarm.Service{"svc_1": testService()}},
		&fakeDomainService{},
		&fakeNoDockerAPI{},
		&fakeEnvVarService{},
		&fakeNetworkService{},
		&fakeSettingMounts{},
		&fakeSSLService{},
		&fakeExportVolumeService{descs: map[string]*volumeservice.AppMountDesc{
			"/var/lib/postgresql/data": {AppKey: "backend", Own: true, Subpath: "data", VolumeID: "vol_setting_1"},
			"/shared":                  {AppKey: "backend", Own: true, Subpath: "cache", VolumeID: "gvol_1"},
		}},
		&fakeTaskQueue{},
	)

	// The seam does what the repository's SQL would: return the settings this
	// scope defines, and none it merely inherits.
	impl, _ := svc.(*service)
	impl.loadOwned = func(
		_ context.Context, _ database.IDB,
		scopes []base.ObjectScopeType, objectID string,
	) ([]*entity.Setting, error) {
		var out []*entity.Setting
		for _, setting := range all {
			if !containsScope(scopes, setting.Scope) {
				continue
			}
			if setting.ObjectID != objectID {
				continue
			}
			out = append(out, fresh(t, setting))
		}
		return out, nil
	}
	impl.loadByIDs = func(_ context.Context, _ database.IDB, ids []string) ([]*entity.Setting, error) {
		var out []*entity.Setting
		for _, setting := range all {
			if slices.Contains(ids, setting.ID) {
				out = append(out, fresh(t, setting))
			}
		}
		return out, nil
	}
	// Found as the repository's scope filters would: global settings everywhere,
	// a project's or an env's where that scope sees them.
	impl.findRef = func(
		_ context.Context, _ database.IDB, scope *entity.ObjectScope, ref *specmodel.ExternalRef,
	) (*entity.Setting, error) {
		visible := func(setting *entity.Setting) bool {
			switch setting.Scope {
			case base.ObjectScopeGlobal:
				return true
			case base.ObjectScopeProject:
				return setting.ObjectID == scope.ProjectID
			case base.ObjectScopeProjectEnv:
				return setting.ObjectID == scope.ProjectEnvID
			case base.ObjectScopeHivepaas, base.ObjectScopeUser, base.ObjectScopeApp:
			}
			return false
		}
		for _, match := range []func(*entity.Setting) bool{
			func(setting *entity.Setting) bool { return ref.ID != "" && setting.ID == ref.ID },
			func(setting *entity.Setting) bool {
				return ref.Name != "" && setting.Name == ref.Name && (ref.Kind == "" || setting.Kind == ref.Kind)
			},
		} {
			for _, setting := range all {
				if string(setting.Type) == ref.Type && visible(setting) && match(setting) {
					return setting, nil
				}
			}
		}
		return nil, nil
	}
	impl.nodeExists = func(_ context.Context, _ database.IDB, nodeID string) (bool, error) {
		return nodeID == "node_1", nil
	}
	return svc
}

// fresh is a setting as a query returns it: a row of its own, so that what one
// reader does to its parsed data does not reach the next reader - export clears
// the secrets of what it renders.
func fresh(t *testing.T, setting *entity.Setting) *entity.Setting {
	t.Helper()
	cp, err := setting.Clone(false)
	assert.NoError(t, err)
	return cp
}

func containsScope(scopes []base.ObjectScopeType, want base.ObjectScopeType) bool {
	for _, scope := range scopes {
		if scope == want {
			return true
		}
	}
	return false
}

func runExport(t *testing.T, mode specmodel.SecretsMode, passphrase string) (string, *specmodel.Report) {
	t.Helper()
	return runExportAt(t, entity.NewObjectScopeGlobal(), mode, passphrase)
}

func runExportAt(
	t *testing.T, scope *entity.ObjectScope, mode specmodel.SecretsMode, passphrase string,
) (string, *specmodel.Report) {
	t.Helper()
	svc := exportFixture(t)
	dir := t.TempDir()

	resp, err := svc.Export(context.Background(), nil, &specservice.ExportReq{
		Scope:       scope,
		SecretsMode: mode,
		Passphrase:  passphrase,
		WorkDir:     dir,
	})
	assert.NoError(t, err)
	assert.FileExists(t, resp.Path)
	return resp.Path, resp.Report
}

func TestExportProducesThePerEnvLayout(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	listing, err := exec.Command("tar", "-tzf", path).Output()
	assert.NoError(t, err)
	for _, want := range []string{
		"spec.yaml",
		"global.yaml",
		"projects/project_a/project.yaml",
		"projects/project_a/envs/dev.yaml",
	} {
		assert.Contains(t, string(listing), want)
	}
}

func TestExportExcludesTheHivePaaSProject(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	listing, err := exec.Command("tar", "-tzf", path).Output()
	assert.NoError(t, err)
	assert.NotContains(t, string(listing), "projects/hivepaas",
		"that project holds HivePaaS's own stack")
}

// The index must be complete before any document is written, or an app setting
// referencing a global one would resolve or not depending on walk order.
func TestExportResolvesCrossScopeReferences(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")
	env := readFromArchive(t, path, "projects/project_a/envs/dev.yaml")

	assert.Contains(t, env, "global/sslCerts/localhost",
		"the app's certificate reference became a path into global.yaml")
	assert.NotContains(t, env, "cert_1", "and no raw identifier survived")
}

func TestExportOmitsSecretsInOmitMode(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")
	env := readFromArchive(t, path, "projects/project_a/envs/dev.yaml")

	assert.Contains(t, env, "DB_PASSWORD", "the key stays so a reader sees a secret is missing")
	assert.NotContains(t, env, "hunter2")
	assert.NotContains(t, env, "hpenc", "and no installation-specific ciphertext either")
}

func TestExportRevealsSecretsInPlaintextMode(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModePlaintext, "")
	env := readFromArchive(t, path, "projects/project_a/envs/dev.yaml")
	assert.Contains(t, env, "hunter2")
}

func TestExportSkipsApiKeysAndPreviewApps(t *testing.T) {
	path, report := runExport(t, specmodel.SecretsModeOmit, "")

	global := readFromArchive(t, path, "global.yaml")
	assert.NotContains(t, global, "apiKeys", "an imported hash authenticates nobody")

	env := readFromArchive(t, path, "projects/project_a/envs/dev.yaml")
	assert.NotContains(t, env, "backend-pr-42", "preview apps are not configuration")

	codes := map[string]int{}
	for _, issue := range report.Issues {
		codes[issue.Code]++
	}
	assert.Positive(t, codes[specmodel.CodeTypeSkipped], "the api key is reported")
	assert.Positive(t, codes[specmodel.CodePreviewAppSkipped], "so is the preview app")
}

// An app with no service is normal, not an error: two of five user apps in a
// development installation are in that state.
func TestExportHandlesAnUndeployedApp(t *testing.T) {
	path, report := runExport(t, specmodel.SecretsModeOmit, "")
	env := readFromArchive(t, path, "projects/project_a/envs/dev.yaml")

	assert.Contains(t, env, "frontend", "the app is still exported")

	var found bool
	for _, issue := range report.Issues {
		if issue.Code == specmodel.CodeServiceUnavailable && strings.Contains(issue.Path, "frontend") {
			found = true
			assert.Equal(t, specmodel.SeverityFixable, issue.Severity)
		}
	}
	assert.True(t, found, "and the report says why it has no deployment block")
}

func TestExportEncryptedModeProducesAnAgeFile(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeEncrypted, "correct horse battery staple")

	assert.True(t, strings.HasSuffix(path, ".tar.gz.age"))
	head, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(head), "age-encryption.org/"))

	// The unencrypted archive must not be left behind beside it.
	assert.NoFileExists(t, strings.TrimSuffix(path, ".age"))
}

func TestExportRefusesEncryptedWithoutAPassphrase(t *testing.T) {
	svc := exportFixture(t)
	_, err := svc.Export(context.Background(), nil, &specservice.ExportReq{
		Scope:       entity.NewObjectScopeGlobal(),
		SecretsMode: specmodel.SecretsModeEncrypted,
		WorkDir:     t.TempDir(),
	})
	assert.Error(t, err)
}

func TestExportRefusesAnUnknownMode(t *testing.T) {
	svc := exportFixture(t)
	_, err := svc.Export(context.Background(), nil, &specservice.ExportReq{
		Scope:       entity.NewObjectScopeGlobal(),
		SecretsMode: specmodel.SecretsMode("none"),
		WorkDir:     t.TempDir(),
	})
	assert.Error(t, err, "none was renamed to omit and must not be silently accepted")
}

func readFromArchive(t *testing.T, archive, name string) string {
	t.Helper()
	dir := t.TempDir()
	out, err := exec.Command("tar", "-xzf", archive, "-C", dir).CombinedOutput()
	assert.NoError(t, err, string(out))

	content, err := os.ReadFile(filepath.Join(dir, name))
	assert.NoError(t, err)
	return string(content)
}

// Two exports of unchanged data must differ only in the manifest's timestamp,
// or the format is useless for review and for git.
func TestExportIsDeterministicApartFromTheTimestamp(t *testing.T) {
	first, _ := runExport(t, specmodel.SecretsModeOmit, "")
	second, _ := runExport(t, specmodel.SecretsModeOmit, "")

	for _, name := range []string{
		"global.yaml",
		"projects/project_a/project.yaml",
		"projects/project_a/envs/dev.yaml",
	} {
		assert.Equal(t,
			readFromArchive(t, first, name),
			readFromArchive(t, second, name),
			"%s differs between two exports of the same data", name)
	}

	// The manifest differs only where it should.
	firstManifest := stripLine(readFromArchive(t, first, "spec.yaml"), "exportedAt:")
	secondManifest := stripLine(readFromArchive(t, second, "spec.yaml"), "exportedAt:")
	assert.Equal(t, firstManifest, secondManifest)
}

func stripLine(content, prefix string) string {
	var kept []string
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, prefix) {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

// The detail belongs in the bundle, where somebody opening the archive later
// can still read it without having kept the HTTP response.
func TestExportWritesTheReportIntoTheBundle(t *testing.T) {
	path, report := runExport(t, specmodel.SecretsModeOmit, "")
	assert.NotEmpty(t, report.Issues, "precondition: this fixture skips things")

	content := readFromArchive(t, path, "report.yaml")
	assert.Contains(t, content, "issues:")
	assert.Contains(t, content, specmodel.CodePreviewAppSkipped)
	assert.Contains(t, content, specmodel.CodeServiceUnavailable)
}

// fakeSSLService records the certificates whose files were written.
type fakeSSLService struct {
	sslservice.Service
	written []string
}

func (f *fakeSSLService) WriteCertFiles(_ bool, settings ...*entity.Setting) error {
	for _, setting := range settings {
		f.written = append(f.written, setting.ID)
	}
	return nil
}

// fakeProjectService records what apply persists, and creates a project with a
// default webhook and notification, as the real one does.
type fakeProjectService struct {
	projectservice.Service
	persisted *projectservice.PersistingProjectData
}

func (f *fakeProjectService) PrepareNewProject(
	_ context.Context, req *projectservice.NewProjectReq, out *projectservice.PersistingProjectData,
) error {
	out.UpsertingProjects = append(out.UpsertingProjects, req.Project)
	webhook := &entity.Setting{
		ID: "webhook_" + req.Project.ID, Type: base.SettingTypeRepoWebhook, Scope: base.ObjectScopeProject,
		ObjectID: req.Project.ID, Name: "default", Status: base.SettingStatusActive, Default: true,
	}
	webhook.MustSetData(&entity.RepoWebhook{Secret: entity.NewEncryptedField("generated")})
	out.UpsertingSettings = append(out.UpsertingSettings, webhook)
	return nil
}

func (f *fakeProjectService) PersistProjectData(
	_ context.Context, _ database.IDB, data *projectservice.PersistingProjectData,
) error {
	f.persisted = data
	return nil
}

func TestExportKeepsProjectAndEnvConfigFiles(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	project := readFromArchive(t, path, "projects/project_a/project.yaml")
	assert.Contains(t, project, "configFiles:")
	assert.Contains(t, project, "level: project")
	assert.Contains(t, project, "inheritable: true", "whether apps get it travels with it")

	env := readFromArchive(t, path, "projects/project_a/envs/dev.yaml")
	assert.Contains(t, env, "dev-only.yaml")
	assert.Contains(t, env, "level: env")
}
