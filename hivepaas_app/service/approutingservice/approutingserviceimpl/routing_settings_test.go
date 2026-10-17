package approutingserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// testRetryDelay only has to be non-zero: the waiting itself is not what these
// tests are about.
const testRetryDelay = time.Millisecond

// errOutOfSequence is swarm refusing an update, built the way it reaches the
// service: the docker client's error for the daemon's 500, wrapped by
// hperrors.NewInfra in the docker manager.
func errOutOfSequence() error {
	return hperrors.NewInfra(errdefs.ErrInternal.WithMessage(
		"Error response from daemon: rpc error: code = Unknown desc = update out of sequence"))
}

// pushedUpdate is what one ServiceUpdate call carried.
type pushedUpdate struct {
	version uint64
	labels  map[string]string
	nets    []swarm.NetworkAttachmentConfig
	config  *swarm.UpdateConfig
}

// fakeSwarm hands out the service at a version that moves on with every inspect,
// as swarm's allocator moves it on after a create, and answers the updates with
// the errors it is given, in turn, until they run out.
type fakeSwarm struct {
	docker.Manager
	version  uint64
	inspects int
	updates  []pushedUpdate
	errs     []error
}

func (f *fakeSwarm) ServiceInspect(
	_ context.Context, serviceID string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	f.inspects++
	f.version++
	return &client.ServiceInspectResult{Service: swarm.Service{
		ID:   serviceID,
		Meta: swarm.Meta{Version: swarm.Version{Index: f.version}},
		Spec: swarm.ServiceSpec{Annotations: swarm.Annotations{Labels: map[string]string{"owner": "app"}}},
	}}, nil
}

func (f *fakeSwarm) ServiceUpdate(
	_ context.Context, _ string, version *swarm.Version, spec *swarm.ServiceSpec, _ ...docker.ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	f.updates = append(f.updates, pushedUpdate{
		version: version.Index,
		labels:  maps.Clone(spec.Labels),
		nets:    spec.TaskTemplate.Networks,
		config:  spec.UpdateConfig,
	})
	if len(f.errs) == 0 {
		return &client.ServiceUpdateResult{}, nil
	}
	err := f.errs[0]
	f.errs = f.errs[1:]
	return nil, err
}

// fakeTraefik writes a label onto the spec, as the real one writes the app's
// Traefik labels, so a test can see whether an update carried it.
type fakeTraefik struct {
	traefikservice.Service
	calls int
	err   error
}

func (f *fakeTraefik) ApplyAppConfig(
	_ context.Context, _ database.IDB, req *traefikservice.ApplyAppConfigReq,
) (*traefikservice.ApplyAppConfigResp, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	req.Service.Spec.Labels["traefik.enable"] = "true"
	return &traefikservice.ApplyAppConfigResp{Service: req.Service}, nil
}

// fakeNetworks attaches the service to the routing network, as the real one does
// for an app exposed publicly.
type fakeNetworks struct {
	networkservice.Service
	calls int
}

func (f *fakeNetworks) UpdateAppGlobalRoutingNetwork(
	_ context.Context, _ *entity.App, service *swarm.Service, _ *entity.AppRoutingSettings,
) error {
	f.calls++
	service.Spec.TaskTemplate.Networks = []swarm.NetworkAttachmentConfig{{Target: "routing-net"}}
	return nil
}

type fakeCertFiles struct {
	sslservice.Service
	calls int
}

func (f *fakeCertFiles) WriteCertFiles(_ bool, _ ...*entity.Setting) error {
	f.calls++
	return nil
}

type routingFakes struct {
	swarm    *fakeSwarm
	traefik  *fakeTraefik
	networks *fakeNetworks
	certs    *fakeCertFiles
}

func newRoutingTest() (*service, *routingFakes) {
	fakes := &routingFakes{
		swarm:    &fakeSwarm{version: 10},
		traefik:  &fakeTraefik{},
		networks: &fakeNetworks{},
		certs:    &fakeCertFiles{},
	}
	svc := &service{
		dockerManager:          fakes.swarm,
		networkService:         fakes.networks,
		settingService:         &recordingSettingService{},
		sslService:             fakes.certs,
		traefikService:         fakes.traefik,
		routingApplyRetryDelay: testRetryDelay,
	}
	return svc, fakes
}

func routingReq() *approutingservice.ApplyAppRoutingReq {
	return &approutingservice.ApplyAppRoutingReq{
		App:             &entity.App{ID: "app-1", ServiceID: "svc-1"},
		RoutingSettings: &entity.AppRoutingSettings{Port: 80, ExposePublicly: true},
	}
}

// Saving routing right after a deployment, or applying it to a service created a
// moment ago, races swarm's allocator for the service's version. The update that
// loses is not a failure of the settings, so it is applied again - and from a
// fresh inspect, with every step that writes into the spec run again on it.
func TestApplyRoutingSettingsRetriesAnUpdateOutOfSequence(t *testing.T) {
	svc, fakes := newRoutingTest()
	fakes.swarm.errs = []error{errOutOfSequence()}

	resp, err := svc.ApplyRoutingSettings(context.Background(), nil, routingReq())

	assert.NoError(t, err)
	assert.Equal(t, 2, fakes.swarm.inspects, "the retry starts from a fresh inspect")
	if assert.Len(t, fakes.swarm.updates, 2) {
		assert.Equal(t, uint64(11), fakes.swarm.updates[0].version)
		assert.Equal(t, uint64(12), fakes.swarm.updates[1].version, "the retry carries the version swarm holds now")

		// What the steps wrote has to be on the update that goes through, not only
		// on the one swarm refused.
		retried := fakes.swarm.updates[1]
		assert.Equal(t, "true", retried.labels["traefik.enable"])
		assert.Equal(t, "app", retried.labels["owner"], "what the service had is kept")
		assert.Equal(t, []swarm.NetworkAttachmentConfig{{Target: "routing-net"}}, retried.nets)
	}
	assert.Equal(t, 2, fakes.traefik.calls)
	assert.Equal(t, 2, fakes.networks.calls)
	assert.Equal(t, 2, fakes.certs.calls)
	if assert.NotNil(t, resp) {
		assert.Equal(t, uint64(12), resp.Service.Version.Index, "the service returned is the one pushed")
	}
}

// The refusal is recognized however deeply it is wrapped on the way up, and only
// by what swarm says - another infra error with the same code is not it.
func TestIsUpdateOutOfSequence(t *testing.T) {
	refused := errOutOfSequence()
	assert.ErrorIs(t, refused, hperrors.ErrInfraInternal, "the error is built as the docker manager builds it")

	assert.True(t, isUpdateOutOfSequence(refused))
	assert.True(t, isUpdateOutOfSequence(hperrors.Wrap(refused)))
	assert.True(t, isUpdateOutOfSequence(hperrors.Wrap(refused).WithExtraDetail("while updating")))
	assert.True(t, isUpdateOutOfSequence(fmt.Errorf("applying routing: %w", hperrors.Wrap(refused))))

	assert.False(t, isUpdateOutOfSequence(nil))
	assert.False(t, isUpdateOutOfSequence(hperrors.NewInfra(errdefs.ErrInternal.WithMessage(
		"Error response from daemon: rpc error: code = Unknown desc = network routing-net not found"))))
	assert.False(t, isUpdateOutOfSequence(hperrors.ErrInfraInternal))
}

// Any other failure is returned at once: it would fail the same way again, and
// most callers are holding a transaction while they wait.
func TestApplyRoutingSettingsDoesNotRetryOtherErrors(t *testing.T) {
	t.Run("an update refused for another reason", func(t *testing.T) {
		svc, fakes := newRoutingTest()
		refused := hperrors.NewInfra(errdefs.ErrInvalidArgument.WithMessage(
			"Error response from daemon: rpc error: code = InvalidArgument desc = port 80 is in use"))
		fakes.swarm.errs = []error{refused}

		_, err := svc.ApplyRoutingSettings(context.Background(), nil, routingReq())

		assert.ErrorIs(t, err, refused)
		assert.Len(t, fakes.swarm.updates, 1)
		assert.Equal(t, 1, fakes.swarm.inspects)
	})

	t.Run("a step before the update", func(t *testing.T) {
		svc, fakes := newRoutingTest()
		errTraefik := errors.New("traefik service not found")
		fakes.traefik.err = errTraefik

		_, err := svc.ApplyRoutingSettings(context.Background(), nil, routingReq())

		assert.ErrorIs(t, err, errTraefik)
		assert.Equal(t, 1, fakes.traefik.calls)
		assert.Empty(t, fakes.swarm.updates)
	})
}

// A service that keeps moving is given up on after a few attempts, with the error
// of the last one.
func TestApplyRoutingSettingsGivesUpAfterTheRetries(t *testing.T) {
	svc, fakes := newRoutingTest()
	errs := make([]error, routingApplyRetryMax+1)
	for i := range errs {
		errs[i] = errOutOfSequence()
	}
	fakes.swarm.errs = append([]error{}, errs...)

	_, err := svc.ApplyRoutingSettings(context.Background(), nil, routingReq())

	assert.ErrorIs(t, err, errs[len(errs)-1])
	assert.NotErrorIs(t, err, errs[0])
	assert.True(t, isUpdateOutOfSequence(err), "the caller is told what swarm said")
	assert.Len(t, fakes.swarm.updates, routingApplyRetryMax+1)
	assert.Equal(t, routingApplyRetryMax+1, fakes.swarm.inspects)
}

// A caller that gives up stops the retries: the wait does not outlast it.
func TestApplyRoutingSettingsStopsRetryingWhenTheContextEnds(t *testing.T) {
	svc, fakes := newRoutingTest()
	svc.routingApplyRetryDelay = time.Hour
	fakes.swarm.errs = []error{errOutOfSequence()}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := svc.ApplyRoutingSettings(ctx, nil, routingReq())

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Len(t, fakes.swarm.updates, 1)
}

// A service handed in by the caller is used for the first attempt, which saves an
// inspect. After a refusal it is exactly the stale copy that must not be pushed
// again, so the retry inspects - and leaves the caller's request as it was.
func TestApplyRoutingSettingsDoesNotReuseTheCallersServiceOnRetry(t *testing.T) {
	svc, fakes := newRoutingTest()
	fakes.swarm.errs = []error{errOutOfSequence()}
	callers := &swarm.Service{
		ID:   "svc-1",
		Meta: swarm.Meta{Version: swarm.Version{Index: 5}},
		Spec: swarm.ServiceSpec{Annotations: swarm.Annotations{Labels: map[string]string{}}},
	}
	req := routingReq()
	req.Service = callers

	resp, err := svc.ApplyRoutingSettings(context.Background(), nil, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, fakes.swarm.inspects, "only the retry inspects")
	if assert.Len(t, fakes.swarm.updates, 2) {
		assert.Equal(t, uint64(5), fakes.swarm.updates[0].version, "the first attempt uses the caller's service")
		assert.Equal(t, uint64(11), fakes.swarm.updates[1].version)
	}
	if assert.NotNil(t, resp) {
		assert.NotSame(t, callers, resp.Service)
	}
	assert.Same(t, callers, req.Service, "the caller's request is not rewritten")
}

// The HivePaaS app's routing, and its revert, roll back an update that fails.
// Those terms are part of the update the apply retries, so they are on every
// attempt rather than only the first.
func TestApplyRoutingSettingsSetsTheRollbackTermsOnEveryAttempt(t *testing.T) {
	svc, fakes := newRoutingTest()
	fakes.swarm.errs = []error{errOutOfSequence()}
	req := routingReq()
	req.RollbackOnFailure = true

	_, err := svc.ApplyRoutingSettings(context.Background(), nil, req)

	assert.NoError(t, err)
	if assert.Len(t, fakes.swarm.updates, 2) {
		for _, update := range fakes.swarm.updates {
			if assert.NotNil(t, update.config) {
				assert.Equal(t, swarm.UpdateFailureActionRollback, update.config.FailureAction)
				assert.InDelta(t, updateMaxFailureRatio, update.config.MaxFailureRatio, 0)
			}
		}
	}

	t.Run("and only when asked", func(t *testing.T) {
		svc, fakes := newRoutingTest()

		_, err := svc.ApplyRoutingSettings(context.Background(), nil, routingReq())

		assert.NoError(t, err)
		if assert.Len(t, fakes.swarm.updates, 1) {
			assert.Nil(t, fakes.swarm.updates[0].config)
		}
	})
}
