package registryauthserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) RecordRenewal(ctx context.Context, db database.IDB, authIDs []string) (*entity.Task, error) {
	job, err := s.activeRenewalJob(ctx, db)
	if err != nil || job == nil {
		return nil, err
	}
	now := s.now()
	// A task's run time is unique for its job: one a scheduled run already has
	// is moved by a second.
	task, err := s.schedJobService.CreateSchedJobTask(job, now.Add(time.Second), now)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(authIDs) > 0 {
		targets := make(entity.ObjectIDSlice, 0, len(authIDs))
		for _, id := range authIDs {
			targets = append(targets, &entity.ObjectID{ID: id})
		}
		task.MustSetArgs(&entity.TaskRegistryAuthRenewalArgs{TargetAuths: targets})
	}
	if err = s.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

func (s *service) RecordRenewalForKeyAuth(
	ctx context.Context,
	db database.IDB,
	keyAuthID string,
) (*entity.Task, error) {
	auths, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeRegistryAuth),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.data->>'kind' = ?", base.RegistryAuthKindAWSECR),
		bunex.SelectWhere("setting.data->'ecr'->'keyAuth'->>'id' = ?", keyAuthID),
		bunex.SelectColumns("id"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(auths) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(auths))
	for _, auth := range auths {
		ids = append(ids, auth.ID)
	}
	return s.RecordRenewal(ctx, db, ids)
}

// activeRenewalJob is the renewal's job when it and its setting are on; nil
// when either is off or not made yet.
func (s *service) activeRenewalJob(ctx context.Context, db database.IDB) (*entity.Setting, error) {
	scope := entity.NewObjectScopeGlobal()
	renewal, err := s.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeRegistryAuthRenewal, true)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	job, err := s.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeSchedJob, true,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeRegistryAuthRenewal),
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", renewal.ID),
	)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return job, nil
}
