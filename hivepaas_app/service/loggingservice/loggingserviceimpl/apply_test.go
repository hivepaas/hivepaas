package loggingserviceimpl

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

func apply(t *testing.T, s *service, cfg *entity.LoggingSettings,
	req *loggingservice.SettingApplyReq) (*loggingservice.SettingApplyResp, error) {
	t.Helper()
	if req == nil {
		req = &loggingservice.SettingApplyReq{}
	}
	req.Setting = storedSetting(t, cfg)
	return s.Apply(context.Background(), nil, req)
}

// provisionedCommand is the command the app with the key was provisioned with.
func provisionedCommand(t *testing.T, s *service, key string) string {
	t.Helper()
	for _, req := range appsOf(s).provisioned {
		if req.Key == key {
			return req.Doc.Deployment.Source["command"].(string)
		}
	}
	t.Fatalf("%s was not provisioned", key)
	return ""
}

func argsOf(t *testing.T, command string) []string {
	t.Helper()
	args, err := executil.CmdSplit(command)
	if err != nil {
		t.Fatalf("CmdSplit: %v", err)
	}
	return args
}

func TestApplyWithNoSettingDoesNothing(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	resp, err := s.Apply(context.Background(), nil, &loggingservice.SettingApplyReq{})

	assert.NoError(t, err)
	assert.Empty(t, resp.Tasks)
	assert.Empty(t, appsOf(s).provisioned)
}

// The backend first: the collector needs somewhere to write. Both are apps of
// their own environment, whose first deployments come back to be scheduled once
// the transaction commits.
func TestApplyProvisionsTheBackendBeforeTheCollector(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	resp, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	apps := appsOf(s)
	if assert.Len(t, apps.provisioned, 2) {
		assert.Equal(t, backendAppKey, apps.provisioned[0].Key)
		assert.Equal(t, collectorAppKey, apps.provisioned[1].Key)
		assert.Equal(t, loggingEnv, apps.provisioned[0].Env)
		assert.Equal(t, loggingEnv, apps.provisioned[1].Env)
	}
	assert.Len(t, resp.Tasks, 2)
	assert.NotNil(t, resp.Cleanup, "what provisioning made in docker has to be undoable")
}

// The ids are written back for the dashboard to link to.
func TestApplyRemembersTheApps(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	upserted := settingsOf(s).upserted
	if assert.Len(t, upserted, 1) {
		cfg, err := upserted[0].AsLoggingSettings()
		assert.NoError(t, err)
		assert.Equal(t, "app-"+backendAppKey, cfg.BackendAppID)
		assert.Equal(t, "app-"+collectorAppKey, cfg.CollectorAppID)
	}
}

// The data volume is named by its setting's id, which the app's mount build
// turns into the volume docker knows - rather than handing docker the id, which
// it would take for the name of a volume to create.
func TestApplyMountsTheChosenVolumeThroughTheApp(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	doc := appsOf(s).provisioned[0].Doc
	if assert.NotNil(t, doc.Deployment.Storage) {
		mnt, ok := doc.Deployment.Storage.Mounts[victorialogs.DataPath]
		assert.True(t, ok, "the logs have to be on the volume")
		assert.Equal(t, mount.TypeVolume, mnt.Type)
		assert.Equal(t, "vol-1", mnt.Source)
	}
	backend := appsOf(s).docker.inspected["svc-"+backendAppKey]
	assert.Empty(t, backend.Spec.TaskTemplate.ContainerSpec.Mounts,
		"nothing mounts the volume behind the app's back")
}

// VictoriaLogs has no authentication, so the API reaches it over the stack's
// internal network and nothing reaches it over the routing one.
func TestApplyLetsOnlyTheAPIQueryTheBackend(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	spec := appsOf(s).docker.inspected["svc-"+backendAppKey].Spec
	assert.Contains(t, spec.TaskTemplate.Networks,
		swarm.NetworkAttachmentConfig{Target: base.NetworkHivepaasLocal, Aliases: []string{backendAlias}})
	for _, attachment := range spec.TaskTemplate.Networks {
		assert.NotEqual(t, base.NetworkGlobalRouting, attachment.Target)
	}
	assert.Equal(t, logDriverLocal, spec.TaskTemplate.LogDriver.Name, "the stack hides its own logs")
}

// One collector per node, reading the node's container logs and nothing more.
func TestApplyRunsTheCollectorOnEveryNode(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	spec := appsOf(s).docker.inspected["svc-"+collectorAppKey].Spec
	assert.NotNil(t, spec.Mode.Global)
	assert.Nil(t, spec.Mode.Replicated, "setting both modes is rejected by the daemon")
	assert.Equal(t, []mount.Mount{{
		Type: mount.TypeBind, Source: vlagent.ContainersPath, Target: vlagent.ContainersPath, ReadOnly: true,
	}}, spec.TaskTemplate.ContainerSpec.Mounts)
	for _, attachment := range spec.TaskTemplate.Networks {
		assert.NotEqual(t, base.NetworkHivepaasLocal, attachment.Target,
			"the collector runs on every node and must not reach the database")
	}
	assert.Equal(t, logDriverLocal, spec.TaskTemplate.LogDriver.Name)
}

func TestApplyPointsTheCollectorAtTheBackend(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	command := provisionedCommand(t, s, collectorAppKey)
	assert.Contains(t, argsOf(t, command),
		"-remoteWrite.url=http://"+backendAlias+":9428"+victorialogs.IngestPath)
}

// Both apps run capped, which is what lets them be protected from the OOM killer.
func TestApplyGivesNewAppsTheirLimits(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, enabledConfig(), &loggingservice.SettingApplyReq{
		BackendResources: &logging.Resources{CPULimit: 1, MemoryLimit: (2 * unit.GB).Bytes()},
	})

	assert.NoError(t, err)
	backend := appsOf(s).provisioned[0].Doc.Deployment.Resources
	assert.Equal(t, 2*unit.GB, backend.Limits.Memory)
	assert.InDelta(t, 1.0, backend.Limits.CPUs, 0)
	assert.Equal(t, int64(base.OomScoreAdjSystemAddon), backend.Capabilities.OomScoreAdj)

	collector := appsOf(s).provisioned[1].Doc.Deployment.Resources
	assert.Equal(t, logging.DefaultCollectorMemoryLimit, collector.Limits.Memory)
	assert.Equal(t, int64(base.OomScoreAdjSystemAddon), collector.Capabilities.OomScoreAdj)
}

func TestApplyGivesANewBackendTheDefaultMemoryLimit(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	assert.Equal(t, logging.DefaultBackendMemoryLimit, appsOf(s).provisioned[0].Doc.Deployment.Resources.Limits.Memory)
}

// A credential on the command line is readable by anyone who can read the app's
// deployment settings, so it goes into a secret file instead.
func TestApplyKeepsForwardCredentialsInSecrets(t *testing.T) {
	cfg := enabledConfig()
	cfg.Forwards = []entity.LoggingForward{{
		Name: "company", Format: vlagent.FormatJSONLine,
		Endpoint: entity.LoggingEndpoint{URL: "http://logs.example.com/ingest",
			BearerToken: entity.NewEncryptedField("TOKEN")},
	}}
	s := newTestService(&fakeDocker{}, nil)

	_, err := apply(t, s, cfg, nil)

	assert.NoError(t, err)
	command := provisionedCommand(t, s, collectorAppKey)
	assert.NotContains(t, command, "TOKEN")
	secrets, _ := appsOf(s).provisioned[1].Doc.Settings["secrets"].(map[string]any)
	token, _ := secrets["REMOTE_WRITE_1_BEARER_TOKEN"].(map[string]any)
	assert.Equal(t, "TOKEN", token["value"])
}

// Saving the same configuration again changes nothing and deploys nothing.
func TestApplyTwiceDeploysNothingNew(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	resp, err := apply(t, s, enabledConfig(), nil)

	assert.NoError(t, err)
	assert.Len(t, appsOf(s).provisioned, 2, "provisioned once")
	assert.Empty(t, appsOf(s).redeployed)
	assert.Empty(t, resp.Tasks)
}

// Retention is an argument of the backend, and only the backend.
func TestApplyRedeploysTheBackendWhenRetentionChanges(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	cfg := enabledConfig()
	cfg.Backend.VictoriaLogs.Retention = timeutil.Duration(14 * 24 * time.Hour)
	resp, err := apply(t, s, cfg, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{backendAppKey}, appsOf(s).redeployed)
	if assert.Len(t, resp.Tasks, 1) {
		assert.Equal(t, "redeploy-"+backendAppKey, resp.Tasks[0].ID)
	}
}

// A new forward changes the collector's arguments and its credential files.
func TestApplyReconcilesTheCollector(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	cfg := enabledConfig()
	cfg.Forwards = []entity.LoggingForward{{
		Name: "company", Format: vlagent.FormatJSONLine,
		Endpoint: entity.LoggingEndpoint{URL: "http://logs.example.com/ingest",
			Password: entity.NewEncryptedField("PASSWORD"), Username: "shipper"},
	}}
	_, err = apply(t, s, cfg, nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{collectorAppKey}, appsOf(s).redeployed)
	files := appsOf(s).secrets[collectorAppKey]
	if assert.Len(t, files, 1) {
		assert.Equal(t, "PASSWORD", files[0].Value)
	}
}

// A running backend's limits are its service's: a save writes them there.
func TestApplyWritesLimitsOntoARunningBackend(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	_, err = apply(t, s, enabledConfig(), &loggingservice.SettingApplyReq{
		BackendResources: &logging.Resources{CPULimit: 0.5, MemoryLimit: (3 * unit.GB).Bytes()},
	})

	assert.NoError(t, err)
	spec := appsOf(s).docker.inspected["svc-"+backendAppKey].Spec
	assert.Equal(t, (3 * unit.GB).Bytes(), spec.TaskTemplate.Resources.Limits.MemoryBytes)
	assert.Equal(t, int64(500_000_000), spec.TaskTemplate.Resources.Limits.NanoCPUs)
	assert.Equal(t, int64(base.OomScoreAdjSystemAddon), spec.TaskTemplate.ContainerSpec.OomScoreAdj)
}

// Switching off takes apps down, which a save must not do unasked.
func TestApplyRefusesToTakeAppsDownUnasked(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	cfg := enabledConfig()
	cfg.Enabled = false
	_, err = apply(t, s, cfg, nil)

	assert.ErrorIs(t, err, hperrors.ErrLoggingAppStillRunning)
	assert.Empty(t, appsOf(s).removed)
}

// The collector goes first, so that nothing ships into a store that is going
// away - and only the backend's storage is the stored logs.
func TestApplyRemovesTheCollectorBeforeTheBackend(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	cfg := enabledConfig()
	cfg.Enabled = false
	// What the usecase carries over from the stored setting.
	cfg.BackendAppID, cfg.CollectorAppID = "app-"+backendAppKey, "app-"+collectorAppKey
	resp, err := apply(t, s, cfg, &loggingservice.SettingApplyReq{RemoveApp: true, RemoveStorage: true})

	assert.NoError(t, err)
	assert.True(t, resp.RemovedApps)
	assert.Equal(t, []string{collectorAppKey, backendAppKey}, appsOf(s).removed)
	assert.Equal(t, map[string]bool{collectorAppKey: false, backendAppKey: true}, appsOf(s).removedStorage)

	last := settingsOf(s).upserted[len(settingsOf(s).upserted)-1].MustAsLoggingSettings()
	assert.Empty(t, last.BackendAppID)
	assert.Empty(t, last.CollectorAppID)
}

// Handing the backend to somebody else's store removes HivePaaS's, and the
// collector then writes there - with the store's credentials, in a secret.
func TestApplyHandsTheBackendToAnExternalStore(t *testing.T) {
	s := newTestService(&fakeDocker{}, nil)
	_, err := apply(t, s, enabledConfig(), nil)
	assert.NoError(t, err)

	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.Ingest = &entity.LoggingEndpoint{
		URL: "http://theirs:9428/internal/insert", BearerToken: entity.NewEncryptedField("THEIRS"),
	}
	_, err = apply(t, s, cfg, &loggingservice.SettingApplyReq{RemoveApp: true})

	assert.NoError(t, err)
	assert.Equal(t, []string{backendAppKey}, appsOf(s).removed)
	command := appsOf(s).apps[collectorAppKey].GetSettingByType(base.SettingTypeAppDeployment).
		MustAsAppDeploymentSettings().Command
	args := argsOf(t, command)
	assert.Contains(t, args, "-remoteWrite.url=http://theirs:9428/internal/insert")
	assert.True(t, slices.ContainsFunc(args, func(arg string) bool {
		return strings.HasPrefix(arg, "-remoteWrite.bearerTokenFile="+vlagent.SecretsDir)
	}), "the external store's credential reaches the collector")
	assert.NotContains(t, command, "THEIRS")
}

// A backend nothing can query is refused rather than deployed.
func TestApplyRefusesWhenTheInternalNetworkIsMissing(t *testing.T) {
	fd := &fakeDocker{}
	s := newTestService(fd, nil)
	fd.networks = nil

	_, err := apply(t, s, enabledConfig(), nil)

	assert.ErrorIs(t, err, hperrors.ErrLoggingAPINetworkMissing)
	assert.Empty(t, appsOf(s).provisioned)
}
