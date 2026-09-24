package appsettingsuc

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// revealGate answers a reveal as told and keeps what it was asked about.
type revealGate struct {
	permission.Manager
	err      error
	subjects []*permission.RevealSubject
}

func (g *revealGate) AuthorizeSecretReveal(
	_ context.Context, _ database.IDB, _ *basedto.Auth, subject *permission.RevealSubject,
) error {
	g.subjects = append(g.subjects, subject)
	return g.err
}

// inspectingClusterService hands back one service and counts the reads.
type inspectingClusterService struct {
	clusterservice.Service
	inspected int
}

func (f *inspectingClusterService) ServiceInspect(_ context.Context, _ string, _ bool) (*swarm.Service, error) {
	f.inspected++
	return &swarm.Service{Spec: swarm.ServiceSpec{
		Annotations:  swarm.Annotations{Labels: map[string]string{"hivepaas.app.id": "app-1", "team": "web"}},
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
	}}, nil
}

func containerSettingsUC(gate *revealGate) (*UC, *inspectingClusterService) {
	cluster := &inspectingClusterService{}
	app := &entity.App{ID: "app-1", Name: "Shop", ServiceID: "svc-1"}
	return &UC{
		appService:        &ownerLookupAppService{apps: map[string]*entity.App{app.ID: app}},
		clusterService:    cluster,
		permissionManager: gate,
	}, cluster
}

func containerSettingsReq(reveal bool) *appsettingsdto.GetAppContainerSettingsReq {
	return &appsettingsdto.GetAppContainerSettingsReq{ProjectID: "p1", AppID: "app-1", RevealSystemLabels: reveal}
}

// System labels can carry what traefik is told, credentials among it, so they
// take what secrets take: the operator's switch, the capability, and a record.
func TestRevealingSystemLabelsPassesTheSecretRevealGate(t *testing.T) {
	gate := &revealGate{}
	uc, _ := containerSettingsUC(gate)

	resp, err := uc.GetAppContainerSettings(context.Background(), &basedto.Auth{}, containerSettingsReq(true))

	assert.NoError(t, err)
	assert.Contains(t, resp.Data.ServiceLabels, "hivepaas.app.id")
	if assert.Len(t, gate.subjects, 1) {
		subject := gate.subjects[0]
		assert.Equal(t, base.ObjectScopeApp, subject.Scope)
		assert.Equal(t, "app-1", subject.ObjectID)
		assert.Equal(t, base.ResourceTypeApp, subject.ResType)
		assert.Equal(t, "Shop", subject.ResName)
		assert.Contains(t, subject.Detail, "systemLabels")
	}
}

// A refused reveal reads nothing of the service.
func TestARefusedRevealOfSystemLabelsReadsNothing(t *testing.T) {
	gate := &revealGate{err: hperrors.Wrap(hperrors.ErrUnauthorized)}
	uc, cluster := containerSettingsUC(gate)

	_, err := uc.GetAppContainerSettings(context.Background(), &basedto.Auth{}, containerSettingsReq(true))

	assert.True(t, errors.Is(err, hperrors.ErrUnauthorized), "%v", err)
	assert.Zero(t, cluster.inspected)
}

// Without the reveal nothing is asked, and nothing is recorded.
func TestContainerSettingsWithoutTheRevealAskNothing(t *testing.T) {
	gate := &revealGate{}
	uc, _ := containerSettingsUC(gate)

	resp, err := uc.GetAppContainerSettings(context.Background(), &basedto.Auth{}, containerSettingsReq(false))

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"team": "web"}, resp.Data.ServiceLabels)
	assert.Empty(t, gate.subjects)
}
