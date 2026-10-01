package envself

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func database(engine string, set ...string) *Facts {
	facts := &Facts{Category: base.AppCategoryDatabase, Engine: engine, Set: map[string]bool{}}
	for _, name := range set {
		facts.Set[name] = true
	}
	return facts
}

var allSet = []string{base.AppSystemEnvVarDatabaseName, base.AppSystemEnvVarUser, base.AppSystemEnvVarPassword,
	base.AppSystemEnvVarRootPassword}

func keys(vars []*Var) []string {
	out := make([]string, 0, len(vars))
	for _, v := range vars {
		out = append(out, v.Key)
	}
	return out
}

func TestPostgresRefersToTheAppsOwnCredentials(t *testing.T) {
	s := Suggest(EngineOf("postgres"), database("postgres", allSet...))
	assert.Equal(t, []string{"POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD"}, keys(s.Vars))
	assert.Equal(t, "${HIVEPAAS_PASSWORD}", s.Vars[2].Value)
	assert.Empty(t, s.Warnings)
	assert.Empty(t, s.Command)
	assert.True(t, s.Engine.InitOnly)
}

func TestEveryReferenceIsToAVariableKindSettingsPublish(t *testing.T) {
	published := map[string]bool{}
	for _, name := range append(allSet, base.AppSystemEnvVarPasswordURLEncoded) {
		published["${"+name+"}"] = true
	}
	for _, engine := range Engines {
		assert.True(t, len(engine.vars) > 0 || engine.command != "", "%s suggests nothing", engine.ID)
		for _, v := range engine.vars {
			if v.needs != "" {
				assert.True(t, published[v.value], "%s: %s refers to %s", engine.ID, v.key, v.value)
			}
		}
	}
}

func TestAMissingCredentialIsSaid(t *testing.T) {
	s := Suggest(EngineOf("mysql"), database("mysql", base.AppSystemEnvVarDatabaseName, base.AppSystemEnvVarUser,
		base.AppSystemEnvVarPassword))
	if assert.Len(t, s.Warnings, 1) {
		assert.Contains(t, s.Warnings[0], "root password")
		assert.Contains(t, s.Warnings[0], "HIVEPAAS_ROOT_PASSWORD")
	}
	// Postgres has no root password to miss.
	s = Suggest(EngineOf("postgres"), database("postgres", base.AppSystemEnvVarDatabaseName,
		base.AppSystemEnvVarUser, base.AppSystemEnvVarPassword))
	assert.Empty(t, s.Warnings)
}

func TestAnAppThatIsNoDatabaseIsTold(t *testing.T) {
	s := Suggest(EngineOf("postgres"), &Facts{Category: base.AppCategoryWebapp})
	assert.Len(t, s.Vars, 3, "the variables are still suggested")
	if assert.Len(t, s.Warnings, 1) {
		assert.Contains(t, s.Warnings[0], "category is webapp")
	}
	s = Suggest(EngineOf("postgres"), &Facts{})
	assert.Contains(t, s.Warnings[0], "category is not set")
}

func TestAnotherEngineThanAppKindsIsTold(t *testing.T) {
	s := Suggest(EngineOf("mariadb"), database("mysql", allSet...))
	if assert.Len(t, s.Warnings, 1) {
		assert.Contains(t, s.Warnings[0], "App Kind names the engine mysql, not MariaDB")
	}
}

func TestRedisIsACommand(t *testing.T) {
	cache := &Facts{Category: base.AppCategoryCache, Engine: "redis",
		Set: map[string]bool{base.AppSystemEnvVarPassword: true}}
	s := Suggest(EngineOf("redis"), cache)
	assert.Empty(t, s.Vars)
	assert.Contains(t, s.Command, `redis-server --requirepass "$HIVEPAAS_PASSWORD"`)
	assert.NotContains(t, s.Command, "${HIVEPAAS_PASSWORD}", "a reference would put the password in the command")
	assert.Empty(t, s.Warnings)

	cache.Set = map[string]bool{}
	s = Suggest(EngineOf("redis"), cache)
	if assert.Len(t, s.Warnings, 1) {
		assert.Contains(t, s.Warnings[0], "no password")
	}
}

func TestEnginesAppKindMayName(t *testing.T) {
	assert.Equal(t, "postgres", EngineOf("TimescaleDB").ID)
	assert.Equal(t, "postgres", EngineOf("postgresql").ID)
	assert.Equal(t, "mongodb", EngineOf("mongo").ID)
	assert.Nil(t, EngineOf("cassandra"))
	assert.Nil(t, EngineOf(""))
}
