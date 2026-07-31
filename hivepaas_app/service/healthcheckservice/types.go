package healthcheckservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type HealthcheckReq struct {
	*queue.PeriodicExecData
	Healthcheck *entity.PeriodicHealthcheck
}

type HealthcheckResp struct {
}
