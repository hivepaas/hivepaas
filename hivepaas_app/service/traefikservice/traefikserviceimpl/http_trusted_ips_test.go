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
