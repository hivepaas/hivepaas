package envvarserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func envByKey(t *testing.T, envs []*envvarservice.EnvVar) map[string]*entity.EnvVar {
	t.Helper()
	byKey := map[string]*entity.EnvVar{}
	for _, env := range envs {
		byKey[env.Key] = env.EnvVar
	}
	return byKey
}

func TestKindEnvVars_Database(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryDatabase,
		Engine:   "postgres",
		Database: &entity.AppKindDatabase{
			DbName:       "app",
			Username:     "app",
			Password:     entity.NewEncryptedField("s3cret"),
			RootPassword: entity.NewEncryptedField("r00t"),
			SSLMode:      base.DatabaseSslModeDisable,
		},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Equal(t, "app", byKey[base.AppSystemEnvVarUser].Value)
	assert.Equal(t, "s3cret", byKey[base.AppSystemEnvVarPassword].Value)
	assert.Equal(t, "app", byKey[base.AppSystemEnvVarDatabaseName].Value)
	assert.True(t, byKey[base.AppSystemEnvVarPassword].IsShared)
	assert.False(t, byKey[base.AppSystemEnvVarRootPassword].IsShared,
		"the root password stays with the app that owns it")
}

func TestKindEnvVars_Cache(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryCache,
		Engine:   "redis",
		Cache: &entity.AppKindCache{
			Password: entity.NewEncryptedField("s3cret"),
		},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Equal(t, "s3cret", byKey[base.AppSystemEnvVarPassword].Value)
	assert.True(t, byKey[base.AppSystemEnvVarPassword].IsShared,
		"other apps connect to the cache with this password")

	// A cache names no user, selects no database and has no root account.
	for _, key := range []string{base.AppSystemEnvVarUser, base.AppSystemEnvVarDatabaseName,
		base.AppSystemEnvVarRootPassword, base.AppSystemEnvVarSSLMode} {
		assert.NotContains(t, byKey, key)
	}
}

func TestKindEnvVars_CacheTuningIsPublishedToTheAppAlone(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryCache,
		Engine:   "redis",
		Cache: &entity.AppKindCache{
			MaxMemory:       256 * unit.MB,
			EvictionRule:    "allkeys-lru",
			PersistenceMode: "aof",
		},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Equal(t, "268435456", byKey[base.AppSystemEnvVarMaxMemory].Value,
		"the ceiling goes out in bytes, which is what a shell can hand to the server")
	assert.Equal(t, "allkeys-lru", byKey[base.AppSystemEnvVarEvictionRule].Value)
	assert.Equal(t, "aof", byKey[base.AppSystemEnvVarPersistenceMode].Value)

	for _, key := range []string{base.AppSystemEnvVarMaxMemory, base.AppSystemEnvVarEvictionRule,
		base.AppSystemEnvVarPersistenceMode} {
		assert.False(t, byKey[key].IsShared, "%s is the cache's own tuning, of no use to another app", key)
	}
}

func TestKindEnvVars_CacheWithoutTuning(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryCache,
		Engine:   "redis",
		Cache:    &entity.AppKindCache{},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	// Nothing chosen reads as nothing, and the command decides what the engine's
	// own default is - here zero, which is how Redis spells "no ceiling".
	byKey := envByKey(t, envs)
	assert.Equal(t, "0", byKey[base.AppSystemEnvVarMaxMemory].Value)
	assert.Equal(t, "", byKey[base.AppSystemEnvVarEvictionRule].Value)
	assert.Equal(t, "", byKey[base.AppSystemEnvVarPersistenceMode].Value)
}

func TestKindEnvVars_CacheWithoutPassword(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryCache,
		Engine:   "memcached",
		Cache:    &entity.AppKindCache{},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Contains(t, byKey, base.AppSystemEnvVarPassword,
		"a cache without authentication still publishes the variable, empty")
	assert.Equal(t, "", byKey[base.AppSystemEnvVarPassword].Value)
}

func TestKindEnvVars_Storage(t *testing.T) {
	kind := &entity.AppKindSettings{
		Category: base.AppCategoryStorage,
		Engine:   "seaweedfs",
		Storage: &entity.AppKindStorage{
			KeyID:  "AKIAEXAMPLE",
			Secret: entity.NewEncryptedField("s3cret"),
			Bucket: "app",
			Region: "us-east-1",
		},
	}

	envs, err := kindEnvVars(kind)
	assert.NoError(t, err)

	byKey := envByKey(t, envs)
	assert.Equal(t, "AKIAEXAMPLE", byKey[base.AppSystemEnvVarKeyID].Value)
	assert.Equal(t, "s3cret", byKey[base.AppSystemEnvVarSecret].Value)
	assert.Equal(t, "app", byKey[base.AppSystemEnvVarBucket].Value)
	assert.Equal(t, "us-east-1", byKey[base.AppSystemEnvVarRegion].Value)
	assert.True(t, byKey[base.AppSystemEnvVarSecret].IsShared,
		"both halves of the pair are shared: it is the only way into the store")

	// A store has no user, no database and no root account.
	for _, key := range []string{base.AppSystemEnvVarUser, base.AppSystemEnvVarPassword,
		base.AppSystemEnvVarDatabaseName, base.AppSystemEnvVarRootPassword} {
		assert.NotContains(t, byKey, key)
	}
}

func TestKindEnvVars_NothingToPublish(t *testing.T) {
	tests := map[string]*entity.AppKindSettings{
		"no kind setting":            nil,
		"webapp":                     {Category: base.AppCategoryWebapp, Webapp: &entity.AppKindWebapp{}},
		"database without its block": {Category: base.AppCategoryDatabase},
		"cache without its block":    {Category: base.AppCategoryCache},
		"storage without its block":  {Category: base.AppCategoryStorage},
	}

	for name, kind := range tests {
		t.Run(name, func(t *testing.T) {
			envs, err := kindEnvVars(kind)
			assert.NoError(t, err)
			assert.Empty(t, envs)
		})
	}
}

// What a kind publishes is read in two places: here, and by the template
// renderer deciding what ${{ deps.<name>.ref.VAR }} may name. This keeps them
// from drifting apart.
func TestKindEnvVars_SharedNamesMatchBase(t *testing.T) {
	kinds := map[base.AppCategory]*entity.AppKindSettings{
		base.AppCategoryDatabase: {Category: base.AppCategoryDatabase, Database: &entity.AppKindDatabase{}},
		base.AppCategoryCache:    {Category: base.AppCategoryCache, Cache: &entity.AppKindCache{}},
		base.AppCategoryStorage:  {Category: base.AppCategoryStorage, Storage: &entity.AppKindStorage{}},
		base.AppCategoryWebapp:   {Category: base.AppCategoryWebapp, Webapp: &entity.AppKindWebapp{}},
	}
	for _, category := range base.AllAppCategories {
		envs, err := kindEnvVars(kinds[category])
		assert.NoError(t, err)

		var shared []string
		for _, env := range envs {
			if env.IsShared {
				shared = append(shared, env.Key)
			}
		}
		assert.ElementsMatch(t, base.AppKindSharedEnvVars(category), shared, string(category))
	}
}

func TestDockerAPIEnvVarsNameTheSocketOnlyForAnAppWithAccess(t *testing.T) {
	assert.Empty(t, dockerAPIEnvVars(false))
	envs := dockerAPIEnvVars(true)
	if assert.Len(t, envs, 1) {
		assert.Equal(t, base.AppSystemEnvVarDockerHost, envs[0].Key)
		assert.Equal(t, "unix:///var/run/hivepaas/docker.sock", envs[0].Value)
		assert.False(t, envs[0].IsShared, "another app has no use for this app's socket")
	}
}
