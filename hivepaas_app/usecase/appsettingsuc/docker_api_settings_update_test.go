package appsettingsuc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

// recordingDockerAPIService applies access the way the real one does, from
// whether the app has it, and records every call in order.
type recordingDockerAPIService struct {
	dockerapiservice.Service
	access bool
	host   bool
	calls  []string
}

func (f *recordingDockerAPIService) SyncAgents(context.Context) error {
	f.calls = append(f.calls, "sync")
	return nil
}

func (f *recordingDockerAPIService) ApplyToService(
	_ context.Context, _ database.IDB, appID string, spec *swarm.ServiceSpec,
) error {
	f.calls = append(f.calls, "apply")
	if f.host {
		dockerapiservice.Detach(spec, "net-"+appID)
		dockerapiservice.AttachHost(spec)
		return nil
	}
	if f.access {
		dockerapiservice.Attach(spec, appID, "net-"+appID)
	} else {
		dockerapiservice.Detach(spec, "net-"+appID)
	}
	return nil
}

func (f *recordingDockerAPIService) RemoveAppObjects(_ context.Context, appID string) error {
	f.calls = append(f.calls, "remove "+appID)
	return nil
}

type recordingEnvVarService struct {
	envvarservice.Service
	calls *[]string
}

func (f *recordingEnvVarService) BuildEnvVarsForAllAppsInScope(
	_ context.Context, _ database.IDB, scope *entity.ObjectScope, _ bool, _ []string, _, _ bool,
) ([]*envvarservice.AppEnvVarData, error) {
	return []*envvarservice.AppEnvVarData{{App: scope.App}}, nil
}

func (f *recordingEnvVarService) ApplyEnvVarsForApps(
	_ context.Context, _ database.IDB, data []*envvarservice.AppEnvVarData, _, _ bool,
) map[int]error {
	*f.calls = append(*f.calls, "env "+data[0].App.ID)
	return nil
}

func dockerAPIAccess() *entity.AppDockerAPISettings {
	return &entity.AppDockerAPISettings{Images: []string{"autobase/automation"}, SharedDirs: []string{"/data/ansible"}}
}

func dockerAPIService() *swarm.Service {
	return &swarm.Service{ID: "svc-1", Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Mounts: []mount.Mount{
			{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/data"},
			{Type: mount.TypeBind, Source: "/srv/cache", Target: "/cache"},
			dockerapiservice.SocketMount("app-1"),
		}},
	}}}
}

// Access where there was none, or more of it, takes what giving it from a
// template takes; less of it takes only the app's own Write, which the route
// already asked for.
func TestDockerAPIGrantAsksForTheClusterOnlyWhenItWidens(t *testing.T) {
	narrower := dockerAPIAccess()
	narrower.SharedDirs = nil
	for name, tc := range map[string]struct {
		prev, next *entity.AppDockerAPISettings
		allow      bool
		asked      bool
		refused    bool
	}{
		"granted by an operator who may": {next: dockerAPIAccess(), allow: true, asked: true},
		"granted by one who may not":     {next: dockerAPIAccess(), asked: true, refused: true},
		"narrowed":                       {prev: dockerAPIAccess(), next: narrower},
		"turned off":                     {prev: dockerAPIAccess()},
	} {
		perms := &answeringPermissionManager{allow: tc.allow}
		uc := &UC{permissionManager: perms}

		err := uc.checkDockerAPIGrant(context.Background(), nil, &basedto.Auth{}, tc.prev, tc.next)

		assert.Equal(t, tc.refused, errors.Is(err, hperrors.ErrUnauthorized), name)
		if !tc.asked {
			assert.Empty(t, perms.checked, name)
			continue
		}
		if assert.Len(t, perms.checked, 1, name) {
			check, ok := perms.checked[0].(*permission.ModuleAccessCheck)
			if assert.True(t, ok, name) {
				assert.Equal(t, base.ResourceModuleCluster, check.Module, name)
				assert.Equal(t, base.ActionTypeWrite, check.Action, name)
			}
		}
	}
}

// A child is given a shared directory through the volume the app mounts it
// from; a directory on a bind, or on the socket, has none.
func TestDockerAPIProblemWantsSharedDirsOnTheAppsVolumes(t *testing.T) {
	for dir, wantProblem := range map[string]bool{
		"/data/ansible":     false,
		"/data":             false,
		"/cache/jobs":       true,
		"/var/run/hivepaas": true,
	} {
		access := dockerAPIAccess()
		access.SharedDirs = []string{dir}

		problem := dockerAPIProblem(access, dockerAPIService())

		assert.Equal(t, wantProblem, problem != "", "%s: %s", dir, problem)
	}
	assert.Contains(t, dockerAPIProblem(&entity.AppDockerAPISettings{}, dockerAPIService()), "images")
}

// Turning access off keeps what it allowed, for turning it on again.
func TestNextDockerAPISettingKeepsWhatAccessTurnedOffAllowed(t *testing.T) {
	app := &entity.App{ID: "app-1"}
	now := time.Now()
	row := nextDockerAPISetting(app, nil, dockerAPIAccess(), now)
	assert.Equal(t, base.SettingStatusActive, row.Status)
	assert.Equal(t, 1, row.UpdateVer)

	off := nextDockerAPISetting(app, row, nil, now)

	assert.Equal(t, base.SettingStatusDisabled, off.Status)
	assert.Equal(t, 2, off.UpdateVer)
	assert.Equal(t, dockerAPIAccess(), off.MustAsAppDockerAPISettings())
}

// The agents serve the app before its new task looks for the socket, and the
// app's HIVEPAAS_DOCKER_HOST appears once the service has the socket.
func TestApplyingGrantedAccessSyncsTheAgentsThenGivesTheServiceItsSocket(t *testing.T) {
	dockerAPI := &recordingDockerAPIService{access: true}
	dockerManager := &serviceUpdatingDockerManager{}
	uc := &UC{dockerAPIService: dockerAPI, dockerManager: dockerManager,
		envVarService: &recordingEnvVarService{calls: &dockerAPI.calls}}
	service := dockerAPIService()
	service.Spec.TaskTemplate.ContainerSpec.Mounts = service.Spec.TaskTemplate.ContainerSpec.Mounts[:2]

	err := uc.applyAppDockerAPI(context.Background(), &updateAppDockerAPIData{
		App: &entity.App{ID: "app-1"}, Service: service, Setting: &entity.Setting{}, Next: dockerAPIAccess(),
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"sync", "apply", "env app-1"}, dockerAPI.calls)
	if assert.NotNil(t, dockerManager.updatedSpec) {
		assert.Contains(t, dockerManager.updatedSpec.TaskTemplate.ContainerSpec.Mounts,
			dockerapiservice.SocketMount("app-1"))
	}
}

// Taking access away takes the socket off the service and removes what the
// app's children left, at once: a child must not outlive the access it came
// from.
func TestApplyingRevokedAccessDetachesTheServiceAndRemovesTheChildren(t *testing.T) {
	dockerAPI := &recordingDockerAPIService{}
	dockerManager := &serviceUpdatingDockerManager{}
	uc := &UC{dockerAPIService: dockerAPI, dockerManager: dockerManager,
		envVarService: &recordingEnvVarService{calls: &dockerAPI.calls}}

	err := uc.applyAppDockerAPI(context.Background(), &updateAppDockerAPIData{
		App: &entity.App{ID: "app-1"}, Service: dockerAPIService(), Setting: &entity.Setting{},
		Prev: dockerAPIAccess(),
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"apply", "env app-1", "remove app-1"}, dockerAPI.calls)
	if assert.NotNil(t, dockerManager.updatedSpec) {
		assert.NotContains(t, dockerManager.updatedSpec.TaskTemplate.ContainerSpec.Mounts,
			dockerapiservice.SocketMount("app-1"))
	}
}

// A changed policy is the agents' alone: the service already has its socket.
func TestApplyingAChangedPolicyOnlySyncsTheAgents(t *testing.T) {
	dockerAPI := &recordingDockerAPIService{access: true}
	dockerManager := &serviceUpdatingDockerManager{}
	uc := &UC{dockerAPIService: dockerAPI, dockerManager: dockerManager}

	err := uc.applyAppDockerAPI(context.Background(), &updateAppDockerAPIData{
		App: &entity.App{ID: "app-1"}, Service: dockerAPIService(), Setting: &entity.Setting{},
		Prev: dockerAPIAccess(), Next: dockerAPIAccess(),
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"sync"}, dockerAPI.calls)
	assert.Nil(t, dockerManager.updatedSpec)
}

func adminAuth(admin bool) *basedto.Auth {
	role := base.UserRoleMember
	if admin {
		role = base.UserRoleAdmin
	}
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "usr-1", Role: role}}}
}

func privilegedAppsSwitch(t *testing.T, on bool) {
	t.Helper()
	prev := config.Current()
	config.SetCurrent(&config.Config{Security: config.Security{AllowPrivilegedApps: on}})
	t.Cleanup(func() { config.SetCurrent(prev) })
}

// The switch comes first: an administrator can turn it on, a member cannot.
func TestHostModeIsBlockedByTheSwitchThenByNotBeingAnAdministrator(t *testing.T) {
	for name, tc := range map[string]struct {
		on, admin bool
		want      string
	}{
		"off, for an administrator": {admin: true, want: hostModeBlockedBySwitch},
		"off, for a member":         {want: hostModeBlockedBySwitch},
		"on, for a member":          {on: true, want: hostModeBlockedByAdmin},
		"on, for an administrator":  {on: true, admin: true},
	} {
		cfg := &config.Config{Security: config.Security{AllowPrivilegedApps: tc.on}}
		assert.Equal(t, tc.want, hostModeBlockedBy(cfg, adminAuth(tc.admin)), name)
	}
}

// The node's own socket takes the switch and an administrator, and Write on the
// Cluster module decides nothing about it; leaving host mode takes nothing.
func TestHostModeIsGrantedOnlyByAnAdministratorWithTheSwitchOn(t *testing.T) {
	host := &entity.AppDockerAPISettings{Mode: entity.DockerAPIModeHost}
	for name, tc := range map[string]struct {
		on, admin  bool
		prev, next *entity.AppDockerAPISettings
		refused    bool
	}{
		"entered with the switch off":   {admin: true, prev: dockerAPIAccess(), next: host, refused: true},
		"entered by a member":           {on: true, next: host, refused: true},
		"entered by an administrator":   {on: true, admin: true, prev: dockerAPIAccess(), next: host},
		"left for the proxy by anybody": {prev: host, next: dockerAPIAccess()},
		"turned off by anybody":         {prev: host},
	} {
		privilegedAppsSwitch(t, tc.on)
		perms := &answeringPermissionManager{allow: true}
		uc := &UC{permissionManager: perms}

		err := uc.checkDockerAPIGrant(context.Background(), nil, adminAuth(tc.admin), tc.prev, tc.next)

		assert.Equal(t, tc.refused, errors.Is(err, hperrors.ErrUnauthorized), name)
		assert.Empty(t, perms.checked, name)
	}
}

// Host mode keeps the proxy's policy without using it, so nothing of it is held
// against the service.
func TestDockerAPIProblemInHostModeLooksAtNoPolicy(t *testing.T) {
	access := dockerAPIAccess()
	access.Mode = entity.DockerAPIModeHost
	access.SharedDirs = []string{"/nowhere"}

	assert.Empty(t, dockerAPIProblem(access, dockerAPIService()))
}

// Moving from the proxy to the node's socket changes the service and the
// environment, after the agents have stopped serving the app.
func TestApplyingHostModeGivesTheServiceTheNodesSocket(t *testing.T) {
	dockerAPI := &recordingDockerAPIService{host: true}
	dockerManager := &serviceUpdatingDockerManager{}
	uc := &UC{dockerAPIService: dockerAPI, dockerManager: dockerManager,
		envVarService: &recordingEnvVarService{calls: &dockerAPI.calls}}

	err := uc.applyAppDockerAPI(context.Background(), &updateAppDockerAPIData{
		App: &entity.App{ID: "app-1"}, Service: dockerAPIService(), Setting: &entity.Setting{},
		Prev: dockerAPIAccess(), Next: &entity.AppDockerAPISettings{Mode: entity.DockerAPIModeHost},
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"sync", "apply", "env app-1"}, dockerAPI.calls)
	if assert.NotNil(t, dockerManager.updatedSpec) {
		mounts := dockerManager.updatedSpec.TaskTemplate.ContainerSpec.Mounts
		assert.Contains(t, mounts, dockerapiservice.HostSocketMount())
		assert.NotContains(t, mounts, dockerapiservice.SocketMount("app-1"))
	}
}

func TestApplyingHostModeUnchangedOnlySyncsTheAgents(t *testing.T) {
	dockerAPI := &recordingDockerAPIService{host: true}
	dockerManager := &serviceUpdatingDockerManager{}
	uc := &UC{dockerAPIService: dockerAPI, dockerManager: dockerManager}
	host := &entity.AppDockerAPISettings{Mode: entity.DockerAPIModeHost}

	err := uc.applyAppDockerAPI(context.Background(), &updateAppDockerAPIData{
		App: &entity.App{ID: "app-1"}, Service: dockerAPIService(), Setting: &entity.Setting{},
		Prev: host, Next: host,
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"sync"}, dockerAPI.calls)
	assert.Nil(t, dockerManager.updatedSpec)
}
