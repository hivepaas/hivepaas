package engines

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// The engines HivePaaS has are the ones templates may name, speaking the API
// versions templates are told they speak: the table in the template format and
// the engines here are held to each other.
func TestTheEnginesMatchTheTemplateFormat(t *testing.T) {
	names := make([]templatemodel.ManagedEngine, 0)
	for _, engine := range all {
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
	engine, found := Lookup(templatemodel.ManagedEnginePostgres)
	assert.True(t, found)
	assert.Equal(t, templatemodel.ManagedEnginePostgres, engine.Name())

	_, found = Lookup("oracle")
	assert.False(t, found)
}

// Postgres keeps what it had before engines said what they can do: replicas,
// a verifier app, and 63-byte point names.
func TestPostgresCanDoWhatItDid(t *testing.T) {
	engine, _ := Lookup(templatemodel.ManagedEnginePostgres)

	caps := engine.Capabilities()
	assert.True(t, caps.Replicas)
	assert.True(t, caps.VerifierApp)
	assert.Equal(t, 63, engine.PointNameMax())
}

// MariaDB has no replicas, and runs its restore test in a verifier app too.
func TestMariaDBCanDoWhatItCan(t *testing.T) {
	engine, found := Lookup(templatemodel.ManagedEngineMariaDB)
	assert.True(t, found)

	caps := engine.Capabilities()
	assert.False(t, caps.Replicas)
	assert.True(t, caps.VerifierApp)
	assert.Equal(t, 63, engine.PointNameMax())
}
