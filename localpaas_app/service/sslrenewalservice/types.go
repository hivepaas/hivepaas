package sslrenewalservice

import (
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
)

type SSLRenewalReq struct {
	*queue.TaskExecData
	CronJob *entity.Setting
}

type SSLRenewalResp struct {
	SkipResultNotification bool
}
