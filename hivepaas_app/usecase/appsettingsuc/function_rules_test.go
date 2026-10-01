package appsettingsuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func deploymentOf(method base.DeploymentMethod, source *entity.DeploymentFunctionSource) *entity.AppDeploymentSettings {
	return &entity.AppDeploymentSettings{ActiveMethod: method, FunctionSource: source}
}

// A function is deployed from its code, and nothing but a function is
// deployed as one or keeps a function's source.
func TestAFunctionIsDeployedAsOne(t *testing.T) {
	source := &entity.DeploymentFunctionSource{Runtime: base.FunctionRuntimeNode24}

	assert.NoError(t, checkDeploymentOfKind(true, deploymentOf(base.DeploymentMethodFunction, source)))
	assert.NoError(t, checkDeploymentOfKind(false, deploymentOf(base.DeploymentMethodRepo, nil)))
	assert.ErrorIs(t, checkDeploymentOfKind(true, deploymentOf(base.DeploymentMethodImage, source)),
		hperrors.ErrDeploymentMethodFunctionRequired)
	assert.ErrorIs(t, checkDeploymentOfKind(false, deploymentOf(base.DeploymentMethodFunction, source)),
		hperrors.ErrDeploymentMethodFunctionUnallowed)
	assert.ErrorIs(t, checkDeploymentOfKind(false, deploymentOf(base.DeploymentMethodImage, source)),
		hperrors.ErrDeploymentMethodFunctionUnallowed)
}

// A function is created as one and stays one.
func TestAFunctionStaysAFunction(t *testing.T) {
	function := &entity.AppKindSettings{Category: base.AppCategoryFunction}
	webapp := &entity.AppKindSettings{Category: base.AppCategoryWebapp}

	assert.NoError(t, checkKindCategoryChange(function, base.AppCategoryFunction))
	assert.NoError(t, checkKindCategoryChange(webapp, base.AppCategoryDatabase))
	assert.NoError(t, checkKindCategoryChange(nil, base.AppCategoryWebapp))
	assert.ErrorIs(t, checkKindCategoryChange(function, base.AppCategoryWebapp),
		hperrors.ErrAppKindFunctionUnchangeable)
	assert.ErrorIs(t, checkKindCategoryChange(webapp, base.AppCategoryFunction),
		hperrors.ErrAppKindFunctionUnchangeable)
	assert.ErrorIs(t, checkKindCategoryChange(nil, base.AppCategoryFunction),
		hperrors.ErrAppKindFunctionUnchangeable)
}

// A function's routing points at the port its runtime listens on, whatever
// the request said.
func TestAFunctionsRoutingPointsAtItsRuntimesPort(t *testing.T) {
	routing := &entity.AppRoutingSettings{Port: 3000, Domains: []*entity.AppDomain{
		{Domain: "fn.example.com", ContainerPort: 3000}, {Domain: "fn2.example.com", ContainerPort: 9000},
	}}

	fixFunctionRouting(routing)

	assert.Equal(t, base.FunctionPort, routing.Port)
	for _, domain := range routing.Domains {
		assert.Equal(t, base.FunctionPort, domain.ContainerPort)
	}
}
