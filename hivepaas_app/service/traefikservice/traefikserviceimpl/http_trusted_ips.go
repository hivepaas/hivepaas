package traefikserviceimpl

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

const (
	defaultServiceUpdateRetryMax = 2

	// The two arguments ApplyTrustedIPsToWebEntrypoints writes and
	// WebEntrypointsCarryTrustedIPs reads back. Named once, because a reader that
	// looked at a different key from the writer would report every change as
	// missing.
	epWebTrustedIPsArg       = "entrypoints.web.forwardedheaders.trustedips"
	epWebsecureTrustedIPsArg = "entrypoints.websecure.forwardedheaders.trustedips"
)

//nolint:gocognit
func (s *service) ApplyTrustedIPsToWebEntrypoints(
	ctx context.Context,
	req *traefikservice.ApplyTrustedIPsReq,
) (resp *traefikservice.ApplyTrustedIPsResp, err error) {
	svc, err := s.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if svc == nil {
		return nil, nil
	}
	resp = &traefikservice.ApplyTrustedIPsResp{
		Service: svc,
	}

	sort.Strings(req.TrustedIPs)
	trustedIPsStr := strings.Join(req.TrustedIPs, ",")
	epWeb := epWebTrustedIPsArg
	epWebsecure := epWebsecureTrustedIPsArg

	applyFunc := func(_ int, svc *swarm.Service) (bool, error) {
		resp.Service = svc
		if svc.Spec.TaskTemplate.ContainerSpec == nil {
			svc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
		}

		hasEpWebTrustedIPs := false
		hasEpWebsecureTrustedIPs := false
		hasChanges := false

		existingArgs := svc.Spec.TaskTemplate.ContainerSpec.Args
		newArgs := make([]string, 0, len(existingArgs)+2) //nolint:mnd
		for _, arg := range existingArgs {
			key, val, valid := traefikhelper.ParseCommandArg(arg)
			if !valid {
				newArgs = append(newArgs, arg)
				continue
			}
			if key != epWeb && key != epWebsecure {
				newArgs = append(newArgs, arg) // keeps other args
				continue
			}

			if key == epWeb {
				hasEpWebTrustedIPs = true
			} else {
				hasEpWebsecureTrustedIPs = true
			}

			// No trusted IPs means the proxy declaration was withdrawn, and the trust
			// has to go with it. Left in place, Traefik keeps honoring X-Forwarded-For
			// from whoever connects - so once the proxy is gone, callers set their own
			// address, and every check keyed on it (rate limits, IP allowlists) is
			// answering about an address the caller chose.
			if trustedIPsStr == "" {
				hasChanges = true
				continue
			}
			if val != trustedIPsStr {
				newArgs = append(newArgs, fmt.Sprintf("--%s=%s", key, trustedIPsStr))
				hasChanges = true
				continue
			}
			newArgs = append(newArgs, arg)
		}

		if !hasEpWebTrustedIPs && trustedIPsStr != "" {
			newArgs = append(newArgs, fmt.Sprintf("--%s=%s", epWeb, trustedIPsStr))
			hasChanges = true
		}
		if !hasEpWebsecureTrustedIPs && trustedIPsStr != "" {
			newArgs = append(newArgs, fmt.Sprintf("--%s=%s", epWebsecure, trustedIPsStr))
			hasChanges = true
		}
		if !hasChanges {
			return false, nil
		}

		svc.Spec.TaskTemplate.ContainerSpec.Args = newArgs

		if svc.Spec.UpdateConfig == nil {
			svc.Spec.UpdateConfig = &swarm.UpdateConfig{}
		}
		svc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
		svc.Spec.UpdateConfig.MaxFailureRatio = 0.5

		return true, nil
	}

	if req.SkipUpdatingServiceInDocker {
		_, err = applyFunc(0, svc)
	} else {
		err = s.dockerManager.ServiceUpdateFunc(ctx, svc.ID, svc, applyFunc, defaultServiceUpdateRetryMax, 0)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

// WebEntrypointsCarryTrustedIPs reports whether traefik is running with exactly
// these trusted IPs on both web entrypoints.
//
// It reads the live service spec, which is the only place that answers honestly.
// The database records what HivePaaS asked for; swarm records what it settled on,
// and the two part company when an update fails its healthcheck and
// failure_action: rollback puts the previous command back. A confirmation of a
// change that was rolled back would vouch for a configuration nobody is running.
//
// Both entrypoints have to agree, because both are written together: one carrying
// the new value and the other the old is a half-applied update, which is no more
// confirmable than none of it.
func (s *service) WebEntrypointsCarryTrustedIPs(ctx context.Context, trustedIPs []string) (bool, error) {
	svc, err := s.GetTraefikSwarmService(ctx)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if svc == nil || svc.Spec.TaskTemplate.ContainerSpec == nil {
		return false, nil
	}

	return argsCarryTrustedIPs(svc.Spec.TaskTemplate.ContainerSpec.Args, trustedIPs), nil
}

// argsCarryTrustedIPs is the comparison on its own, away from docker.
func argsCarryTrustedIPs(args, trustedIPs []string) bool {
	// Sorted and joined the way the writer does it, so the comparison is against
	// the exact string that would have been written.
	wanted := append([]string(nil), trustedIPs...)
	sort.Strings(wanted)
	wantedStr := strings.Join(wanted, ",")

	found := map[string]string{}
	for _, arg := range args {
		key, val, valid := traefikhelper.ParseCommandArg(arg)
		if !valid {
			continue
		}
		if key == epWebTrustedIPsArg || key == epWebsecureTrustedIPsArg {
			found[key] = val
		}
	}

	// An empty list is applied by removing the arguments, so their absence is
	// what "no trusted IPs" looks like - not an argument with an empty value.
	if wantedStr == "" {
		return len(found) == 0
	}
	return found[epWebTrustedIPsArg] == wantedStr && found[epWebsecureTrustedIPsArg] == wantedStr
}
