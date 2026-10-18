package artifactserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/services/docker"
)

// artifactSettings lists the app-artifacts settings it has.
type artifactSettings struct {
	repository.SettingRepo
	settings []*entity.Setting
}

func (f *artifactSettings) List(context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.settings, nil, nil
}

type listedApps struct {
	repository.AppRepo
	apps []*entity.App
}

func (f *listedApps) ListByIDs(context.Context, database.IDB, string, []string,
	...bunex.SelectQueryOption) ([]*entity.App, error) {
	return f.apps, nil
}

// placedDocker has a service per id, placed by its constraints; others are gone.
type placedDocker struct {
	*fakeDocker
	constraints map[string][]string
}

func (f *placedDocker) ServiceInspect(_ context.Context, serviceID string,
	_ ...docker.ServiceInspectOption) (*client.ServiceInspectResult, error) {
	constraints, found := f.constraints[serviceID]
	if !found {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return &client.ServiceInspectResult{Service: swarm.Service{Spec: *specWith(constraints...)}}, nil
}

func artifactsOf(appID string) *entity.Setting {
	setting := &entity.Setting{ID: "artifacts-" + appID, Type: base.SettingTypeAppArtifacts, ObjectID: appID,
		Status: base.SettingStatusActive}
	setting.MustSetData(walgSettings())
	return setting
}

// Each app's artifacts go where its service may start a task now: a pinned one's
// node, every like node for one on shared storage. An app whose nodes would need
// two builds fails alone; one whose service is gone, or that has none, is not
// had.
func TestReconcileAllHasEachAppsArtifactsWhereItMayRun(t *testing.T) {
	svc, agent := newTestService("")
	svc.dockerManager = &placedDocker{fakeDocker: svc.dockerManager.(*fakeDocker), constraints: map[string][]string{
		"svc-pinned": {"node.id==n2"},
		"svc-shared": {"node.platform.arch==x86_64"},
		"svc-mixed":  nil,
	}}
	svc.settingRepo = &artifactSettings{settings: []*entity.Setting{artifactsOf("pinned"), artifactsOf("shared"),
		artifactsOf("mixed"), artifactsOf("gone"), artifactsOf("none"), artifactsOf("deleted")}}
	svc.appRepo = &listedApps{apps: []*entity.App{
		{ID: "pinned", Key: "db", ServiceID: "svc-pinned"}, {ID: "shared", Key: "db2", ServiceID: "svc-shared"},
		{ID: "mixed", Key: "db3", ServiceID: "svc-mixed"}, {ID: "gone", Key: "db4", ServiceID: "svc-gone"},
		{ID: "none", Key: "db5"},
	}}

	resp, err := svc.ReconcileAll(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Apps)
	assert.Equal(t, []string{"n2:10001", "n1:10001"}, agent.addrs)
	if assert.Len(t, resp.Failed, 1) {
		assert.Equal(t, "db3", resp.Failed[0].App.Key)
		assert.ErrorIs(t, resp.Failed[0].Err, hperrors.ErrArtifactArchUnsupported)
	}
}
