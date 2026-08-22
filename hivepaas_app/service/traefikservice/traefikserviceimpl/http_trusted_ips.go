package traefikserviceimpl

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/traefikhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

const (
	defaultServiceUpdateRetryMax = 2
)

//nolint:gocognit
func (s *service) ApplyTrustedIPsToWebEntrypoints(
	ctx context.Context,
	req *traefikservice.ApplyTrustedIPsReq,
) (resp *traefikservice.ApplyTrustedIPsResp, err error) {
	svc, err := s.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if svc == nil {
		return nil, nil
	}
	resp = &traefikservice.ApplyTrustedIPsResp{
		Service: svc,
	}

	sort.Strings(req.TrustedIPs)
	trustedIPsStr := strings.Join(req.TrustedIPs, ",")
	epWeb := "entrypoints.web.forwardedheaders.trustedips"
	epWebsecure := "entrypoints.websecure.forwardedheaders.trustedips"

	err = s.dockerManager.ServiceUpdateFunc(ctx, svc.ID, svc,
		func(_ int, svc *swarm.Service) (bool, error) {
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
				if key == epWeb {
					hasEpWebTrustedIPs = true
					if val != trustedIPsStr && trustedIPsStr != "" {
						newArgs = append(newArgs, fmt.Sprintf("--%s=%s", epWeb, trustedIPsStr))
						hasChanges = true
						continue
					}
				}
				if key == epWebsecure {
					hasEpWebsecureTrustedIPs = true
					if val != trustedIPsStr && trustedIPsStr != "" {
						newArgs = append(newArgs, fmt.Sprintf("--%s=%s", epWebsecure, trustedIPsStr))
						hasChanges = true
						continue
					}
				}
				newArgs = append(newArgs, arg) // keeps other args
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
		}, defaultServiceUpdateRetryMax, 0)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
