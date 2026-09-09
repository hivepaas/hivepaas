package taskservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type GetTaskReq struct {
	Scope    *entity.ObjectScope
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
	Scope     *entity.ObjectScope
	Types     []base.TaskType
	TargetIDs []string
	Statuses  []base.TaskStatus
	FromDate  timeutil.Date
	ToDate    timeutil.Date
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
	Scope    *entity.ObjectScope
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
