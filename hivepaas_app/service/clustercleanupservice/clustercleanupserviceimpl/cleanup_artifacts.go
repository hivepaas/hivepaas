package clustercleanupserviceimpl

import (
	"context"
	"path/filepath"
	"slices"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) KeptArtifacts(ctx context.Context) ([]string, error) {
	resp, err := s.dockerManager.ServiceList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	var kept []string
	for i := range resp.Items {
		svc := &resp.Items[i]
		for _, spec := range []*swarm.ServiceSpec{&svc.Spec, svc.PreviousSpec} {
			if spec == nil || spec.TaskTemplate.ContainerSpec == nil {
				continue
			}
			for j := range spec.TaskTemplate.ContainerSpec.Mounts {
				if m := &spec.TaskTemplate.ContainerSpec.Mounts[j]; artifactservice.IsArtifactMount(m) {
					kept = append(kept, filepath.Base(m.Source))
				}
			}
		}
	}
	slices.Sort(kept)
	return slices.Compact(kept), nil
}

// cleanupArtifacts removes the artifacts of this node that no service binds
// and nothing has used for the retention. They go with images - files a node
// keeps for what it runs - and only when the manager said which are used.
func cleanupArtifacts(data *clusterCleanupData) error {
	cacheDir := artifactservice.CacheHostDir()
	if cacheDir == "" {
		return nil
	}
	return cleanupArtifactsIn(data, volumeservice.HostPathPrefix+cacheDir, time.Now())
}

func cleanupArtifactsIn(data *clusterCleanupData, dir string, now time.Time) error {
	if !data.ArtifactsKnown || data.CleanupImages == base.CleanupFlagFalse || !data.CleanupSettings.PruneImages {
		return nil
	}
	retention := data.CleanupSettings.GeneralRetention.ToDuration()
	deleted, reclaimed, err := artifactservice.PruneCache(dir, data.KeepArtifacts, now.Add(-retention))
	data.Output.ArtifactsDeleted = deleted
	data.Output.SpaceReclaimed += reclaimed
	if err != nil {
		data.Output.ArtifactsPruneError = err.Error()
		return hperrors.Wrap(err)
	}
	return nil
}
