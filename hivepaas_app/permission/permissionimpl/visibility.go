package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

type visibility struct {
	manager *manager
	db      database.IDB
	auth    *basedto.Auth
	answers map[string]bool
}

func (p *manager) NewVisibility(db database.IDB, auth *basedto.Auth) permission.Visibility {
	return &visibility{manager: p, db: db, auth: auth, answers: map[string]bool{}}
}

func (v *visibility) AllowsModule(
	ctx context.Context,
	module base.ResourceModule,
	action base.ActionType,
) (bool, error) {
	return v.ask(ctx, "module/"+string(module)+"/"+string(action), &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		Module:          module,
	})
}

func (v *visibility) AllowsProjectEnv(
	ctx context.Context,
	projectID, env string,
	action base.ActionType,
) (bool, error) {
	envID := projecthelper.CalcProjectEnvID(projectID, env)
	if envID == "" {
		// Without an env the check would ask about the project and every env in
		// it, and answer yes for a grant on any one of them.
		return false, nil
	}
	return v.ask(ctx, "env/"+envID+"/"+string(action), &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		ProjectEnv:      &envID,
	})
}

func (v *visibility) ask(ctx context.Context, key string, check permission.AccessCheck) (bool, error) {
	if answer, ok := v.answers[key]; ok {
		return answer, nil
	}
	// CheckAccess writes what it finds onto the auth it is given, for the lists
	// that follow a check. A question asked here must leave nothing on the
	// caller's for its own lists to read, so each is asked on a copy.
	scratch := &basedto.Auth{User: v.auth.User}
	answer, err := v.manager.CheckAccess(ctx, v.db, scratch, check)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	v.answers[key] = answer
	return answer, nil
}
