package sslrenewalservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type SSLRenewalReq struct {
	*queue.TaskExecData
	SchedJob        *entity.Setting
	RenewalSettings *entity.SSLRenewal
}

type SSLRenewalResp struct {
	SkipResultNotification bool
}
