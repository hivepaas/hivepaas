package loggingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// fakeBackend records the request and the endpoint it was built for.
type fakeBackend struct {
	logging.Backend
	endpoint logging.Endpoint
	got      *logging.QueryReq
}

func (f *fakeBackend) Query(_ context.Context, req *logging.QueryReq) (*logging.QueryResp, error) {
	f.got = req
	return &logging.QueryResp{}, nil
}

func withFakeBackend(s *service) *fakeBackend {
	fb := &fakeBackend{}
	s.newBackend = func(_ logging.BackendType, cfg *logging.BackendConfig) (logging.Backend, error) {
		fb.endpoint = cfg.VictoriaLogs.Endpoint
		return fb, nil
	}
	return fb
}

func TestQueryAppLogsScopesToTheLoadedApp(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	fb := withFakeBackend(s)

	_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"},
		&loggingservice.AppLogQuery{Contains: "boom", Levels: []string{"error"}, Limit: 50})
	assert.NoError(t, err)

	if assert.NotNil(t, fb.got) {
		assert.Equal(t, []logging.FieldMatch{{Field: "attrs." + appservice.LabelLogAppID, Value: "APP1"}}, fb.got.Match)
		assert.Equal(t, "boom", fb.got.Contains)
		assert.Equal(t, []string{"error"}, fb.got.Levels)
		assert.Equal(t, 50, fb.got.Limit)
	}
	assert.Equal(t, backendBaseURL(), fb.endpoint.URL, "a managed backend is reached by service name")
}

func TestQueryAppLogsRefusesAnAppWithoutAnID(t *testing.T) {
	s := newTestService(&fakeDocker{}, storedSetting(t, enabledConfig()))
	fb := withFakeBackend(s)

	_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{}, &loggingservice.AppLogQuery{Limit: 1})
	assert.Error(t, err)
	assert.Nil(t, fb.got, "an empty id would match every line with no app at all")
}

func TestQueryAppLogsWhenDisabled(t *testing.T) {
	cfg := enabledConfig()
	cfg.Enabled = false
	for _, setting := range []*entity.Setting{nil, storedSetting(t, cfg)} {
		s := newTestService(&fakeDocker{}, setting)
		withFakeBackend(s)
		_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"},
			&loggingservice.AppLogQuery{Limit: 1})
		assert.ErrorIs(t, err, loggingservice.ErrNotEnabled)
	}
}

func TestQueryAppLogsUsesTheQueryEndpointOfAnUnmanagedBackend(t *testing.T) {
	cfg := enabledConfig()
	cfg.Backend.Managed = false
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "http://theirs:9428/insert"}
	s := newTestService(&fakeDocker{}, storedSetting(t, cfg))
	withFakeBackend(s)

	_, err := s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"}, &loggingservice.AppLogQuery{Limit: 1})
	assert.ErrorIs(t, err, loggingservice.ErrQueryEndpointMissing)

	cfg.Backend.Query = &entity.LoggingEndpoint{URL: "http://theirs:9428"}
	s = newTestService(&fakeDocker{}, storedSetting(t, cfg))
	fb := withFakeBackend(s)
	_, err = s.QueryAppLogs(context.Background(), nil, &entity.App{ID: "APP1"}, &loggingservice.AppLogQuery{Limit: 1})
	assert.NoError(t, err)
	assert.Equal(t, "http://theirs:9428", fb.endpoint.URL)
}

func (f *fakeDocker) withService(id string, spec swarm.ServiceSpec) *fakeDocker {
	if f.inspected == nil {
		f.inspected = map[string]swarm.Service{}
	}
	f.inspected[id] = swarm.Service{ID: id, Spec: spec}
	return f
}

func collectibleSpec(appID string) swarm.ServiceSpec {
	return swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Labels: map[string]string{appservice.LabelLogAppID: appID}},
		LogDriver:     appservice.DefaultLogDriver(),
	}}
}

func TestAppHistory(t *testing.T) {
	disabled := enabledConfig()
	disabled.Enabled = false
	noApps := enabledConfig()
	noApps.Sources = entity.LoggingSources{HivePaaS: true}
	unmanagedNoQuery := enabledConfig()
	unmanagedNoQuery.Backend.Managed = false
	unmanagedNoQuery.Backend.Ingest = &entity.LoggingEndpoint{URL: "http://theirs/insert"}

	local := collectibleSpec("APP1")
	local.TaskTemplate.LogDriver = &swarm.Driver{Name: "local"}
	noIdentity := collectibleSpec("SOURCE-APP") // cloned before the deep-copy fix

	cases := []struct {
		name    string
		setting *entity.Setting
		spec    *swarm.ServiceSpec
		want    loggingservice.AppHistory
	}{
		{"never configured", nil, nil, loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDisabled}},
		{"disabled", storedSetting(t, disabled), nil,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDisabled}},
		{"apps not collected", storedSetting(t, noApps), nil,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonAppsNotCollected}},
		{"no query endpoint", storedSetting(t, unmanagedNoQuery), nil,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonNoQueryEndpoint}},
		{"local driver", storedSetting(t, enabledConfig()), &local,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDriverUnreadable}},
		{"wrong identity", storedSetting(t, enabledConfig()), &noIdentity,
			loggingservice.AppHistory{Reason: loggingservice.HistoryReasonIdentityMissing}},
		{"collected", storedSetting(t, enabledConfig()), new(collectibleSpec("APP1")),
			loggingservice.AppHistory{Available: true}},
		// The service is gone: what it logged is exactly what history is for.
		{"service removed", storedSetting(t, enabledConfig()), nil, loggingservice.AppHistory{Available: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fd := &fakeDocker{}
			if tc.spec != nil {
				fd.withService("svc-app", *tc.spec)
			}
			s := newTestService(fd, tc.setting)
			got, err := s.AppHistory(context.Background(), nil, &entity.App{ID: "APP1", ServiceID: "svc-app"})
			assert.NoError(t, err)
			assert.Equal(t, tc.want, *got)
		})
	}
}
