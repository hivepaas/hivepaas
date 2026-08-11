package clustercleanupserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/go-units"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) cleanupBuildCache(
	ctx context.Context,
	data *clusterCleanupData,
) (err error) {
	clusterCleanup := data.CleanupSettings
	isForce := data.CleanupBuildCache == base.CleanupFlagForce
	isScheduled := data.CleanupBuildCache != base.CleanupFlagFalse && clusterCleanup.PruneBuildCache
	if !isForce && !isScheduled {
		return nil
	}

	buildCacheRetention := clusterCleanup.BuildCacheRetention.ToDuration()
	if buildCacheRetention == 0 {
		buildCacheRetention = clusterCleanup.GeneralRetention.ToDuration()
	}

	// 1. Prune Docker daemon build cache
	var dockerPruneOpts []docker.BuildCachePruneOption
	if isForce {
		dockerPruneOpts = append(dockerPruneOpts, func(o *client.BuildCachePruneOptions) {
			o.All = true
		})
	} else if buildCacheRetention > 0 {
		dockerPruneOpts = append(dockerPruneOpts, func(o *client.BuildCachePruneOptions) {
			docker.FilterAdd(&o.Filters, "until", buildCacheRetention.String())
		})
	}

	resp, e := s.dockerManager.BuildCachePrune(ctx, dockerPruneOpts...)
	if e != nil {
		data.Output.BuildCachesPruneError = e.Error()
		err = errors.Join(err, e)
	} else if resp != nil {
		report := &resp.Report
		data.Output.BuildCachesDeleted += len(report.CachesDeleted)
		data.Output.SpaceReclaimed += report.SpaceReclaimed
	}

	// 2. Prune Custom Builder
	pruneArgs := []string{"buildx", "prune", "--builder", base.HivepaasGlobalBuilder, "--force"}
	if isForce {
		pruneArgs = append(pruneArgs, "--all")
	} else if buildCacheRetention > 0 {
		hours := int(buildCacheRetention.Hours())
		pruneArgs = append(pruneArgs, "--filter", fmt.Sprintf("unused-for=%dh", hours))
	}

	pruneCmd := exec.CommandContext(ctx, "docker", pruneArgs...)
	out, pruneErr := pruneCmd.CombinedOutput()
	outStr := reflectutil.UnsafeBytesToStr(out)
	if pruneErr != nil {
		if !strings.Contains(outStr, "no builder") && !strings.Contains(outStr, "not found") {
			data.Output.BuildCachesPruneError = outStr
			err = errors.Join(err, apperrors.Wrap(pruneErr).WithMsgLog("%s", outStr))
		}
	} else {
		deletedCount, spaceReclaimed := parseBuildxPruneOutput(outStr)
		data.Output.BuildCachesDeleted += deletedCount
		data.Output.SpaceReclaimed += spaceReclaimed
	}

	return apperrors.Wrap(err)
}

func parseBuildxPruneOutput(outStr string) (int, uint64) {
	lines := strings.Split(outStr, "\n")
	deletedCount := 0
	var spaceReclaimed uint64

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "Total:") {
			totalStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "Total:"))
			if totalStr != "" {
				if b, err := units.FromHumanSize(totalStr); err == nil && b > 0 {
					spaceReclaimed = uint64(b)
				}
			}
			continue
		}

		// Skip header
		if strings.HasPrefix(trimmed, "ID") || strings.Contains(trimmed, "RECLAIMABLE") {
			continue
		}

		deletedCount++
	}

	return deletedCount, spaceReclaimed
}
