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
}
