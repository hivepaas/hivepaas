package imagebuildagentuc

import (
	"context"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

const (
	logBatchThresholdPeriod = 500 * time.Millisecond
	logBatchMaxItem         = 50
)

func (uc *UC) ImageBuild(
	ctx context.Context,
	req *imagebuildagentdto.ImageBuildReq,
) (*imagebuildagentdto.ImageBuildResp, error) {
	if req.TaskExecData == nil {
		req.TaskExecData = &queue.TaskExecData{
			Task: &entity.Task{ID: req.TaskID},
		}
	} else if req.Task == nil {
		req.Task = &entity.Task{ID: req.TaskID}
	}

	if req.App == nil && req.AppID != "" {
		app, err := uc.appService.LoadApp(ctx, uc.db, "", req.AppID, true, true,
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectRelation("ProjectEnv"),
		)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		req.App = app
	}

	var wg sync.WaitGroup
	var logChan *batchrecvchan.Chan[*tasklog.LogFrame]

	if req.SendLog != nil {
		logChan = batchrecvchan.NewChan[*tasklog.LogFrame](batchrecvchan.Options{
			ThresholdPeriod: logBatchThresholdPeriod,
			MaxItem:         logBatchMaxItem,
		})

		req.LogStore = tasklog.NewForwardStore("task:"+req.TaskID+":log",
			func(ctx context.Context, frames []*tasklog.LogFrame) error {
				logChan.Send(frames...)
				return nil
			})

		wg.Add(1)
		go func() {
			defer wg.Done()
			for frames := range logChan.Receiver() {
				if len(frames) > 0 {
					if err := req.SendLog(frames); err != nil {
						uc.logger.Warnf("Failed to send logs to caller: %v", err)
					}
				}
			}
		}()
	}

	resp, err := uc.imageBuildService.ImageBuild(ctx, uc.db, &req.ImageBuildReq)

	if logChan != nil {
		_ = logChan.Close()
		wg.Wait()
	}

	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuildagentdto.ImageBuildResp{
		ImageBuildResp: *resp,
	}, nil
}
