package approutingserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

// The HivePaaS app has to be swept first.
//
// The trial this sweep runs inside exists to guard reachability of that one app.
// Left until last, the operator could be handed a working dashboard and a confirm
// button while their own labels still carried the old depth - vouching for a
// change that had not been applied to the thing they were vouching through.
func TestSweepOrderPutsThePrimaryAppFirst(t *testing.T) {
	apps := []*entity.App{{ID: "app-c"}, {ID: "app-a"}, {ID: "hivepaas"}, {ID: "app-b"}}

	ordered := sweepTargets(apps, "hivepaas", false)

	ids := make([]string, 0, len(ordered))
	for _, app := range ordered {
		ids = append(ids, app.ID)
	}
	assert.Equal(t, []string{"hivepaas", "app-a", "app-b", "app-c"}, ids)

	// The input must not be reordered under the caller.
	assert.Equal(t, "app-c", apps[0].ID)
}

// The rest of the order is stable so a sweep that fails partway is repeatable
// rather than random - the operator retrying should hit the same apps in the
// same sequence, not a new one.
func TestSweepOrderIsStableWithoutAPrimary(t *testing.T) {
	apps := []*entity.App{{ID: "b"}, {ID: "a"}, {ID: "c"}}

	first := sweepTargets(apps, "", false)
	second := sweepTargets(apps, "", false)

	assert.Equal(t, first[0].ID, second[0].ID)
	assert.Equal(t, "a", first[0].ID)
	assert.Equal(t, "c", first[2].ID)
}

// An app with neither a rate limit nor an allowlist carries no ip-strategy depth,
// so the sweep has nothing to rewrite on it. Getting this wrong in the permissive
// direction costs a pointless service update per app; getting it wrong in the
// other direction leaves an app reading the wrong forwarded position with nothing
// to notice.
func TestSweepSkipsAppsThatDoNotReadClientAddresses(t *testing.T) {
	withRouting := func(data string) *entity.App {
		return &entity.App{
			ID: "app",
			Settings: []*entity.Setting{{
				ID:   "set",
				Type: "app-routing",
				Data: data,
			}},
		}
	}

	plain := withRouting(`{"exposePublicly":true,"domains":[{"domain":"x.example.com"}]}`)
	settings, err := plain.GetSettingByType("app-routing").AsAppRoutingSettings()
	assert.NoError(t, err)
	assert.False(t, settings.UsesClientIP())

	guarded := withRouting(
		`{"exposePublicly":true,"domains":[{"domain":"x.example.com",` +
			`"clientConfig":{"enabled":true,"allowedIPs":["10.0.0.0/8"]}}]}`)
	settings, err = guarded.GetSettingByType("app-routing").AsAppRoutingSettings()
	assert.NoError(t, err)
	assert.True(t, settings.UsesClientIP())
}

// Leniency about missing references is asymmetric on purpose, and the asymmetry
// is the whole point: a revert has to survive a reference that was deleted while
// the change was on trial, because refusing leaves the operator in the
// configuration they are trying to escape. An apply must not, because dropping
// the reference there takes the basic auth off an app that is meant to have it.
//
// It also has to be requested through the flag. Pre-loading the references
// leniently does nothing: the apply reloads them and re-queries exactly the ones
// a lenient load left out - which is how this was wrong to begin with.
func TestMissingReferencesAreToleratedOnlyOnTheWayBack(t *testing.T) {
	t.Run("the revert asks for leniency", func(t *testing.T) {
		req := revertApplyReq(&entity.App{ID: "app"}, &entity.AppRoutingSettings{})
		assert.True(t, req.SkipMissingRefObjects)
		assert.True(t, req.SkipUpdatingService, "the revert pushes the service itself, afterwards")
	})

	t.Run("the forward sweep does not", func(t *testing.T) {
		req := sweepApplyReq(&entity.App{ID: "app"}, &entity.AppRoutingSettings{}, false)
		assert.False(t, req.SkipMissingRefObjects)
		assert.False(t, req.SkipUpdatingService, "the sweep exists to push labels onto the service")
	})

	t.Run("a sweep running inside a revert does", func(t *testing.T) {
		req := sweepApplyReq(&entity.App{ID: "app"}, &entity.AppRoutingSettings{}, true)
		assert.True(t, req.SkipMissingRefObjects)
	})
}

// recordingSettingService answers only the two calls this test is about.
//
// The embedded interface is nil on purpose: anything else reaching it panics,
// which is the assertion that loadAppRoutingData touches nothing else.
type recordingSettingService struct {
	settingservice.Service
	strictCalls  int
	lenientCalls int
}

func (r *recordingSettingService) LoadRefObjectsByIDs(
	_ context.Context, _ database.IDB, pRefObjects **entity.RefObjects,
	_ *entity.ObjectScope, _ bool, _ *entity.RefObjectIDs,
) error {
	r.strictCalls++
	*pRefObjects = entity.NewRefObjects()
	return nil
}

func (r *recordingSettingService) LoadRefObjectsByIDsSkipMissing(
	_ context.Context, _ database.IDB, pRefObjects **entity.RefObjects,
	_ *entity.ObjectScope, _ bool, _ *entity.RefObjectIDs,
) error {
	r.lenientCalls++
	*pRefObjects = entity.NewRefObjects()
	return nil
}

// Setting the flag is not enough - loadAppRoutingData has to act on it.
//
// This is the half that was broken: the revert asked for leniency by pre-loading
// the references with the lenient call, and the apply then reloaded them with the
// strict one, re-querying exactly the references the lenient load had left out.
// The flag was added to fix that, so the flag has to be what selects the loader.
func TestLoadAppRoutingDataHonoursTheLeniencyFlag(t *testing.T) {
	run := func(skipMissing bool) *recordingSettingService {
		recorder := &recordingSettingService{}
		svc := &service{settingService: recorder}
		data := &applyAppRoutingData{ApplyAppRoutingReq: &approutingservice.ApplyAppRoutingReq{
			App:             &entity.App{ID: "app", ServiceID: "svc"},
			RoutingSettings: &entity.AppRoutingSettings{},
			// Non-nil so nothing tries to reach docker.
			Service:               &swarm.Service{},
			SkipMissingRefObjects: skipMissing,
		}}
		assert.NoError(t, svc.loadAppRoutingData(context.Background(), nil, data))
		return recorder
	}

	strict := run(false)
	assert.Equal(t, 1, strict.strictCalls)
	assert.Equal(t, 0, strict.lenientCalls, "an apply must not tolerate a reference that is gone")

	lenient := run(true)
	assert.Equal(t, 1, lenient.lenientCalls)
	assert.Equal(t, 0, lenient.strictCalls, "a revert must not be aborted by a reference that is gone")
}

// While a change is on trial only the HivePaaS app carries the new depth: it is
// the only one the trial asks a question about, and keeping the rest out of it
// keeps the undo down to a single app. The fan-out happens once somebody
// confirms.
func TestSweepIsNarrowedWhileAChangeIsOnTrial(t *testing.T) {
	apps := []*entity.App{{ID: "app-b"}, {ID: "hivepaas"}, {ID: "app-a"}}

	onTrial := sweepTargets(apps, "hivepaas", true)
	assert.Len(t, onTrial, 1)
	assert.Equal(t, "hivepaas", onTrial[0].ID)

	confirmed := sweepTargets(apps, "hivepaas", false)
	assert.Len(t, confirmed, 3)
	assert.Equal(t, "hivepaas", confirmed[0].ID, "the primary app still goes first")

	// A primary that is not in the list must not silently sweep everything.
	assert.Empty(t, sweepTargets(apps, "missing", true))
}
