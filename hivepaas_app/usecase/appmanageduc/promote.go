package appmanageduc

import (
	"context"
	"fmt"
	"slices"
	"time"

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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// fenceWait is how long a fenced primary's tasks are given to stop: a task on a
// node that does not answer never is seen stopped, and nothing is promoted then.
const fenceWait = 90 * time.Second

// promotion is a promotion's parts, read before anything stops.
type promotion struct {
	loaded *loadedManaged
	spec   *managedmodel.PromoteSpec
	member *entity.ManagedMember
	// promoted is the member's app; fenced is what writes now, which is fenced.
	promoted *entity.App
	fenced   *entity.App
}

// PromoteManagedMember promotes a replica to the cluster's primary (design §9,
// "Manual promotion"), the steps run by runPromotion. The cluster's app stays
// the cluster's: its setting, its jobs, and the addresses its consumers name.
func (uc *UC) PromoteManagedMember(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.PromoteManagedMemberReq,
) (*appmanageddto.PromoteManagedMemberResp, error) {
	err := uc.licenseService.CheckFeatures(ctx, []string{string(base.LicenseFeatureManagedDatabases)})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	p, err := uc.planPromotion(ctx, req)
	if err != nil {
		return nil, err
	}
	steps := &promotionRun{uc: uc, auth: auth, req: req, p: p, meta: &basedto.Meta{}}
	if err = runPromotion(ctx, steps, p.promoted.Name, p.fenced.Name); err != nil {
		return nil, err
	}
	data := &appmanageddto.PromoteManagedMemberDataResp{
		Writer:    &basedto.ObjectIDResp{ID: p.promoted.ID},
		Fenced:    &basedto.ObjectIDResp{ID: p.fenced.ID},
		Restarted: []*basedto.ObjectIDResp{},
	}
	for _, app := range steps.restarted {
		data.Restarted = append(data.Restarted, &basedto.ObjectIDResp{ID: app.ID})
	}
	return &appmanageddto.PromoteManagedMemberResp{Meta: steps.meta, Data: data}, nil
}

// promotionSteps are a promotion's steps, which runPromotion runs in order.
type promotionSteps interface {
	// fence keeps what writes now from starting its server, in its settings.
	fence(ctx context.Context) error
	// stop stops it, and says whether it was seen stopped in time.
	stop(ctx context.Context) (bool, error)
	// applyFence puts the fence on its stopped service.
	applyFence(ctx context.Context)
	// promote promotes the member: its exit code, nil when it did not run.
	promote(ctx context.Context) (*int, error)
	// allow lets the replicas into the promoted member.
	allow(ctx context.Context)
	// unfence lifts the fence, and starts what was fenced again.
	unfence(ctx context.Context) error
	// record makes the member the cluster's primary in the settings, and points
	// the other replicas at it.
	record(ctx context.Context) error
	// restart restarts the other replicas, to follow it.
	restart(ctx context.Context)
}

// runPromotion runs a promotion's steps in order: what writes now is fenced and
// seen stopped; the member is promoted, and lets the replicas in; it is recorded
// as the primary; the other replicas restart to follow it. A step that fails stops the promotion there.
// Before the member is promoted - what was fenced not seen stopped, or the
// member certainly not promoted - the fence is lifted. Once whether the member
// was promoted is not known, or once it was, the fence stays: two primaries
// are never risked, and promoting again finishes the promotion.
func runPromotion(ctx context.Context, steps promotionSteps, member, fenced string) error {
	if err := steps.fence(ctx); err != nil {
		return err
	}
	stopped, err := steps.stop(ctx)
	if err != nil || !stopped {
		reason := fenced + " could not be seen stopped, so nothing was promoted"
		if err != nil {
			reason += ": " + hperrors.GetErrorDetail(err, "")
		}
		return unfenced(ctx, steps, reason, fenced)
	}
	steps.applyFence(ctx)
	code, err := steps.promote(ctx)
	switch {
	case err == nil && code != nil && *code == 0:
	case code == nil || *code == managedmodel.PromoteExitNotPromoted:
		reason := member + " was not promoted"
		if err != nil {
			reason += ": " + hperrors.GetErrorDetail(err, "")
		}
		return unfenced(ctx, steps, reason, fenced)
	default:
		return promoteFailed(fmt.Sprintf("whether %s was promoted is not known (it exited with %d); %s stays "+
			"fenced, and promoting %s again finishes the promotion", member, *code, fenced, member))
	}
	steps.allow(ctx)
	if err = steps.record(ctx); err != nil {
		return promoteFailed(fmt.Sprintf("%s is the primary, but recording it failed: %s; %s stays fenced, and "+
			"promoting %s again finishes the promotion", member, hperrors.GetErrorDetail(err, ""), fenced, member))
	}
	steps.restart(ctx)
	return nil
}

// unfenced lifts the fence of a promotion that did not happen, and fails with
// the reason.
func unfenced(ctx context.Context, steps promotionSteps, reason, fenced string) error {
	if err := steps.unfence(ctx); err != nil {
		return promoteFailed(fmt.Sprintf("%s; lifting %s's fence failed too: %s - start it again from its page "+
			"once its HIVEPAAS_FENCED variable is gone", reason, fenced, hperrors.GetErrorDetail(err, "")))
	}
	return promoteFailed(reason + "; " + fenced + " runs again")
}

func promoteFailed(reason string) error {
	return hperrors.Wrap(hperrors.ErrAppManagedPromoteFailed).WithParam("Reason", reason)
}

func (uc *UC) planPromotion(ctx context.Context, req *appmanageddto.PromoteManagedMemberReq) (*promotion, error) {
	loaded, err := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	p := &promotion{loaded: loaded}
	if pg := loaded.managed.Postgres; pg != nil {
		p.member = pg.Member(req.Member.ID)
	}
	paths, err := uc.mountPaths(ctx, uc.db, loaded.app.ID)
	if err != nil {
		return nil, err
	}
	p.spec, err = loaded.engine.Promote(&managedmodel.PromoteReq{Primary: loaded.app, Managed: loaded.managed,
		Member: p.member, MountPaths: paths})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if p.promoted, err = uc.appService.LoadApp(ctx, uc.db, loaded.app.ProjectID, p.member.App.ID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...)); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if p.fenced, err = uc.managedService.WriterOf(ctx, uc.db, loaded.app, loaded.managed); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return p, nil
}

// promotionRun is a promotion's steps, run against the database and the swarm.
type promotionRun struct {
	uc   *UC
	auth *basedto.Auth
	req  *appmanageddto.PromoteManagedMemberReq
	p    *promotion
	meta *basedto.Meta
	// fenceEnv and replicasEnv are the variables built to apply; restarted are
	// the replicas pointed at the new primary.
	fenceEnv    []*envvarservice.AppEnvVarData
	replicasEnv []*envvarservice.AppEnvVarData
	restarted   []*entity.App
}

func (r *promotionRun) fence(ctx context.Context) error {
	p := r.p
	return hperrors.Wrap(transaction.Execute(ctx, r.uc.db, func(db database.Tx) error {
		changed, err := r.uc.setEnv(ctx, db, p.fenced, p.spec.FenceEnv, nil)
		if err != nil {
			return err
		}
		if err = r.uc.lockEnvKeys(ctx, db, p.fenced, envKeysOf(p.spec.FenceEnv), nil); err != nil {
			return err
		}
		if !changed {
			return nil
		}
		if r.fenceEnv, err = r.uc.buildConsumersEnv(ctx, db, p.loaded.app, []*entity.App{p.fenced}); err != nil {
			return err
		}
		return r.uc.recordReplica(ctx, db, r.auth, p.loaded.app, "promotionFenced", p.fenced.ID)
	}))
}

func (r *promotionRun) stop(ctx context.Context) (bool, error) {
	stopped, err := r.uc.managedService.StopAndWait(ctx, r.p.fenced, fenceWait)
	return stopped, hperrors.Wrap(err)
}

// applyFence applies the fence once stopped: the update starts nothing at no
// replicas.
func (r *promotionRun) applyFence(ctx context.Context) {
	for i, err := range r.uc.envVarService.ApplyEnvVarsForApps(ctx, r.uc.db, r.fenceEnv, true, false) {
		r.meta.Warning += fmt.Sprintf("'%v' is stopped, but its fence could not be applied to it: %v\n",
			r.fenceEnv[i].App.Name, err.Error())
	}
}

func (r *promotionRun) promote(ctx context.Context) (*int, error) {
	var code *int
	err := transaction.Execute(ctx, r.uc.db, func(db database.Tx) error {
		var runErr error
		code, runErr = r.uc.managedService.RunCommand(ctx, db, r.p.promoted, r.p.spec.Promote)
		if runErr != nil && code == nil {
			return hperrors.Wrap(runErr)
		}
		return nil
	})
	return code, hperrors.Wrap(err)
}

// allow lets the replicas into the promoted member. It fails no promotion: the
// replicas follow it from the archive meanwhile.
func (r *promotionRun) allow(ctx context.Context) {
	password, err := r.p.loaded.managed.Postgres.ReplicationPassword.GetPlain()
	if err == nil && password != "" {
		err = transaction.Execute(ctx, r.uc.db, func(db database.Tx) error {
			return r.uc.managedService.AllowReplicas(ctx, db, r.p.promoted, r.p.spec.Allow, password)
		})
	}
	if err != nil {
		r.meta.Warning += fmt.Sprintf("The replicas could not be let into '%v', and follow it from the archive "+
			"until they are: %v\n", r.p.promoted.Name, hperrors.GetErrorDetail(err, ""))
	}
}

func (r *promotionRun) unfence(ctx context.Context) error {
	p := r.p
	var envData []*envvarservice.AppEnvVarData
	err := transaction.Execute(ctx, r.uc.db, func(db database.Tx) error {
		if _, err := r.uc.setEnv(ctx, db, p.fenced, nil, envKeysOf(p.spec.FenceEnv)); err != nil {
			return err
		}
		if err := r.uc.lockEnvKeys(ctx, db, p.fenced, nil, envKeysOf(p.spec.FenceEnv)); err != nil {
			return err
		}
		var err error
		envData, err = r.uc.buildConsumersEnv(ctx, db, p.loaded.app, []*entity.App{p.fenced})
		return err
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if errs := r.uc.envVarService.ApplyEnvVarsForApps(ctx, r.uc.db, envData, true, false); len(errs) > 0 {
		for _, applyErr := range errs {
			return hperrors.Wrap(applyErr)
		}
	}
	return hperrors.Wrap(r.uc.managedService.Start(ctx, p.fenced))
}

func (r *promotionRun) record(ctx context.Context) error {
	return hperrors.Wrap(transaction.Execute(ctx, r.uc.db, func(db database.Tx) error {
		var err error
		r.restarted, r.replicasEnv, err = r.uc.recordPromotion(ctx, db, r.auth, r.req, r.p)
		return err
	}))
}

// restart restarts the other replicas with the new primary to follow. The
// promoted one is not restarted: it is the primary already, and its marker keeps
// it so.
func (r *promotionRun) restart(ctx context.Context) {
	for i, err := range r.uc.envVarService.ApplyEnvVarsForApps(ctx, r.uc.db, r.replicasEnv, true, true) {
		r.meta.Warning += fmt.Sprintf("Replica '%v' could not be restarted to follow it: %v\n",
			r.replicasEnv[i].App.Name, err.Error())
	}
}

// recordPromotion records the member as the cluster's primary: the engine moves
// its jobs and fences what wrote; the member sheds its standby variables, and the
// other replicas are pointed at it. It returns the replicas to restart, and
// their variables built.
func (uc *UC) recordPromotion(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	req *appmanageddto.PromoteManagedMemberReq,
	p *promotion,
) ([]*entity.App, []*envvarservice.AppEnvVarData, error) {
	loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
	if err != nil {
		return nil, nil, err
	}
	now := timeutil.NowUTC()
	jobs := make(map[string]*entity.Setting, len(loaded.jobs))
	for _, job := range loaded.jobs {
		jobs[job.ID] = job
	}
	if member := loaded.managed.Postgres.Member(p.member.App.ID); member == nil ||
		member.Role != entity.ManagedRolePrimary {
		resp, err := loaded.engine.Reconfigure(&managedmodel.ReconfigureReq{App: loaded.app,
			ManagedID: loaded.setting.ID, Managed: loaded.managed, Jobs: jobs, Now: now,
			Change: &managedmodel.SettingsChange{Writer: &p.member.App.ID}})
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		if err = loaded.setting.SetData(loaded.managed); err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		loaded.setting.UpdateVer++
		if err = uc.storeReconfigured(ctx, db, loaded.setting, resp, now); err != nil {
			return nil, nil, err
		}
	}
	// What made the member a standby goes: on a restart it would be one again.
	if _, err = uc.setEnv(ctx, db, p.promoted, nil, p.spec.StandbyEnvKeys); err != nil {
		return nil, nil, err
	}
	if err = uc.lockEnvKeys(ctx, db, p.promoted, nil, p.spec.StandbyEnvKeys); err != nil {
		return nil, nil, err
	}
	var restarted []*entity.App
	for _, member := range loaded.managed.Postgres.Members {
		if member.Role != entity.ManagedRoleReplica {
			continue
		}
		app, err := uc.appService.LoadApp(ctx, db, loaded.app.ProjectID, member.App.ID, false, false,
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...))
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		changed, err := uc.setEnv(ctx, db, app, [][2]string{{p.spec.StandbyHostKey, p.promoted.Key}}, nil)
		if err != nil {
			return nil, nil, err
		}
		if changed {
			restarted = append(restarted, app)
		}
	}
	envData, err := uc.buildConsumersEnv(ctx, db, loaded.app, restarted)
	if err != nil {
		return nil, nil, err
	}
	return restarted, envData, hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db,
		&auditservice.Entry{
			Type:     base.AuditLogTypeAppUpdate,
			Scope:    base.ObjectScopeApp,
			ObjectID: loaded.app.ID,
			Source:   base.AuditLogSourceAPIUpdate,
			Section:  auditSectionManaged,
			Auth:     auth,
			ResType:  base.ResourceTypeApp,
			ResID:    loaded.app.ID,
			ResName:  loaded.app.Name,
			Detail: auditdetail.New().Set("promoted", p.promoted.ID).Set("fenced", p.fenced.ID).
				Set("restarted", len(restarted)).String(),
		}))
}

// setEnv sets variables of an app's own, literal, and removes others, in its
// env-var setting; it says whether that changed anything.
func (uc *UC) setEnv(ctx context.Context, db database.Tx, app *entity.App, set [][2]string,
	remove []string) (bool, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", app.ID),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return false, hperrors.NewNotFound("Variables of " + app.Name)
	}
	setting := settings[0]
	vars, err := setting.AsEnvVars()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	next, changed := changedEnv(vars.Data, set, remove)
	if !changed {
		return false, nil
	}
	vars.Data = next
	if err = setting.SetData(vars); err != nil {
		return false, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	return true, hperrors.Wrap(uc.settingRepo.Update(ctx, db, setting,
		bunex.UpdateColumns("data", "update_ver", "updated_at")))
}

// changedEnv is vars with set's set, as literal runtime variables, and remove's
// removed; and whether that differs from vars.
func changedEnv(vars []*entity.EnvVar, set [][2]string, remove []string) ([]*entity.EnvVar, bool) {
	changed := false
	out := make([]*entity.EnvVar, 0, len(vars)+len(set))
	for _, env := range vars {
		if slices.Contains(remove, env.Key) {
			changed = true
			continue
		}
		out = append(out, env)
	}
	for _, pair := range set {
		found := false
		for _, env := range out {
			if env.Key != pair[0] || env.IsBuild || env.IsShared {
				continue
			}
			found = true
			if env.Value != pair[1] || !env.IsLiteral {
				env.Value, env.IsLiteral, changed = pair[1], true, true
			}
		}
		if !found {
			out = append(out, &entity.EnvVar{Key: pair[0], Value: pair[1], IsLiteral: true})
			changed = true
		}
	}
	return out, changed
}

// lockEnvKeys adds keys to, and removes keys from, the variables an app's engine
// setting keeps its value of: a person changes none of them.
func (uc *UC) lockEnvKeys(ctx context.Context, db database.Tx, app *entity.App, add, remove []string) error {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", app.ID),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppManaged),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil || len(settings) == 0 {
		return hperrors.Wrap(err)
	}
	setting := settings[0]
	managed, err := setting.AsAppManagedSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	keys := slices.DeleteFunc(slices.Clone(managed.Locks.EnvKeys), func(key string) bool {
		return slices.Contains(remove, key)
	})
	for _, key := range add {
		if !slices.Contains(keys, key) {
			keys = append(keys, key)
		}
	}
	if slices.Equal(keys, managed.Locks.EnvKeys) {
		return nil
	}
	managed.Locks.EnvKeys = keys
	return uc.saveManagedSetting(ctx, db, setting, managed, timeutil.NowUTC())
}

func envKeysOf(pairs [][2]string) []string {
	keys := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		keys = append(keys, pair[0])
	}
	return keys
}
