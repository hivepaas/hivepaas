package sysupdateservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SysUpdateReq struct {
	*queue.TaskExecData
}

type SysUpdateResp struct {
}
