package imagescanserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

const (
	// A scan that fails is tried once more, half a minute later: the runner of
	// a node that just joined, a server fetching its DB, are both a matter of
	// seconds.
	scanTaskMaxRetry   = 1
	scanTaskRetryDelay = timeutil.Duration(30 * time.Second)
)

func (s *service) RecordScan(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.RecordScanReq,
) (*imagescanservice.RecordScanResp, error) {
	if _, err := s.enabled(ctx, db); err != nil {
		return nil, err
	}
	pending, _, err := s.taskRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeImageScan),
		bunex.SelectWhere("task.object_id = ?", req.App.ID),
		bunex.SelectWhere("task.status = ?", base.TaskStatusNotStarted),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(pending) > 0 {
		return &imagescanservice.RecordScanResp{Task: pending[0], Pending: true}, nil
	}

	timeNow := timeutil.NowUTC()
	task := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeApp,
		ObjectID: req.App.ID,
		Type:     base.TaskTypeImageScan,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityDefault,
			MaxRetry:   scanTaskMaxRetry,
			RetryDelay: scanTaskRetryDelay,
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     timeNow,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	if err = task.SetArgs(&entity.TaskImageScanArgs{AppID: req.App.ID, Trigger: req.Trigger,
		Force: req.Force}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = s.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &imagescanservice.RecordScanResp{Task: task}, nil
}
