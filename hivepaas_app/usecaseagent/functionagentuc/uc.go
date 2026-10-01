// Package functionagentuc is what an agent does for functions: their test runs,
// on its node.
package functionagentuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// Runner runs a test run on this node; functiontest.Runner is the one an agent
// has.
type Runner interface {
	Run(ctx context.Context, req *functiontest.RunReq) (*functiontest.RunResp, error)
}

type UC struct {
	logger logging.Logger
	runner Runner
}

func New(
	logger logging.Logger,
	dockerManager docker.Manager,
	imageBuildService imagebuildservice.Service,
) *UC {
	return &UC{
		logger: logger,
		runner: &functiontest.Runner{
			Docker:  dockerManager,
			Builder: functiontest.NewLibrariesBuilder(imageBuildService),
			TempDir: base.BaseTempDirDefault,
		},
	}
}

// WithRunner makes the agent run test runs with runner.
func (uc *UC) WithRunner(runner Runner) *UC {
	uc.runner = runner
	return uc
}

// TestRun runs a function's test run on this node.
func (uc *UC) TestRun(ctx context.Context, req *functiontest.RunReq) (resp *functiontest.RunResp, err error) {
	// The agent's calls have no recovery of their own: a panic here would take
	// the agent down with every build and run on it.
	defer func() {
		if r := recover(); r != nil {
			resp, err = nil, hperrors.NewPanic(r)
		}
	}()
	resp, err = uc.runner.Run(ctx, req)
	return resp, hperrors.Wrap(err)
}
