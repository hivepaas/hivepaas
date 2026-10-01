package appuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	fnservice "github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// fakeApps loads one app.
type fakeApps struct {
	appservice.Service
	app *entity.App
}

func (f *fakeApps) LoadApp(
	_ context.Context, _ database.IDB, _, _ string, _, _ bool, _ ...bunex.SelectQueryOption,
) (*entity.App, error) {
	return f.app, nil
}

// fakeFunctions runs test runs, and keeps the one it was asked for.
type fakeFunctions struct {
	fnservice.Service
	got      *fnservice.TestRunReq
	deadline time.Time
}

func (f *fakeFunctions) TestRun(
	ctx context.Context, _ database.IDB, req *fnservice.TestRunReq,
) (*functiontest.RunResp, error) {
	f.got = req
	f.deadline, _ = ctx.Deadline()
	return &functiontest.RunResp{
		Outcome: functiontest.OutcomeOK, Status: 200, Body: []byte("hi"), Logs: "hello\n",
		LockFiles: []*entity.FunctionFile{{Path: "package-lock.json", Content: "{}"}},
	}, nil
}

func appWithSettings(t *testing.T, category base.AppCategory, source *entity.DeploymentFunctionSource) *entity.App {
	t.Helper()
	kind := &entity.Setting{Type: base.SettingTypeAppKind, Status: base.SettingStatusActive}
	assert.NoError(t, kind.SetData(&entity.AppKindSettings{Category: category}))
	deployment := &entity.Setting{Type: base.SettingTypeAppDeployment, Status: base.SettingStatusActive}
	assert.NoError(t, deployment.SetData(&entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodFunction,
		FunctionSource: source}))
	return &entity.App{ID: "app-1", Name: "hello", Settings: []*entity.Setting{kind, deployment}}
}

func testRunFunctionReq() *appdto.TestRunFunctionReq {
	req := &appdto.TestRunFunctionReq{
		ProjectID: "p", ProjectEnvID: "e", AppID: "app-1",
		Code: &appsettingsdto.FunctionInlineCodeReq{Files: []*appsettingsdto.FunctionFileReq{
			{Path: "index.js", Content: "export default () => 'hi'"},
		}},
	}
	_ = req.ModifyRequest()
	return req
}

// A test run is the function's: its saved settings, the code sent, the request
// sent; and what comes back is answered as it came, within a time limit.
func TestATestRunCallsTheFunctionWithTheCodeSent(t *testing.T) {
	source := &entity.DeploymentFunctionSource{Runtime: base.FunctionRuntimeNode24,
		Timeout: timeutil.Duration(30 * time.Second)}
	functions := &fakeFunctions{}
	uc := &UC{appService: &fakeApps{app: appWithSettings(t, base.AppCategoryFunction, source)},
		functionService: functions}

	resp, err := uc.TestRunFunction(context.Background(), &basedto.Auth{}, testRunFunctionReq())

	assert.NoError(t, err)
	assert.Equal(t, base.FunctionRuntimeNode24, functions.got.Source.Runtime)
	assert.Equal(t, "index.js", functions.got.Files[0].Path)
	assert.Equal(t, "GET", functions.got.Request.Method)
	assert.WithinDuration(t, time.Now().Add(testRunTimeout), functions.deadline, time.Minute)
	data := resp.Data
	assert.Equal(t, "ok", data.Outcome)
	assert.Equal(t, 200, data.Status)
	assert.Equal(t, []byte("hi"), data.Body)
	assert.Equal(t, "hello\n", data.Logs)
	assert.Equal(t, "package-lock.json", data.LockFiles[0].Path)
}

// Only a function has test runs.
func TestAnAppThatIsNotAFunctionHasNoTestRun(t *testing.T) {
	uc := &UC{appService: &fakeApps{app: appWithSettings(t, base.AppCategoryWebapp, nil)},
		functionService: &fakeFunctions{}}

	_, err := uc.TestRunFunction(context.Background(), &basedto.Auth{}, testRunFunctionReq())

	assert.ErrorIs(t, err, hperrors.ErrAppNotFunction)
}
