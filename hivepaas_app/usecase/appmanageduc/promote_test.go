package appmanageduc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// fakeSteps records the steps run, each failing as asked.
type fakeSteps struct {
	ran        []string
	fenceErr   error
	stopped    bool
	stopErr    error
	exit       *int
	promoteErr error
	unfenceErr error
	recordErr  error
}

func (f *fakeSteps) fence(context.Context) error {
	f.ran = append(f.ran, "fence")
	return f.fenceErr
}

func (f *fakeSteps) stop(context.Context) (bool, error) {
	f.ran = append(f.ran, "stop")
	return f.stopped, f.stopErr
}

func (f *fakeSteps) applyFence(context.Context) { f.ran = append(f.ran, "applyFence") }

func (f *fakeSteps) promote(context.Context) (*int, error) {
	f.ran = append(f.ran, "promote")
	return f.exit, f.promoteErr
}

func (f *fakeSteps) allow(context.Context) { f.ran = append(f.ran, "allow") }

func (f *fakeSteps) unfence(context.Context) error {
	f.ran = append(f.ran, "unfence")
	return f.unfenceErr
}

func (f *fakeSteps) record(context.Context) error {
	f.ran = append(f.ran, "record")
	return f.recordErr
}

func (f *fakeSteps) restart(context.Context) { f.ran = append(f.ran, "restart") }

func (f *fakeSteps) moveAlias(context.Context) { f.ran = append(f.ran, "moveAlias") }

func exitCode(code int) *int { return &code }

func runSteps(steps *fakeSteps) error {
	return runPromotion(context.Background(), steps, "db-r1", "db")
}

func TestAPromotionRunsItsStepsInOrder(t *testing.T) {
	steps := &fakeSteps{stopped: true, exit: exitCode(0)}

	err := runSteps(steps)

	assert.NoError(t, err)
	assert.Equal(t, []string{"fence", "stop", "applyFence", "promote", "allow", "record", "moveAlias", "restart"},
		steps.ran)
}

// Before the member is promoted, a failure lifts the fence; once whether it was
// is not known, or once it was, the fence stays.
func TestAPromotionStopsAtTheStepThatFails(t *testing.T) {
	for name, tc := range map[string]struct {
		steps  *fakeSteps
		ran    []string
		reason string
	}{
		"the fence not written": {&fakeSteps{fenceErr: hperrors.NewNotFound("Variables of db")},
			[]string{"fence"}, ""},
		"not seen stopped": {&fakeSteps{stopped: false},
			[]string{"fence", "stop", "unfence"}, "db could not be seen stopped, so nothing was promoted; db runs again"},
		"stopping failed": {&fakeSteps{stopErr: hperrors.Wrap(hperrors.ErrInfraActionFailed)},
			[]string{"fence", "stop", "unfence"}, "db could not be seen stopped"},
		"certainly not promoted": {&fakeSteps{stopped: true, exit: exitCode(3)},
			[]string{"fence", "stop", "applyFence", "promote", "unfence"}, "db-r1 was not promoted; db runs again"},
		"the promotion did not run": {&fakeSteps{stopped: true,
			promoteErr: hperrors.Wrap(hperrors.ErrActiveContainerNotFound)},
			[]string{"fence", "stop", "applyFence", "promote", "unfence"}, "db-r1 was not promoted"},
		"not known whether promoted": {&fakeSteps{stopped: true, exit: exitCode(1)},
			[]string{"fence", "stop", "applyFence", "promote"}, "whether db-r1 was promoted is not known"},
		"promoted, not recorded": {&fakeSteps{stopped: true, exit: exitCode(0),
			recordErr: hperrors.Wrap(hperrors.ErrInternal)},
			[]string{"fence", "stop", "applyFence", "promote", "allow", "record"}, "db-r1 is the primary, but recording"},
		"the fence not lifted either": {&fakeSteps{stopped: true, exit: exitCode(3),
			unfenceErr: hperrors.Wrap(hperrors.ErrInfraActionFailed)},
			[]string{"fence", "stop", "applyFence", "promote", "unfence"}, "lifting db's fence failed too"},
	} {
		err := runSteps(tc.steps)

		assert.Error(t, err, name)
		assert.Equal(t, tc.ran, tc.steps.ran, name)
		if tc.reason != "" {
			assert.ErrorIs(t, err, hperrors.ErrAppManagedPromoteFailed, name)
			info, _ := hperrors.ParseError(err, "en")
			assert.Contains(t, info.Detail, tc.reason, name)
		}
	}
}

func TestChangedEnv(t *testing.T) {
	vars := []*entity.EnvVar{
		{Key: "POSTGRES_DB", Value: "app"},
		{Key: "HIVEPAAS_STANDBY_HOST", Value: "db", IsLiteral: true},
		{Key: "HIVEPAAS_STANDBY_NAME", Value: "db-r2", IsLiteral: true},
	}

	next, changed := changedEnv(vars, [][2]string{{"HIVEPAAS_STANDBY_HOST", "db-r1"}}, nil)
	assert.True(t, changed)
	assert.Equal(t, "db-r1", next[1].Value)

	next, changed = changedEnv(next, [][2]string{{"HIVEPAAS_STANDBY_HOST", "db-r1"}}, nil)
	assert.False(t, changed, "already so")
	assert.Len(t, next, 3)

	next, changed = changedEnv(next, [][2]string{{"HIVEPAAS_FENCED", "db-r1 was promoted"}},
		[]string{"HIVEPAAS_STANDBY_HOST", "HIVEPAAS_STANDBY_NAME"})
	assert.True(t, changed)
	assert.Equal(t, []string{"POSTGRES_DB", "HIVEPAAS_FENCED"}, keysOfEnv(next))
	assert.True(t, next[1].IsLiteral)
}

func keysOfEnv(vars []*entity.EnvVar) []string {
	out := make([]string, 0, len(vars))
	for _, env := range vars {
		out = append(out, env.Key)
	}
	return out
}
