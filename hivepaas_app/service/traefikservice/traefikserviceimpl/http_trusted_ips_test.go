package traefikserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func traefikSvcWithArgs(args ...string) *swarm.Service {
	return &swarm.Service{Spec: swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Args: args}},
	}}
}

func applyTrustedIPs(t *testing.T, svc *swarm.Service, trustedIPs []string) []string {
	t.Helper()
	s := &service{dockerManager: &fakeDockerManager{traefikSvc: svc}}

	resp, err := s.ApplyTrustedIPsToWebEntrypoints(context.Background(), &traefikservice.ApplyTrustedIPsReq{
		TrustedIPs:                  trustedIPs,
		SkipUpdatingServiceInDocker: true,
	})
	assert.NoError(t, err)
	return resp.Service.Spec.TaskTemplate.ContainerSpec.Args
}

func TestApplyTrustedIPsAddsTheArgs(t *testing.T) {
	args := applyTrustedIPs(t, traefikSvcWithArgs("--entrypoints.web.address=:80"),
		[]string{"10.0.0.0/8"})

	assert.Contains(t, args, "--entrypoints.web.forwardedheaders.trustedips=10.0.0.0/8")
	assert.Contains(t, args, "--entrypoints.websecure.forwardedheaders.trustedips=10.0.0.0/8")
	assert.Contains(t, args, "--entrypoints.web.address=:80", "unrelated args must survive")
}

func TestApplyTrustedIPsReplacesTheArgs(t *testing.T) {
	args := applyTrustedIPs(t, traefikSvcWithArgs(
		"--entrypoints.web.forwardedheaders.trustedips=1.1.1.1",
		"--entrypoints.websecure.forwardedheaders.trustedips=1.1.1.1",
	), []string{"10.0.0.0/8"})

	assert.Contains(t, args, "--entrypoints.web.forwardedheaders.trustedips=10.0.0.0/8")
	assert.NotContains(t, args, "--entrypoints.web.forwardedheaders.trustedips=1.1.1.1")
}

// Withdrawing the proxy has to reach Traefik. Left in place, the entrypoint keeps
// honoring X-Forwarded-For from whoever connects - so once the proxy is gone,
// callers write their own address and every check keyed on it is answering about
// a value the caller chose.
func TestApplyTrustedIPsRemovesTheArgsWhenThereAreNone(t *testing.T) {
	args := applyTrustedIPs(t, traefikSvcWithArgs(
		"--entrypoints.web.address=:80",
		"--entrypoints.web.forwardedheaders.trustedips=1.1.1.1",
		"--entrypoints.websecure.forwardedheaders.trustedips=1.1.1.1",
	), nil)

	for _, arg := range args {
		assert.NotContains(t, arg, "forwardedheaders.trustedips")
	}
	assert.Contains(t, args, "--entrypoints.web.address=:80", "unrelated args must survive")
}

func TestApplyTrustedIPsWithNothingToRemove(t *testing.T) {
	args := applyTrustedIPs(t, traefikSvcWithArgs("--entrypoints.web.address=:80"), nil)
	assert.Equal(t, []string{"--entrypoints.web.address=:80"}, args)
}

// This is what stands between an operator and confirming a change swarm has
// already thrown away. It has to answer about the live command exactly the way
// ApplyTrustedIPsToWebEntrypoints writes it - sorted, comma-joined, on both web
// entrypoints - or a confirmation is refused for changes that are perfectly live,
// which is the failure that makes people want the mechanism turned off.
func TestArgsCarryTrustedIPs(t *testing.T) {
	const (
		web       = "--entrypoints.web.forwardedheaders.trustedips="
		websecure = "--entrypoints.websecure.forwardedheaders.trustedips="
	)
	other := []string{"traefik", "--log.level=INFO", "--entrypoints.web.address=:80"}
	with := func(vals ...string) []string { return append(append([]string(nil), other...), vals...) }

	t.Run("both entrypoints carrying the value", func(t *testing.T) {
		assert.True(t, argsCarryTrustedIPs(with(web+"10.0.0.0/8", websecure+"10.0.0.0/8"),
			[]string{"10.0.0.0/8"}))
	})

	// The writer sorts before joining, so the expectation has to as well. Without
	// it a two-entry list would compare unequal roughly half the time.
	t.Run("order of the wanted list does not matter", func(t *testing.T) {
		args := with(web+"1.1.1.1,10.0.0.0/8", websecure+"1.1.1.1,10.0.0.0/8")
		assert.True(t, argsCarryTrustedIPs(args, []string{"10.0.0.0/8", "1.1.1.1"}))
		assert.True(t, argsCarryTrustedIPs(args, []string{"1.1.1.1", "10.0.0.0/8"}))
	})

	// The rollback case: swarm restored the previous command, so traefik carries
	// the old value while the setting row carries the new one.
	t.Run("a rolled back value is not live", func(t *testing.T) {
		assert.False(t, argsCarryTrustedIPs(with(web+"192.168.0.0/16", websecure+"192.168.0.0/16"),
			[]string{"10.0.0.0/8"}))
	})

	// Half-applied is not applied. Both are written together, so one carrying the
	// new value and the other the old is a state nobody asked for.
	t.Run("one entrypoint short is not live", func(t *testing.T) {
		assert.False(t, argsCarryTrustedIPs(with(web+"10.0.0.0/8"), []string{"10.0.0.0/8"}))
		assert.False(t, argsCarryTrustedIPs(with(web+"10.0.0.0/8", websecure+"192.168.0.0/16"),
			[]string{"10.0.0.0/8"}))
	})

	// Withdrawing the proxy is applied by removing the arguments, so their absence
	// is what an empty list looks like. Reading it any other way would refuse
	// every confirmation of a change that turns trusted IPs off.
	t.Run("an empty list means the arguments are gone", func(t *testing.T) {
		assert.True(t, argsCarryTrustedIPs(other, nil))
		assert.True(t, argsCarryTrustedIPs(other, []string{}))
		assert.False(t, argsCarryTrustedIPs(with(web+"10.0.0.0/8", websecure+"10.0.0.0/8"), nil))
	})
}
