package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// The interval is kept within what a 12-hour token allows, whatever is stored.
func TestRegistryAuthRenewalIntervalStaysWithinItsBounds(t *testing.T) {
	for stored, want := range map[time.Duration]time.Duration{
		0:                RegistryAuthRenewalIntervalDefault,
		30 * time.Minute: time.Hour,
		3 * time.Hour:    3 * time.Hour,
		24 * time.Hour:   10 * time.Hour,
	} {
		renewal := &RegistryAuthRenewal{Schedule: SchedJobSchedule{Interval: timeutil.Duration(stored)}}
		assert.Equal(t, want, renewal.Interval(), stored.String())
	}
}

// A new installation's renewal runs every 6 hours from the next hour, and tells
// the default notification of a failure.
func TestNewRegistryAuthRenewal(t *testing.T) {
	renewal := NewRegistryAuthRenewal(time.Date(2026, 10, 2, 12, 34, 0, 0, time.UTC))
	assert.Equal(t, 6*time.Hour, renewal.Interval())
	assert.Equal(t, time.Date(2026, 10, 2, 13, 0, 0, 0, time.UTC), renewal.Schedule.InitialTime)
	assert.NoError(t, renewal.Schedule.IsValid())
	assert.True(t, renewal.Notification.FailureUseDefault)
	assert.False(t, renewal.Notification.SuccessUseDefault)
}

// The credential a service pulls with is the active method's.
func TestServiceRegistryAuthIDFollowsTheActiveMethod(t *testing.T) {
	settings := &AppDeploymentSettings{
		ImageSource:    &DeploymentImageSource{RegistryAuth: ObjectID{ID: "image"}},
		RepoSource:     &DeploymentRepoSource{PushToRegistry: ObjectID{ID: "repo"}},
		FunctionSource: &DeploymentFunctionSource{PushToRegistry: ObjectID{ID: "function"}},
	}
	for method, want := range map[base.DeploymentMethod]string{
		base.DeploymentMethodImage:    "image",
		base.DeploymentMethodRepo:     "repo",
		base.DeploymentMethodFunction: "function",
	} {
		settings.ActiveMethod = method
		assert.Equal(t, want, settings.ServiceRegistryAuthID(), string(method))
	}
	assert.Empty(t, (&AppDeploymentSettings{ActiveMethod: base.DeploymentMethodRepo}).ServiceRegistryAuthID())
}
