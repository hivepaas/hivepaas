package apptemplateuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// checkAppRefs resolves what each app parameter names. Rendering only saw the
// shape of a key, because the template service has no database; whether an app
// of that name is here, and whether it is the kind the template asked for, is
// answered now - before anything is created.
//
// Left unchecked, a name with a typo in it provisions cleanly and then waits
// forever for a database that was never there, with the reason in a log nobody
// is watching.
func (uc *UC) checkAppRefs(
	ctx context.Context,
	req *apptemplatedto.CreateAppFromTemplateReq,
	rendered *apptemplateservice.RenderResp,
) error {
	for _, param := range rendered.Template.Parameters {
		if param == nil || param.Type != templatemodel.ParamTypeApp {
			continue
		}
		value := rendered.Result.Params[param.Name]
		if value == nil || value.Value == nil || value.Text() == "" {
			continue // optional, and nothing was given
		}
		if err := uc.checkAppRef(ctx, req, param, value.Text()); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UC) checkAppRef(
	ctx context.Context,
	req *apptemplatedto.CreateAppFromTemplateReq,
	param *templatemodel.Parameter,
	key string,
) error {
	apps, _, err := uc.appRepo.List(ctx, uc.db, req.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", req.ProjectEnvID),
		bunex.SelectWhere("app.key = ?", key),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).WithParam("Name", param.Name).
			WithExtraDetail("%s: this environment has no app called %q", param.Name, key)
	}

	engine, err := uc.appEngine(ctx, apps[0])
	if err != nil {
		return err
	}
	// An app that declares no engine is accepted: most were not made from a
	// template, and saying nothing is not the same as saying something else.
	if param.Engine == "" || engine == "" || engine == param.Engine {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).WithParam("Name", param.Name).
		WithExtraDetail("%s: %q is a %s app, and this asks for %s", param.Name, key, engine, param.Engine)
}

// appEngine is what an app says it runs, empty when it says nothing - which is
// every app created without a template, since nothing else writes a kind.
func (uc *UC) appEngine(ctx context.Context, app *entity.App) (string, error) {
	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	setting := settinghelper.FindSettingByType(settings, base.SettingTypeAppKind)
	if setting == nil {
		return "", nil
	}
	kind, err := setting.AsAppKindSettings()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return kind.Engine, nil
}
