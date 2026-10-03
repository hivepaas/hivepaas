package appsettingsuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
)

// An app is paused only when none of the signals it scales on can be read;
// one that can is enough.
func TestUnreadableSignals(t *testing.T) {
	both := &entity.AppAutoscale{RequestsTarget: 10, CPUTarget: 70}
	onCPU := &entity.AppAutoscale{CPUTarget: 70}

	assert.Empty(t, unreadableSignals(both, &appautoscaleservice.Check{Requests: "not-exposed"}))
	assert.Empty(t, unreadableSignals(both, &appautoscaleservice.Check{CPU: "no-cpu-limit"}))
	assert.Equal(t, []signalReason{{"requests", "not-exposed"}, {"CPU", "no-cpu-limit"}},
		unreadableSignals(both, &appautoscaleservice.Check{Requests: "not-exposed", CPU: "no-cpu-limit"}))
	assert.Equal(t, []signalReason{{"CPU", "no-cpu-limit"}},
		unreadableSignals(onCPU, &appautoscaleservice.Check{CPU: "no-cpu-limit"}))
	assert.Empty(t, unreadableSignals(onCPU, &appautoscaleservice.Check{Requests: "not-exposed"}),
		"requests it does not scale on")
	assert.Empty(t, unreadableSignals(&entity.AppAutoscale{}, &appautoscaleservice.Check{CPU: "x"}))
}

// What cannot be read is said as a person reads it, every signal.
func TestUnreadableText(t *testing.T) {
	text := unreadableText([]signalReason{{"requests", "access-log-unlabelled"}, {"CPU", "no-cpu-limit"}})
	assert.Equal(t, "its requests cannot be read, as Traefik's access log is written in an older form; "+
		"an administrator saves System → Traefik → Config Options once, with Access Log on, and Traefik restarts "+
		"briefly; and its CPU cannot be read, as it has neither a CPU limit nor a CPU reservation to measure its "+
		"CPU by; set one in its Resources", text)
}

// The refusals read as sentences: the app's name and why, nothing left unset.
func TestAutoscaleRefusalsRead(t *testing.T) {
	unreadable := hperrors.Wrap(hperrors.ErrAutoscaleUnreadable).WithParam("Name", "web").
		WithParam("Reason", unreadableText([]signalReason{{"requests", "not-exposed"}})).Build(translation.LangEn)
	assert.Equal(t, "Autoscale cannot be turned on for 'web': its requests cannot be read, as it has no domain, "+
		"so none of its requests go through Traefik", unreadable.Detail)

	noSignal := hperrors.Wrap(hperrors.ErrAutoscaleNoSignal).WithParam("Name", "web").Build(translation.LangEn)
	assert.Equal(t, "Autoscale scales 'web' on its requests, its CPU or both: set Requests Per Instance or CPU "+
		"Target", noSignal.Detail)

	hostPorts := hperrors.Wrap(hperrors.ErrAutoscaleHostPorts).WithParam("Name", "web").Build(translation.LangEn)
	assert.NotContains(t, hostPorts.Detail, "<no value>")
	assert.Contains(t, hostPorts.Detail, "'web' publishes a port in host mode")
}
