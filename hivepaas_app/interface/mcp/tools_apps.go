package mcp

import (
	"context"
	"net/url"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

type appInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env     string `json:"env" jsonschema:"the env's name, such as production"`
	App     string `json:"app" jsonschema:"the app's key, name or id; list_apps lists them"`
}

func (in appInput) resolve(ctx context.Context, call *Call) (*appRef, error) {
	return resolveApp(ctx, call, in.Project, in.Env, in.App)
}

// readServiceTasks is appsettingsuc's service tasks of an app: its containers.
func readServiceTasks(ctx context.Context, call *Call, ref *appRef) ([]*appsettingsdto.ServiceTaskResp, error) {
	var resp appsettingsdto.GetAppServiceTasksResp
	if err := call.Get(ctx, ref.path("/service-tasks"), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// withoutUserinfo is a repository URL without the credentials one may carry,
// as https://user:token@host/repo does.
func withoutUserinfo(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	u.User = nil
	return u.String()
}
