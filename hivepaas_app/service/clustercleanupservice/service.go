package clustercleanupservice

import (
	"context"
)

type Service interface {
	Cleanup(ctx context.Context, req *ClusterCleanupReq) (*ClusterCleanupResp, error)
}
