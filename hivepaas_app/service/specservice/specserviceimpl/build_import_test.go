package specserviceimpl

import (
	"context"
	"slices"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const importDocYAML = `
deployment:
  source:
    activeMethod: repo
    repoSource: {repoUrl: "https://github.com/acme/api", repoRef: main}
  container:
    serviceLabels: {team: platform}
    hostname: api
    user: "1000"
    readOnly: true
    healthcheck: {enabled: true, mode: CMD-SHELL, command: curl -f localhost, retries: 3}
    logDriver: {name: json-file, options: {max-size: 10m}}
  resources:
    limits: {cpus: 1, memory: 512mb}
    memory: {shmSize: 64mb}
  networks:
    attachments: [{name: shop_prod, aliases: [api]}]
    hostsFileEntries: [{address: 10.0.0.5, hostnames: [db]}]
  service:
    modeSpec: {mode: replicated, serviceReplicas: 2}
    placement: {constraints: ["node.labels.zone==eu"]}
  storage:
    mounts:
      /data: {type: volume, source: vol-1, volumeOptions: {subpath: data}}
    dockerMounts:
      /etc/localtime: {type: bind, source: /etc/localtime, readOnly: true}
settings:
  secrets:
    DB_PASSWORD: {id: 01JSECRET, key: DB_PASSWORD, value: hunter2}
`

// An exported document, built in import mode and read back with export's own
// mapping, comes back as it went in: builder and export agree about every block.
func TestBuildAppImportsAnExportedDocument(t *testing.T) {
	useDataKey(t)
	volumes := &fakeBuildVolumeService{}
	svc := &service{volumeService: volumes}
	req := buildReq(t, importDocYAML)
	req.Import = true

	resp, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	out := mapSwarmService(&swarm.Service{Spec: *req.Spec}, nil)
	doc := req.Doc.Deployment
	doc.Container.Image = req.Spec.TaskTemplate.ContainerSpec.Image // never written
	assert.Equal(t, doc.Container, out.Container)
	assert.Equal(t, doc.Networks, out.Networks)
	assert.Equal(t, doc.Service, out.Service)
	assert.Equal(t, doc.Resources, out.Resources)

	if assert.Len(t, volumes.req.New, 1) {
		assert.Equal(t, "vol-1", volumes.req.New[0].Source)
		assert.Equal(t, "data", volumes.req.New[0].VolumeOptions.Subpath)
	}
	assert.Equal(t, []mount.Mount{{Type: mount.TypeBind, Source: "/etc/localtime", Target: "/etc/localtime",
		ReadOnly: true}}, volumes.req.Kept)

	secret := slices.IndexFunc(resp.Settings, func(s *entity.Setting) bool { return s.Type == base.SettingTypeSecret })
	assert.GreaterOrEqual(t, secret, 0, "the entry's id is taken out before it is decoded")
}

// A template still cannot say what only an export can.
func TestBuildAppKeepsTheTemplateGate(t *testing.T) {
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	_, err := svc.BuildApp(context.Background(), nil, buildReq(t, "deployment:\n  container: {user: root}\n"))
	assert.Error(t, err)
}

// An export with no deployment - an app never deployed - builds no service block.
func TestBuildAppImportsAnAppNeverDeployed(t *testing.T) {
	useDataKey(t)
	svc := &service{volumeService: &fakeBuildVolumeService{}}
	req := buildReq(t, "settings:\n  routing: {port: 80}\n")
	req.Import = true

	_, err := svc.BuildApp(context.Background(), nil, req)

	assert.NoError(t, err)
	assert.Equal(t, "busybox:latest", req.Spec.TaskTemplate.ContainerSpec.Image)
}
