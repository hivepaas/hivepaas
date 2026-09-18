package specserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

const buildDocYAML = `
deployment:
  source:
    activeMethod: image
    imageSource: {image: "postgres:17.6-alpine3.22"}
    command: postgres -c max_connections=200
    workingDir: /
  storage:
    mounts:
      /var/lib/postgresql/data: {type: volume, source: vol-1, volumeOptions: {subpath: data}}
  container:
    healthcheck: {enabled: true, mode: CMD-SHELL, command: pg_isready -U app, interval: 10s, retries: 5}
  resources:
    reservations: {cpus: 0.5, memory: 256mb}
    limits: {cpus: 1, memory: 512mb, pids: 100}
settings:
  kind:
    category: database
    engine: postgres
    version: "17"
    database: {dbName: app, username: app, password: s3cret}
  envVars:
    data:
      - {k: POSTGRES_PASSWORD, v: "${HIVEPAAS_PASSWORD}"}
  routing: {port: 5432}
`

type fakeBuildVolumeService struct {
	volumeservice.Service
	req *volumeservice.BuildAppMountsReq
}

func (f *fakeBuildVolumeService) BuildAppMounts(
	_ context.Context, _ database.IDB, req *volumeservice.BuildAppMountsReq,
) (*volumeservice.BuildAppMountsResp, error) {
	f.req = req
	mounts := make([]mount.Mount, 0, len(req.New))
	for _, m := range req.New {
		mounts = append(mounts, mount.Mount{Type: m.Type, Source: "docker-" + m.Source, Target: m.Target})
	}
	return &volumeservice.BuildAppMountsResp{Mounts: mounts}, nil
}

func buildDoc(t *testing.T, text string) *specmodel.AppDoc {
	t.Helper()
	doc := &specmodel.AppDoc{}
	assert.NoError(t, yaml.Unmarshal([]byte(text), doc))
	return doc
}

func buildReq(t *testing.T, text string) *specservice.BuildAppReq {
	t.Helper()
	return &specservice.BuildAppReq{
		App: &entity.App{
			ID: "app-1", Key: "db",
			Project:    &entity.Project{ID: "p1", Key: "shop"},
			ProjectEnv: &entity.ProjectEnv{ID: "p1:dev", Key: "dev", Name: "dev"},
		},
		Doc: buildDoc(t, text),
		Spec: &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{Image: "busybox:latest"},
		}},
		TimeNow: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
	}
}

func TestBuildAppBuildsEverySupportedBlock(t *testing.T) {
	useDataKey(t)
	volumes := &fakeBuildVolumeService{}
	svc := &service{volumeService: volumes}
	req := buildReq(t, buildDocYAML)

	resp, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	byType := map[base.SettingType]*entity.Setting{}
	for _, setting := range resp.Settings {
		assert.Equal(t, base.ObjectScopeApp, setting.Scope)
		assert.Equal(t, "app-1", setting.ObjectID)
		assert.Equal(t, base.SettingStatusActive, setting.Status)
		assert.Equal(t, req.TimeNow, setting.CreatedAt)
		byType[setting.Type] = setting
	}
	assert.Len(t, byType, 4)

	deployment := byType[base.SettingTypeAppDeployment].MustAsAppDeploymentSettings()
	assert.Equal(t, base.DeploymentMethodImage, deployment.ActiveMethod)
	assert.Equal(t, "postgres:17.6-alpine3.22", deployment.ImageSource.Image)
	assert.Equal(t, "postgres -c max_connections=200", deployment.Command)

	kind := byType[base.SettingTypeAppKind].MustAsAppKindSettings()
	assert.Equal(t, base.AppCategoryDatabase, kind.Category)
	password, err := kind.Database.Password.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "s3cret", password)

	assert.Equal(t, "${HIVEPAAS_PASSWORD}", byType[base.SettingTypeEnvVar].MustAsEnvVars().Data[0].Value)
	assert.Equal(t, 5432, byType[base.SettingTypeAppRouting].MustAsAppRoutingSettings().Port)

	containerSpec := req.Spec.TaskTemplate.ContainerSpec
	assert.Equal(t, []string{"CMD-SHELL", "pg_isready -U app"}, containerSpec.Healthcheck.Test,
		"a shell command stays one string, as the shell needs it")
	assert.Equal(t, 10*time.Second, containerSpec.Healthcheck.Interval)
	assert.Equal(t, 5, containerSpec.Healthcheck.Retries)

	limits := req.Spec.TaskTemplate.Resources.Limits
	assert.Equal(t, int64(1_000_000_000), limits.NanoCPUs)
	assert.Equal(t, int64(512<<20), limits.MemoryBytes)
	assert.Equal(t, int64(100), limits.Pids)
	assert.Equal(t, int64(256<<20), req.Spec.TaskTemplate.Resources.Reservations.MemoryBytes)

	assert.Equal(t, "docker-vol-1", containerSpec.Mounts[0].Source)
	assert.Equal(t, "/var/lib/postgresql/data", containerSpec.Mounts[0].Target)
	assert.Equal(t, "data", volumes.req.New[0].VolumeOptions.Subpath)
	assert.Equal(t, "busybox:latest", containerSpec.Image, "the image arrives with the deployment")
}

func TestBuildAppAlwaysGivesAVolumeMountOptions(t *testing.T) {
	useDataKey(t)
	volumes := &fakeBuildVolumeService{}
	svc := &service{volumeService: volumes}

	_, err := svc.BuildApp(context.Background(), nil,
		buildReq(t, "deployment:\n  storage:\n    mounts:\n      /data: {type: volume, source: vol-1}\n"))

	assert.NoError(t, err)
	assert.NotNil(t, volumes.req.New[0].VolumeOptions,
		"without options a non-bind volume mount would get no app subpath and share the volume's root")
}

func TestBuildAppSplitsACmdHealthcheck(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t,
		"deployment:\n  container:\n    healthcheck: {enabled: true, mode: CMD, command: \"pg_isready -U app\"}\n")

	_, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	assert.Equal(t, []string{"CMD", "pg_isready", "-U", "app"}, req.Spec.TaskTemplate.ContainerSpec.Healthcheck.Test)
}

// Export maps a service back into an AppDoc; BuildApp maps an AppDoc into a
// service. For the swarm-side blocks the two must agree, or a template and an
// export of the app it created would disagree about the same app.
func TestBuildAppAgreesWithExport(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t, buildDocYAML)

	resp, err := svc.BuildApp(context.Background(), nil, req)
	assert.NoError(t, err)

	assert.Equal(t, req.Doc.Deployment.Container.Healthcheck,
		mapHealthcheck(req.Spec.TaskTemplate.ContainerSpec.Healthcheck))
	assert.Equal(t, req.Doc.Deployment.Resources, mapResources(&req.Spec.TaskTemplate))

	for _, setting := range resp.Settings {
		if setting.Type != base.SettingTypeAppDeployment {
			continue
		}
		body, err := renderSetting(setting, newRefIndex(), specmodel.SecretsModeOmit)
		assert.NoError(t, err)
		for key, value := range req.Doc.Deployment.Source {
			assert.Equal(t, value, body[key], key)
		}
	}
}

func TestBuilderRegistryCoversEveryBuildableBlock(t *testing.T) {
	registered := slices.Sorted(maps.Keys((&service{}).builders()))
	assert.Equal(t, slices.Sorted(slices.Values(specmodel.BuildableBlocks)), registered)
}

func TestBuildAppRefuses(t *testing.T) {
	cases := map[string]struct {
		doc  string
		want error
	}{
		"no image": {
			"deployment:\n  source: {activeMethod: image}\n", hperrors.ErrSpecBlockInvalid,
		},
		"reserved env var": {
			"settings:\n  envVars: {data: [{k: HIVEPAAS_PASSWORD, v: x}]}\n", hperrors.ErrSpecBlockInvalid,
		},
		"duplicate env var": {
			"settings:\n  envVars: {data: [{k: A, v: x}, {k: A, v: y}]}\n", hperrors.ErrSpecBlockInvalid,
		},
		"system env var": {
			"settings:\n  envVars: {data: [{k: A, v: x, system: true}]}\n", hperrors.ErrSpecBlockInvalid,
		},
		"unknown kind category": {
			"settings:\n  kind: {category: queue}\n", hperrors.ErrSpecBlockInvalid,
		},
		"unknown kind field": {
			"settings:\n  kind: {category: cache, flavor: x}\n", hperrors.ErrSpecBlockInvalid,
		},
		"port out of range": {
			"settings:\n  routing: {port: 70000}\n", hperrors.ErrSpecBlockInvalid,
		},
		"unbuildable block": {
			"deployment:\n  networks: {dnsConfig: {nameservers: [1.1.1.1]}}\n", hperrors.ErrSpecBlockUnsupported,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			useDataKey(t)
			svc := &service{volumeService: &fakeBuildVolumeService{}}
			_, err := svc.BuildApp(context.Background(), nil, buildReq(t, tc.doc))
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

const buildSecretsDocYAML = `
deployment:
  source:
    activeMethod: image
    imageSource: {image: "keycloak:26.7.4"}
settings:
  secrets:
    KC_BOOTSTRAP_ADMIN_PASSWORD: {value: hunter2}
    LICENSE:
      value: a-license
      swarmRef: {file: {name: /run/secrets/license, mode: 400}}
  configFiles:
    prometheus.yml:
      content: "scrape_interval: 30s\n"
      swarmRef: {file: {name: /etc/prometheus/prometheus.yml, mode: 444}}
`

func TestBuildAppBuildsSecretsAndConfigFiles(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}

	resp, err := svc.BuildApp(context.Background(), nil, buildReq(t, buildSecretsDocYAML))

	assert.NoError(t, err)
	byName := map[string]*entity.Setting{}
	for _, setting := range resp.Settings {
		byName[string(setting.Type)+"/"+setting.Name] = setting
	}

	admin := byName["secret/KC_BOOTSTRAP_ADMIN_PASSWORD"]
	if admin == nil {
		t.Fatal("a secret is built under the name it is listed by, which is what ${...} refers to")
	}
	secret, err := admin.AsSecret()
	assert.NoError(t, err)
	assert.Equal(t, "KC_BOOTSTRAP_ADMIN_PASSWORD", secret.Key)
	value, err := secret.Value.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", value)
	assert.NotContains(t, admin.Data, "hunter2", "a secret is stored encrypted")
	assert.Nil(t, secret.SwarmRef, "a secret with no file is read through the environment, not mounted")

	license, err := byName["secret/LICENSE"].AsSecret()
	assert.NoError(t, err)
	assert.Equal(t, "/run/secrets/license", license.SwarmRef.File.Name)

	configFile, err := byName["config-file/prometheus.yml"].AsConfigFile()
	assert.NoError(t, err)
	assert.Equal(t, "prometheus.yml", configFile.Name)
	assert.Equal(t, "scrape_interval: 30s\n", configFile.Content)
	assert.Equal(t, "/etc/prometheus/prometheus.yml", configFile.SwarmRef.File.Name)
}

// buildErrorDetail is the explanation a person reads; Error() carries the code.
func buildErrorDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestBuildAppRefusesSecretsAndConfigFilesItCannotApply(t *testing.T) {
	useDataKey(t)
	cases := map[string]struct {
		doc  string
		want string
	}{
		"key disagrees with the name": {
			"settings:\n  secrets:\n    A: {key: B, value: x}\n", "does not match the name",
		},
		"name disagrees with the key": {
			"settings:\n  configFiles:\n    a.conf: {name: b.conf, content: x}\n", "does not match the name",
		},
		"reserved secret key": {
			"settings:\n  secrets:\n    HIVEPAAS_PASSWORD: {value: x}\n", "reserved for HivePaaS",
		},
	}
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.BuildApp(context.Background(), nil, buildReq(t, tc.doc))
			assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
			assert.Contains(t, buildErrorDetail(t, err), tc.want)
		})
	}
}

// routingDocYAML is the routing block a web app template writes: one domain,
// taken from a parameter that a person may leave empty.
const routingDocYAML = `
deployment:
  source: {activeMethod: image, imageSource: {image: "ghost:6.64.0-alpine"}}
settings:
  routing:
    port: 2368
    exposePublicly: true
    domains:
      - {domain: "%s", enabled: true, forceHttps: true}
`

func buildRoutingSettings(t *testing.T, domain string) *entity.AppRoutingSettings {
	t.Helper()
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	resp, err := svc.BuildApp(context.Background(), nil, buildReq(t, fmt.Sprintf(routingDocYAML, domain)))
	assert.NoError(t, err)
	setting := settinghelper.FindSettingByType(resp.Settings, base.SettingTypeAppRouting)
	assert.NotNil(t, setting)
	stored := &entity.Setting{Type: base.SettingTypeAppRouting, Data: setting.Data}
	return stored.MustAsAppRoutingSettings()
}

func TestBuildAppKeepsADomainItWasGiven(t *testing.T) {
	routing := buildRoutingSettings(t, "blog.example.com")

	assert.Equal(t, 2368, routing.Port)
	assert.True(t, routing.ExposePublicly)
	assert.Equal(t, []string{"blog.example.com"}, routing.GetActiveDomainNames())
	assert.Equal(t, base.NetworkProtocolHTTP, routing.Domains[0].Protocol, "http unless the template says otherwise")
	assert.True(t, routing.Domains[0].ForceHttps)
}

// The domain parameter is optional, and an app created without one has to come
// out exactly as an app whose template never mentioned a domain.
func TestBuildAppDropsADomainLeftEmpty(t *testing.T) {
	routing := buildRoutingSettings(t, "")

	assert.Equal(t, 2368, routing.Port)
	assert.Empty(t, routing.Domains)
	assert.False(t, routing.ExposePublicly, "there is no address for being exposed to mean")
}

func TestBuildAppRefusesAWildcardAsAnAddress(t *testing.T) {
	svc := &service{volumeService: &fakeBuildVolumeService{}}

	_, err := svc.BuildApp(context.Background(), nil,
		buildReq(t, fmt.Sprintf(routingDocYAML, "*.example.com")))

	assert.ErrorIs(t, err, hperrors.ErrSpecBlockInvalid)
	assert.Contains(t, buildErrorDetail(t, err), "*.example.com")
}
