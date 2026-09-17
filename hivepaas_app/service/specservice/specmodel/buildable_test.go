package specmodel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// buildableDocYAML uses every block and field phase 1 can build.
const buildableDocYAML = `
deployment:
  source:
    activeMethod: image
    imageSource: {image: "postgres:17.6-alpine3.22"}
    command: postgres -c max_connections=200
    workingDir: /
  storage:
    mounts:
      /var/lib/postgresql/data: {type: volume, source: vol-1, readOnly: false, volumeOptions: {subpath: data}}
  container:
    healthcheck: {enabled: true, mode: CMD-SHELL, command: pg_isready, interval: 10s, retries: 5}
  resources:
    reservations: {cpus: 0.5, memory: 256mb}
    limits: {cpus: 1, memory: 512mb, pids: 100}
settings:
  kind: {category: database, engine: postgres}
  envVars: {data: [{k: A, v: b}]}
  routing: {port: 5432}
`

func decodeDoc(t *testing.T, text string) *AppDoc {
	t.Helper()
	doc := &AppDoc{}
	assert.NoError(t, yaml.Unmarshal([]byte(text), doc))
	return doc
}

func buildableErrorDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestCheckBuildableAcceptsTheSupportedSubset(t *testing.T) {
	assert.NoError(t, CheckBuildable(decodeDoc(t, buildableDocYAML)))
	assert.NoError(t, CheckBuildable(&AppDoc{}))
	assert.NoError(t, CheckBuildable(nil))
}

func TestPresentBlocks(t *testing.T) {
	assert.Equal(t, BuildableBlocks, PresentBlocks(decodeDoc(t, buildableDocYAML)))
	assert.Equal(t, []Block{BlockSettingsKind}, PresentBlocks(decodeDoc(t, "settings:\n  kind: {category: cache}\n")))
	assert.Empty(t, PresentBlocks(&AppDoc{}))
}

func TestCheckBuildableRefusesTheRest(t *testing.T) {
	cases := map[string]struct {
		doc  string
		path string
	}{
		"app name":         {"name: db\n", "name"},
		"repository build": {"deployment:\n  source:\n    activeMethod: repo\n", "deployment.source.activeMethod"},
		"registry auth": {"deployment:\n  source:\n    imageSource: {image: x, registryAuth: {id: r}}\n",
			"deployment.source.imageSource.registryAuth"},
		"unknown source key": {"deployment:\n  source:\n    preDeploymentCommand: x\n",
			"deployment.source.preDeploymentCommand"},
		"container user": {"deployment:\n  container:\n    user: root\n", "deployment.container.user"},
		"memory swap":    {"deployment:\n  resources:\n    memory: {swap: 1gb}\n", "deployment.resources.memory"},
		"generic resources": {"deployment:\n  resources:\n    reservations: {genericResources: [{kind: gpu, value: '1'}]}\n",
			"deployment.resources.reservations.genericResources"},
		"bind mount": {"deployment:\n  storage:\n    mounts:\n      /data: {type: bind, source: /srv}\n",
			"deployment.storage.mounts./data.type"},
		"volume labels": {"deployment:\n  storage:\n    mounts:\n" +
			"      /data: {type: volume, source: v, volumeOptions: {labels: {a: b}}}\n",
			"deployment.storage.mounts./data.volumeOptions.labels"},
		"networks":        {"deployment:\n  networks:\n    dnsConfig: {nameservers: [1.1.1.1]}\n", "deployment.networks"},
		"service mode":    {"deployment:\n  service:\n    modeSpec: {mode: global}\n", "deployment.service"},
		"config files":    {"settings:\n  configFiles: {}\n", "settings.configFiles"},
		"routing domains": {"settings:\n  routing: {port: 80, domains: []}\n", "settings.routing.domains"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := CheckBuildable(decodeDoc(t, tc.doc))
			assert.ErrorIs(t, err, hperrors.ErrSpecBlockUnsupported)
			assert.Contains(t, buildableErrorDetail(t, err), tc.path)
		})
	}
}
