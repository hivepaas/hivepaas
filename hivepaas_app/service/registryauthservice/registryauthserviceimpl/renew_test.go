package registryauthserviceimpl

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/pkg/authconfig"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	ecrHost   = "123456789012.dkr.ecr.eu-west-1.amazonaws.com"
	otherHost = "210987654321.dkr.ecr.us-east-1.amazonaws.com"
)

// settingStore holds the credentials and the apps' deployment settings: List
// answers whichever the query asks for, GetByID and Update a credential's row.
type settingStore struct {
	repository.SettingRepo
	auths       []*entity.Setting
	deployments []*entity.Setting
	keys        map[string]*entity.Setting
}

func (f *settingStore) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	var rows []*entity.Setting
	sql := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&rows), opts...).String()
	if strings.Contains(sql, "'app-deployment'") {
		return f.deployments, nil, nil
	}
	for _, setting := range f.auths {
		if strings.Contains(sql, "setting.id IN") && !strings.Contains(sql, "'"+setting.ID+"'") {
			continue
		}
		rows = append(rows, setting)
	}
	return rows, nil, nil
}

func (f *settingStore) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, typ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if typ == base.SettingTypeKeyAuth {
		return keyAuthIn(f.keys, id)
	}
	for _, setting := range f.auths {
		if setting.ID == id {
			copied := *setting
			return &copied, nil
		}
	}
	return nil, hperrors.NewNotFound("Setting")
}

func (f *settingStore) Update(_ context.Context, _ database.IDB, setting *entity.Setting,
	_ ...bunex.UpdateQueryOption) error {
	for i, row := range f.auths {
		if row.ID == setting.ID {
			f.auths[i] = setting
		}
	}
	return nil
}

type appStore struct {
	repository.AppRepo
	apps []*entity.App
}

func (f *appStore) ListByIDs(_ context.Context, _ database.IDB, _ string, ids []string,
	_ ...bunex.SelectQueryOption) ([]*entity.App, error) {
	var res []*entity.App
	for _, app := range f.apps {
		for _, id := range ids {
			if app.ID == id {
				res = append(res, app)
			}
		}
	}
	return res, nil
}

// fakeSwarm has services by id, and records each update: the spec sent, and the
// credential sent with it.
type fakeSwarm struct {
	docker.Manager
	services map[string]*swarm.Service
	failing  map[string]error
	updates  map[string]*swarmUpdate
}

type swarmUpdate struct {
	spec swarm.ServiceSpec
	auth string
}

func (f *fakeSwarm) ServiceUpdateFunc(_ context.Context, serviceID string, _ *swarm.Service,
	fn func(int, *swarm.Service) (bool, error), _ int, _ time.Duration, options ...docker.ServiceUpdateOption) error {
	service := f.services[serviceID]
	if service == nil {
		return hperrors.NewNotFound("Service")
	}
	read := *service
	ok, err := fn(0, &read)
	if err != nil || !ok {
		return err
	}
	if err = f.failing[serviceID]; err != nil {
		return hperrors.NewInfra(err)
	}
	opts := client.ServiceUpdateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	if f.updates == nil {
		f.updates = map[string]*swarmUpdate{}
	}
	f.updates[serviceID] = &swarmUpdate{spec: read.Spec, auth: opts.EncodedRegistryAuth}
	return nil
}

func swarmService(id, image string) *swarm.Service {
	return &swarm.Service{ID: id, Spec: swarm.ServiceSpec{
		Annotations:  swarm.Annotations{Name: id},
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Image: image}},
	}}
}

func ecrAuth(t *testing.T, id, address, token string, expiresAt time.Time) *entity.Setting {
	t.Helper()
	setting := ecrSetting(t, token, expiresAt)
	setting.ID, setting.Name, setting.Status = id, id, base.SettingStatusActive
	auth := setting.MustAsRegistryAuth()
	auth.Address = address
	if _, region, ok := entity.ParseECRAddress(address); ok {
		auth.ECR.Region = region
	}
	setting.MustSetData(auth)
	return setting
}

func deploymentOf(appID string, settings *entity.AppDeploymentSettings) *entity.Setting {
	setting := &entity.Setting{ID: "dep-" + appID, Type: base.SettingTypeAppDeployment, ObjectID: appID}
	setting.MustSetData(settings)
	return setting
}

func imageApp(authID string) *entity.AppDeploymentSettings {
	return &entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodImage,
		ImageSource: &entity.DeploymentImageSource{Image: "x", RegistryAuth: entity.ObjectID{ID: authID}}}
}

func repoApp(authID string) *entity.AppDeploymentSettings {
	return &entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodRepo,
		RepoSource: &entity.DeploymentRepoSource{PushToRegistry: entity.ObjectID{ID: authID}}}
}

func newRenewService(store *settingStore, apps *appStore, sw *fakeSwarm, ecr *fakeECR) *service {
	svc := newTestService(nil, ecr)
	svc.settingRepo, svc.appRepo, svc.dockerManager = store, apps, sw
	return svc
}

func passwordOf(t *testing.T, header string) string {
	t.Helper()
	auth, err := authconfig.Decode(header)
	assert.NoError(t, err)
	return auth.Password
}

func renew(svc *service, interval time.Duration, targets ...string) (*registryauthservice.RenewResp, error) {
	return svc.Renew(context.Background(), database.Tx{}, &registryauthservice.RenewReq{
		Interval: interval, TargetAuths: targets})
}

// The services pulling with an ECR credential - an image app's, a repository
// app's that pushes there - get a new token, their spec sent back as read. An
// app whose active method does not use the credential is left alone.
func TestRenewHandsEveryServicePullingWithACredentialANewToken(t *testing.T) {
	store := &settingStore{
		auths: []*entity.Setting{ecrAuth(t, "ra1", ecrHost, "old", now.Add(2*time.Hour))},
		deployments: []*entity.Setting{
			deploymentOf("app1", imageApp("ra1")),
			deploymentOf("app2", repoApp("ra1")),
			// Its image source names ra1, but it is deployed from a repository.
			deploymentOf("app3", &entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodRepo,
				ImageSource: &entity.DeploymentImageSource{RegistryAuth: entity.ObjectID{ID: "ra1"}},
				RepoSource:  &entity.DeploymentRepoSource{}}),
		},
	}
	apps := &appStore{apps: []*entity.App{
		{ID: "app1", Name: "web", ServiceID: "svc1"},
		{ID: "app2", Name: "api", ServiceID: "svc2"},
		{ID: "app3", Name: "other", ServiceID: "svc3"},
	}}
	sw := &fakeSwarm{services: map[string]*swarm.Service{
		"svc1": swarmService("svc1", ecrHost+"/web:1@sha256:abc"),
		"svc2": swarmService("svc2", ecrHost+"/api:2"),
		"svc3": swarmService("svc3", ecrHost+"/other:3"),
	}}
	ecr := &fakeECR{}

	resp, err := renew(newRenewService(store, apps, sw, ecr), 6*time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, 1, ecr.calls)
	assert.Len(t, sw.updates, 2)
	for _, id := range []string{"svc1", "svc2"} {
		assert.Equal(t, "tok-1", passwordOf(t, sw.updates[id].auth))
		assert.Equal(t, sw.services[id].Spec, sw.updates[id].spec, "the spec goes back as it was read")
	}
	assert.Nil(t, sw.updates["svc3"])
	assert.Equal(t, []string{"ra1"}, resp.Output.Auths.ToIDStringSlice())
	assert.Len(t, resp.Output.Services, 2)
	assert.Empty(t, resp.Output.Failures)
}

// A kept token that outlives the next run by the margin is handed over as it
// is; the interval decides what is long enough.
func TestRenewUsesAKeptTokenThatOutlivesTheNextRun(t *testing.T) {
	newStore := func() *settingStore {
		return &settingStore{
			auths:       []*entity.Setting{ecrAuth(t, "ra1", ecrHost, "kept", now.Add(5*time.Hour))},
			deployments: []*entity.Setting{deploymentOf("app1", imageApp("ra1"))},
		}
	}
	apps := &appStore{apps: []*entity.App{{ID: "app1", Name: "web", ServiceID: "svc1"}}}

	sw, ecr := &fakeSwarm{services: map[string]*swarm.Service{"svc1": swarmService("svc1", ecrHost+"/web")}},
		&fakeECR{}
	_, err := renew(newRenewService(newStore(), apps, sw, ecr), 3*time.Hour)
	assert.NoError(t, err)
	assert.Zero(t, ecr.calls)
	assert.Equal(t, "kept", passwordOf(t, sw.updates["svc1"].auth))

	sw, ecr = &fakeSwarm{services: map[string]*swarm.Service{"svc1": swarmService("svc1", ecrHost+"/web")}},
		&fakeECR{}
	_, err = renew(newRenewService(newStore(), apps, sw, ecr), 6*time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, 1, ecr.calls)
	assert.Equal(t, "tok-1", passwordOf(t, sw.updates["svc1"].auth))
}

// Keys AWS refuses fail their credential - kept in the output, and the run - but
// the other credentials' services are still renewed.
func TestRenewKeepsGoingPastARefusedCredential(t *testing.T) {
	refused := ecrAuth(t, "ra2", ecrHost, "", time.Time{})
	refusedAuth := refused.MustAsRegistryAuth()
	refusedAuth.ECR.KeyAuth.ID = "ka-revoked"
	refused.MustSetData(refusedAuth)
	store := &settingStore{
		keys: map[string]*entity.Setting{
			"ka1":        keyAuth("ka1", "AKIAEXAMPLE000000000", "s3cret", 1, base.SettingStatusActive),
			"ka-revoked": keyAuth("ka-revoked", "AKIAREVOKED000000000", "s3cret", 1, base.SettingStatusActive),
		},
		auths: []*entity.Setting{ecrAuth(t, "ra1", ecrHost, "", time.Time{}), refused},
		deployments: []*entity.Setting{
			deploymentOf("app1", imageApp("ra1")),
			deploymentOf("app2", imageApp("ra2")),
		},
	}
	apps := &appStore{apps: []*entity.App{
		{ID: "app1", Name: "web", ServiceID: "svc1"},
		{ID: "app2", Name: "api", ServiceID: "svc2"},
	}}
	sw := &fakeSwarm{services: map[string]*swarm.Service{
		"svc1": swarmService("svc1", ecrHost+"/web"),
		"svc2": swarmService("svc2", ecrHost+"/api"),
	}}

	resp, err := renew(newRenewService(store, apps, sw, &fakeECR{}), 6*time.Hour)
	assert.Error(t, err)
	assert.Len(t, sw.updates, 1)
	assert.NotNil(t, sw.updates["svc1"])
	assert.Len(t, resp.Output.Failures, 1)
	assert.Equal(t, "ra2", resp.Output.Failures[0].Auth)
	assert.Empty(t, resp.Output.Failures[0].App)
	assert.Contains(t, resp.Output.Failures[0].Error, "were refused")
}

// A service whose update fails is a failure of its app; one whose image is in
// another registry - its credential changed, not deployed - or that is gone is
// skipped.
func TestRenewSkipsServicesItShouldNotTouch(t *testing.T) {
	store := &settingStore{
		auths: []*entity.Setting{ecrAuth(t, "ra1", ecrHost, "", time.Time{})},
		deployments: []*entity.Setting{
			deploymentOf("app1", imageApp("ra1")),
			deploymentOf("app2", imageApp("ra1")),
			deploymentOf("app3", imageApp("ra1")),
			deploymentOf("app4", imageApp("ra1")),
		},
	}
	apps := &appStore{apps: []*entity.App{
		{ID: "app1", Name: "elsewhere", ServiceID: "svc1"},
		{ID: "app2", Name: "gone", ServiceID: "svc-gone"},
		{ID: "app3", Name: "failing", ServiceID: "svc3"},
		{ID: "app4", Name: "never-deployed"},
	}}
	sw := &fakeSwarm{
		services: map[string]*swarm.Service{
			"svc1": swarmService("svc1", "ghcr.io/acme/web"),
			"svc3": swarmService("svc3", ecrHost+"/api"),
		},
		failing: map[string]error{"svc3": errors.New("rpc error: update out of sequence")},
	}

	resp, err := renew(newRenewService(store, apps, sw, &fakeECR{}), 6*time.Hour)
	assert.Error(t, err)
	assert.Empty(t, sw.updates)
	assert.Empty(t, resp.Output.Services)
	assert.Len(t, resp.Output.Failures, 1)
	assert.Equal(t, "app3", resp.Output.Failures[0].App)
}

// With no ECR credential a run does nothing: no AWS call, no update.
func TestRenewWithNoCredentialDoesNothing(t *testing.T) {
	sw, ecr := &fakeSwarm{}, &fakeECR{}
	resp, err := renew(newRenewService(&settingStore{}, &appStore{}, sw, ecr), 6*time.Hour)
	assert.NoError(t, err)
	assert.Zero(t, ecr.calls)
	assert.Empty(t, sw.updates)
	assert.Empty(t, resp.Output.Auths)
}

// A run for one credential, as on saving its keys, renews that one only.
func TestRenewForTargetsOnly(t *testing.T) {
	store := &settingStore{
		auths: []*entity.Setting{
			ecrAuth(t, "ra1", ecrHost, "", time.Time{}),
			ecrAuth(t, "ra2", otherHost, "", time.Time{}),
		},
	}
	ecr := &fakeECR{}
	resp, err := renew(newRenewService(store, &appStore{}, &fakeSwarm{}, ecr), 6*time.Hour, "ra2")
	assert.NoError(t, err)
	assert.Equal(t, 1, ecr.calls)
	assert.Equal(t, []string{"ra2"}, resp.Output.Auths.ToIDStringSlice())
}
