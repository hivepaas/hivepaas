package schedjobuc

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// checkJobTypeInScope says which job types a scope holds: a job sequence lives
// in an app or a project env, and an env holds sequences only, for now.
func checkJobTypeInScope(scopeType base.ObjectScopeType, jobType base.SchedJobType) error {
	isSequence := jobType == base.SchedJobTypeJobSequence
	switch {
	case isSequence && scopeType != base.ObjectScopeApp && scopeType != base.ObjectScopeProjectEnv:
		return hperrors.NewArgumentInvalid("jobType").
			WithExtraDetail("a job sequence belongs to an app or a project env")
	case !isSequence && scopeType == base.ObjectScopeProjectEnv:
		return hperrors.NewArgumentInvalid("jobType").
			WithExtraDetail("a project env's scheduled jobs are job sequences")
	}
	return nil
}

// verifyingRefIDs is what the setting base checks exists, active, in the job's
// own scope. A sequence's members are left to checkSequenceMembers: an env
// sequence's members are its apps' jobs, which the env's scope does not show,
// and a disabled member is allowed - the run skips it.
func verifyingRefIDs(job *entity.SchedJob) *entity.RefObjectIDs {
	refIDs := job.GetRefObjectIDs()
	members := job.Sequence.MemberIDs()
	if len(members) == 0 {
		return refIDs
	}
	refIDs.RefSettingIDs = gofn.Filter(refIDs.RefSettingIDs, func(id string) bool {
		return !gofn.Contain(members, id)
	})
	return refIDs
}

// checkSequenceMembers refuses a sequence whose members are not all scheduled
// jobs within its reach: the app's own, for an app sequence; those of the
// env's apps, for an env sequence.
func (uc *UC) checkSequenceMembers(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	job *entity.SchedJob,
) error {
	memberIDs := job.Sequence.MemberIDs()
	if len(memberIDs) == 0 {
		return nil
	}
	members, _, err := uc.SettingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSchedJob),
		bunex.SelectWhere("setting.id IN (?)", bunex.List(memberIDs)),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	byID := make(map[string]*entity.Setting, len(members))
	appIDs := make([]string, 0, len(members))
	for _, member := range members {
		byID[member.ID] = member
		if member.Scope == base.ObjectScopeApp {
			appIDs = append(appIDs, member.ObjectID)
		}
	}
	apps, err := uc.appService.LoadApps(ctx, db, "", gofn.ToSet(appIDs), false, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	appByID := make(map[string]*entity.App, len(apps))
	for _, app := range apps {
		appByID[app.ID] = app
	}
	for i, step := range job.Sequence.Steps {
		member := byID[step.Job.ID]
		var app *entity.App
		if member != nil {
			app = appByID[member.ObjectID]
		}
		if problem := memberProblem(scope, member, app); problem != "" {
			return hperrors.NewArgumentInvalid(fmt.Sprintf("sequence.steps[%d].job", i)).
				WithExtraDetail("step %d: %s", i+1, problem)
		}
	}
	return nil
}

// memberProblem says why a job cannot be a step of a sequence in the scope, or
// nothing when it can. app is the member's app, when it has one.
func memberProblem(scope *entity.ObjectScope, member *entity.Setting, app *entity.App) string {
	switch {
	case member == nil:
		return "the scheduled job is not found"
	case member.Kind == string(base.SchedJobTypeJobSequence):
		return "a sequence cannot run another sequence"
	case member.Scope != base.ObjectScopeApp:
		return "the scheduled job is not an app's job"
	case scope.ScopeType == base.ObjectScopeApp && member.ObjectID != scope.AppID:
		return "the scheduled job belongs to another app"
	case scope.ScopeType == base.ObjectScopeProjectEnv && (app == nil || app.ProjectEnvID != scope.ProjectEnvID):
		return "the scheduled job belongs to an app outside this env"
	}
	return ""
}

// loadSequenceMembers adds to refObjects the jobs the sequences among
// jobSettings run, and their apps, whatever their scope: an env sequence's
// members are its apps' jobs, which loading its references in its own scope
// does not reach. What is gone is left out; the response says "missing".
func (uc *UC) loadSequenceMembers(
	ctx context.Context,
	db database.IDB,
	refObjects *entity.RefObjects,
	jobSettings ...*entity.Setting,
) error {
	var memberIDs []string
	for _, setting := range jobSettings {
		if setting.Kind != string(base.SchedJobTypeJobSequence) {
			continue
		}
		job, err := setting.AsSchedJob()
		if err != nil {
			return hperrors.Wrap(err)
		}
		memberIDs = append(memberIDs, job.Sequence.MemberIDs()...)
	}
	if len(memberIDs) == 0 || refObjects == nil {
		return nil
	}
	err := uc.SettingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &refObjects, nil, false,
		&entity.RefObjectIDs{RefSettingIDs: gofn.ToSet(memberIDs)})
	if err != nil {
		return hperrors.Wrap(err)
	}
	var appIDs []string
	for _, id := range memberIDs {
		if member := refObjects.RefSettings[id]; member != nil && member.Scope == base.ObjectScopeApp {
			appIDs = append(appIDs, member.ObjectID)
		}
	}
	if len(appIDs) == 0 {
		return nil
	}
	err = uc.SettingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &refObjects, nil, false,
		&entity.RefObjectIDs{RefAppIDs: gofn.ToSet(appIDs)})
	return hperrors.Wrap(err)
}
