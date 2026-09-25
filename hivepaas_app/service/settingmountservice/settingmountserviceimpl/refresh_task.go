package settingmountserviceimpl

import (
	"context"
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

const (
	refreshMaxRetry   = 3
	refreshRetryDelay = timeutil.Duration(time.Minute)
)

func (s *service) RecordRefresh(
	ctx context.Context, db database.IDB, settings ...*entity.Setting,
) (*entity.Task, error) {
	var appIDs, sourceIDs []string
	for _, setting := range settings {
		switch {
		case setting == nil:
		case setting.Type == base.SettingTypeAppSettingMount:
			appIDs = append(appIDs, setting.ObjectID)
		case settingmountservice.IsSourceType(setting.Type):
			sourceIDs = append(sourceIDs, setting.ID)
		}
	}
	if len(sourceIDs) > 0 {
		readers, err := s.loadReaders(ctx, db, sourceIDs)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		appIDs = append(appIDs, readers...)
	}
	appIDs = gofn.ToSet(gofn.ToSliceSkippingZero(appIDs...))
	if len(appIDs) == 0 {
		return nil, nil
	}
	slices.Sort(appIDs)

	now := timeutil.NowUTC()
	task := &entity.Task{
		ID:     gofn.Must(ulid.NewStringULID()),
		Scope:  base.ObjectScopeGlobal,
		Type:   base.TaskTypeSettingMountRefresh,
		Status: base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityDefault,
			MaxRetry:   refreshMaxRetry,
			RetryDelay: refreshRetryDelay,
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := task.SetArgs(&entity.TaskSettingMountRefreshArgs{AppIDs: appIDs}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := s.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

// loadReadersFromRepo is the apps whose entries mount one of sourceIDs, through
// the links every setting reference writes.
func (s *service) loadReadersFromRepo(ctx context.Context, db database.IDB, sourceIDs []string) ([]string, error) {
	links, _, err := s.resLinkRepo.List(ctx, db, nil,
		bunex.SelectWhere("res_link.src_type = ?", base.ResourceTypeSetting),
		bunex.SelectWhere("res_link.dst_type = ?", base.ResourceTypeSetting),
		bunex.SelectWhere("res_link.dst_id IN (?)", bunex.List(sourceIDs)),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(links) == 0 {
		return nil, nil
	}
	entryIDs := make([]string, 0, len(links))
	for _, link := range links {
		entryIDs = append(entryIDs, link.SrcID)
	}
	entries, err := s.settingRepo.ListByIDs(ctx, db, nil, entryIDs, false,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appIDs := make([]string, 0, len(entries))
	for _, e := range entries {
		appIDs = append(appIDs, e.ObjectID)
	}
	return appIDs, nil
}
