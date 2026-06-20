package taskservice

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
)

type GetTaskReq struct {
	ID       string
	Type     base.TaskType
	TargetID string

	ExtraSelectOpts []bunex.SelectQueryOption
	SkipQueryCache  bool
}

type GetTaskResp struct {
	Task     *entity.Task
	TaskInfo *cacheentity.TaskInfo
}

type ListTaskReq struct {
	Scope     *base.ObjectScope
	TargetIDs []string
	Statuses  []base.TaskStatus
	Search    string
	Paging    basedto.Paging

	ExtraSelectOpts []bunex.SelectQueryOption
	SkipQueryCache  bool
}

type ListTaskResp struct {
	PagingMeta  *basedto.PagingMeta
	Tasks       []*entity.Task
	TaskInfoMap map[string]*cacheentity.TaskInfo
}

type GetTaskLogsReq struct {
	TaskID   string
	Follow   bool
	Since    time.Time
	Duration time.Duration
	Tail     int

	LogBatchThresholdPeriod time.Duration
	LogBatchMaxFrame        int
	LogSessionTimeout       time.Duration
}

type GetTaskLogsResp struct {
	StaticLogs       []*tasklog.LogFrame
	LogsStream       <-chan []*tasklog.LogFrame
	LogsStreamCloser func() error
}
