package appsettingsuc

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func functionContainerReq() *appsettingsdto.UpdateAppContainerSettingsReq {
	grace := timeutil.Duration(time.Second)
	return &appsettingsdto.UpdateAppContainerSettingsReq{BaseContainerSettings: &appsettingsdto.BaseContainerSettings{
		Image: "fn:dev-1234567", Command: "node server.js", WorkingDir: "/srv", User: "1000",
		StopGracePeriod: &grace,
		Healthcheck:     &appsettingsdto.Healthcheck{Enabled: true, Mode: "CMD", Command: "true"},
	}}
}

func functionContainerData(source *entity.DeploymentFunctionSource) *updateAppContainerSettingsData {
	return &updateAppContainerSettingsData{
		App: &entity.App{ID: "app-1"},
		Service: &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
		}}},
		FunctionSource: source,
	}
}

// A function's container settings keep what is fixed for a function - the
// runtime's command, working directory and health check, a stop that lets
// running calls finish - and take the rest from the request.
func TestAFunctionsContainerSettingsKeepWhatIsFixedForIt(t *testing.T) {
	data := functionContainerData(&entity.DeploymentFunctionSource{Timeout: timeutil.Duration(30 * time.Second)})

	assert.NoError(t, (&UC{}).prepareUpdatingAppContainerSettings(functionContainerReq(), data))

	contSpec := data.Service.Spec.TaskTemplate.ContainerSpec
	assert.Nil(t, contSpec.Command)
	assert.Nil(t, contSpec.Args)
	assert.Empty(t, contSpec.Dir)
	assert.Nil(t, contSpec.Healthcheck)
	assert.Equal(t, 40*time.Second, *contSpec.StopGracePeriod)
	assert.Equal(t, "1000", contSpec.User)
	assert.Equal(t, "fn:dev-1234567", contSpec.Image)
}

// Another app's are the request's, as before.
func TestAnAppsContainerSettingsAreTheRequests(t *testing.T) {
	data := functionContainerData(nil)

	assert.NoError(t, (&UC{}).prepareUpdatingAppContainerSettings(functionContainerReq(), data))

	contSpec := data.Service.Spec.TaskTemplate.ContainerSpec
	assert.Equal(t, []string{"node", "server.js"}, append(contSpec.Command, contSpec.Args...))
	assert.Equal(t, "/srv", contSpec.Dir)
	assert.NotNil(t, contSpec.Healthcheck)
	assert.Equal(t, time.Second, *contSpec.StopGracePeriod)
}
