package envlink

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func groupIDs(groups []*Group) []string {
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, g.ID)
	}
	return ids
}

func varsOf(g *Group) map[string]string {
	out := map[string]string{}
	for _, v := range g.Vars {
		out[v.Key] = v.Value
	}
	return out
}

var ready = &Facts{HasPort: true, HasPassword: true}

func TestEngineFamilies(t *testing.T) {
	for engine, want := range map[string]string{
		"postgres": FamilyPostgres, "PostgreSQL": FamilyPostgres, "timescaledb": FamilyPostgres,
		"mariadb": FamilyMySQL, "mongo": FamilyMongoDB, "valkey": FamilyRedis, "memcached": FamilyMemcached,
		"clickhouse": "", "": "",
	} {
		assert.Equal(t, want, EngineFamily(engine), engine)
	}
}

func TestPostgresGetsAURLTheirPartsAndLibpq(t *testing.T) {
	groups := Suggest(&Target{Key: "pg", Name: "PG", Category: base.AppCategoryDatabase, Engine: "postgres"}, ready)
	assert.Equal(t, []string{"connection-url", "individual", "libpq"}, groupIDs(groups))
	assert.True(t, groups[0].Recommended)
	assert.Equal(t,
		"postgres://${pg.HIVEPAAS_USER}:${pg.HIVEPAAS_PASSWORD_URLENCODED}@${pg.HIVEPAAS_HOST}:${pg.HIVEPAAS_PORT}"+
			"/${pg.HIVEPAAS_DATABASE_NAME}?sslmode=${pg.HIVEPAAS_SSL_MODE}",
		varsOf(groups[0])["DATABASE_URL"])
	assert.Equal(t, "${pg.HIVEPAAS_PASSWORD}", varsOf(groups[1])["DB_PASSWORD"], "a part is the raw password")
	assert.Equal(t, "${pg.HIVEPAAS_HOST}", varsOf(groups[2])["PGHOST"])
}

func TestEachCategoryAndFamilyHasARecipe(t *testing.T) {
	db, cache := base.AppCategoryDatabase, base.AppCategoryCache
	urlAndParts := []string{"connection-url", "individual"}
	for name, tc := range map[string]struct {
		target *Target
		ids    []string
	}{
		"mysql":        {&Target{Key: "db", Category: db, Engine: "mariadb"}, urlAndParts},
		"mongodb":      {&Target{Key: "db", Category: db, Engine: "mongodb"}, urlAndParts},
		"other db":     {&Target{Key: "db", Category: db, Engine: "clickhouse"}, []string{"individual"}},
		"redis":        {&Target{Key: "c", Category: cache, Engine: "dragonfly"}, urlAndParts},
		"memcached":    {&Target{Key: "c", Category: cache, Engine: "memcached"}, []string{"servers"}},
		"other cache":  {&Target{Key: "c", Category: cache, Engine: "hazelcast"}, []string{"individual"}},
		"storage":      {&Target{Key: "s3", Category: base.AppCategoryStorage, Engine: "minio"}, []string{"s3"}},
		"webapp":       {&Target{Key: "api", Category: base.AppCategoryWebapp}, []string{"addresses"}},
		"without kind": {&Target{Key: "api"}, []string{"addresses"}},
	} {
		groups := Suggest(tc.target, ready)
		assert.Equal(t, tc.ids, groupIDs(groups), name)
		assert.True(t, groups[0].Recommended, name)
	}
}

// Every variable a recipe names is one the target's kind shares, or one every
// app shares: a suggestion never refers to something that is not there.
func TestRecipesNameOnlySharedVariables(t *testing.T) {
	for _, target := range []*Target{
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "postgres"},
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "mysql"},
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "mongodb"},
		{Key: "a", Category: base.AppCategoryDatabase, Engine: "x"},
		{Key: "a", Category: base.AppCategoryCache, Engine: "redis"},
		{Key: "a", Category: base.AppCategoryCache, Engine: "memcached"},
		{Key: "a", Category: base.AppCategoryCache, Engine: "x"},
		{Key: "a", Category: base.AppCategoryStorage},
		{Key: "a", Category: base.AppCategoryWebapp},
	} {
		allowed := map[string]bool{}
		for _, name := range append(append([]string{}, base.AppCommonSharedEnvVars...),
			base.AppKindSharedEnvVars(target.Category)...) {
			allowed[name] = true
		}
		for _, g := range Suggest(target, ready) {
			for _, v := range g.Vars {
				for _, name := range referencedNames(v.Value, target.Key) {
					assert.True(t, allowed[name], "%s/%s names %s", target.Category, target.Engine, name)
				}
			}
		}
	}
}

func TestRedisWithoutAPasswordHasNoCredentialsInItsURL(t *testing.T) {
	groups := Suggest(&Target{Key: "cache", Category: base.AppCategoryCache, Engine: "redis"},
		&Facts{HasPort: true, HasPassword: false})
	assert.Equal(t, "redis://${cache.HIVEPAAS_HOST}:${cache.HIVEPAAS_PORT}/0", varsOf(groups[0])["REDIS_URL"])
}

func TestATargetWithoutAPortIsWarnedAbout(t *testing.T) {
	groups := Suggest(&Target{Key: "pg", Name: "PG", Category: base.AppCategoryDatabase, Engine: "postgres"},
		&Facts{HasPort: false, HasPassword: true})
	for _, g := range groups {
		assert.NotEmpty(t, g.Warnings, g.ID)
		assert.Contains(t, g.Warnings[0], "PG has no container port")
	}
}

func TestATargetWithoutAKindSaysWhy(t *testing.T) {
	groups := Suggest(&Target{Key: "api", Name: "API"}, ready)
	assert.Contains(t, groups[0].Warnings, "API declares no kind, so only its addresses and shared variables are offered.")
}

func TestTheTargetsSharedVariablesComeLast(t *testing.T) {
	groups := Suggest(&Target{Key: "api", Name: "API", Category: base.AppCategoryWebapp},
		&Facts{HasPort: true, SharedVars: []string{"TOKEN_URL", "API_VERSION"}})
	last := groups[len(groups)-1]
	assert.Equal(t, "shared", last.ID)
	assert.Equal(t, "Shared variables of API", last.Title)
	assert.Equal(t, []string{"API_VERSION", "TOKEN_URL"}, []string{last.Vars[0].Key, last.Vars[1].Key})
	assert.Equal(t, "${api.API_VERSION}", last.Vars[0].Value)
}

func TestEnvKeyOfAKey(t *testing.T) {
	assert.Equal(t, "API", envKeyOf("api"))
	assert.Equal(t, "MY_API", envKeyOf("my-api"))
	assert.Equal(t, "APP_2FA_APP", envKeyOf("2fa-app"))
}

func TestTargetsLeaveOutTheAppItselfAndPreviews(t *testing.T) {
	self := &entity.App{ID: "a1", Key: "web", Name: "Web", ProjectEnvID: "e1"}
	apps := []*entity.App{
		self,
		{ID: "a2", Key: "pg", Name: "Postgres", ProjectEnvID: "e1"},
		{ID: "a3", Key: "web-pr-1", Name: "PR", ProjectEnvID: "e1", ParentID: "a1"},
		{ID: "a4", Key: "cache", Name: "Cache", ProjectEnvID: "e1"},
		{ID: "a5", Key: "other", Name: "Other env", ProjectEnvID: "e2"},
	}
	kinds := map[string]*entity.AppKindSettings{"a2": {Category: base.AppCategoryDatabase, Engine: "postgres"}}

	targets := Targets(self, apps, kinds)

	assert.Len(t, targets, 2)
	assert.Equal(t, "Cache", targets[0].Name, "ordered by name")
	assert.Equal(t, &Target{ID: "a2", Key: "pg", Name: "Postgres", Category: base.AppCategoryDatabase,
		Engine: "postgres"}, targets[1])
}

func TestFactsOfASharedEnvironment(t *testing.T) {
	system := func(key, value string) *envvarservice.EnvVar {
		return &envvarservice.EnvVar{EnvVar: &entity.EnvVar{Key: key, Value: value, IsShared: true, IsSystem: true}}
	}
	facts := FactsOf([]*envvarservice.EnvVar{
		system(base.AppSystemEnvVarPort, "5432"),
		system(base.AppSystemEnvVarPassword, ""),
		{EnvVar: &entity.EnvVar{Key: "API_VERSION", Value: "2", IsShared: true}},
	})
	assert.Equal(t, &Facts{HasPort: true, HasPassword: false, SharedVars: []string{"API_VERSION"}}, facts)
}
