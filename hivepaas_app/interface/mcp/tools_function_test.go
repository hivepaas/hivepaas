package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func inlineFunctionSettings(content string) *appsettingsdto.DeploymentSettingsResp {
	return &appsettingsdto.DeploymentSettingsResp{
		ActiveMethod: base.DeploymentMethodFunction,
		FunctionSource: &appsettingsdto.DeploymentFunctionSourceResp{
			Runtime: base.FunctionRuntimeNode24,
			Code: &appsettingsdto.FunctionCodeResp{Inline: &appsettingsdto.FunctionInlineCodeResp{
				Files: []*appsettingsdto.FunctionFileResp{{Path: "index.js", Content: content}},
			}},
		},
	}
}

// A redeploy of a function sees its code: code changed since the plan is a
// source changed since the plan.
func TestARedeployOfAFunctionSeesItsCode(t *testing.T) {
	before := deploySourceOf(inlineFunctionSettings("export default () => 1"))
	after := deploySourceOf(inlineFunctionSettings("export default () => 2"))

	assert.Equal(t, string(base.DeploymentMethodFunction), before.Method)
	assert.NotEmpty(t, before.Code)
	assert.NotEqual(t, before, after)
	assert.Equal(t, "the function's code", before.describe())
}

func TestARedeployOfAFunctionInARepositorySeesItsCommit(t *testing.T) {
	source := deploySourceOf(&appsettingsdto.DeploymentSettingsResp{
		ActiveMethod: base.DeploymentMethodFunction,
		FunctionSource: &appsettingsdto.DeploymentFunctionSourceResp{
			Runtime: base.FunctionRuntimeGo127,
			Code: &appsettingsdto.FunctionCodeResp{Repo: &appsettingsdto.FunctionRepoCodeResp{
				RepoURL: "https://token@github.com/acme/fns.git", RepoRef: "refs/heads/main", CommitHash: "abc123",
			}},
		},
	})

	assert.Equal(t, "https://github.com/acme/fns.git", source.Repo)
	assert.Equal(t, "abc123", source.Commit)
	assert.Equal(t, "https://github.com/acme/fns.git refs/heads/main", source.describe())
}
