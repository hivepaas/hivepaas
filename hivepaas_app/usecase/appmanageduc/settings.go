package appmanageduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

func (uc *UC) GetManagedSettings(
	ctx context.Context,
	_ *basedto.Auth,
	req *appmanageddto.GetManagedSettingsReq,
) (*appmanageddto.GetManagedSettingsResp, error) {
	loaded, err := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	return &appmanageddto.GetManagedSettingsResp{Data: appmanageddto.TransformManagedSettings(loaded.view())}, nil
}

// UpdateManagedSettings changes when a managed app's jobs run and what they are
// told: the engine rewrites its setting and the jobs it made, and a job whose
// schedule moved is scheduled again.
func (uc *UC) UpdateManagedSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.UpdateManagedSettingsReq,
) (*appmanageddto.UpdateManagedSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
		if err != nil {
			return err
		}
		changed, rescheduled, err := reconfigureManaged(loaded, req)
		if err != nil {
			return err
		}
		now := timeutil.NowUTC()
		for _, setting := range append([]*entity.Setting{loaded.setting}, changed...) {
			setting.UpdatedAt = now
			if err = uc.settingRepo.Update(ctx, db, setting,
				bunex.UpdateColumns("data", "update_ver", "updated_at")); err != nil {
				return hperrors.Wrap(err)
			}
		}
		for _, setting := range changed {
			moved := false
			for _, one := range rescheduled {
				moved = moved || one.ID == setting.ID
			}
			if err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting, moved); err != nil {
				return hperrors.Wrap(err)
			}
		}
		return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
			Type:     base.AuditLogTypeAppUpdate,
			Scope:    base.ObjectScopeApp,
			ObjectID: loaded.app.ID,
			Source:   base.AuditLogSourceAPIUpdate,
			Section:  "managed",
			Auth:     auth,
			ResType:  base.ResourceTypeApp,
			ResID:    loaded.app.ID,
			ResName:  loaded.app.Name,
			Detail:   auditdetail.New().Set("changedJobs", len(changed)).String(),
		}))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmanageddto.UpdateManagedSettingsResp{}, nil
}

// reconfigureManaged has the engine apply a change to the app's setting, which it
// stores back, and to its jobs, which it returns as rewritten and as moved.
func reconfigureManaged(
	loaded *loadedManaged,
	req *appmanageddto.UpdateManagedSettingsReq,
) (changed, rescheduled []*entity.Setting, err error) {
	if req.UpdateVer != loaded.setting.UpdateVer {
		return nil, nil, hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	jobs := make(map[string]*entity.Setting, len(loaded.jobs))
	for _, job := range loaded.jobs {
		jobs[job.ID] = job
	}
	resp, err := loaded.engine.Reconfigure(&managedmodel.ReconfigureReq{
		Managed: loaded.managed,
		Jobs:    jobs,
		Change: &managedmodel.SettingsChange{
			RetainFull:          req.RetainFull,
			BaseBackupSchedule:  req.BaseBackupSchedule,
			RestoreTestSchedule: req.RestoreTestSchedule,
			RestoreTestQuery:    req.RestoreTestQuery,
		},
	})
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if err = loaded.setting.SetData(loaded.managed); err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	loaded.setting.UpdateVer++
	return resp.Changed, resp.Rescheduled, nil
}

// view is what the settings are read from.
func (l *loadedManaged) view() *appmanageddto.ManagedSettingsView {
	view := &appmanageddto.ManagedSettingsView{Setting: l.setting, Managed: l.managed,
		Jobs: make(map[string]*entity.SchedJob, len(l.jobs))}
	for _, setting := range l.jobs {
		if job, err := setting.AsSchedJob(); err == nil {
			view.Jobs[setting.ID] = job
		}
	}
	return view
}
