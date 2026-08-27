package traefiksettingsuc

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

func (uc *UC) UpdateConfigOptions(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.UpdateConfigOptionsReq,
) (*traefiksettingsdto.UpdateConfigOptionsResp, error) {
	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	newCommandArgs := uc.buildStartupCommand(req, traefikSvc)

	// 3. Update container spec & rollback policy
	if traefikSvc.Spec.TaskTemplate.ContainerSpec == nil {
		traefikSvc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	traefikSvc.Spec.TaskTemplate.ContainerSpec.Args = newCommandArgs

	if traefikSvc.Spec.UpdateConfig == nil {
		traefikSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	traefikSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	traefikSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

	// 4. Update Traefik Swarm Service via dockerManager
	_, err = uc.dockerManager.ServiceUpdate(ctx, traefikSvc.ID, &traefikSvc.Version, &traefikSvc.Spec)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &traefiksettingsdto.UpdateConfigOptionsResp{}, nil
}

//nolint:gocognit
func (uc *UC) buildStartupCommand(
	req *traefiksettingsdto.UpdateConfigOptionsReq,
	traefikSvc *swarm.Service,
) []string {
	newArgs := make([]string, 0, 20) //nolint:mnd

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

	// 2. Append updated user-configured settings and args from StartupCommand
	if req.StartupCommand != nil {
		cmd := req.StartupCommand
		if cmd.LogLevel != "" {
			newArgs = append(newArgs, "--log=true")
			if strings.ToLower(cmd.LogLevel) != "default" {
				newArgs = append(newArgs, fmt.Sprintf("--log.level=%s", cmd.LogLevel))
			}
		}
		if cmd.AccessLog {
			newArgs = append(newArgs, "--accesslog=true")
		}
		if cmd.HTTP3 {
			newArgs = append(newArgs, "--entrypoints.websecure.http3=true")
		}
		if cmd.FastProxy {
			newArgs = append(newArgs, "--experimental.fastproxy=true")
		}

		for _, kv := range cmd.ParsedArgs {
			switch kv[0] {
			case "log", "log.level", "accesslog", "entrypoints.websecure.http3", "experimental.fastproxy":
				continue
			}
			switch {
			case len(kv) >= 3 && kv[2] != "":
				newArgs = append(newArgs, kv[2])
			case len(kv) >= 2 && kv[1] != "":
				newArgs = append(newArgs, fmt.Sprintf("--%s=%s", kv[0], kv[1]))
			case len(kv) >= 1 && kv[0] != "":
				newArgs = append(newArgs, fmt.Sprintf("--%s", kv[0]))
			}
		}
	}

	return newArgs
}
