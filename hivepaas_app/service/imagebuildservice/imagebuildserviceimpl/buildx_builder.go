package imagebuildserviceimpl

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/services/docker"
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
			return hperrors.Wrap(err).WithMsgLog("%s", reflectutil.UnsafeBytesToStr(out))
		}
	}

	// Bootstrap builder to ensure the buildkit container is running
	bootstrapCmd := exec.CommandContext(ctx, "docker", "buildx", "inspect", "--bootstrap", builderName)
	if out, err := bootstrapCmd.CombinedOutput(); err != nil {
		return hperrors.Wrap(err).WithMsgLog("%s", reflectutil.UnsafeBytesToStr(out))
	}

	// Update container resource limits (CPUs, Memory) dynamically if settings are provided
	if res != nil && (res.CPUs > 0 || res.Mem > 0 || res.MemSwap > 0) {
		resList, err := s.dockerManager.ContainerList(ctx, func(opts *client.ContainerListOptions) {
			opts.All = true
			docker.FilterAdd(&opts.Filters, "label", fmt.Sprintf("com.docker.buildx.builder=%s", builderName))
		})
		if err == nil && resList != nil && len(resList.Items) > 0 {
			var updateRes container.Resources
			if res.CPUs > 0 {
				updateRes.NanoCPUs = int64(res.CPUs) * docker.UnitCPUNano //nolint:gosec
			}
			if res.Mem > 0 {
				updateRes.Memory = res.Mem.Bytes()
			}
			if res.MemSwap > 0 {
				updateRes.MemorySwap = res.MemSwap.Bytes()
			}
			for _, item := range resList.Items {
				_, _ = s.dockerManager.ContainerUpdate(ctx, item.ID, func(opts *client.ContainerUpdateOptions) {
					opts.Resources = &updateRes
				})
			}
		}
	}

	return nil
}
