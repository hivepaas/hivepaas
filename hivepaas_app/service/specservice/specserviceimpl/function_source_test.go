package specserviceimpl

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func inlineSource(files ...*entity.FunctionFile) *entity.DeploymentFunctionSource {
	if len(files) == 0 {
		files = []*entity.FunctionFile{{Path: "index.js", Content: "export default () => ({})"}}
	}
	return &entity.DeploymentFunctionSource{
		Runtime: base.FunctionRuntimeNode24,
		Code:    entity.FunctionCode{Inline: &entity.FunctionInlineCode{Files: files}},
	}
}

func repoSource(url string) *entity.DeploymentFunctionSource {
	return &entity.DeploymentFunctionSource{
		Runtime: base.FunctionRuntimeGo127,
		Code: entity.FunctionCode{Repo: &entity.FunctionRepoCode{
			RepoType: base.RepoTypeGit, RepoURL: url, RepoRef: "main",
		}},
	}
}

// functionSourceCases are sources the API takes and sources it refuses, by
// what each changes of a valid one.
func functionSourceCases() map[string]*entity.DeploymentFunctionSource {
	many := make([]*entity.FunctionFile, 0, base.FunctionInlineCodeMaxFiles+1)
	for i := range base.FunctionInlineCodeMaxFiles + 1 {
		many = append(many, &entity.FunctionFile{Path: fmt.Sprintf("f%d.js", i), Content: "x"})
	}
	with := func(change func(*entity.DeploymentFunctionSource)) *entity.DeploymentFunctionSource {
		source := inlineSource()
		change(source)
		return source
	}
	return map[string]*entity.DeploymentFunctionSource{
		"valid inline":     inlineSource(),
		"valid repository": repoSource("https://github.com/acme/fns.git"),
		"unknown runtime":  with(func(s *entity.DeploymentFunctionSource) { s.Runtime = "cobol" }),
		"wrong extension":  with(func(s *entity.DeploymentFunctionSource) { s.Entrypoint.File = "main.py" }),
		"typescript":       with(func(s *entity.DeploymentFunctionSource) { s.Entrypoint.File = "src/index.ts" }),
		"typescript in python": with(func(s *entity.DeploymentFunctionSource) {
			s.Runtime, s.Entrypoint.File = base.FunctionRuntimePython313, "index.ts"
		}),
		"handler": with(func(s *entity.DeploymentFunctionSource) { s.Entrypoint.Handler = "1x" }),
		"no code": with(func(s *entity.DeploymentFunctionSource) { s.Code.Inline = nil }),
		"both codes": with(func(s *entity.DeploymentFunctionSource) {
			s.Code.Repo = repoSource("https://github.com/acme/fns.git").Code.Repo
		}),
		"too many files": inlineSource(many...),
		"reserved file":  inlineSource(&entity.FunctionFile{Path: ".hivepaas/Dockerfile", Content: "x"}),
		"file path":      inlineSource(&entity.FunctionFile{Path: "a b.js", Content: "x"}),
		"same file twice": inlineSource(&entity.FunctionFile{Path: "index.js", Content: "x"},
			&entity.FunctionFile{Path: "index.js", Content: "y"}),
		"code too large": inlineSource(&entity.FunctionFile{Path: "index.js",
			Content: strings.Repeat("x", int(base.FunctionInlineCodeMaxSize)+1)}),
		"inline in a directory": with(func(s *entity.DeploymentFunctionSource) { s.Code.Dir = "fns" }),
		"package":               with(func(s *entity.DeploymentFunctionSource) { s.SystemPackages = []string{"Bad Pkg"} }),
		"timeout": with(func(s *entity.DeploymentFunctionSource) {
			s.Timeout = timeutil.Duration(20 * time.Minute)
		}),
		"concurrency":    with(func(s *entity.DeploymentFunctionSource) { s.MaxConcurrency = 2000 }),
		"body size":      with(func(s *entity.DeploymentFunctionSource) { s.MaxBodySize = 200 * unit.MB }),
		"repository URL": repoSource("not a url"),
	}
}

// What an import accepts of a function's source is what the API accepts: the
// two are written apart - a service does not import a DTO - and held together
// here.
func TestAnImportedFunctionSourceIsCheckedAsTheAPIChecksIt(t *testing.T) {
	for name, source := range functionSourceCases() {
		raw, err := json.Marshal(source)
		assert.NoError(t, err)
		req := &appsettingsdto.DeploymentFunctionSourceReq{}
		assert.NoError(t, json.Unmarshal(raw, req), name)
		req.Normalize()
		apiRefuses := len(vld.Validate(req.Validate("functionSource")...)) > 0

		normalizeFunctionSource(source)
		problems := functionSourceProblems(source)

		assert.Equal(t, apiRefuses, len(problems) > 0, "%s: problems %v", name, problems)
	}
}

// A Node.js function written in TypeScript is imported as the API takes it.
func TestAnImportedTypeScriptFunctionIsValid(t *testing.T) {
	source := inlineSource(&entity.FunctionFile{Path: "index.ts", Content: "export default () => ({})"})
	source.Entrypoint.File = "index.ts"

	normalizeFunctionSource(source)

	assert.Empty(t, functionSourceProblems(source))
}
