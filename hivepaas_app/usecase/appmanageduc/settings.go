package appmanageduc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
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
// told, and where its restore test runs: the engine rewrites its setting and the
// jobs it made, and a job whose schedule moved is scheduled again. A move into a
// verifier app makes the app; a move back deletes it.
func (uc *UC) UpdateManagedSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.UpdateManagedSettingsReq,
) (_ *appmanageddto.UpdateManagedSettingsResp, err error) {
	// Before the transaction: the verifier's artifacts are downloaded onto its
	// node, which is not something to hold one open for.
	plan, err := uc.planVerifier(ctx, req)
	if err != nil {
		return nil, err
	}
	var made *appprovisionservice.ProvisionAppsResp
	committed := false
	defer func() {
		// The verifier's service is created inside the transaction but is not
		// part of it.
		if err != nil && !committed && made != nil && made.Cleanup != nil {
			err = errors.Join(err, made.Cleanup(context.WithoutCancel(ctx)))
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		made, txErr = uc.updateManagedSettings(ctx, db, auth, req, plan)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true
	if made != nil {
		// Once the verifier's rows are committed: the deployment reads them.
		for _, one := range made.Apps {
			if err = uc.taskQueue.ScheduleTask(ctx, one.DeploymentTask); err != nil {
				return nil, hperrors.Wrap(err)
			}
			if err = uc.taskQueue.ScheduleTask(ctx, one.CertTasks...); err != nil {
				return nil, hperrors.Wrap(err)
			}
		}
	}
	return &appmanageddto.UpdateManagedSettingsResp{}, nil
}

func (uc *UC) updateManagedSettings(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	req *appmanageddto.UpdateManagedSettingsReq,
	plan *verifierPlan,
) (made *appprovisionservice.ProvisionAppsResp, err error) {
	loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
	if err != nil {
		return nil, err
	}
	// Before anything is made for a stale request; reconfiguring checks it too.
	if req.UpdateVer != loaded.setting.UpdateVer {
		return nil, hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	move, deleting, err := uc.restoreTestMove(ctx, db, auth, loaded, req, plan, &made)
	if err != nil {
		return made, err
	}
	changed, rescheduled, err := reconfigureManaged(loaded, req, move)
	if err != nil {
		return made, err
	}
	now := timeutil.NowUTC()
	for _, setting := range append([]*entity.Setting{loaded.setting}, changed...) {
		setting.UpdatedAt = now
		if err = uc.settingRepo.Update(ctx, db, setting,
			bunex.UpdateColumns("kind", "data", "update_ver", "updated_at")); err != nil {
			return made, hperrors.Wrap(err)
		}
	}
	for _, setting := range changed {
		moved := false
		for _, one := range rescheduled {
			moved = moved || one.ID == setting.ID
		}
		if err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting, moved); err != nil {
			return made, hperrors.Wrap(err)
		}
	}
	// Last: a deletion reaches past the database, and nothing after it may fail
	// the transaction it cannot be rolled back with.
	if deleting != "" {
		if err = uc.deleteVerifier(ctx, db, loaded.app, deleting); err != nil {
			return made, err
		}
	}
	detail := auditdetail.New().Set("changedJobs", len(changed))
	if move != nil {
		detail = detail.Set("restoreTest", req.RestoreTest.Mode)
	}
	return made, hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: loaded.app.ID,
		Source:   base.AuditLogSourceAPIUpdate,
		Section:  "managed",
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    loaded.app.ID,
		ResName:  loaded.app.Name,
		Detail:   detail.String(),
	}))
}

// restoreTestMove is the move of the restore test a request asks for, nil for
// none: into the verifier it makes - returned through made, for the caller to
// clean up - or back into the container, which returns the verifier to delete
// once the move is stored.
func (uc *UC) restoreTestMove(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	loaded *loadedManaged,
	req *appmanageddto.UpdateManagedSettingsReq,
	plan *verifierPlan,
	made **appprovisionservice.ProvisionAppsResp,
) (*managedmodel.RestoreTestChange, string, error) {
	pg := loaded.managed.Postgres
	if req.RestoreTest == nil || pg == nil || req.RestoreTest.Mode == pg.RestoreTest.Mode && plan == nil {
		return nil, "", nil
	}
	if plan == nil {
		return &managedmodel.RestoreTestChange{}, pg.RestoreTest.VerifierApp.ID, nil
	}
	// Checked again under the lock: two moves at once make one verifier.
	if pg.RestoreTest.Mode == entity.ManagedRestoreTestVerifier {
		return nil, "", hperrors.Wrap(hperrors.ErrAppManagedVerifierExists)
	}
	var err error
	if *made, err = uc.provisionVerifier(ctx, db, auth, loaded, plan); err != nil {
		return nil, "", err
	}
	return &managedmodel.RestoreTestChange{Verifier: (*made).Apps[0].App, NodeID: plan.nodeID,
		KeyAuth: plan.keyAuth}, "", nil
}

// reconfigureManaged has the engine apply a change to the app's setting, which it
// stores back, and to its jobs, which it returns as rewritten and as moved.
func reconfigureManaged(
	loaded *loadedManaged,
	req *appmanageddto.UpdateManagedSettingsReq,
	move *managedmodel.RestoreTestChange,
) (changed, rescheduled []*entity.Setting, err error) {
	if req.UpdateVer != loaded.setting.UpdateVer {
		return nil, nil, hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	jobs := make(map[string]*entity.Setting, len(loaded.jobs))
	for _, job := range loaded.jobs {
		jobs[job.ID] = job
	}
	resp, err := loaded.engine.Reconfigure(&managedmodel.ReconfigureReq{
		App:     loaded.app,
		Managed: loaded.managed,
		Jobs:    jobs,
		Change: &managedmodel.SettingsChange{
			RetainFull:          req.RetainFull,
			BaseBackupSchedule:  req.BaseBackupSchedule,
			RestoreTestSchedule: req.RestoreTestSchedule,
			RestoreTestQuery:    req.RestoreTestQuery,
			RestoreTest:         move,
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
