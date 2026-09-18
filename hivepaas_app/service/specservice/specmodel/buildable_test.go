package specmodel

import (
	"errors"
	"fmt"
	"strings"
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
  secrets:
    ADMIN_PASSWORD: {value: hunter2}
    LICENSE: {value: abc, base64: false, swarmRef: {file: {name: /run/secrets/license, uid: "0", gid: "0", mode: 400}}}
  configFiles:
    postgresql.conf:
      content: "max_connections = 200\n"
      swarmRef: {file: {name: /etc/postgresql/postgresql.conf, mode: 444}}
  routing:
    port: 5432
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
	assert.Equal(t, []Block{BlockSettingsSecrets},
		PresentBlocks(decodeDoc(t, "settings:\n  secrets:\n    A: {value: x}\n")))
	assert.Empty(t, PresentBlocks(decodeDoc(t, "settings:\n  secrets: {}\n")), "an empty block builds nothing")
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
		"networks":     {"deployment:\n  networks:\n    dnsConfig: {nameservers: [1.1.1.1]}\n", "deployment.networks"},
		"service mode": {"deployment:\n  service:\n    modeSpec: {mode: global}\n", "deployment.service"},
		"secret swarm id": {"settings:\n  secrets:\n    A: {value: x, swarmRef: {secretId: abc}}\n",
			"settings.secrets.A.swarmRef.secretId"},
		"config swarm id": {"settings:\n  configFiles:\n    a.conf: {content: x, swarmRef: {configId: abc}}\n",
			"settings.configFiles.a.conf.swarmRef.configId"},
		"routing reset": {"settings:\n  routing: {port: 80, reset: true}\n", "settings.routing.reset"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := CheckBuildable(decodeDoc(t, tc.doc))
			assert.ErrorIs(t, err, hperrors.ErrSpecBlockUnsupported)
			assert.Contains(t, buildableErrorDetail(t, err), tc.path)
		})
	}
}

// A template can put a generated password where it is encrypted at rest, and
// mount a file; what it cannot do is name docker objects that do not exist yet.
func TestCheckBuildableRefusesOversizedOrMalformedSecretsAndConfigs(t *testing.T) {
	cases := map[string]struct {
		doc  string
		path string
	}{
		"too many secrets": {
			"settings:\n  secrets:\n" + func() string {
				out := ""
				for i := range MaxSettingsPerBlock + 1 {
					out += fmt.Sprintf("    S%d: {value: x}\n", i)
				}
				return out
			}(),
			"settings.secrets",
		},
		"secret too large": {
			"settings:\n  secrets:\n    A: {value: \"" + strings.Repeat("x", MaxSecretValueBytes+1) + "\"}\n",
			"settings.secrets.A.value",
		},
		"config too large": {
			"settings:\n  configFiles:\n    a.conf: {content: \"" +
				strings.Repeat("x", MaxConfigFileBytes+1) + "\"}\n",
			"settings.configFiles.a.conf.content",
		},
		"relative mount path": {
			"settings:\n  configFiles:\n    a.conf: {content: x, swarmRef: {file: {name: etc/a.conf}}}\n",
			"settings.configFiles.a.conf.swarmRef.file.name",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := CheckBuildable(decodeDoc(t, tc.doc))
			assert.Error(t, err)
			assert.Contains(t, buildableErrorDetail(t, err), tc.path)
		})
	}
}

func TestCheckBuildableAcceptsRoutingDomains(t *testing.T) {
	doc := decodeDoc(t, buildableDocYAML+`    domains:
      - {domain: app.example.com, enabled: true, protocol: http, forceHttps: true}
    exposePublicly: true
`)
	assert.NoError(t, CheckBuildable(doc))
}

func TestCheckBuildableRefusesWhatADomainMustNotCarry(t *testing.T) {
	cases := map[string]string{
		"a certificate chosen by hand": "    domains:\n      - {domain: app.example.com, sslCert: {id: s1}}\n",
		"basic authentication":         "    domains:\n      - {domain: app.example.com, basicAuth: {enabled: true}}\n",
		"a path rewrite":               "    domains:\n      - {domain: app.example.com, pathRewriteConfig: {}}\n",
		"a redirect": "    domains:\n" +
			"      - {domain: app.example.com, domainRedirect: other.example.com}\n",
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			err := CheckBuildable(decodeDoc(t, buildableDocYAML+extra))
			assert.ErrorIs(t, err, hperrors.ErrSpecBlockUnsupported)
		})
	}
}

func TestCheckBuildableCapsRoutingDomains(t *testing.T) {
	entries := "    domains:\n"
	for i := range MaxRoutingDomains + 1 {
		entries += fmt.Sprintf("      - {domain: app%d.example.com}\n", i)
	}

	err := CheckBuildable(decodeDoc(t, buildableDocYAML+entries))

	assert.ErrorIs(t, err, hperrors.ErrSpecBlockUnsupported)
}
