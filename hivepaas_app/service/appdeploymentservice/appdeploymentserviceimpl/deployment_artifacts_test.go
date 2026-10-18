package appdeploymentserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type fakeArtifacts struct {
	artifactservice.Service
	artifacts *entity.AppArtifactsSettings
	mounts    []mount.Mount
	err       error
	specs     []*swarm.ServiceSpec
}

func (f *fakeArtifacts) ArtifactsOf(context.Context, database.IDB, string) (*entity.AppArtifactsSettings, error) {
	return f.artifacts, nil
}

func (f *fakeArtifacts) MountsFor(_ context.Context, _ *entity.AppArtifactsSettings,
	spec *swarm.ServiceSpec) ([]mount.Mount, error) {
	f.specs = append(f.specs, spec)
	return f.mounts, f.err
}

type inspectingDocker struct {
	docker.Manager
	spec swarm.ServiceSpec
}

func (d *inspectingDocker) ServiceInspect(context.Context, string,
	...docker.ServiceInspectOption) (*client.ServiceInspectResult, error) {
	return &client.ServiceInspectResult{Service: swarm.Service{Spec: d.spec}}, nil
}

func artifactsData() *appDeploymentData {
	return &appDeploymentData{App: &entity.App{ID: "app-1", ServiceID: "svc-1"}}
}

// An app with artifacts has them on its nodes, as the service's placement
// says, before the service is updated.
func TestPrepareArtifactsEnsuresThemWhereTheServiceRuns(t *testing.T) {
	walg := mount.Mount{Type: mount.TypeBind, Source: "/c/aa", Target: "/opt/hivepaas/bin/wal-g", ReadOnly: true}
	artifacts := &fakeArtifacts{artifacts: &entity.AppArtifactsSettings{Artifacts: []*entity.AppArtifact{{}}},
		mounts: []mount.Mount{walg}}
	placement := swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		Placement: &swarm.Placement{Constraints: []string{"node.id==n1"}}}}
	s := &service{artifactService: artifacts, dockerManager: &inspectingDocker{spec: placement}}

	mounts, err := s.prepareArtifacts(context.Background(), nil, artifactsData())

	assert.NoError(t, err)
	assert.Equal(t, []mount.Mount{walg}, mounts)
	assert.Equal(t, []string{"node.id==n1"}, artifacts.specs[0].TaskTemplate.Placement.Constraints)
}

// Unlike the Docker API, a failure stops the deployment: the task would not
// start without the file.
func TestPrepareArtifactsFailsTheDeployment(t *testing.T) {
	artifacts := &fakeArtifacts{artifacts: &entity.AppArtifactsSettings{Artifacts: []*entity.AppArtifact{{}}},
		err: hperrors.Wrap(hperrors.ErrArtifactHashMismatch)}
	s := &service{artifactService: artifacts, dockerManager: &inspectingDocker{}}

	_, err := s.prepareArtifacts(context.Background(), nil, artifactsData())

	assert.ErrorIs(t, err, hperrors.ErrArtifactHashMismatch)
}

// An app without artifacts asks nothing, and the service keeps none.
func TestPrepareArtifactsOfAnAppWithNone(t *testing.T) {
	artifacts := &fakeArtifacts{}
	s := &service{artifactService: artifacts, dockerManager: &inspectingDocker{}}

	mounts, err := s.prepareArtifacts(context.Background(), nil, artifactsData())

	assert.NoError(t, err)
	assert.Empty(t, mounts)
	assert.Empty(t, artifacts.specs)
}
