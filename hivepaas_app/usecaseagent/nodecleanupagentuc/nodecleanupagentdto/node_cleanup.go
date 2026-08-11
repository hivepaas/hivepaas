package nodecleanupagentdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
)

type NodeCleanupReq struct {
	TaskID string
	clustercleanupservice.ClusterCleanupReq
}

type NodeCleanupResp struct {
	entity.ClusterNodeCleanupOutput
}
