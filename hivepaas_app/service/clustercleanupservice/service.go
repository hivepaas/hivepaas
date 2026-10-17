package clustercleanupservice

import (
	"context"
)

type Service interface {
	Cleanup(ctx context.Context, req *ClusterCleanupReq) (*ClusterCleanupResp, error)
	// KeptImages are the images the cleanup keeps on every node, used or not:
	// those the swarm's services name, as they run and as they ran before their
	// last update, and those of this release of HivePaaS. Only a manager can
	// answer it.
	KeptImages(ctx context.Context) ([]string, error)
	// KeptArtifacts are the artifacts the cleanup keeps on every node: the files
	// the swarm's services bind, as they run and as they ran before their last
	// update. Only a manager can answer it.
	KeptArtifacts(ctx context.Context) ([]string, error)
}
