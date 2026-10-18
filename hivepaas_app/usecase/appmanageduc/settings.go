package appmanageduc

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
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
	if err = uc.checkDumpRepo(ctx, db, loaded.app, req.DumpRepo); err != nil {
		return nil, err
	}
	move, deleting, err := uc.restoreTestMove(ctx, db, auth, loaded, req, plan, &made)
	if err != nil {
		return made, err
	}
	now := timeutil.NowUTC()
	reconfigured, err := reconfigureManaged(loaded, req, move, now)
	if err != nil {
		return made, err
	}
	if err = uc.storeReconfigured(ctx, db, loaded.setting, reconfigured, now); err != nil {
		return made, err
	}
	// Last: a deletion reaches past the database, and nothing after it may fail
	// the transaction it cannot be rolled back with.
	if deleting != "" {
		if err = uc.deleteVerifier(ctx, db, loaded.app, deleting); err != nil {
			return made, err
		}
	}
	detail := auditdetail.New().Set("changedJobs", len(reconfigured.Changed)+len(reconfigured.Created)+
		len(reconfigured.Deleted))
	if req.DumpRepo != nil {
		detail = detail.Set("dumpRepo", req.DumpRepo.ID)
	}
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
// stores back, and to its jobs, which it returns as rewritten, moved, made and
// removed.
func reconfigureManaged(
	loaded *loadedManaged,
	req *appmanageddto.UpdateManagedSettingsReq,
	move *managedmodel.RestoreTestChange,
	now time.Time,
) (*managedmodel.ReconfigureResp, error) {
	if req.UpdateVer != loaded.setting.UpdateVer {
		return nil, hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	jobs := make(map[string]*entity.Setting, len(loaded.jobs))
	for _, job := range loaded.jobs {
		jobs[job.ID] = job
	}
	var dumpRepo *string
	if req.DumpRepo != nil {
		dumpRepo = &req.DumpRepo.ID
	}
	resp, err := loaded.engine.Reconfigure(&managedmodel.ReconfigureReq{
		App:       loaded.app,
		ManagedID: loaded.setting.ID,
		Managed:   loaded.managed,
		Jobs:      jobs,
		Change: &managedmodel.SettingsChange{
			RetainFull:          req.RetainFull,
			BaseBackupSchedule:  req.BaseBackupSchedule,
			RestoreTestSchedule: req.RestoreTestSchedule,
			RestoreTestQuery:    req.RestoreTestQuery,
			RestoreTest:         move,
			DumpRepo:            dumpRepo,
			DumpSchedule:        req.DumpSchedule,
		},
		Now: now,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = loaded.setting.SetData(loaded.managed); err != nil {
		return nil, hperrors.Wrap(err)
	}
	loaded.setting.UpdateVer++
	return resp, nil
}

// storeReconfigured stores the engine's setting and its jobs as the change left
// them, and schedules them again: a job rewritten keeps its coming runs unless
// its schedule moved, one made gets its first, and one removed loses them, as
// when a person deletes a job.
func (uc *UC) storeReconfigured(
	ctx context.Context,
	db database.Tx,
	managed *entity.Setting,
	resp *managedmodel.ReconfigureResp,
	now time.Time,
) error {
	for _, setting := range append([]*entity.Setting{managed}, resp.Changed...) {
		setting.UpdatedAt = now
		if err := uc.settingRepo.Update(ctx, db, setting,
			bunex.UpdateColumns("kind", "data", "update_ver", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
	}
	for _, setting := range resp.Changed {
		if err := uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting,
			slices.Contains(resp.Rescheduled, setting)); err != nil {
			return hperrors.Wrap(err)
		}
	}
	for _, setting := range resp.Created {
		if err := uc.settingRepo.Insert(ctx, db, setting); err != nil {
			return hperrors.Wrap(err)
		}
		if err := uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting, false); err != nil {
			return hperrors.Wrap(err)
		}
	}
	for _, setting := range resp.Deleted {
		setting.DeletedAt = now
		setting.UpdateVer++
		if err := uc.settingRepo.Update(ctx, db, setting, bunex.UpdateColumns("deleted_at", "update_ver")); err != nil {
			return hperrors.Wrap(err)
		}
		if err := uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting, true); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// checkDumpRepo refuses a repository to dump into that is not a backup
// repository the app's environment can use, or that is not active.
func (uc *UC) checkDumpRepo(ctx context.Context, db database.IDB, app *entity.App, repo *basedto.ObjectIDReq) error {
	if repo == nil || repo.ID == "" {
		return nil
	}
	_, envKey := projecthelper.ParseProjectEnvID(app.ProjectEnvID)
	setting, err := uc.settingRepo.GetByID(ctx, db, entity.NewObjectScopeProjectEnv(app.ProjectID, envKey),
		base.SettingTypeBackupRepo, repo.ID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !setting.IsActive() {
		return hperrors.NewArgumentInvalid("dumpRepo").WithExtraDetail("the backup repository is not active")
	}
	return nil
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
