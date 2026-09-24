package dockerapiserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
)

type fakeSettingRepo struct {
	repository.SettingRepo
	settings []*entity.Setting
}

func (f *fakeSettingRepo) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.settings, nil, nil
}

type fakeAppRepo struct {
	repository.AppRepo
	apps []*entity.App
}

func (f *fakeAppRepo) ListByIDs(_ context.Context, _ database.IDB, _ string, ids []string,
	_ ...bunex.SelectQueryOption) ([]*entity.App, error) {
	var out []*entity.App
	for _, app := range f.apps {
		for _, id := range ids {
			if app.ID == id {
				out = append(out, app)
			}
		}
	}
	return out, nil
}

type fakeNetworkService struct {
	networkservice.Service
}

func (fakeNetworkService) GetProjectNetworkName(project *entity.Project, env string) string {
	return project.Key + "_" + env + "_net"
}

func dockerAPISetting(appID, data string) *entity.Setting {
	return &entity.Setting{ID: "s-" + appID, Scope: base.ObjectScopeApp, ObjectID: appID,
		Type: base.SettingTypeAppDockerAPI, Status: base.SettingStatusActive, Data: data}
}

func appIn(id string) *entity.App {
	return &entity.App{ID: id, ServiceID: "svc-" + id, Project: &entity.Project{Key: "shop"},
		ProjectEnv: &entity.ProjectEnv{Name: "prod"}}
}

func TestPoliciesComeFromTheAppsSettings(t *testing.T) {
	svc := &service{
		settingRepo: &fakeSettingRepo{settings: []*entity.Setting{
			dockerAPISetting("runner", `{"images":["*"],"networks":["env"],`+
				`"allow":["exec","files","volumes","networks","nestedSocket"]}`),
			dockerAPISetting("autobase", `{"images":["autobase/automation:2.11.0"],`+
				`"sharedDirs":["/var/lib/autobase/ansible"],"limits":{"containers":3,"memory":2147483648}}`),
			// An app deleted since: its setting row is still there for a moment.
			dockerAPISetting("gone", `{"images":["*"]}`),
		}},
		appRepo:        &fakeAppRepo{apps: []*entity.App{appIn("runner"), appIn("autobase")}},
		networkService: fakeNetworkService{},
	}

	policies, err := svc.Policies(context.Background(), nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, []*dockerproxy.Policy{
		{
			AppID: "runner", ServiceID: "svc-runner", Images: []string{"*"},
			Network: "hp-dapi-runner", Networks: []string{"shop_prod_net"}, SocketVolume: "hp-dapi-sock-runner",
			Allow: []dockerproxy.Group{dockerproxy.GroupExec, dockerproxy.GroupFiles, dockerproxy.GroupVolumes,
				dockerproxy.GroupNetworks, dockerproxy.GroupNestedSocket},
			Limits: dockerproxy.Limits{Containers: 5, Memory: 1 << 30, NanoCPUs: 1_000_000_000},
		},
		{
			AppID: "autobase", ServiceID: "svc-autobase", Images: []string{"autobase/automation:2.11.0"},
			SharedDirs: []string{"/var/lib/autobase/ansible"},
			Network:    "hp-dapi-autobase", SocketVolume: "hp-dapi-sock-autobase",
			Limits: dockerproxy.Limits{Containers: 3, Memory: 2 << 30, NanoCPUs: 1_000_000_000},
		},
	}, policies)
}
