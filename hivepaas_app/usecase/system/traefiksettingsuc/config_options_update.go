package traefiksettingsuc

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/traefikhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

func (uc *UC) UpdateConfigOptions(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.UpdateConfigOptionsReq,
) (*traefiksettingsdto.UpdateConfigOptionsResp, error) {
	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	newArgs := make([]string, 0)

	// 1. Preserve binary executable name ("traefik") and any non-settable system args
	if traefikSvc.Spec.TaskTemplate.ContainerSpec != nil {
		for _, arg := range traefikSvc.Spec.TaskTemplate.ContainerSpec.Args {
			arg = strings.TrimSpace(arg)
			key, _, valid := traefikhelper.ParseCommandArg(arg)
			if key == "traefik" || (valid && !base.IsTraefikCmdArgSettable(key)) {
				newArgs = append(newArgs, arg)
			}
		}
	}

	// 2. Append updated user-configured settable args
	for _, kv := range req.ParsedCommandArgs {
		switch {
		case len(kv) >= 3 && kv[2] != "":
			newArgs = append(newArgs, kv[2])
		case len(kv) >= 2 && kv[1] != "":
			newArgs = append(newArgs, fmt.Sprintf("--%s=%s", kv[0], kv[1]))
		case len(kv) >= 1 && kv[0] != "":
			newArgs = append(newArgs, fmt.Sprintf("--%s", kv[0]))
		}
	}

	// 3. Update container spec & rollback policy
	if traefikSvc.Spec.TaskTemplate.ContainerSpec == nil {
		traefikSvc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	traefikSvc.Spec.TaskTemplate.ContainerSpec.Args = newArgs

	if traefikSvc.Spec.UpdateConfig == nil {
		traefikSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	traefikSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	traefikSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

	// 4. Update Traefik Swarm Service via dockerManager
	_, err = uc.dockerManager.ServiceUpdate(ctx, traefikSvc.ID, &traefikSvc.Version, &traefikSvc.Spec)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &traefiksettingsdto.UpdateConfigOptionsResp{}, nil
}
