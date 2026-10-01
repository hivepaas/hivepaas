// Package functionservice is what HivePaaS does with functions beyond their
// deployment: their test runs. Its packages hold the parts: functionbuild the
// Dockerfiles, functioncontainer what is fixed for a function's container,
// functiontest a test run on a node.
package functionservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
)

type Service interface {
	// TestRun calls a function once with code not yet saved, on a build node
	// chosen as for a build, and says what came back.
	TestRun(ctx context.Context, db database.IDB, req *TestRunReq) (*functiontest.RunResp, error)
}

type TestRunReq struct {
	// App is the function, with its Project and ProjectEnv.
	App *entity.App
	// Source is the function's saved source: its runtime, entrypoint, Debian
	// packages and limits. Its code is Files.
	Source  *entity.DeploymentFunctionSource
	Files   []*entity.FunctionFile
	Request *functiontest.Request
}
