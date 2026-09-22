package apptemplateserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
)

func renderStack(t *testing.T) *apptemplateservice.RenderResp {
	t.Helper()
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	resp, err := svc.Render(context.Background(), &apptemplateservice.RenderReq{
		Name:             "demostack",
		AppName:          "shop",
		DependencyParams: map[string]map[string]any{"db": {"dataVolume": "vol-db"}},
	})
	assert.NoError(t, err)
	return resp
}

func envVarsOf(t *testing.T, result any) map[string]string {
	t.Helper()
	doc, ok := result.(map[string]any)
	assert.True(t, ok)
	entries, ok := doc["data"].([]any)
	assert.True(t, ok)
	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		pair, _ := entry.(map[string]any)
		key, _ := pair["k"].(string)
		value, _ := pair["v"].(string)
		out[key] = value
	}
	return out
}

// The apps of one stack are the name the person gave for the primary component
// and that name with the component's added for the rest - the same shape a
// dependency's app has.
func TestServiceRenderNamesTheComponentApps(t *testing.T) {
	resp := renderStack(t)

	assert.Len(t, resp.Components, 2)
	names := map[string]string{}
	for _, component := range resp.Components {
		names[component.Name] = component.AppName
	}
	assert.Equal(t, map[string]string{"auth": "shop-auth", "gw": "shop"}, names)
}

// needs decides the order, not the order they are declared in: auth is declared
// second and rendered first, because the gateway reads what it shares.
func TestServiceRenderOrdersComponentsByNeeds(t *testing.T) {
	resp := renderStack(t)

	assert.Equal(t, "auth", resp.Components[0].Name)
	assert.Equal(t, "gw", resp.Components[1].Name)
	assert.True(t, resp.Components[1].Primary)
}

// The primary component's render is the response's own Result, so a caller that
// knows nothing about components still finds the app the person asked for.
func TestServiceRenderReturnsThePrimaryComponentAsTheResult(t *testing.T) {
	resp := renderStack(t)

	assert.Equal(t, "demogw:1.0.0", resp.Result.Image)
	assert.Equal(t, resp.Components[1].Result, resp.Result)
}

// Each component gets its own image from the version's table.
func TestServiceRenderGivesEachComponentItsOwnImage(t *testing.T) {
	resp := renderStack(t)

	images := map[string]string{}
	for _, component := range resp.Components {
		images[component.Name] = component.Result.Image
	}
	assert.Equal(t, map[string]string{"auth": "demoauth:1.0.0", "gw": "demogw:1.0.0"}, images)
}

// One generated secret reaches every component. This is the reason components
// exist rather than one template per process: as separate templates, the person
// would paste this value into each of them by hand.
func TestServiceRenderSharesOneGeneratedSecretAcrossComponents(t *testing.T) {
	resp := renderStack(t)

	gateway := envVarsOf(t, resp.Components[1].Result.Doc.Settings["envVars"])
	auth := envVarsOf(t, resp.Components[0].Result.Doc.Settings["envVars"])

	assert.NotEmpty(t, gateway["JWT_SECRET"])
	assert.Equal(t, gateway["JWT_SECRET"], auth["JWT_SECRET"], "one secret, generated once")
}

// One database, read by the component that needs it. Declaring the dependency on
// the template rather than on each component is what keeps a stack from creating
// one database per process.
func TestServiceRenderSharesOneDependencyAcrossComponents(t *testing.T) {
	resp := renderStack(t)

	assert.Len(t, resp.Dependencies, 1, "one database for the whole stack")
	assert.Equal(t, "shop-db", resp.Dependencies[0].AppName)

	auth := envVarsOf(t, resp.Components[0].Result.Doc.Settings["envVars"])
	assert.Equal(t, "${shop-db.HIVEPAAS_PASSWORD}", auth["DB_PASSWORD"])
}

// A component reaches another by key, and reads what it shares as an ordinary
// environment reference - the same two spellings a dependency has.
func TestServiceRenderResolvesReferencesBetweenComponents(t *testing.T) {
	resp := renderStack(t)

	gateway := envVarsOf(t, resp.Components[1].Result.Doc.Settings["envVars"])
	assert.Equal(t, "shop-auth", gateway["AUTH_KEY"])
	assert.Equal(t, "${shop-auth.HIVEPAAS_HOST}", gateway["AUTH_HOST"])
}

// Every component records its own base, because phase 2 merges each app against
// the render it came from.
func TestServiceRenderGivesEachComponentItsOwnBase(t *testing.T) {
	resp := renderStack(t)

	first, second := resp.Components[0].Result, resp.Components[1].Result
	assert.NotEmpty(t, first.BaseSHA256)
	assert.NotEmpty(t, second.BaseSHA256)
	assert.NotEqual(t, first.BaseSHA256, second.BaseSHA256)
}

// A generated secret is not in the base, which is stored: the base is what a
// later merge compares against, and it must never hold one.
func TestServiceRenderKeepsSecretsOutOfEveryComponentBase(t *testing.T) {
	resp := renderStack(t)

	gateway := envVarsOf(t, resp.Components[1].Result.Doc.Settings["envVars"])
	for _, component := range resp.Components {
		assert.NotContains(t, string(component.Result.Base), gateway["JWT_SECRET"])
	}
}
