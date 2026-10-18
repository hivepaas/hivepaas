package managedserviceimpl

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// The engines HivePaaS has are the ones templates may name, speaking the API
// versions templates are told they speak: the table in the template format and
// the engines here are held to each other.
func TestTheRegistryMatchesTheTemplateFormat(t *testing.T) {
	registry := newRegistry()

	names := make([]templatemodel.ManagedEngine, 0)
	for _, engine := range registry.engines {
		names = append(names, engine.Name())
		contract, known := templatemodel.ManagedEngines[engine.Name()]
		if assert.True(t, known, "%s is not in templatemodel.ManagedEngines", engine.Name()) {
			minAPI, maxAPI := engine.APIRange()
			assert.Equal(t, contract.MinAPI, minAPI, engine.Name())
			assert.Equal(t, contract.MaxAPI, maxAPI, engine.Name())
		}
	}
	for name := range templatemodel.ManagedEngines {
		assert.True(t, slices.Contains(names, name), "no engine for %s", name)
	}
}

func TestEngineLooksUpByName(t *testing.T) {
	registry := newRegistry()

	engine, found := registry.Engine(templatemodel.ManagedEnginePostgres)
	assert.True(t, found)
	assert.Equal(t, templatemodel.ManagedEnginePostgres, engine.Name())

	_, found = registry.Engine("oracle")
	assert.False(t, found)
}
