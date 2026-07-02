package healthcheckservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type HealthcheckReq struct {
	*queue.HealthcheckExecData
}

type HealthcheckResp struct {
}
