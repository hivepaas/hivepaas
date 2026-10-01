package appuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// testRunTimeout bounds a test run, its libraries' install included: the first
// run of a function may pull its runtime's image and install its libraries.
const testRunTimeout = 15 * time.Minute

// TestRunFunction calls a function once with the code the editor has, not yet
// saved, on a build node. It runs the code with the function's variables and
// secrets: only someone who can change the function may.
func (uc *UC) TestRunFunction(
	ctx context.Context,
	_ *basedto.Auth,
	req *appdto.TestRunFunctionReq,
) (*appdto.TestRunFunctionResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppKind, base.SettingTypeAppDeployment),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !entity.IsFunctionKind(app.GetSettingByType(base.SettingTypeAppKind)) {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFunction).WithParam("Name", app.Name)
	}
	var source *entity.DeploymentFunctionSource
	if setting := app.GetSettingByType(base.SettingTypeAppDeployment); setting != nil {
		deployment, err := setting.AsAppDeploymentSettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		source = deployment.FunctionSource
	}
	if source == nil {
		return nil, hperrors.Wrap(hperrors.ErrUnconfigured).WithParam("Name", "Function source")
	}

	ctx, cancel := context.WithTimeout(ctx, testRunTimeout)
	defer cancel()
	resp, err := uc.functionService.TestRun(ctx, uc.db, &functionservice.TestRunReq{
		App:     app,
		Source:  source,
		Files:   req.Code.ToEntity().Files,
		Request: req.ToRunRequest(),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	data := &appdto.TestRunFunctionDataResp{
		Outcome:        string(resp.Outcome),
		Status:         resp.Status,
		Headers:        resp.Headers,
		Body:           resp.Body,
		BodyTruncated:  resp.BodyTruncated,
		RequestID:      resp.RequestID,
		DurationMs:     resp.DurationMs,
		Logs:           resp.Logs,
		LogsTruncated:  resp.LogsTruncated,
		Error:          resp.Error,
		ExitCode:       resp.ExitCode,
		LibrariesBuilt: resp.LibrariesBuilt,
		LibrariesLog:   resp.LibrariesLog,
	}
	for _, f := range resp.LockFiles {
		data.LockFiles = append(data.LockFiles, &appsettingsdto.FunctionFileResp{Path: f.Path, Content: f.Content})
	}
	return &appdto.TestRunFunctionResp{Data: data}, nil
}
