package sysupdateservice

import (
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type SysUpdateReq struct {
	*queue.TaskExecData
}

type SysUpdateResp struct {
}
