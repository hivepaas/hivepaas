package imagebuildserviceimpl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

func (s *service) ensureCustomBuilder(
	ctx context.Context,
	builderName string,
	res *entity.ImageBuildResourceSettings,
) error {
	// Check if builder already exists
	inspectCmd := exec.CommandContext(ctx, "docker", "buildx", "inspect", builderName)
	if err := inspectCmd.Run(); err != nil {
		// Builder does not exist, create it with docker-container driver
		createCmd := exec.CommandContext(ctx, "docker", "buildx", "create",
			"--name", builderName,
			"--driver", "docker-container",
		)
		if out, err := createCmd.CombinedOutput(); err != nil {
			return apperrors.Wrap(err).WithMsgLog("%s", reflectutil.UnsafeBytesToStr(out))
		}
	}

	// Bootstrap builder to ensure the buildkit container is running
	bootstrapCmd := exec.CommandContext(ctx, "docker", "buildx", "inspect", "--bootstrap", builderName)
	if out, err := bootstrapCmd.CombinedOutput(); err != nil {
		return apperrors.Wrap(err).WithMsgLog("%s", reflectutil.UnsafeBytesToStr(out))
	}

	// Update container resource limits (CPUs, Memory) dynamically if settings are provided
	if res != nil && (res.CPUs > 0 || res.Mem > 0 || res.MemSwap > 0) { //nolint:nestif
		//nolint:gosec
		out, err := exec.CommandContext(ctx, "docker", "ps", "-a",
			"--filter", fmt.Sprintf("label=com.docker.buildx.builder=%s", builderName),
			"--format", "{{.ID}}",
		).Output()
		if err == nil {
			cids := strings.Fields(reflectutil.UnsafeBytesToStr(out))
			if len(cids) > 0 {
				updateArgs := []string{"update"}
				if res.CPUs > 0 {
					updateArgs = append(updateArgs, "--cpus", fmt.Sprintf("%d", res.CPUs))
				}
				if res.Mem > 0 {
					updateArgs = append(updateArgs, "--memory", fmt.Sprintf("%d", res.Mem.Bytes()))
				}
				if res.MemSwap > 0 {
					updateArgs = append(updateArgs, "--memory-swap", fmt.Sprintf("%d", res.MemSwap.Bytes()))
				}
				updateArgs = append(updateArgs, cids...)
				_ = exec.CommandContext(ctx, "docker", updateArgs...).Run()
			}
		}
	}

	return nil
}
