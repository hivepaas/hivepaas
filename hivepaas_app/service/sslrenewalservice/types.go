package sslrenewalservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type SSLRenewalReq struct {
	*queue.TaskExecData
	RenewalJobSetting *entity.Setting
	RenewalSettings   *entity.SSLRenewal
}

type SSLRenewalResp struct {
	SkipResultNotification bool
}
