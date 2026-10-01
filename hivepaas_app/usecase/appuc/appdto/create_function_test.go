package appdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func createFunctionReq(source *appsettingsdto.DeploymentFunctionSourceReq) *CreateFunctionReq {
	return &CreateFunctionReq{
		ProjectID:    "01J0000000000000000000PRJ1",
		ProjectEnvID: "01J0000000000000000000ENV1",
		AppBaseReq:   &AppBaseReq{Name: "  hello  ", Status: base.AppStatusActive},
		Source:       source,
	}
}

// pathsOf are the fields validation errors are on.
func pathsOf(errs hperrors.ValidationErrors) []string {
	var paths []string
	for _, inner := range errs.Build(translation.LangEn).InnerErrors {
		paths = append(paths, inner.Path)
	}
	return paths
}

// A function is created from a name and its source, which takes its runtime's
// defaults before it is checked.
func TestAFunctionIsCreatedFromItsSource(t *testing.T) {
	req := createFunctionReq(&appsettingsdto.DeploymentFunctionSourceReq{
		Runtime: base.FunctionRuntimeNode24,
		Code: appsettingsdto.FunctionCodeReq{Inline: &appsettingsdto.FunctionInlineCodeReq{
			Files: []*appsettingsdto.FunctionFileReq{{Path: "index.js", Content: "export default () => ({})"}},
		}},
	})

	assert.NoError(t, req.ModifyRequest())

	assert.Empty(t, req.Validate())
	assert.Equal(t, "hello", req.Name)
	assert.Equal(t, "index.js", req.Source.Entrypoint.File)
	assert.Equal(t, base.FunctionContractV1, req.Source.Contract)
}

func TestAFunctionWithoutASourceIsRefused(t *testing.T) {
	req := createFunctionReq(nil)

	assert.NoError(t, req.ModifyRequest())

	assert.Equal(t, []string{"source"}, pathsOf(req.Validate()))
}

// What is wrong with a function's source is said of the source's own field.
func TestAFunctionsSourceIsCheckedAsItsSettingsWouldBe(t *testing.T) {
	req := createFunctionReq(&appsettingsdto.DeploymentFunctionSourceReq{
		Runtime: base.FunctionRuntimeNode24,
		Code: appsettingsdto.FunctionCodeReq{Inline: &appsettingsdto.FunctionInlineCodeReq{
			Files: []*appsettingsdto.FunctionFileReq{{Path: "index.js", Content: "export default () => ({})"}},
		}},
		SystemPackages: []string{"ffmpeg; rm -rf /"},
	})

	assert.NoError(t, req.ModifyRequest())

	assert.Equal(t, []string{"source.systemPackages[0]"}, pathsOf(req.Validate()))
}

// A request that names no function is refused as one, not taken without a name.
func TestAFunctionWithoutANameIsRefused(t *testing.T) {
	req := createFunctionReq(&appsettingsdto.DeploymentFunctionSourceReq{
		Runtime: base.FunctionRuntimeNode24,
		Code: appsettingsdto.FunctionCodeReq{Inline: &appsettingsdto.FunctionInlineCodeReq{
			Files: []*appsettingsdto.FunctionFileReq{{Path: "index.js", Content: "export default () => ({})"}},
		}},
	})
	req.AppBaseReq = nil

	assert.NoError(t, req.ModifyRequest())

	assert.ElementsMatch(t, []string{"name", "status"}, pathsOf(req.Validate()))
}

// A function may ask for a domain to be routed at, written one way.
func TestAFunctionMayAskForADomain(t *testing.T) {
	req := createFunctionReq(nodeSource())
	req.Domain = "  Hello.Example.com. "

	assert.NoError(t, req.ModifyRequest())

	assert.Empty(t, req.Validate())
	assert.Equal(t, "hello.example.com", req.Domain)
}

func TestAFunctionsDomainIsADomainName(t *testing.T) {
	for _, domain := range []string{"not a domain", "*.example.com", "http://hello.example.com"} {
		req := createFunctionReq(nodeSource())
		req.Domain = domain

		assert.NoError(t, req.ModifyRequest())

		assert.Equal(t, []string{"domain"}, pathsOf(req.Validate()), domain)
	}
}

func nodeSource() *appsettingsdto.DeploymentFunctionSourceReq {
	return &appsettingsdto.DeploymentFunctionSourceReq{
		Runtime: base.FunctionRuntimeNode24,
		Code: appsettingsdto.FunctionCodeReq{Inline: &appsettingsdto.FunctionInlineCodeReq{
			Files: []*appsettingsdto.FunctionFileReq{{Path: "index.js", Content: "export default () => ({})"}},
		}},
	}
}
